// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2023 The Ebitengine Authors

#include <dlfcn.h>

struct nested_libdl
{
    nested_libdl() {
        // Fails for sure because a symbol cannot be named like this 
        dlsym(RTLD_DEFAULT, "@/*<>");
    }
};

static nested_libdl test;
