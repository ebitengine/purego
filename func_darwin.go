package purego

import (
	"errors"
	"github.com/ebitengine/purego/internal/strings"
	"reflect"
	"unsafe"
)

// Func takes a handle to a shared object returned from Dlopen, the name of a C function in that
// shared object and a pointer to a function representing the calling convention of the C function.
// fptr will be set to a function that when called will call the C function given by name with the
// parameters passed in the correct registers and stack.
// An error is produced if the name symbol cannot be found in handle or if the type is not a function
// pointer or if the function returns more than 1 value.
func Func(handle uintptr, name string, fptr interface{}) error {
	sym := Dlsym(handle, name)
	if sym == 0 {
		return errors.New("purego: couldn't find symbol" + Dlerror())
	}
	fn := reflect.ValueOf(fptr).Elem()
	ty := fn.Type()
	if ty.Kind() != reflect.Func {
		return errors.New("purego: fptr must be a function pointer")
	}
	if ty.NumOut() > 1 {
		return errors.New("purego: function can only return zero or one values")
	}
	v := reflect.MakeFunc(ty, func(args []reflect.Value) (results []reflect.Value) {
		var sysargs = make([]uintptr, len(args))
		for i, v := range args {
			switch v.Kind() {
			case reflect.String:
				sysargs[i] = uintptr(unsafe.Pointer(strings.CString(v.String()))) // TODO: keep alive
			case reflect.Uintptr, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				sysargs[i] = uintptr(v.Uint())
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				sysargs[i] = uintptr(v.Int())
			case reflect.Ptr, reflect.UnsafePointer:
				sysargs[i] = v.Pointer() // TODO: keep alive
			case reflect.Func:
				sysargs[i] = NewCallback(v.Interface())
			default:
				panic("purego: unsupported kind: " + v.Kind().String())
			}
		}
		r1, _, _ := SyscallN(sym, sysargs...)
		if ty.NumOut() == 0 {
			return nil
		}
		outType := ty.Out(0)
		v := reflect.New(outType).Elem()
		switch outType.Kind() {
		case reflect.Uintptr, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			v.SetUint(uint64(r1))
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			v.SetInt(int64(r1))
		case reflect.Bool:
			v.SetBool(r1 != 0)
		case reflect.Ptr:
			// We take the address and then dereference it to trick go vet from creating a possible miss-use of unsafe.Pointer
			v = reflect.NewAt(outType, *(*unsafe.Pointer)(unsafe.Pointer(&r1))).Elem()
		default:
			panic("purego: unsupported return kind: " + outType.Kind().String())
		}
		return []reflect.Value{v}
	})
	fn.Set(v)
	return nil
}
