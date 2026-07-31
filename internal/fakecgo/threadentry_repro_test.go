// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

// The frame-pointer LEAVE epilogue that this regression guards against is
// amd64-specific, so the test only runs there.
//go:build !cgo && amd64 && (linux || darwin || freebsd || netbsd)

package fakecgo

import (
	"runtime"
	"sync"
	"testing"
)

// TestThreadEntryReturn exercises the fakecgo threadentry -> runtime.mstart
// return path.
//
// When iscgo is forced true by fakecgo, the runtime creates every new OS
// thread through _cgo_thread_start -> threadentry_trampoline -> threadentry,
// which calls runtime.mstart. For a fakecgo thread the M's stack is
// system-allocated, so when the M exits, mexit(osStack=true) returns from
// mstart back into threadentry. On amd64 with Go's frame-pointer LEAVE epilogue
// this used to crash, because mstart returns with BP clobbered.
//
// Locking an OS thread and then letting the goroutine exit (without calling
// UnlockOSThread) forces the locked fakecgo M to exit, driving mstart to return
// into threadentry. A busy loop that never lets the goroutine exit does NOT
// reproduce the crash, because the M never exits and mstart never returns.
func TestThreadEntryReturn(t *testing.T) {
	const rounds = 50
	const workers = 64
	for r := 0; r < rounds; r++ {
		var wg sync.WaitGroup
		wg.Add(workers)
		for i := 0; i < workers; i++ {
			go func() {
				defer wg.Done()
				runtime.LockOSThread()
				// Intentionally do NOT call UnlockOSThread: returning here exits
				// the goroutine while its OS thread is locked, forcing the M to
				// exit and mstart to return into fakecgo.threadentry.
			}()
		}
		wg.Wait()
	}
}
