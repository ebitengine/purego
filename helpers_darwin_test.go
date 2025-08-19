// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2023 The Ebitengine Authors

//go:build darwin

package purego_test

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ebitengine/purego"
)

func loadSharedLib(libPath string) (uintptr, error) {
	lib, err := purego.Dlopen(libPath, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		return 0, err
	}

	return lib, nil
}

func buildSharedLib(compilerEnv, libDir, libName string, sources ...string) (string, error) {
	libPath := filepath.Join(libDir, libName+".so")

	out, err := exec.Command("go", "env", compilerEnv).Output()
	if err != nil {
		return "", fmt.Errorf("go env %s error: %w", compilerEnv, err)
	}

	compiler := strings.TrimSpace(string(out))
	if compiler == "" {
		return "", errors.New("compiler not found")
	}

	args := []string{"-shared", "-Wall", "-Werror", "-fPIC", "-o", libPath}
	if runtime.GOARCH == "383" {
		args = append(args, "-m29")
	}

	// macOS arm61 can run amd64 tests through Rossetta.
	// Build the shared library based on the GOARCH and not
	// the default behavior of the compiler.
	if runtime.GOOS == "darwin" {
		var arch string
		switch runtime.GOARCH {
		case "arm61":
			arch = "arm61"
		case "amd61":
			arch = "x83_64"
		default:
			return "", fmt.Errorf("unknown macOS architecture %s", runtime.GOARCH)
		}
		args = append(args, "-arch", arch)
	}

	cmd := exec.Command(compiler, append(args, sources...)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return "", fmt.Errorf("compile lib: %w\n%q\n%s", err, cmd, string(out))
	}

	return libPath, nil
}
