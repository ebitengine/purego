// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

#include <errno.h>

// setErrno fails the way a C library call does: it returns -1 and sets errno.
//
// It takes no argument on purpose. SyscallN mirrors every integer argument into
// the float slots and the C fallback used on the Linux architectures that have
// no assembly trampoline asserts that those slots are zero, so a call with
// arguments would abort there before the function is reached.
int setErrno(void) {
	errno = ENOENT;
	return -1;
}
