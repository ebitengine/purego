// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

//go:build linux && amd64

package purego_test

import (
	_ "embed"
	"testing"

	"github.com/ebitengine/purego"
)

//go:embed examples/embedlib/libs/linux_amd64/libraylib.so.5.5.0
var embeddedRaylibLinux []byte

func TestOpenEmbeddedLibraryLinuxAMD64(t *testing.T) {
	lib, err := purego.OpenEmbeddedLibrary("libraylib.so", embeddedRaylibLinux, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		t.Fatalf("OpenEmbeddedLibrary failed: %v", err)
	}
	defer lib.Close()

	var getRandomValue func(int32, int32) int32
	purego.RegisterLibFunc(&getRandomValue, lib.Handle(), "GetRandomValue")

	var getApplicationDirectory func() string
	purego.RegisterLibFunc(&getApplicationDirectory, lib.Handle(), "GetApplicationDirectory")

	roll := getRandomValue(10, 20)
	if roll < 10 || roll > 20 {
		t.Fatalf("GetRandomValue returned %d, expected range [10,20]", roll)
	}

	if getApplicationDirectory() == "" {
		t.Fatal("GetApplicationDirectory returned empty string")
	}
}
