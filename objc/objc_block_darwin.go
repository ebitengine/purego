// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2024 The Ebitengine Authors

package objc

import (
	"fmt"
	"reflect"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
)

const (
	// end-goal of these defaults is to get an Objectve-C memory-managed block object,
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
type blockDescriptor struct {
	_         uintptr
	Size      uintptr
	_         uintptr
	Dispose   uintptr
	Signature *uint8
}

// blockLayout is the Go representation of the structure abstracted by a block pointer.
// From the Objective-C point of view, a pointer to this struct is equivalent to an ID that
// references a block.
type blockLayout struct {
	Isa        Class
	Flags      uint32
	_          uint32
	Invoke     uintptr
	Descriptor *blockDescriptor
}

/*
blockCache is a thread safe cache of block layouts.

The function closures themselves are kept alive by caching them internally until the Objective-C runtime indicates that
they can be released (presumably when the reference count reaches zero.) This approach is used instead of appending the function
object to the block allocation, where it is out of the visible domain of Go's GC.
*/
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

// newBlockFunctionCache initilizes a new blockFunctionCache
func newBlockFunctionCache() *blockFunctionCache {
	return &blockFunctionCache{functions: map[Block]reflect.Value{}}
}

/*
blockCache is a thread safe cache of block layouts.

It takes advantage of the block being the first argument of a block call being the block closure,
only invoking purego.NewCallback() when it encounters a new function type (rather than on for every block creation.)
This should mitigate block creations putting pressure on the callback limit.
*/
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
	// but altered to panic on error, and to only accep a block-type signature.
	if (typ == nil) || (typ.Kind() != reflect.Func) {
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

	if (typ.NumIn() == 0) || (typ.In(0) != reflect.TypeOf(Block(0))) {
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

// get layout retrieves a blockLayout VALUE constructed with the supplied function type
// It will panic if the type is not a valid block function.
func (b *blockCache) GetLayout(typ reflect.Type) blockLayout {
	b.Lock()
	defer b.Unlock()

	// return the cached layout, if it exists.
	if layout, ok := b.layouts[typ]; ok {
		return layout
	}

	// otherwise: create a layout, and populate it with the default templates
	layout := b.layoutTemplate
	layout.Descriptor = &blockDescriptor{}
	(*layout.Descriptor) = b.descriptorTemplate

	// getting the signature now will panic on invalid types before we invest in creating a callback.
	layout.Descriptor.Signature = b.encode(typ)

	// create a global callback.
	// this single callback can dispatch to any function with the same signature,
	// since the user-provided functions are associated with the actual block allocations.
	layout.Invoke = purego.NewCallback(
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

// newBlockCache initilizes a block cache.
// It should not be called until AFTER libobjc is fully initialized.
func newBlockCache() *blockCache {
	cache := &blockCache{
		descriptorTemplate: blockDescriptor{
			Size: unsafe.Sizeof(blockLayout{}),
		},
		layoutTemplate: blockLayout{
			Isa:   GetClass(blockBaseClass),
			Flags: blockFlags,
		},
		layouts:   map[reflect.Type]blockLayout{},
		Functions: newBlockFunctionCache(),
	}
	cache.descriptorTemplate.Dispose = purego.NewCallback(cache.Functions.Delete)
	return cache
}

// blocks is the global block cache
var blocks *blockCache

// Block is an opaque pointer to an Objective-C object containing a function with its associated closure.
type Block ID

// Copy creates a copy of a block on the Objective-C heap (or increments the reference count if already on the heap.)
// Use Block.Release() to free the copy when it is no longer in use.
func (b Block) Copy() Block {
	return _Block_copy(b)
}

// GetImplementation populates a function pointer with the implementation of a Block.
// Function will panic if the Block is not kept alive while it is in use
// (possibly by using Block.Copy()).
func (b Block) GetImplementation(fptr any) {
	// there is a runtime function imp_implementationWithBlock that could have been used instead,
	// but experimentation has shown the returned implementation doesn't actually work as expected.
	// also, it creates a new copy of the block which must be freed independently,
	// which would have made this implementation more complicated than necessary.
	// we know a block ID is actually a pointer to a blockLayout struct, so we'll take advantage of that.
	if b != 0 {
		if cfn := (*(**blockLayout)(unsafe.Pointer(&b))).Invoke; cfn != 0 {
			purego.RegisterFunc(fptr, cfn)
		}
	}
}

// Invoke is calls the implementation of a block.
func (b Block) Invoke(args ...any) {
	InvokeBlock[struct{}](b, args...)
}

// Release decrements the Block's reference count, and if it is the last reference, frees it.
func (b Block) Release() {
	_Block_release(b)
}

// NewBlock takes a Go function that takes a Block as its first argument.
// It returns an Block that can be called by Objective-C code.
// The function panics if an error occurs.
// Use Block.Release() to free this block when it is no longer in use.
func NewBlock(fn interface{}) Block {
	// get or create a block layout for the callback.
	layout := blocks.GetLayout(reflect.TypeOf(fn))
	// we created the layout in Go memory, so we'll copy it to a newly-created Objectve-C object.
	block := Block(unsafe.Pointer(&layout)).Copy()
	// associate the fn with the block we created before returning it.
	return blocks.Functions.Store(block, reflect.ValueOf(fn))
}

// InvokeBlock is a convenience method for calling the implementation of a block.
func InvokeBlock[T any](block Block, args ...any) T {
	block = block.Copy()
	defer block.Release()

	var invoke func(Block, ...any) T
	block.GetImplementation(&invoke)
	return invoke(block, args...)
}
