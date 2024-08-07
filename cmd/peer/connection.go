package peer

import (
	"fmt"
	"net"

	"github.com/SotaUeda/gobgp/packets"
)

// 通信に関する処理を担当する構造体です。
// TcpConnectionを張ったり、
// Messageのデータを送受信したりします。
type Connection struct {
	conn *net.TCPConn
}

const BGP_PORT = 179 // BGPは179番ポートで固定

func NewConnection(c *Config) (*Connection, error) {
	var (
		conn = &net.TCPConn{}
		err  error
	)
	switch c.Mode {
	case Active:
		conn, err = connectRemoteAddress(c)
	case Passive:
		conn, err = waitRemoteAddress(c)
	default:
		err = fmt.Errorf("config mode is undefined")
	}
	if err != nil {
		return nil, err
	}
	err = conn.SetWriteBuffer(1500)
	if err != nil {
		return nil, err
	}
	return &Connection{conn}, nil
}

// Writer, Readerを実装した方がよりGoらしい？
func (c *Connection) Send(m packets.Message) error {
	buf, err := m.ToBytes()
	if err != nil {
		fmt.Printf("MessageのByte変換に失敗しました: %v\n", err)
		return err
	}
	_, err = c.conn.Write(buf)
	if err != nil {
		fmt.Printf("メッセージの送信に失敗しました: %v\n", err)
		return err
	}
	return nil
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
