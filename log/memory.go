package log

import "runtime"

func Memory(namespace string) {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	Info(namespace,
		"mem.used: %.2fMb, mem.released: %.2fMb, mem.sys: %.2fMb, objects: %.2fk, gc.num: %d",
		mb(mem.Alloc), mb(mem.HeapReleased), mb(mem.HeapSys), float64(mem.HeapObjects)/1000, mem.NumGC)
}

func mb(b uint64) float64 {
	return float64(b) / 1024 / 1024
}
