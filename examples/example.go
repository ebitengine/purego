//go:build darwin || windows
// +build darwin windows

package main

import (
	"fmt"
	_ "github.com/ebitengine/purego"
)

func main() {
	fmt.Println("Hello World")
}
