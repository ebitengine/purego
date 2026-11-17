// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

#include <pthread.h>
#include <stdint.h>

// Invoke callbacks with enough arguments to exercise both integer registers and
// stack arguments. Keeping the loop in C isolates native-to-Go callback costs.
uintptr_t callWordCallback(uintptr_t callback, int count,
                           const uintptr_t *args, uintptr_t iterations) {
    uintptr_t result = 0;
    for (uintptr_t i = 0; i < iterations; i++) {
        switch (count) {
        case 0:
            result += ((uintptr_t (*)(void))callback)();
            break;
        case 1:
            result += ((uintptr_t (*)(uintptr_t))callback)(args[0]);
            break;
        case 2:
            result += ((uintptr_t (*)(uintptr_t, uintptr_t))callback)(args[0], args[1]);
            break;
        case 3:
            result += ((uintptr_t (*)(uintptr_t, uintptr_t, uintptr_t))callback)(args[0], args[1], args[2]);
            break;
        case 4:
            result += ((uintptr_t (*)(uintptr_t, uintptr_t, uintptr_t, uintptr_t))callback)(
                args[0], args[1], args[2], args[3]);
            break;
        case 5:
            result += ((uintptr_t (*)(uintptr_t, uintptr_t, uintptr_t, uintptr_t, uintptr_t))callback)(
                args[0], args[1], args[2], args[3], args[4]);
            break;
        case 6:
            result += ((uintptr_t (*)(uintptr_t, uintptr_t, uintptr_t, uintptr_t, uintptr_t, uintptr_t))callback)(
                args[0], args[1], args[2], args[3], args[4], args[5]);
            break;
        case 7:
            result += ((uintptr_t (*)(uintptr_t, uintptr_t, uintptr_t, uintptr_t, uintptr_t, uintptr_t, uintptr_t))callback)(
                args[0], args[1], args[2], args[3], args[4], args[5], args[6]);
            break;
        case 8:
            result += ((uintptr_t (*)(uintptr_t, uintptr_t, uintptr_t, uintptr_t, uintptr_t, uintptr_t, uintptr_t, uintptr_t))callback)(
                args[0], args[1], args[2], args[3], args[4], args[5], args[6], args[7]);
            break;
        case 9:
            result += ((uintptr_t (*)(uintptr_t, uintptr_t, uintptr_t, uintptr_t, uintptr_t, uintptr_t, uintptr_t, uintptr_t, uintptr_t))callback)(
                args[0], args[1], args[2], args[3], args[4], args[5], args[6], args[7], args[8]);
            break;
        }
    }
    return result;
}

struct word_callback_thread {
    uintptr_t callback;
    uintptr_t result;
};

static void *runWordCallback(void *arg) {
    struct word_callback_thread *state = arg;
    state->result = ((uintptr_t (*)(void))state->callback)();
    return NULL;
}

int callWordCallbackOnThread(uintptr_t callback, uintptr_t *result) {
    struct word_callback_thread state = {callback, 0};
    pthread_t thread;
    int rc = pthread_create(&thread, NULL, runWordCallback, &state);
    if (rc != 0) {
        return rc;
    }
    rc = pthread_join(thread, NULL);
    if (rc == 0) {
        *result = state.result;
    }
    return rc;
}
