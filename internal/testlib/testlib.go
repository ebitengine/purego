// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

// Package testlib builds the C and Objective-C libraries used by purego's tests.
package testlib

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// BuildSharedLib compiles sources into a shared library at libFile using the C compiler named
// by the go env variable compilerEnv (such as "CC" or "CXX"). The library is built for GOARCH,
// and Objective-C (.m) sources are linked against Foundation.
func BuildSharedLib(tb testing.TB, compilerEnv, libFile string, sources ...string) error {
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

	out, err := exec.Command("go", "env", compilerEnv).Output()
	if err != nil {
		return fmt.Errorf("go env %s error: %w", compilerEnv, err)
	}

	compiler := strings.TrimSpace(string(out))
	if compiler == "" {
		return errors.New("compiler not found")
	}

	args := []string{"-shared", "-Wall", "-Werror", "-fPIC", "-o", libFile}
	if runtime.GOARCH == "386" {
		args = append(args, "-m32")
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
		args = append(args, "-arch", arch)
	}
	for _, src := range sources {
		if filepath.Ext(src) == ".m" {
			args = append(args, "-framework", "Foundation")
			break
		}
	}
	cmd := exec.Command(compiler, append(args, sources...)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("compile lib: %w\n%q\n%s", err, cmd, string(out))
	}

	return nil
}
