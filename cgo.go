//go:build cgo
// +build cgo

package purego

// if CGO_ENABLED=1 import the Cgo runtime to ensure that it is set up properly.
// This is required since some frameworks need TLS setup the C way which Go doesn't do.
import _ "runtime/cgo"
