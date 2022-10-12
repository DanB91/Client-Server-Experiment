//go:build !release

package main

import (
	"log"
)

const (
	Kibibyte = 1024
	KiB      = Kibibyte
	Mebibyte = Kibibyte * 1024
	MiB      = Mebibyte
	Gibibyte = Mebibyte * 1024
	GiB      = Gibibyte
	Tebibyte = Gibibyte * 1024
	TiB      = Tebibyte
	Pebibyte = Tebibyte * 1024
	PiB      = Pebibyte
	Exbibyte = Pebibyte * 1024
	EiB      = Exbibyte
)

type Integer interface {
	byte | uint16 | uint32 | uint64 | int16 | int32 | int64 | int | uint
}
type RingBuffer[T any] struct {
	store      []T
	head, tail int
}

func NewRingBuffer[T any](capacity int) RingBuffer[T] {
	capacity = NextPowerOf2(capacity)
	store := make([]T, capacity)
	return RingBuffer[T]{store: store}
}
func (self *RingBuffer[T]) Put(entry T) {
	self.store[self.tail] = entry
	self.tail++
	self.tail &= len(self.store) - 1
	if self.tail == self.head {
		self.head++
		self.head &= len(self.store) - 1
	}
}
func (self *RingBuffer[T]) Iterator() func() (*T, bool) {
	i := self.head
	return func() (*T, bool) {
		if i == self.tail {
			return nil, false
		}
		ret := &self.store[i]
		i++
		i &= len(self.store) - 1
		return ret, true
	}
}
func NextPowerOf2[T Integer](n T) T {
	if IsPowerOf2(n) {
		return n
	}
	count := 0
	val := n
	for val > 0 {
		val >>= 1
		count += 1
	}
	return 1 << count
}
func IsPowerOf2[T Integer](n T) bool {
	return (n & (n - 1)) == 0
}

func println(f string, args ...any) {
	log.Printf(f+"\n", args...)
}

func fatal(f string, args ...any) {
	log.Fatalf(f+"\n", args...)
}

func assert(cond bool, msg string, args ...any) {
	if !cond {
		log.Panicf(msg, args...)
	}
}
