// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2023 The Ebitengine Authors

#ifdef __APPLE__
#include <dlfcn.h>
#define DLSYM(F, N) dlsym(F, N)
#elif _WIN32
#define WIN32_LEAN_AND_MEAN
#include "Windows.h"
#define DLSYM(F, N) LoadLibraryA(N)
#endif

struct nested_libdl
{
    nested_libdl() {
        // Fails for sure because a symbol cannot be named like this
        DLSYM(DEFAULT, "@/*<>");
    }
};

static nested_libdl test;
