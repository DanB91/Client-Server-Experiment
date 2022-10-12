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
	{
		conn, err := net.Dial("tcp", ADDRESS)
		if err != nil {
			t.Fatal(err)
		}

		command_bytes := [COMMAND_SIZE]byte{}
		binary.LittleEndian.PutUint64(command_bytes[:], CMD_POST_MESSAGE)

		if delete_connection := write(conn, command_bytes[:]); delete_connection {
			t.FailNow()
		}

		if delete_connection := write_string(conn, "user"); delete_connection {
			t.FailNow()
		}
		if delete_connection := write_string(conn, "Hello!!"); delete_connection {
			t.FailNow()
		}
		conn.Close()

	}
	{
		conn, err := net.Dial("tcp", ADDRESS)
		if err != nil {
			t.Fatal(err)
		}
		command_bytes := [COMMAND_SIZE]byte{}
		binary.LittleEndian.PutUint64(command_bytes[:], CMD_GET_CHAT)

		if delete_connection := write(conn, command_bytes[:]); delete_connection {
			t.FailNow()
		}
		message, has_data, delete_connection := read_string(conn)
		if !has_data || delete_connection {
			t.FailNow()
		}

		if message != "user: Hello!!\n" {
			t.FailNow()
		}

	}
	{
		conn, err := net.Dial("tcp", ADDRESS)
		if err != nil {
			t.Fatal(err)
		}

		command_bytes := [COMMAND_SIZE]byte{}
		binary.LittleEndian.PutUint64(command_bytes[:], CMD_POST_MESSAGE)

		if delete_connection := write(conn, command_bytes[:]); delete_connection {
			t.FailNow()
		}

		if delete_connection := write_string(conn, "user"); delete_connection {
			t.FailNow()
		}
		if delete_connection := write_string(conn, "Good Bye!"); delete_connection {
			t.FailNow()
		}
		conn.Close()

	}
	{
		conn, err := net.Dial("tcp", ADDRESS)
		if err != nil {
			t.Fatal(err)
		}
		command_bytes := [COMMAND_SIZE]byte{}
		binary.LittleEndian.PutUint64(command_bytes[:], CMD_GET_CHAT)

		if delete_connection := write(conn, command_bytes[:]); delete_connection {
			t.FailNow()
		}
		message, has_data, delete_connection := read_string(conn)
		if !has_data || delete_connection {
			t.FailNow()
		}

		if message != "user: Hello!!\nuser: Good Bye!\n" {
			t.FailNow()
		}

	}
}
