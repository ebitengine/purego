// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package objc

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"structs"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
)

const (
	// The end-goal of these defaults is to get an Objective-C memory-managed block object
	// that won't try to free() a Go pointer, but will call our custom blockFunctionCache.Delete()
	// when the reference count drops to zero, so the associated function is also unreferenced.

	// blockBaseClass is the name of the class that block objects will be initialized with.
	// A new block is a template that Block_copy always relocates to the
	// Objective-C heap, so the stack-block class is the accurate name here,
	// as the isa of the copied block becomes __NSMallocBlock__ anyway.
	blockBaseClass = "__NSStackBlock__"
	// blockFlags is the set of flags that block objects will be initialized with.
	blockFlags = blockHasCopyDispose | blockHasSignature

	// blockHasCopyDispose is a flag that tells the Objective-C runtime the block exports Copy and/or Dispose helpers.
	blockHasCopyDispose = 1 << 25

	// blockHasSignature is a flag that tells the Objective-C runtime the block exports a function signature.
	blockHasSignature = 1 << 30
)

// blockDescriptor is the Go representation of an Objective-C block descriptor.
// It is a component to be referenced by blockDescriptor.
//
// The layout of this struct matches Block_literal_1 described in https://clang.llvm.org/docs/Block-ABI-Apple.html#high-level
type blockDescriptor struct {
	_         structs.HostLayout
	_         uintptr
	size      uintptr
	_         uintptr
	dispose   uintptr
	signature *uint8
}

// blockLayout is the Go representation of the structure abstracted by a block pointer.
// From the Objective-C point of view, a pointer to this struct is equivalent to an ID that
// references a block.
//
// The layout of this struct matches __block_literal_1 described in https://clang.llvm.org/docs/Block-ABI-Apple.html#high-level
type blockLayout struct {
	_          structs.HostLayout
	isa        Class
	flags      uint32
	_          uint32
	invoke     uintptr
	descriptor *blockDescriptor
}

// blockFunctionCache is a thread safe cache of block layouts.
//
// The function closures themselves are kept alive by caching them internally until the Objective-C runtime indicates that
// they can be released (presumably when the reference count reaches zero). This approach is used instead of appending the function
// object to the block allocation, where it is out of the visible domain of Go's GC.
type blockFunctionCache struct {
	mutex     sync.RWMutex
	functions map[Block]reflect.Value
}

// Load retrieves a function (in the form of a reflect.Value, so Call can be invoked) associated with the key Block.
func (b *blockFunctionCache) Load(key Block) reflect.Value {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return b.functions[key]
}

// Store associates a function (in the form of a reflect.Value) with the key Block.
func (b *blockFunctionCache) Store(key Block, value reflect.Value) Block {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.functions[key] = value
	return key
}

// Delete removed the function associated with the key Block.
func (b *blockFunctionCache) Delete(key Block) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	delete(b.functions, key)
}

// newBlockFunctionCache initializes a new blockFunctionCache
func newBlockFunctionCache() *blockFunctionCache {
	return &blockFunctionCache{functions: map[Block]reflect.Value{}}
}

// blockCache is a thread safe cache of block layouts.
//
// It takes advantage of the block being the first argument of a block call being the block closure,
// only invoking [github.com/ebitengine/purego.NewCallback] when it encounters a new function type (rather than on for every block creation).
// This should mitigate block creations putting pressure on the callback limit.
type blockCache struct {
	sync.Mutex
	descriptorTemplate blockDescriptor
	layoutTemplate     blockLayout
	layouts            map[reflect.Type]blockLayout
	Functions          *blockFunctionCache
}

