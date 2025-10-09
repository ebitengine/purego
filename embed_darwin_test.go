// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

//go:build darwin

package purego_test

import (
	_ "embed"
	"testing"

	"github.com/ebitengine/purego"
)

//go:embed examples/embedlib/libs/macos/libraylib.5.5.0.dylib
var embeddedRaylibDarwin []byte

func TestOpenEmbeddedLibraryDarwin(t *testing.T) {
	lib, err := purego.OpenEmbeddedLibrary("libraylib.dylib", embeddedRaylibDarwin, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		t.Fatalf("OpenEmbeddedLibrary failed: %v", err)
	}
	defer lib.Close()

	var getRandomValue func(int32, int32) int32
	purego.RegisterLibFunc(&getRandomValue, lib.Handle(), "GetRandomValue")

	var getApplicationDirectory func() string
	purego.RegisterLibFunc(&getApplicationDirectory, lib.Handle(), "GetApplicationDirectory")

	roll := getRandomValue(2, 5)
	if roll < 2 || roll > 5 {
		t.Fatalf("GetRandomValue returned %d, expected range [2,5]", roll)
	}

	if getApplicationDirectory() == "" {
		t.Fatal("GetApplicationDirectory returned empty string")
	}
}
