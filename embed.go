// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package purego

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// EmbeddedLibrary represents a dynamic library that was materialized from an
// embedded byte slice at runtime.
//
// The backing temporary file is automatically cleaned up on Unix style systems
// immediately after the library is opened. On Windows the file is removed when
// Close is called because a loaded DLL cannot be unlinked.
type EmbeddedLibrary struct {
	handle    uintptr
	closeFunc func(uintptr) error
	cleanup   func() error
}

// Handle returns the dynamic library handle that can be used with RegisterLibFunc.
func (l *EmbeddedLibrary) Handle() uintptr {
	if l == nil {
		return 0
	}
	return l.handle
}

// Close releases the library handle and removes the temporary file that was
// created to materialize the embedded library. Close is safe to call multiple
// times.
func (l *EmbeddedLibrary) Close() error {
	if l == nil {
		return nil
	}
	var errs []error
	if l.handle != 0 && l.closeFunc != nil {
		if err := l.closeFunc(l.handle); err != nil {
			errs = append(errs, err)
		}
		l.handle = 0
	}
	if l.cleanup != nil {
		if err := l.cleanup(); err != nil && !errors.Is(err, os.ErrNotExist) {
			errs = append(errs, err)
		}
		l.cleanup = nil
	}
	return errors.Join(errs...)
}

// OpenEmbeddedLibrary takes a byte slice containing a platform specific shared
// library, writes it to a temporary file, and loads the resulting image through
// Dlopen/LoadLibrary. The name parameter is used to infer the desired filename
// extension for the temporary file. The returned EmbeddedLibrary must be
// closed by the caller to avoid leaking the handle.
func OpenEmbeddedLibrary(name string, data []byte, mode int) (*EmbeddedLibrary, error) {
	if len(data) == 0 {
		return nil, errors.New("purego: embedded library data is empty")
	}
	ext := filepath.Ext(name)
	if ext == "" {
		switch runtime.GOOS {
		case "windows":
			ext = ".dll"
		case "darwin":
			ext = ".dylib"
		default:
			ext = ".so"
		}
	}

	pattern := fmt.Sprintf("purego-embedded-*%s", ext)
	f, err := os.CreateTemp("", pattern)
	if err != nil {
		return nil, fmt.Errorf("purego: create temp file: %w", err)
	}
	fileName := f.Name()
	if _, err := f.Write(data); err != nil {
		f.Close()
		os.Remove(fileName)
		return nil, fmt.Errorf("purego: write temp library: %w", err)
	}
	if err := f.Close(); err != nil {
		os.Remove(fileName)
		return nil, fmt.Errorf("purego: close temp library: %w", err)
	}

	if runtime.GOOS != "windows" {
		if err := os.Chmod(fileName, 0o755); err != nil {
			os.Remove(fileName)
			return nil, fmt.Errorf("purego: chmod temp library: %w", err)
		}
	}

	handle, closeFn, err := openEmbeddedHandle(fileName, mode)
	if err != nil {
		os.Remove(fileName)
		return nil, err
	}

	lib := &EmbeddedLibrary{
		handle:    handle,
		closeFunc: closeFn,
	}

	if runtime.GOOS == "windows" {
		lib.cleanup = func() error {
			return os.Remove(fileName)
		}
	} else {
		if err := os.Remove(fileName); err != nil && !errors.Is(err, os.ErrNotExist) {
			// Defer removal to Close when immediate unlinking fails (e.g. unexpected permission error).
			lib.cleanup = func() error {
				return os.Remove(fileName)
			}
		}
	}

	return lib, nil
}
