package peer

import (
	"fmt"
	"net"
)

type Connection struct {
	conn *net.TCPConn
}

const BGP_PORT = 179

func NewConnection(c *Config) (*Connection, error) {
	var (
		conn = &Connection{}
		err  error
	)
	switch c.Mode {
	case Active:
		conn.conn, err = connectRemoteAddress(c)
		return conn, err
	case Passive:
		conn.conn, err = waitRemoteAddress(c)
		return conn, err
	default:
		err = fmt.Errorf("config mode is undefined")
		return conn, err
	}
}

func connectRemoteAddress(c *Config) (*net.TCPConn, error) {
	ladd := &net.TCPAddr{
		IP:   c.LocalIP,
		Port: BGP_PORT,
	}
	radd := &net.TCPAddr{
		IP:   c.RemoteIP,
		Port: BGP_PORT,
	}
	conn, err := net.DialTCP("tcp", ladd, radd)
	if err != nil {
		fmt.Printf("failed to connect on port %d: %v\n", BGP_PORT, err)
		return conn, err
	}
	// TODO: タイムアウト実装
	fmt.Print("connected\n")
	return conn, nil
}

func waitRemoteAddress(c *Config) (*net.TCPConn, error) {
	ladd := &net.TCPAddr{
		IP:   c.LocalIP,
		Port: BGP_PORT,
	}
	listener, err := net.ListenTCP("tcp", ladd)
	if err != nil {
		fmt.Printf("failed to listen on port %d: %v\n", BGP_PORT, err)
		return nil, err
	}
	conn, err := listener.AcceptTCP()
	if err != nil {
		fmt.Printf("failed to accept: %v\n", err)
		return conn, err
	}
	fmt.Print("accepted\n")
	return conn, nil
}
