// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

//go:build go1.28 && (darwin || freebsd || illumos || linux || netbsd)

package purego

// Starting with Go 1.28 runtime/cgo is written in Go and assembly (golang/go#81450)
// and can be imported whatever CGO_ENABLED is set to. It sets up threads and TLS the
// way C expects, which is the whole job of internal/fakecgo on older releases, so it
// replaces both fakecgo (nocgo.go) and the C-compiled internal/cgo (cgo.go).
//
// As in cgo.go, the import has to be explicit: purego does not use `import "C"` in
// this package, so nothing else would pull runtime/cgo into the program.
import _ "runtime/cgo"
