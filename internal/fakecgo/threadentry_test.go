// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

// 386/arm are excluded: the standalone fakecgo test binary resolves pthread
// from libpthread.so.0, which fails the dynamic symbol lookup on 32-bit linux.
//go:build !cgo && (linux || darwin || freebsd || netbsd) && !386 && !arm

package fakecgo

import (
	"runtime"
	"sync"
	"testing"
)

// TestThreadEntryReturn exercises the fakecgo threadentry -> runtime.mstart
// return path, i.e. that an M created by fakecgo can exit cleanly.
//
// When iscgo is forced true by fakecgo, the runtime creates every new OS
// thread through _cgo_thread_start -> threadentry_trampoline -> threadentry,
// which calls runtime.mstart. For a fakecgo thread the M's stack is
// system-allocated, so when the M exits, mexit(osStack=true) returns from
// mstart back into threadentry. This return path has been the source of
// crashes (a frame-pointer LEAVE fault on amd64, a bad indirect call and race
// instrumentation on other platforms), so it is worth exercising everywhere.
//
// Locking an OS thread and then letting the goroutine exit (without calling
// UnlockOSThread) forces the locked fakecgo M to exit, driving mstart to return
// into threadentry. A busy loop that never lets the goroutine exit does NOT
// reproduce the crash, because the M never exits and mstart never returns.
func TestThreadEntryReturn(t *testing.T) {
	// The M exits on the first goroutine return, so a modest amount of churn is
	// enough to catch a broken teardown while staying quick under emulation.
	const rounds = 10
	const workers = 16
	for range rounds {
		var wg sync.WaitGroup
		wg.Add(workers)
		for range workers {
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
