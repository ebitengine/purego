// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

// Package testlib builds the C and Objective-C libraries used by purego's tests.
package testlib

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// Compiler returns the path of the compiler named by `go env compilerEnv`, such as "CC".
func Compiler(compilerEnv string) (string, error) {
	out, err := exec.Command("go", "env", compilerEnv).Output()
	if err != nil {
		return "", fmt.Errorf("go env %s error: %w", compilerEnv, err)
	}
	compiler := strings.TrimSpace(string(out))
	if compiler == "" {
		return "", fmt.Errorf("go env %s is empty", compilerEnv)
	}
	return exec.LookPath(compiler)
}

// BuildSharedLib builds a shared library with the compiler from [Compiler].
// args holds the sources and any extra compiler flags, such as -framework.
func BuildSharedLib(tb testing.TB, compilerEnv, libFile string, args ...string) error {
	tb.Helper()
	// When PUREGO_TEST_PREBUILT_LIBDIR is set, the shared library has been
	// cross-compiled ahead of time and placed in that directory under the
	// base name of libFile. This allows running the tests on a target that
	// has no C toolchain, such as an Android emulator.
	if dir := os.Getenv("PUREGO_TEST_PREBUILT_LIBDIR"); dir != "" {
		data, err := os.ReadFile(filepath.Join(dir, filepath.Base(libFile)))
		if err != nil {
			return fmt.Errorf("prebuilt lib: %w", err)
		}
		if err := os.WriteFile(libFile, data, 0o755); err != nil {
			return fmt.Errorf("prebuilt lib: %w", err)
		}
		return nil
	}

	// Compiling the library needs a C toolchain targeting GOARCH. CI has none
	// for Windows on 386 or arm64, so skip those (the prebuilt path above
	// avoids the toolchain).
	if runtime.GOOS == "windows" {
		switch runtime.GOARCH {
		case "386":
			tb.Skip("need a 386 C toolchain to run this test") // TODO: find a 386 C toolchain for test
		case "arm64":
			tb.Skip("need an arm64 C toolchain to run this test")
		}
	}

	compiler, err := Compiler(compilerEnv)
	if err != nil {
		return err
	}

	flags := []string{"-shared", "-Wall", "-Werror", "-fPIC", "-o", libFile}
	if runtime.GOARCH == "386" {
		flags = append(flags, "-m32")
	}
	// macOS arm64 can run amd64 tests through Rossetta.
	// Build the shared library based on the GOARCH and not
	// the default behavior of the compiler.
	if runtime.GOOS == "darwin" {
		var arch string
		switch runtime.GOARCH {
		case "arm64":
			arch = "arm64"
		case "amd64":
			arch = "x86_64"
		default:
			return fmt.Errorf("unknown macOS architecture %s", runtime.GOARCH)
		}
		flags = append(flags, "-arch", arch)
	}
	cmd := exec.Command(compiler, append(flags, args...)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("compile lib: %w\n%q\n%s", err, cmd, string(out))
	}

	return nil
}
