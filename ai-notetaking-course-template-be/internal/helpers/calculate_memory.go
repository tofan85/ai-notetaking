package helpers

import (
	"runtime"
)

type MemUsage struct {
	AllocBefore uint64
	AllocAfter  uint64
	Used        uint64
}

func TrackMemory() (before uint64) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.Alloc
}

func CalculateMemory(before uint64) MemUsage {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return MemUsage{
		AllocBefore: before,
		AllocAfter:  m.Alloc,
		Used:        m.Alloc - before,
	}
}
