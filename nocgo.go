//go:build !cgo
// +build !cgo

package purego

// if CGO_ENABLED=0 import fakecgo to setup the Cgo runtime correctly.
// This is required since some frameworks need TLS setup the C way which Go doesn't do.
import _ "github.com/ebitengine/purego/internal/fakecgo"
