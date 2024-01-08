// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2023 The Ebitengine Authors

//go:build windows

package purego_test

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/ebitengine/purego"
)

func TestNestedDlopenCall(t *testing.T) {
	libFileName := filepath.Join(t.TempDir(), "libdlnested.dll")
	t.Logf("Build %v", libFileName)

	if err := buildSharedLib("CXX", libFileName, filepath.Join("libdlnested", "nested.cpp")); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(libFileName)

	lib, err := purego.Dlopen(libFileName, purego.RTLD_DEFAULT)
	if err != nil {
		t.Fatalf("Dlopen(%q) failed: %v", libFileName, err)
	}

	err = purego.Dlclose(lib)
	if err != nil {
		t.Fatalf("Dlclose(...) failed: %v", err)
	}
}

func buildSharedLib(compilerEnv, libFile string, sources ...string) error {
	out, err := exec.Command("go", "env", compilerEnv).Output()
	if err != nil {
		return fmt.Errorf("go env %s error: %w", compilerEnv, err)
	}

	compiler := strings.TrimSpace(string(out))
	if compiler == "" {
		return errors.New("compiler not found")
	}

	var args []string
	if runtime.GOOS == "freebsd" {
		args = []string{"-shared", "-Wall", "-Werror", "-fPIC", "-o", libFile}
	} else {
		args = []string{"-shared", "-Wall", "-Werror", "-o", libFile}
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
	cmd := exec.Command(compiler, append(args, sources...)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("compile lib: %w\n%q\n%s", err, cmd, string(out))
	}

	return nil
}