// encode returns a blocks type as if it was given to @encode(typ)
func (*blockCache) encode(typ reflect.Type) *uint8 {
	// this algorithm was copied from encodeFunc,
	// but altered to panic on error, and to only accept a block-type signature.
	if typ == nil || typ.Kind() != reflect.Func {
		panic("objc: not a function")
	}

	var encoding string
	switch typ.NumOut() {
	case 0:
		encoding = encVoid
	default:
		returnType, err := encodeType(typ.Out(0), false)
		if err != nil {
			panic(fmt.Sprintf("objc: %v", err))
		}
		encoding = returnType
	}

	if typ.NumIn() == 0 || typ.In(0) != reflect.TypeFor[Block]() {
		panic(fmt.Sprintf("objc: A Block implementation must take a Block as its first argument; got %v", typ.String()))
	}

	encoding += encId
	for i := 1; i < typ.NumIn(); i++ {
		argType, err := encodeType(typ.In(i), false)
		if err != nil {
			panic(fmt.Sprintf("objc: %v", err))
		}
		encoding = fmt.Sprint(encoding, argType)
	}

	// return the encoding as a C-style string.
	return &append([]uint8(encoding), 0)[0]
}

// getLayout retrieves a blockLayout VALUE constructed with the supplied function type.
// It will panic if the type is not a valid block function.
func (b *blockCache) getLayout(typ reflect.Type) blockLayout {
	b.Lock()
	defer b.Unlock()

	// return the cached layout, if it exists.
	if layout, ok := b.layouts[typ]; ok {
		return layout
	}

	// otherwise: create a layout, and populate it with the default templates
	layout := b.layoutTemplate
	layout.descriptor = &blockDescriptor{}
	*layout.descriptor = b.descriptorTemplate

	// getting the signature now will panic on invalid types before we invest in creating a callback.
	layout.descriptor.signature = b.encode(typ)

	// create a global callback.
	// this single callback can dispatch to any function with the same signature,
	// since the user-provided functions are associated with the actual block allocations.
	layout.invoke = purego.NewCallback(
		reflect.MakeFunc(
			typ,
			func(args []reflect.Value) (results []reflect.Value) {
				block, ok := reflect.TypeAssert[Block](args[0])
				if !ok {
					panic(fmt.Sprintf("objc: block argument is not a block but %s", args[0].Type().String()))
				}
				return b.Functions.Load(block).Call(args)
			},
		).Interface(),
	)

	// store it and return it
	b.layouts[typ] = layout
	return layout
}

// newBlockCache initializes a block cache.
// It should not be called until AFTER libobjc is fully initialized.
func newBlockCache() *blockCache {
	cache := &blockCache{
		descriptorTemplate: blockDescriptor{
			size: unsafe.Sizeof(blockLayout{}),
		},
		layoutTemplate: blockLayout{
			isa:   GetClass(blockBaseClass),
			flags: blockFlags,
		},
		layouts:   map[reflect.Type]blockLayout{},
		Functions: newBlockFunctionCache(),
	}
	cache.descriptorTemplate.dispose = purego.NewCallback(cache.Functions.Delete)
	return cache
}

// theBlocksCache is the global block cache
var theBlocksCache *blockCache

// Block is an opaque pointer to an Objective-C object containing a function with its associated closure.
type Block ID

// Copy creates a copy of a block on the Objective-C heap (or increments the reference count if already on the heap).
// Use [Block.Release] to free the copy when it is no longer in use.
func (b Block) Copy() Block {
	return _Block_copy(b)
}

