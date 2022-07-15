// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebiten Authors

//go:build cgo
// +build cgo

package purego

// if CGO_ENABLED=1 then make sure that the runtime is called to setup TLS properly.

import _ "runtime/cgo"
