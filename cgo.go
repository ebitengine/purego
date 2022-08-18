// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

//go:build cgo || ios
// +build cgo ios

package purego

// if CGO_ENABLED=1 import the Cgo runtime to ensure that it is set up properly.
// This is required since some frameworks need TLS setup the C way which Go doesn't do.
// We currently don't support ios in fakecgo mode so force Cgo or fail
import _ "runtime/cgo"
