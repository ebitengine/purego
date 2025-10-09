// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package main

import (
	_ "embed"
	"fmt"
	"log"
	"runtime"

	"github.com/ebitengine/purego"
)

//go:embed libs/macos/libraylib.5.5.0.dylib
var raylibDarwin []byte

//go:embed libs/linux_amd64/libraylib.so.5.5.0
var raylibLinuxAMD64 []byte

//go:embed libs/windows_amd64/raylib.dll
var raylibWindowsAMD64 []byte

func main() {
	name, data, mode := selectLibrary()
	if len(data) == 0 {
		log.Fatalf("no embedded raylib library for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	lib, err := purego.OpenEmbeddedLibrary(name, data, mode)
	if err != nil {
		log.Fatalf("OpenEmbeddedLibrary failed: %v", err)
	}
	defer lib.Close()

	var getRandomValue func(int32, int32) int32
	purego.RegisterLibFunc(&getRandomValue, lib.Handle(), "GetRandomValue")

	var getApplicationDirectory func() string
	purego.RegisterLibFunc(&getApplicationDirectory, lib.Handle(), "GetApplicationDirectory")

	roll := getRandomValue(1, 6)
	fmt.Printf("GetRandomValue(1, 6) => %d\n", roll)
	fmt.Printf("GetApplicationDirectory() => %q\n", getApplicationDirectory())
}

func selectLibrary() (name string, data []byte, mode int) {
	mode = defaultMode()
	switch runtime.GOOS {
	case "darwin":
		return "libraylib.dylib", raylibDarwin, mode
	case "linux":
		if runtime.GOARCH == "amd64" {
			return "libraylib.so", raylibLinuxAMD64, mode
		}
	case "windows":
		if runtime.GOARCH == "amd64" {
			return "raylib.dll", raylibWindowsAMD64, mode
		}
	}
	return "", nil, mode
}
