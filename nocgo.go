// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebiten Authors

//go:build !cgo
// +build !cgo

package purego

// If CGO_ENABLED=0 is true then we must setup fake cgo
// in order for TLS to work properly.

import _ "github.com/ebitengine/purego/internal/fakecgo"
