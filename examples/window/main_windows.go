// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package main

import (
	"runtime"
	"syscall"
	"unsafe"

	"github.com/ebitengine/purego"
)

const (
	WS_OVERLAPPEDWINDOW = 0x00000000 | 0x00C00000 | 0x00080000 | 0x00040000 | 0x00020000 | 0x00010000
	CW_USEDEFAULT       = ^0x7fffffff
	SW_SHOW             = 5
	WM_DESTROY          = 2
)

type (
	ATOM      uint16
	HANDLE    uintptr
	HINSTANCE HANDLE
	HICON     HANDLE
	HCURSOR   HANDLE
	HBRUSH    HANDLE
	HWND      HANDLE
	HMENU     HANDLE
)

type WNDCLASSEX struct {
	Size       uint32
	Style      uint32
	WndProc    uintptr
	ClsExtra   int32
	WndExtra   int32
	Instance   HINSTANCE
	Icon       HICON
	Cursor     HCURSOR
	Background HBRUSH
	MenuName   *uint16
	ClassName  *uint16
	IconSm     HICON
}

type RECT struct {
	Left, Top, Right, Bottom int32
}

type POINT struct {
	X, Y int32
}

type MSG struct {
	Hwnd    HWND
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      POINT
}

var (
	GetModuleHandle func(modulename *uint16) HINSTANCE
	RegisterClassEx func(w *WNDCLASSEX) ATOM
	CreateWindowEx  func(exStyle uint, className, windowName *uint16,
		style uint, x, y, width, height int, parent HWND, menu HMENU,
		instance HINSTANCE, param unsafe.Pointer) HWND
	AdjustWindowRect func(rect *RECT, style uint, menu bool) bool
	ShowWindow       func(hwnd HWND, cmdshow int) bool
	GetMessage       func(msg *MSG, hwnd HWND, msgFilterMin, msgFilterMax uint32) int
	TranslateMessage func(msg *MSG) bool
	DispatchMessage  func(msg *MSG) uintptr
	DefWindowProc    func(hwnd HWND, msg uint32, wParam, lParam uintptr) uintptr
	PostQuitMessage  func(exitCode int)
)

func init() {
	// Use [syscall.NewLazyDLL] here to avoid external dependencies (#270).
	// For actual use cases, [golang.org/x/sys/windows.NewLazySystemDLL] is recommended.
	kernel32 := syscall.NewLazyDLL("kernel32.dll").Handle()
	purego.RegisterLibFunc(&GetModuleHandle, kernel32, "GetModuleHandleW")

	// Use [syscall.NewLazyDLL] here to avoid external dependencies (#270).
	// For actual use cases, [golang.org/x/sys/windows.NewLazySystemDLL] is recommended.
	user32 := syscall.NewLazyDLL("user32.dll").Handle()
	purego.RegisterLibFunc(&RegisterClassEx, user32, "RegisterClassExW")
	purego.RegisterLibFunc(&CreateWindowEx, user32, "CreateWindowExW")
	purego.RegisterLibFunc(&AdjustWindowRect, user32, "AdjustWindowRect")
	purego.RegisterLibFunc(&ShowWindow, user32, "ShowWindow")
	purego.RegisterLibFunc(&GetMessage, user32, "GetMessageW")
	purego.RegisterLibFunc(&TranslateMessage, user32, "TranslateMessage")
	purego.RegisterLibFunc(&DispatchMessage, user32, "DispatchMessageW")
	purego.RegisterLibFunc(&DefWindowProc, user32, "DefWindowProcW")
	purego.RegisterLibFunc(&PostQuitMessage, user32, "PostQuitMessage")

	runtime.LockOSThread()
}

func main() {
	className, err := syscall.UTF16PtrFromString("Sample Window Class")
	if err != nil {
		panic(err)
	}
	inst := GetModuleHandle(className)

	wc := WNDCLASSEX{
		Size:      uint32(unsafe.Sizeof(WNDCLASSEX{})),
		WndProc:   syscall.NewCallback(wndProc),
		Instance:  inst,
		ClassName: className,
	}

	RegisterClassEx(&wc)

	wr := RECT{
		Left:   0,
		Top:    0,
		Right:  320,
		Bottom: 240,
	}
	title, err := syscall.UTF16PtrFromString("My Title")
	if err != nil {
		panic(err)
	}
	AdjustWindowRect(&wr, WS_OVERLAPPEDWINDOW, false)
	hwnd := CreateWindowEx(
		0, className,
		title,
		WS_OVERLAPPEDWINDOW,
		CW_USEDEFAULT, CW_USEDEFAULT, int(wr.Right-wr.Left), int(wr.Bottom-wr.Top),
		0, 0, inst, nil,
	)
	if hwnd == 0 {
		panic(syscall.GetLastError())
	}

	ShowWindow(hwnd, SW_SHOW)

	var msg MSG
	for GetMessage(&msg, 0, 0, 0) != 0 {
		TranslateMessage(&msg)
		DispatchMessage(&msg)
	}
}

func wndProc(hwnd HWND, msg uint32, wparam, lparam uintptr) uintptr {
	switch msg {
	case WM_DESTROY:
		PostQuitMessage(0)
	}
	return DefWindowProc(hwnd, msg, wparam, lparam)
}
