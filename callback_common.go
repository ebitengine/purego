// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

package purego

import "unsafe"

// maxCb is the maximum number of callbacks
// only increase this if you have added more to the callbackasm function
const maxCB = 2000

// callbackasmABI0 is implemented in zcallback_GOOS_GOARCH.s
var callbackasmABI0 uintptr

const ptrSize = unsafe.Sizeof((*int)(nil))
