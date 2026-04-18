package arena

import (
	"errors"
	"unsafe"
)

type Arena struct {
	MemoryStart unsafe.Pointer
	Size uint64
	ByteSlice []byte
	AllocateStart uint64
}

func (arena *Arena)Init(size uint64) {
	arena.ByteSlice = make([]byte, size)
	arena.Size = size
	arena.MemoryStart = unsafe.Pointer(unsafe.SliceData(arena.ByteSlice))
	arena.AllocateStart = 0
}

func (arena *Arena)Reset() {
	arena.AllocateStart = 0
}

func (arena *Arena)Alloc(size uint64) (unsafe.Pointer, error) {
	if arena.AllocateStart + size > arena.Size {
		return nil, errors.New("Arena ran out of space")
	}

	old := arena.AllocateStart

	arena.AllocateStart += size

	return unsafe.Pointer(uintptr(arena.MemoryStart) + uintptr(old)), nil
}