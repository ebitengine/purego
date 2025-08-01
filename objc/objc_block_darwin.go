// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 The Ebitengine Authors

package objc

import (
	"fmt"
	"reflect"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
)

const (
	// The end-goal of these defaults is to get an Objective-C memory-managed block object
	// that won't try to free() a Go pointer, but will call our custom blockFunctionCache.Delete()
	// when the reference count drops to zero, so the associated function is also unreferenced.

	// blockBaseClass is the name of the class that block objects will be initialized with.
	blockBaseClass = "__NSMallocBlock__"
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

	if typ.NumIn() == 0 || typ.In(0) != reflect.TypeOf(Block(0)) {
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
				return b.Functions.Load(args[0].Interface().(Block)).Call(args)
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

// Invoke calls the implementation of a block.
func (b Block) Invoke(args ...any) {
	fn := theBlocksCache.Functions.Load(b)

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
func InvokeBlock[T any](block Block, args ...any) (result T, err error) {
	block = block.Copy()
	defer block.Release()

	fn := theBlocksCache.Functions.Load(block)
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
	result, ok = callResult[0].Interface().(T)
	if !ok {
		return result, fmt.Errorf("objc: the returned value type %s was not %T", callResult[0].Type().String(), result)
	}

	return result, nil
}
