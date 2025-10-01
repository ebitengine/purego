// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package purego_test

import (
	"errors"
	"syscall"
	"testing"

	_ "github.com/ebitengine/purego"
)

func TestAllThreadsSyscall(t *testing.T) {
	_, _, err := syscall.AllThreadsSyscall(syscall.SYS_FCNTL, 0, 0, 0)
	if err != syscall.ENOTSUP {
		t.Errorf("AllThreadsSyscall should return ENOTSUP, got: %v", err)
	}
}

// TestSetuidEtc performs tests on all of the wrapped system calls
// that mirror to the 9 glibc syscalls with POSIX semantics. The test
// here is considered authoritative and should compile and run
// CGO_ENABLED=0 or 1.
func TestSetuidEtc(t *testing.T) {
	if syscall.Getuid() == 0 {
		// We just want to verify that the syscall wrappers
		// work, not that we can actually change the uid/gid
		// of the test process.
		t.Skip("skipping non-root test as root")
	}
	vs := []struct {
		call string
		fn   func() error
	}{
		{call: "Setegid(1)", fn: func() error { return syscall.Setegid(1) }},
		{call: "Seteuid(1)", fn: func() error { return syscall.Seteuid(1) }},
		{call: "Setgid(1)", fn: func() error { return syscall.Setgid(1) }},
		{call: "Setgroups([]int{0,1,2,3})", fn: func() error { return syscall.Setgroups([]int{0, 1, 2, 3}) }},
		{call: "Setregid(101,0)", fn: func() error { return syscall.Setregid(101, 0) }},
		{call: "Setreuid(1,0)", fn: func() error { return syscall.Setreuid(1, 0) }},
		{call: "Setresgid(101,0,102)", fn: func() error { return syscall.Setresgid(101, 0, 102) }},
		{call: "Setresuid(1,0,2)", fn: func() error { return syscall.Setresuid(1, 0, 2) }},
	}

	for _, v := range vs {
		t.Run(v.call, func(t *testing.T) {
			if err := v.fn(); !errors.Is(err, syscall.EPERM) {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
