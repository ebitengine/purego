// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2023 The Ebitengine Authors

#ifdef _WIN32
#include <libloaderapi.h>
#else
#include <dlfcn.h>
#endif

struct nested_libdl
{
    nested_libdl() {
#ifdef _WIN32
        // Fails for sure because a symbol cannot be named like this
        GetProcAddress(GetModuleHandleA(NULL), "@/*<>");
#else
        // Fails for sure because a symbol cannot be named like this
        dlsym(RTLD_DEFAULT, "@/*<>");
#endif
    }
};

static nested_libdl test;
