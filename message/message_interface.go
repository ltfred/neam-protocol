package message

import "net"

type Message interface {
}

func NewMessage(conn net.Conn) Message {

	return MonitorMessage{}
}
