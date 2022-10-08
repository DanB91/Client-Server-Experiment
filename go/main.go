package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"time"
)

const (
	PORT = 54321
)

var (
	ADDRESS = fmt.Sprintf(":%v", PORT)
)

func main() {
	start_server()
}
func start_server() {
	ctx := context.Background()
	cancel_ctx, cancel := context.WithCancel(ctx)

	go start_service(cancel_ctx)

	_ = cancel

}
func start_service(ctx context.Context) {
	listener, err := net.Listen("tcp", ADDRESS)
	if err != nil {
		fatal("Error starting up input listener %v", err)
	}
	connection_chan := make(chan net.Conn, 128)
	defer close(connection_chan)
	go input_request_handler(ctx, connection_chan)

	defer listener.Close()
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		conn, err := listener.Accept()
		if err != nil {
			println("Error accepting input connection: %v", err)
			continue
		}
		conn.SetDeadline(time.Now().Add(5 * time.Second))
		connection_chan <- conn
	}
}
func input_request_handler(ctx context.Context, connection_chan <-chan net.Conn) {
	connections := map[net.Conn]bool{}
	read := func(conn net.Conn, data []byte) (has_data bool) {
		n, err := conn.Read(data)
		if err != nil {
			if err == os.ErrDeadlineExceeded {
				has_data = false
				return
			}
			println("Error reading input connection %v", err)
			delete(connections, conn)
			has_data = false
			return
		}
		assert(n == len(data), "Wrong number of bytes read Expected: %v, was: %v", data, n)
		has_data = true
		return
	}
	write := func(conn net.Conn, data []byte) {
		n, err := conn.Write(data)
		if err != nil {
			println("Error reading input connection %v", err)
			delete(connections, conn)
			return
		}
		assert(n == len(data), "Wrong number of bytes written Expected: %v, was: %v", data, n)
	}
	for {
		select {
		case new_conn := <-connection_chan:
			connections[new_conn] = true
		case <-ctx.Done():
			return
		default:
		}
	connection_loop:
		for conn := range connections {
			var num_bytes_to_read uint64
			{
				num_bytes_data := [8]byte{}
				if has_data := read(conn, num_bytes_data[:]); !has_data {
					continue connection_loop
				}
				num_bytes_to_read = binary.LittleEndian.Uint64(num_bytes_data[:])
				if num_bytes_to_read >= 512*KiB {
					continue connection_loop
				}

			}
			{
				data := make([]byte, num_bytes_to_read)
				if has_data := read(conn, data); !has_data {
					continue connection_loop
				}

				//TODO: remove echoing
				write(conn, data)

			}
		}

	}

}
