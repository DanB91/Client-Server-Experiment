package main

import (
	"encoding/binary"
	"net"
	"testing"
)

func init() {
	start_server()
}
func TestSendingData(t *testing.T) {
	conn, err := net.Dial("tcp", ADDRESS)
	if err != nil {
		t.Fatal(err)
	}
	data := []byte("Hello!")
	len_data := [8]byte{}
	binary.LittleEndian.PutUint64(len_data[:], uint64(len(data)))
	n, err := conn.Write(len_data[:])
	if err != nil {
		t.Fatal(err)
	}
	if n != len(len_data) {
		t.Fatal()
	}
	n, err = conn.Write(data)
	if err != nil {
		t.Fatal(err)
	}
	if n != len(data) {
		t.Fatal()
	}

	read_data := [256]byte{}
	n, err = conn.Read(read_data[:])
	if err != nil {
		t.Fatal(err)
	}
	if n != len("Hello!") {
		t.Fatal()
	}
	if string(read_data[:n]) != "Hello!" {
		t.Fatal()
	}

}
