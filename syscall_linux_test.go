// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package purego_test

import (
	"fmt"
	"os"
	"runtime"
	"slices"
	"strings"
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
	if syscall.Getuid() != 0 {
		t.Skip("skipping root only test")
	}
	if syscall.Getgid() != 0 {
		t.Skip("skipping the test when root's gid is not default value 0")
	}
	vs := []struct {
		call           string
		fn             func() error
		filter, expect string
	}{
		{call: "Setegid(1)", fn: func() error { return syscall.Setegid(1) }, filter: "Gid:", expect: "\t0\t1\t0\t1"},
		{call: "Setegid(0)", fn: func() error { return syscall.Setegid(0) }, filter: "Gid:", expect: "\t0\t0\t0\t0"},

		{call: "Seteuid(1)", fn: func() error { return syscall.Seteuid(1) }, filter: "Uid:", expect: "\t0\t1\t0\t1"},
		{call: "Setuid(0)", fn: func() error { return syscall.Setuid(0) }, filter: "Uid:", expect: "\t0\t0\t0\t0"},

		{call: "Setgid(1)", fn: func() error { return syscall.Setgid(1) }, filter: "Gid:", expect: "\t1\t1\t1\t1"},
		{call: "Setgid(0)", fn: func() error { return syscall.Setgid(0) }, filter: "Gid:", expect: "\t0\t0\t0\t0"},

		{call: "Setgroups([]int{0,1,2,3})", fn: func() error { return syscall.Setgroups([]int{0, 1, 2, 3}) }, filter: "Groups:", expect: "\t0 1 2 3"},
		{call: "Setgroups(nil)", fn: func() error { return syscall.Setgroups(nil) }, filter: "Groups:", expect: ""},
		{call: "Setgroups([]int{0})", fn: func() error { return syscall.Setgroups([]int{0}) }, filter: "Groups:", expect: "\t0"},

		{call: "Setregid(101,0)", fn: func() error { return syscall.Setregid(101, 0) }, filter: "Gid:", expect: "\t101\t0\t0\t0"},
		{call: "Setregid(0,102)", fn: func() error { return syscall.Setregid(0, 102) }, filter: "Gid:", expect: "\t0\t102\t102\t102"},
		{call: "Setregid(0,0)", fn: func() error { return syscall.Setregid(0, 0) }, filter: "Gid:", expect: "\t0\t0\t0\t0"},

		{call: "Setreuid(1,0)", fn: func() error { return syscall.Setreuid(1, 0) }, filter: "Uid:", expect: "\t1\t0\t0\t0"},
		{call: "Setreuid(0,2)", fn: func() error { return syscall.Setreuid(0, 2) }, filter: "Uid:", expect: "\t0\t2\t2\t2"},
		{call: "Setreuid(0,0)", fn: func() error { return syscall.Setreuid(0, 0) }, filter: "Uid:", expect: "\t0\t0\t0\t0"},

		{call: "Setresgid(101,0,102)", fn: func() error { return syscall.Setresgid(101, 0, 102) }, filter: "Gid:", expect: "\t101\t0\t102\t0"},
		{call: "Setresgid(0,102,101)", fn: func() error { return syscall.Setresgid(0, 102, 101) }, filter: "Gid:", expect: "\t0\t102\t101\t102"},
		{call: "Setresgid(0,0,0)", fn: func() error { return syscall.Setresgid(0, 0, 0) }, filter: "Gid:", expect: "\t0\t0\t0\t0"},

		{call: "Setresuid(1,0,2)", fn: func() error { return syscall.Setresuid(1, 0, 2) }, filter: "Uid:", expect: "\t1\t0\t2\t0"},
		{call: "Setresuid(0,2,1)", fn: func() error { return syscall.Setresuid(0, 2, 1) }, filter: "Uid:", expect: "\t0\t2\t1\t2"},
		{call: "Setresuid(0,0,0)", fn: func() error { return syscall.Setresuid(0, 0, 0) }, filter: "Uid:", expect: "\t0\t0\t0\t0"},
	}

	for i, v := range vs {
		// Generate some thread churn as we execute the tests.
		c := make(chan struct{})
		go killAThread(c)
		close(c)

		if err := v.fn(); err != nil {
			t.Errorf("[%d] %q failed: %v", i, v.call, err)
			continue
		}
		if err := compareStatus(v.filter, v.expect); err != nil {
			t.Errorf("[%d] %q comparison: %v", i, v.call, err)
		}
	}
}

// killAThread locks the goroutine to an OS thread and exits; this
// causes an OS thread to terminate.
func killAThread(c <-chan struct{}) {
	runtime.LockOSThread()
	<-c
}

// compareStatus is used to confirm the contents of the thread
// specific status files match expectations.
func compareStatus(filter, expect string) error {
	expected := filter + expect
	pid := syscall.Getpid()
	fs, err := os.ReadDir(fmt.Sprintf("/proc/%d/task", pid))
	if err != nil {
		return fmt.Errorf("unable to find %d tasks: %v", pid, err)
	}
	expectedProc := fmt.Sprintf("Pid:\t%d", pid)
	foundAThread := false
	for _, f := range fs {
		tf := fmt.Sprintf("/proc/%s/status", f.Name())
		d, err := os.ReadFile(tf)
		if err != nil {
			// There are a surprising number of ways this
			// can error out on linux.  We've seen all of
			// the following, so treat any error here as
			// equivalent to the "process is gone":
			//    os.IsNotExist(err),
			//    "... : no such process",
			//    "... : bad file descriptor.
			continue
		}
		lines := strings.Split(string(d), "\n")
		for _, line := range lines {
			// Different kernel vintages pad differently.
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "Pid:\t") {
				// On loaded systems, it is possible
				// for a TID to be reused really
				// quickly. As such, we need to
				// validate that the thread status
				// info we just read is a task of the
				// same process PID as we are
				// currently running, and not a
				// recently terminated thread
				// resurfaced in a different process.
				if line != expectedProc {
					break
				}
				// Fall through in the unlikely case
				// that filter at some point is
				// "Pid:\t".
			}
			if strings.HasPrefix(line, filter) {
				if line == expected {
					foundAThread = true
					break
				}
				if filter == "Groups:" && strings.HasPrefix(line, "Groups:\t") {
					// https://github.com/golang/go/issues/46145
					// Containers don't reliably output this line in sorted order so manually sort and compare that.
					a := strings.Split(line[8:], " ")
					slices.Sort(a)
					got := strings.Join(a, " ")
					if got == expected[8:] {
						foundAThread = true
						break
					}

				}
				return fmt.Errorf("%q got:%q want:%q (bad) [pid=%d file:'%s' %v]\n", tf, line, expected, pid, string(d), expectedProc)
			}
		}
	}
	if !foundAThread {
		return fmt.Errorf("found no thread /proc/<TID>/status files for process %q", expectedProc)
	}
	return nil
}
