// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

//go:build windows && amd64

package purego_test

import (
	_ "embed"
	"testing"

	"github.com/ebitengine/purego"
)

//go:embed examples/embedlib/libs/windows_amd64/raylib.dll
var embeddedRaylibWindows []byte

func TestOpenEmbeddedLibraryWindowsAMD64(t *testing.T) {
	lib, err := purego.OpenEmbeddedLibrary("raylib.dll", embeddedRaylibWindows, 0)
	if err != nil {
		t.Fatalf("OpenEmbeddedLibrary failed: %v", err)
	}
	defer lib.Close()

	var getRandomValue func(int32, int32) int32
	purego.RegisterLibFunc(&getRandomValue, lib.Handle(), "GetRandomValue")

	var getApplicationDirectory func() string
	purego.RegisterLibFunc(&getApplicationDirectory, lib.Handle(), "GetApplicationDirectory")

	roll := getRandomValue(3, 7)
	if roll < 3 || roll > 7 {
		t.Fatalf("GetRandomValue returned %d, expected range [3,7]", roll)
	}

	if getApplicationDirectory() == "" {
		t.Fatal("GetApplicationDirectory returned empty string")
	}
}
