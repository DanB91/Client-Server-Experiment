package main

import "fmt"

type Command uint64

const (
	PORT = 54320

	COMMAND_SIZE = 8

	CMD_GET_CHAT = iota
	CMD_POST_MESSAGE

	STRING_LEN_SIZE = 8
	STRING_LEN_MAX  = 512 * KiB
)

var (
	ADDRESS = fmt.Sprintf(":%v", PORT)
)
