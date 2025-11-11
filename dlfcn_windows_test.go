// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2023 The Ebitengine Authors

package purego_test

import (
	"github.com/ebitengine/purego"
	"testing"
	"unsafe"
)

func TestMessageBox(t *testing.T) {
	user32, err := purego.Dlopen("user32.dll", purego.RTLD_NOW)
	if err != nil {
		t.Fatal(err)
	}
	defer purego.Dlclose(user32)

	symbol, err := purego.Dlsym(user32, "MessageBoxA")
	if err != nil {
		t.Fatal(err)
	}

	var messageBox func(hwnd unsafe.Pointer, text, caption string, flag uint32)
	purego.RegisterFunc(&messageBox, symbol)

	messageBox(nil, "message", "information", 0)
}
