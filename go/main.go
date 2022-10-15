package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type ServerState struct {
	chat_messages        RingBuffer[ChatMessage]
	lock                 sync.RWMutex
	num_open_connections atomic.Int64
}
type ChatMessage struct {
	username string
	message  string
}

const (
	SERVER_STATE_KEY_VALUE = iota
)

var (
	server_state ServerState
)

func main() {
	nclients := 50
	requests_per_second := 60

	ctx := context.Background()
	cancel_server_context, cancel_server := context.WithCancel(ctx)
	cancel_client_context, cancel_client := context.WithCancel(ctx)
	end_stats_chan := make(chan Stats, nclients)

	start_server(cancel_server_context)
	run_clients(nclients, requests_per_second, cancel_client_context, end_stats_chan)

	<-time.After(30 * time.Second)
	cancel_client()

	total_iters := 0
	for i := 0; i < nclients; i++ {
		stats := <-end_stats_chan
		printfln("Iters for client %v: %v", i, stats.iters)
		total_iters += stats.iters
	}
	printfln("Total iters: %v", total_iters)
	cancel_server()
}
func start_server(ctx context.Context) {
	server_state = ServerState{
		chat_messages: NewRingBuffer[ChatMessage](256),
	}

	go start_service(ctx)

}
func start_service(ctx context.Context) {
	listener, err := net.Listen("tcp", ADDRESS)
	if err != nil {
		fatal("Error starting up input listener %v", err)
	}
	get_chat_command_chan := make(chan net.Conn, 128)
	defer close(get_chat_command_chan)
	post_message_command_chan := make(chan net.Conn, 128)
	defer close(post_message_command_chan)

	go get_chat_command_handler(ctx, get_chat_command_chan)
	go post_message_command_handler(ctx, post_message_command_chan)

	defer listener.Close()
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		conn, err := listener.Accept()

		if err != nil {
			printfln("Error accepting input connection: %v", err)
			continue
		}
		conn.SetDeadline(time.Now().Add(5 * time.Minute))

		server_state.num_open_connections.Add(1)
		printfln("Opening connection from server... open conns: %v", server_state.num_open_connections.Load())
		var command Command
		{
			command_data := [COMMAND_SIZE]byte{}
			has_data, _ := read(conn, command_data[:])
			if !has_data {
				close_conn(conn)
				continue
			}
			command = Command(binary.LittleEndian.Uint64(command_data[:]))
		}
		switch command {
		case CMD_GET_CHAT:
			get_chat_command_chan <- conn
		case CMD_POST_MESSAGE:
			post_message_command_chan <- conn
		}

	}
}
func get_chat_command_handler(ctx context.Context, connection_chan <-chan net.Conn) {
	for {
		select {
		case conn, ok := <-connection_chan:
			if !ok {
				return
			}
			write_string(conn, messages_to_string())
			close_conn(conn)

		case <-ctx.Done():
			return
		}
	}
}
func post_message_command_handler(ctx context.Context, connection_chan <-chan net.Conn) {
	for {
		select {
		case conn, ok := <-connection_chan:
			if !ok {
				return
			}
			username, has_data, delete_connection := read_string(conn)
			if !has_data || delete_connection {
				close_conn(conn)
				continue
			}
			message_string, has_data, delete_connection := read_string(conn)
			if !has_data || delete_connection {
				close_conn(conn)
				continue
			}
			close_conn(conn)

			message := ChatMessage{username: username, message: message_string}

			server_state.lock.Lock()
			server_state.chat_messages.Put(message)
			server_state.lock.Unlock()
		case <-ctx.Done():
			return
		}
	}
}
func close_conn(conn net.Conn) {
	err := conn.Close()
	if err == nil {
		server_state.num_open_connections.Add(-1)
		printfln("closing connection from server... open conns %v", server_state.num_open_connections.Load())
	} else {
		printfln("Error closing connection: %v", err)
	}
}
func messages_to_string() string {
	server_state.lock.RLock()
	defer server_state.lock.RUnlock()

	builder := strings.Builder{}
	it := server_state.chat_messages.Iterator()
	for {
		message, has_value := it()
		if !has_value {
			break
		}
		builder.WriteString(fmt.Sprintf("%v: %v\n", message.username, message.message))
	}
	return builder.String()
}
func read_string(conn net.Conn) (value string, has_data bool, delete_connection bool) {
	length_bytes := [STRING_LEN_SIZE]byte{}
	has_data, delete_connection = read(conn, length_bytes[:])
	if !has_data || delete_connection {
		return
	}
	incoming_string_size := binary.LittleEndian.Uint64(length_bytes[:])
	if incoming_string_size >= STRING_LEN_MAX {
		delete_connection = true
		return
	}
	string_bytes := make([]byte, incoming_string_size)
	has_data, delete_connection = read(conn, string_bytes[:])
	if !has_data || delete_connection {
		return
	}
	value = string(string_bytes)
	has_data = true
	delete_connection = false
	return
}
func read(conn net.Conn, data []byte) (has_data bool, delete_connection bool) {
	n := 0
	for n < len(data) {
		_n, err := conn.Read(data)
		if err != nil {
			if err == os.ErrDeadlineExceeded {
				has_data = false
				delete_connection = false
				return
			}
			printfln("Error reading from connection %v", err)
			delete_connection = true
			has_data = false
			return
		}
		n += _n
	}
	assert(n == len(data), "Wrong number of bytes read Expected: %v, was: %v", len(data), n)
	delete_connection = false
	has_data = true
	return
}
func write(conn net.Conn, data []byte) (delete_connection bool) {
	n := 0
	for n < len(data) {
		_n, err := conn.Write(data)
		if err != nil {
			printfln("Error writing to connection %v", err)
			delete_connection = true
			return
		}
		n += _n
	}
	assert(n == len(data), "Wrong number of bytes written Expected: %v, was: %v", len(data), n)
	delete_connection = false
	return
}
func write_string(conn net.Conn, value string) (delete_connection bool) {
	string_len_bytes := [STRING_LEN_SIZE]byte{}
	binary.LittleEndian.PutUint64(string_len_bytes[:], uint64(len(value)))
	if delete_connection = write(conn, string_len_bytes[:]); delete_connection {
		return
	}
	if delete_connection = write(conn, []byte(value)); delete_connection {
		return
	}

	delete_connection = false
	return
}
