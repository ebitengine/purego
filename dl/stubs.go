// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebiten Authors

//go:build darwin
// +build darwin

package dl

var dlopenABI0 uintptr
var dlsymABI0 uintptr
var dlerrorABI0 uintptr
var dlcloseABI0 uintptr

func dlopen()

func dlerror()

func dlclose()

func dlsym()
