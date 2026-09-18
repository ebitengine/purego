// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

#include <errno.h>

// setErrno fails the way a C library call does: it returns -1 and sets errno.
int setErrno(void) {
	errno = ENOENT;
	return -1;
}
