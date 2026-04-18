package arena

import "unsafe"

type ArenaGroup struct {
	Arenas []*Arena
	CurrentArenaIndex int
}

func NewArenaGroup(baseSize uint64) *ArenaGroup {
	var arena Arena
	arena.Init(baseSize)
	return &ArenaGroup{
		Arenas: []*Arena{&arena},
		CurrentArenaIndex: 0,
	}
}

// Returns []T of the type T you pass it
func AllocSlice[T any](arena_group *ArenaGroup, size uint64) []T {
	total_size := uint64(unsafe.Sizeof(*new(T))) * size

	data_ptr, err := arena_group.Arenas[arena_group.CurrentArenaIndex].Alloc(total_size)

	if err != nil {
		var newArena Arena
		newArena.Init(total_size)
		arena_group.Arenas = append(arena_group.Arenas, &newArena)
		arena_group.CurrentArenaIndex += 1
		data_ptr, _ = newArena.Alloc(total_size)
	}

	return unsafe.Slice((*T)(data_ptr), size)
}

// Reuse clears all inner arenas except the first one.
// Reuse also Resets the first arena.
func (arena_group *ArenaGroup)Reuse() {
	for i := 1; i < len(arena_group.Arenas); i++ {
        arena_group.Arenas[i] = nil // must nil out all inner arenas from the backing slice to make sure the GC collects
    }
	arena_group.Arenas = arena_group.Arenas[:1] // slicing like this still causes the backing slice allocation to point to other elements
	arena_group.CurrentArenaIndex = 0
	arena_group.Arenas[0].Reset()
}

func (arena_group *ArenaGroup)Usage() uint64 {
	count := uint64(0)

	for _, arena := range arena_group.Arenas {
		count += arena.AllocateStart
	}

	return count
}