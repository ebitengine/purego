// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2023 The Ebitengine Authors

//go:build windows

package purego_test

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
)

func buildSharedLib(_, libPath string, sources ...string) error {
	libPath += ".dll"

	arch := "x64"
	if runtime.GOARCH == "386" {
		arch = "x86"
	}

	args := append([]string{"/c", filepath.Join("testdata", "runcl.bat"), arch, libPath}, sources...)
	if out, err := exec.Command("cmd", args...).Output(); err != nil {
		return fmt.Errorf("compile dll: %w\n%s", err, string(out))
	}

	return nil
}