// callForeign calls a block that was not created by [NewBlock] (for example one
// handed to us by Objective-C) through the Blocks ABI: the block is passed as the
// first argument followed by args. The signature is derived from the dynamic
// types of args and from resultType (nil for no result), and is checked against
// the block's own signature when it has one.
// See https://clang.llvm.org/docs/Block-ABI-Apple.html.
func (b Block) callForeign(resultType reflect.Type, args []any) ([]reflect.Value, error) {
	if b == 0 {
		return nil, errors.New("objc: cannot invoke a nil block")
	}
	invoke := (*(**blockLayout)(unsafe.Pointer(&b))).invoke
	if invoke == 0 {
		return nil, errors.New("objc: block has no invoke function")
	}

	key := foreignKey{invoke: invoke, result: resultType, nargs: len(args)}
	reflectedArgs := make([]reflect.Value, len(args)+1)
	reflectedArgs[0] = reflect.ValueOf(b)
	for i, arg := range args {
		if arg == nil {
			return nil, fmt.Errorf("objc: argument %d to a block is nil; pass a typed value such as objc.ID(0)", i)
		}
		reflectedArgs[i+1] = reflect.ValueOf(arg)
		if i < len(key.args) {
			key.args[i] = reflectedArgs[i+1].Type()
		}
	}
	if len(args) > len(key.args) {
		key.typ = foreignFuncOf(reflectedArgs, resultType)
	}
	key.sig, key.hasSig = b.signature()

	call, ok := foreignCalls.Load(key)
	if !ok {
		call = newForeignCall(key, reflectedArgs)
		// key.sig refers to the block's descriptor, which may not outlive the block.
		key.sig = strings.Clone(key.sig)
		call, _ = foreignCalls.LoadOrStore(key, call)
	}
	if err := call.(*foreignCall).err; err != nil {
		return nil, err
	}
	return call.(*foreignCall).fn.Call(reflectedArgs), nil
}

// foreignKey identifies a call made by callForeign: the block's implementation and
// type signature, and the types of the result and arguments of the call.
type foreignKey struct {
	invoke uintptr
	sig    string
	hasSig bool
	result reflect.Type
	nargs  int
	// args holds the types of the arguments, so that the common case can be looked up
	// without building a func type. A call with more arguments uses typ instead.
	args [8]reflect.Type
	typ  reflect.Type
}

// foreignCall is a checked call to a foreign block: either a function that calls
// the block's implementation, or the reason the block cannot be called that way.
type foreignCall struct {
	fn  reflect.Value
	err error
}

// foreignCalls caches calls to foreign blocks, so that a block that is called
// repeatedly, such as an enumeration handler, is only checked and registered once.
var foreignCalls sync.Map // map[foreignKey]*foreignCall

// newForeignCall checks and registers the call identified by key, with the given arguments.
func newForeignCall(key foreignKey, args []reflect.Value) *foreignCall {
	typ := key.typ
	if typ == nil {
		typ = foreignFuncOf(args, key.result)
	}
	if err := checkForeign(key.sig, key.hasSig, typ); err != nil {
		return &foreignCall{err: err}
	}
	fn := reflect.New(typ)
	purego.RegisterFunc(fn.Interface(), key.invoke)
	return &foreignCall{fn: fn.Elem()}
}

// foreignFuncOf returns the type of a function that takes args and returns result, if not nil.
func foreignFuncOf(args []reflect.Value, result reflect.Type) reflect.Type {
	in := make([]reflect.Type, len(args))
	for i, arg := range args {
		in[i] = arg.Type()
	}
	var out []reflect.Type
	if result != nil {
		out = []reflect.Type{result}
	}
	return reflect.FuncOf(in, out, false)
}

// checkForeign reports whether a block with the type signature sig, if hasSig,
// can be called as a function of type typ, whose first argument is the block.
func checkForeign(sig string, hasSig bool, typ reflect.Type) error {
	// goTypes are in the order of a signature: the result (nil to discard it), then the block, then its parameters.
	goTypes := make([]reflect.Type, 1, typ.NumIn()+1)
	if typ.NumOut() > 0 {
		goTypes[0] = typ.Out(0)
	}
	for i := range typ.NumIn() {
		goTypes = append(goTypes, typ.In(i))
	}
	layouts := make([]abiLayout, len(goTypes))
	for i, t := range goTypes {
		if t == nil {
			continue
		}
		var err error
		if layouts[i], err = goLayout(t); err != nil {
			return err
		}
	}
	if !hasSig {
		return nil
	}

	types, err := splitSignature(sig)
	if err != nil || len(types) < 2 {
		return fmt.Errorf("objc: malformed block signature %q", sig)
	}
	if len(types) != len(goTypes) {
		return fmt.Errorf("objc: block callback expects %d arguments, got %d", len(types)-2, len(goTypes)-2)
	}
	if goTypes[0] == nil && (types[0][0] == '{' || types[0][0] == '(') {
		// The caller has to provide the result buffer for a struct returned in memory.
		return fmt.Errorf("objc: block returns %s; use InvokeBlock to receive it", types[0])
	}
	for i, enc := range types {
		if goTypes[i] == nil {
			continue
		}
		what := "result"
		if i > 0 {
			what = fmt.Sprintf("argument %d", i-2)
		}
		want, err := encodingLayout(enc)
		if err != nil {
			return fmt.Errorf("objc: block %s %s: %w", what, enc, err)
		}
		if !want.matches(layouts[i]) {
			return fmt.Errorf("objc: block %s is %s, which does not match %s", what, enc, goTypes[i])
		}
	}
	return nil
}

