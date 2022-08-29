// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

//go:build (arm64 && !go1.18) || (amd64 && !go1.17)
// +build arm64,!go1.18 amd64,!go1.17

package purego

// stackCallingConvention represents all versions of Go that use the stack
// to pass parameters to functions. This is used to circumvent
// the need for ABIInternal tag which is only allowed in the runtime.
const stackCallingConvention = true
