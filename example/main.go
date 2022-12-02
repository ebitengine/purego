// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2022 The Ebitengine Authors

//go:build darwin || linux || windows

package main

import (
	"os"

	_ "github.com/ebitengine/purego"
)

func main() {
	// set and unset an environment variable since this calls into fakecgo.
	err := os.Setenv("TESTING", "SOMETHING")
	if err != nil {
		panic(err)
	}
	err = os.Unsetenv("TESTING")
	if err != nil {
		panic(err)
	}
}