// Invoke calls the implementation of a block, discarding any result.
//
// A block that was not created by [NewBlock], such as a completion handler
// supplied by Objective-C, is called through the Blocks ABI. The Go type of each
// argument must then match the block's C parameter type (for example, int32 for
// an int), and is checked against the block's type signature when it has one.
// Go func arguments are not supported; pass a pointer from [purego.NewCallback]
// instead. Invoke panics on a mismatch, and on a block that returns a struct
// must be called with [InvokeBlock].
func (b Block) Invoke(args ...any) {
	fn := theBlocksCache.Functions.Load(b)
	if !fn.IsValid() {
		if _, err := b.callForeign(nil, args); err != nil {
			panic(err)
		}
		return
	}

	reflectedArgs := make([]reflect.Value, len(args)+1)
	reflectedArgs[0] = reflect.ValueOf(b)
	for i := range args {
		reflectedArgs[i+1] = reflect.ValueOf(args[i])
	}

	fn.Call(reflectedArgs)
}

// Release decrements the Block's reference count, and if it is the last reference, frees it.
func (b Block) Release() {
	_Block_release(b)
}

// NewBlock takes a Go function that takes a Block as its first argument.
// It returns an Block that can be called by Objective-C code.
// The function panics if an error occurs.
// Use [Block.Release] to free this block when it is no longer in use.
func NewBlock(fn any) Block {
	// get or create a block layout for the callback.
	layout := theBlocksCache.getLayout(reflect.TypeOf(fn))
	// we created the layout in Go memory, so we'll copy it to a newly-created Objective-C object.
	block := Block(unsafe.Pointer(&layout)).Copy()
	// associate the fn with the block we created before returning it.
	return theBlocksCache.Functions.Store(block, reflect.ValueOf(fn))
}

// InvokeBlock is a convenience method for calling the implementation of a block.
// The block implementation must return 1 value.
//
// A block that was not created by [NewBlock] is called as described for
// [Block.Invoke], with T as the type of its result. T must match the block's C
// result type, and may be a struct. A mismatch returns an error instead of
// calling the block.
func InvokeBlock[T any](block Block, args ...any) (result T, err error) {
	fn := theBlocksCache.Functions.Load(block)
	if !fn.IsValid() {
		// not one of ours. The block may live on the caller's stack, so it must not be
		// copied: the copy would be a different pointer and is unnecessary for a synchronous call.
		out, err := block.callForeign(reflect.TypeFor[T](), args)
		if err != nil {
			return result, err
		}
		result, _ = reflect.TypeAssert[T](out[0])
		return result, nil
	}

	// NewBlock returns a heap block, which Copy retains rather than moving, so fn still applies.
	block = block.Copy()
	defer block.Release()

	if fn.Type().NumIn() != len(args)+1 {
		return result, fmt.Errorf("objc: block callback expects %d arguments, got %d", fn.Type().NumIn()-1, len(args))
	}

	reflectedArgs := make([]reflect.Value, len(args)+1)
	reflectedArgs[0] = reflect.ValueOf(block)
	for i := range args {
		reflectedArgs[i+1] = reflect.ValueOf(args[i])
	}

	callResult := fn.Call(reflectedArgs)

	var ok bool
	result, ok = reflect.TypeAssert[T](callResult[0])
	if !ok {
		return result, fmt.Errorf("objc: the returned value type %s was not %T", callResult[0].Type().String(), result)
	}

	return result, nil
}
