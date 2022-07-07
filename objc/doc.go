// Package objc is a low-level pure Go objective-c runtime. This package is easy to use incorrectly so it is best
// to use a wrapper that provides the functionality you need in a safer way.
//
// All functions that take a string as an argument are NULL terminated ('\x00'). This is so that there is no
// need to copy the string and put pressure on the GC. The decision to go this route is because objective-c
// calls into the runtime a lot and there would be a lot of time wasted just copying strings around.
package objc
