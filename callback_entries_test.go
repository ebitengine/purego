// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

//go:build darwin || (linux && (386 || amd64 || arm || arm64 || loong64 || ppc64le || riscv64 || (s390x && (cgo || go1.27))))

package purego_test

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/ebitengine/purego"
)

const callbackEntriesChildEnv = "PUREGO_TEST_CALLBACK_ENTRIES_CHILD"

// TestCallbackEntries calls every callback entry and checks that each one
// reaches its own Go function. Callback slots are never released, so the
// callbacks are registered in a child process to keep the slots of this
// process available to other tests.
func TestCallbackEntries(t *testing.T) {
	if os.Getenv(callbackEntriesChildEnv) == "1" {
		callEveryCallbackEntry(t)
		return
	}
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(exe, "-test.run=^TestCallbackEntries$", "-test.v")
	cmd.Env = append(os.Environ(), callbackEntriesChildEnv+"=1")
	if prefix := os.Getenv("QEMU_LD_PREFIX"); prefix != "" {
		// QEMU_LD_PREFIX only tells qemu-user where the target's shared
		// libraries are. The child might not run under the same emulator:
		// an arm64 host with AArch32 support runs a linux/arm binary natively,
		// and the dynamic loader then searches the host's directories. Point
		// the loader at the target's libraries directly so the child starts
		// either way. Under qemu-user the path still resolves to the same
		// libraries.
		libs := filepath.Join(prefix, "lib")
		if old := os.Getenv("LD_LIBRARY_PATH"); old != "" {
			libs += string(filepath.ListSeparator) + old
		}
		cmd.Env = append(cmd.Env, "LD_LIBRARY_PATH="+libs)
	}
	out, err := cmd.CombinedOutput()
	if errors.Is(err, syscall.ENOEXEC) {
		// For example, the test binary runs under a user-mode emulator without
		// binfmt_misc registered, so the child cannot be started.
		t.Skipf("cannot re-execute the test binary: %v", err)
	}
	if err != nil {
		t.Fatalf("child process failed: %v\n%s", err, out)
	}
}

func callEveryCallbackEntry(t *testing.T) {
	callbacks := make([]uintptr, purego.MaxCB)
	for i := range callbacks {
		callbacks[i] = purego.NewCallback(func() uintptr { return uintptr(i) })
	}
	const maxFailures = 20
	failures := 0
	for i, cb := range callbacks {
		if got, _, _ := purego.SyscallN(cb); got != uintptr(i) {
			t.Errorf("callback %d returned %d", i, got)
			if failures++; failures >= maxFailures {
				t.Fatalf("stopping after %d failures", failures)
			}
		}
	}
}
