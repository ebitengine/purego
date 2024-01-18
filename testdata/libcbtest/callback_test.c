// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2023 The Ebitengine Authors

#include <string.h>

typedef int (*callback)(const char *, int);

int callCallback(const void *fp, const char *s) {
    // If the callback corrupts FP, this local variable on the stack will have incorrect value.
    int sentinel = 10101;
    ((callback)(fp))(s, strlen(s));
    return sentinel;
}
