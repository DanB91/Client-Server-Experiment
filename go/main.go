package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

type ServerState struct {
	chat_messages []ChatMessage
	lock          sync.RWMutex
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
	start_server()
}
func start_server() {
	ctx := context.Background()
	server_state = ServerState{
		chat_messages: make([]ChatMessage, 0, 256),
	}
	cancel_ctx, cancel := context.WithCancel(ctx)

	go start_service(cancel_ctx)

	_ = cancel

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
			println("Error accepting input connection: %v", err)
			continue
		}
		conn.SetDeadline(time.Now().Add(5 * time.Second))

		var command Command
		{
			command_data := [COMMAND_SIZE]byte{}
			has_data, _ := read(conn, command_data[:])
			if !has_data {
				conn.Close()
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
			write(conn, []byte(messages_to_string()))
			conn.Close()
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
				conn.Close()
				continue
			}
			message_string, has_data, delete_connection := read_string(conn)
			if !has_data || delete_connection {
				conn.Close()
				continue
			}
			conn.Close()

			message := ChatMessage{username: username, message: message_string}

			server_state.lock.Lock()
			server_state.chat_messages = append(server_state.chat_messages, message)
			server_state.lock.Unlock()
		case <-ctx.Done():
			return
		}
	}
}
func messages_to_string() string {
	server_state.lock.RLock()
	defer server_state.lock.RUnlock()

	builder := strings.Builder{}
	for _, message := range server_state.chat_messages {
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
	n, err := conn.Read(data)
	if err != nil {
		if err == os.ErrDeadlineExceeded {
			has_data = false
			delete_connection = false
			return
		}
		println("Error reading input connection %v", err)
		delete_connection = true
		has_data = false
		return
	}
	assert(n == len(data), "Wrong number of bytes read Expected: %v, was: %v", data, n)
	delete_connection = false
	has_data = true
	return
}
func write(conn net.Conn, data []byte) (delete_connection bool) {
	n, err := conn.Write(data)
	if err != nil {
		println("Error reading input connection %v", err)
		delete_connection = true
		return
	}
	assert(n == len(data), "Wrong number of bytes written Expected: %v, was: %v", data, n)
	delete_connection = false
	return
}
