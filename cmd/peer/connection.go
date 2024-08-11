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
	buf  []byte // 受信用バッファ
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
	return &Connection{conn, nil}, nil
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

// Writer, Readerを実装した方がよりGoらしい？
func (c *Connection) Send(m packets.Message) error {
	b, err := m.ToBytes()
	if err != nil {
		fmt.Printf("MessageのByte変換に失敗しました: %v\n", err)
		return err
	}
	_, err = c.conn.Write(b)
	if err != nil {
		fmt.Printf("メッセージの送信に失敗しました: %v\n", err)
		return err
	}
	return nil
}

// bgp messageを1つ以上受信していれば
// 最古に受信したMessageを返す。
// bgp messageのデータの受信中（半端に受信している）、
// ないしは何も受信していない場合はnilを返す。
// この関数は非同期で呼び出されることを想定している。
func (c *Connection) Recv() (packets.Message, error) {
	if c.buf == nil {
		c.buf = make([]byte, 4096)
	}
	for {
		n, err := c.conn.Read(c.buf)
		if err != nil {
			fmt.Printf("メッセージの受信に失敗しました: %v\n", err)
			return nil, err
		}
		if n == 0 {
			continue
		}
		b, err := c.splitMsgSep()
		if err != nil {
			fmt.Printf("MessageのByte切り出しに失敗しました: %v\n", err)
			return nil, err
		}
		if b == nil {
			continue
		}
		m, err := packets.BytesToMessage(b)
		if err != nil {
			fmt.Printf("MessageのByte変換に失敗しました: %v\n", err)
			return nil, err
		}
		return m, nil
	}
}

// *Connection.bufから1つのbgp messageを切り出す
func (c *Connection) splitMsgSep() ([]byte, error) {
	idx, err := c.getIdxMsgSep()
	if err != nil {
		return nil, err
	}
	if len(c.buf) < idx {
		return nil, nil // まだMessageのSeparateorを表すデータがbufferに入っていない
	}
	c.buf = c.buf[idx:]
	return c.buf[:idx], nil
}

// *Connection.bufのうちどこまでが1つのbgp messageを表すbyteであるかを返す
// BGPヘッダーのLengthフィールドの値を返す
func (c *Connection) getIdxMsgSep() (int, error) {
	minMsgLen := 19 // BGP messageの最小長
	if len(c.buf) < minMsgLen {
		return 0, fmt.Errorf(
			"MessageのSeparateorを表すデータまでbufferに入っていません。"+
				"データの受信が半端であることが想定されます。 buffer: %v", len(c.buf))
	}
	return int(c.buf[16])<<8 + int(c.buf[17]), nil
}
