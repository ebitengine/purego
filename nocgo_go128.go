// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

//go:build !cgo && ((go1.28 && goexperiment.runtimecgopurego) || go1.29) && (darwin || freebsd || illumos || linux || netbsd)

package purego

// Go 1.28 introduces a pure-Go and assembly implementation of runtime/cgo,
// so there's no longer a need for fakecgo and we can implement the true
// runtime/cgo.
import _ "runtime/cgo"
