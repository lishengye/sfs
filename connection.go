package sfs

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

type Connection struct {
	Conn net.Conn
}

func NewConnection(conn net.Conn) *Connection {
	return &Connection{
		Conn: conn,
	}
}

// write data in one msg
func (connection *Connection) SendMsg(data []byte) error {
	sent := 0
	total := len(data) + 4
	head := make([]byte, 4)
	binary.BigEndian.PutUint32(head, uint32(total))
	data = append(head, data...)
	for sent < total {
		size := int(Min(total-sent, 8096))
		n, err := connection.Conn.Write(data[sent : sent+size])
		if err != nil {
			fmt.Println(err)
			return err
		}
		sent += n
	}
	return nil
}

// read one msg
func (connection *Connection) ReceiveMsg(data []byte) error {
	head := make([]byte, 4)
	connection.Conn.Read(head[0:4])
	total := binary.BigEndian.Uint32(head)
	readed := 0
	buf := make([]byte, BUF_SIZE)
	for readed < int(total) {
		n, err := connection.Conn.Read(buf)
		// 先判断n是否为0,再考虑error。参考io.Reader接口
		if n > 0 {
			// append内存不够会自动扩充
			data = append(data, buf[:n]...)
			readed += n
		}
		if err != nil && err != io.EOF {
			fmt.Println(err)
			return err
		}
	}
	return nil
}
