package main

import (
	"context"
	"encoding/binary"
	"log"
	"math/rand"
	"net"
	"sync/atomic"
	"time"
)

type Stats struct {
	iters int
}

var connections_open atomic.Int64

func run_clients(nclients, requests_per_second int, ctx context.Context, end_stats_chan chan<- Stats) {
	if requests_per_second > 0 {
		printfln("Running %v clients at %v requests a second each...", nclients, requests_per_second)
	} else {
		printfln("Running %v clients as fast as possible", nclients)

	}
	for i := 0; i < nclients; i++ {
		go run_client(requests_per_second, 100, 16, end_stats_chan, ctx)
	}
}
func run_client(requests_per_second, message_length, username_length int,
	end_stats_chan chan<- Stats, ctx context.Context) {
	username := rand_string_bytes(username_length)
	message := rand_string_bytes(message_length)
	iters := 0
	defer func() {
		iters_ptr := &iters
		end_stats_chan <- Stats{iters: *iters_ptr}
	}()
	if requests_per_second <= 0 {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			post_message(username, message)
			chat := get_chat()
			iters++
			_ = chat
		}

	} else {
		ticker := time.NewTicker(time.Duration(1_000_000_000/requests_per_second) * time.Nanosecond)
		for {
			select {
			case <-ticker.C:
			case <-ctx.Done():
				return
			}
			post_message(username, message)
			<-ticker.C
			chat := get_chat()
			iters++
			_ = chat

		}

	}

}

func rand_string_bytes(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}

func post_message(username, message string) {
	conn := connect_to_server()
	defer close_server_connection(conn)

	command_bytes := [COMMAND_SIZE]byte{}
	binary.LittleEndian.PutUint64(command_bytes[:], CMD_POST_MESSAGE)

	if delete_connection := write(conn, command_bytes[:]); delete_connection {
		log.Panicf("Failed to write command to server")
	}

	if delete_connection := write_string(conn, username); delete_connection {
		log.Panicf("Failed to write username to server")
	}
	if delete_connection := write_string(conn, message); delete_connection {
		log.Panicf("Failed to write message to server")
	}

}
func get_chat() string {
	conn := connect_to_server()
	defer close_server_connection(conn)
	command_bytes := [COMMAND_SIZE]byte{}
	binary.LittleEndian.PutUint64(command_bytes[:], CMD_GET_CHAT)

	if delete_connection := write(conn, command_bytes[:]); delete_connection {
		log.Panicf("Failed to write command to server")
	}
	message, has_data, delete_connection := read_string(conn)
	if !has_data || delete_connection {
		log.Panicf("Failed to read chat from server")
	}

	return message
}

func connect_to_server() net.Conn {
	conn, err := net.Dial("tcp", ADDRESS)
	if err != nil {
		log.Panicf("Failed to connect to server: %v", err)
	}
	connections_open.Add(1)
	printfln("Opening connection from client: %v", connections_open.Load())
	return conn
}

func close_server_connection(conn net.Conn) {
	conn.Close()
	connections_open.Add(-1)
	printfln("Closing connection from client: %v", connections_open.Load())
}
