package main

import (
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

}
