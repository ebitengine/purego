// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2023 The Ebitengine Authors

//go:build windows

package purego_test

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
)

func loadSharedLib(libPath string) (uintptr, error) {
	lib := syscall.NewLazyDLL(libPath)
	if err := lib.Load(); err != nil {
		return 0, err
	}

	return lib.Handle(), nil
}

func buildSharedLib(_, libDir, libName string, sources ...string) (string, error) {
	arch := "x64"
	if runtime.GOARCH == "386" {
		arch = "x86"
	}

	libPath := filepath.Join(libDir, libName+".dll")
	args := append([]string{"/c", filepath.Join("testdata", "runcl.bat"), arch, libPath}, sources...)
	if out, err := exec.Command("cmd", args...).Output(); err != nil {
		return "", fmt.Errorf("compile dll: %w\n%s", err, string(out))
	} else {
		fmt.Println(string(out))
	}

	return libPath, nil
}
