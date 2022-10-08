//go:build !release

package main

import "log"

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
