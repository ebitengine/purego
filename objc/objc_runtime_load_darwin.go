// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

//go:build go1.28

package objc

// Records libobjc's full install name as a load command. The bindings in
// objc_runtime_bindings_darwin.go use "_" for their library because, at least in
// the golang/go#81450 CLs as of patch set 20 of CL 856832, cmd/cgo keeps only the
// base name of a binding's library ("libobjc.A.dylib"), which dyld cannot find.
//
//go:cgo_import_dynamic _ _ "/usr/lib/libobjc.A.dylib"
