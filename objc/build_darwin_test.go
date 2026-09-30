// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2026 The Ebitengine Authors

package objc_test

import (
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"testing"
)

// buildSharedLib compiles the Objective-C sources into a dynamic library at libFile.
// The library is built for GOARCH rather than the compiler's default, because
// macOS arm64 can run amd64 tests through Rosetta.
func buildSharedLib(tb testing.TB, libFile string, sources ...string) error {
	tb.Helper()

	out, err := exec.Command("go", "env", "CC").Output()
	if err != nil {
		return fmt.Errorf("go env CC error: %w", err)
	}
	compiler := strings.TrimSpace(string(out))
	if compiler == "" {
		return errors.New("compiler not found")
	}

	var arch string
	switch runtime.GOARCH {
	case "arm64":
		arch = "arm64"
	case "amd64":
		arch = "x86_64"
	default:
		return fmt.Errorf("unknown macOS architecture %s", runtime.GOARCH)
	}

	args := []string{"-dynamiclib", "-Wall", "-Werror", "-arch", arch, "-framework", "Foundation", "-o", libFile}
	cmd := exec.Command(compiler, append(args, sources...)...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("compile lib: %w\n%q\n%s", err, cmd, string(out))
	}
	return nil
}
