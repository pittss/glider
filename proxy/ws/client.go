package ws

import (
	"bufio"
	"errors"
	"io"
	"net"
	"net/textproto"
	"strings"

	"github.com/nadoo/glider/common/log"
)

// Client ws client
type Client struct {
	path string
}

// Conn is a connection to ws server
type Conn struct {
	net.Conn
	reader io.Reader
	writer io.Writer
}

// NewClient .
func NewClient() (*Client, error) {
	c := &Client{}
	return c, nil
}

// NewConn .
func (c *Client) NewConn(rc net.Conn, target string) (*Conn, error) {
	conn := &Conn{Conn: rc}
	conn.Handshake()
	return conn, nil
}

// Handshake handshakes with the server using HTTP to request a protocol upgrade
//
// GET /chat HTTP/1.1
// Host: server.example.com
// Upgrade: websocket
// Connection: Upgrade
// Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==
// Origin: http://example.com
// Sec-WebSocket-Protocol: chat, superchat
// Sec-WebSocket-Version: 13
//
// HTTP/1.1 101 Switching Protocols
// Upgrade: websocket
// Connection: Upgrade
// Sec-WebSocket-Accept: s3pPLMBiTxaQ9kYGzzhZRbK+xOo=
// Sec-WebSocket-Protocol: chat
func (c *Conn) Handshake() error {
	c.Conn.Write([]byte("GET / HTTP/1.1\r\n"))
	c.Conn.Write([]byte("Host: echo.websocket.org\r\n"))
	c.Conn.Write([]byte("Upgrade: websocket\r\n"))
	c.Conn.Write([]byte("Connection: Upgrade\r\n"))
	c.Conn.Write([]byte("Origin: http://127.0.0.1\r\n"))
	c.Conn.Write([]byte("Sec-WebSocket-Key: dGhlIHNhbXBsZSBub25jZQ==\r\n"))
	c.Conn.Write([]byte("Sec-WebSocket-Protocol: binary\r\n"))
	c.Conn.Write([]byte("Sec-WebSocket-Version: 13\r\n"))
	c.Conn.Write([]byte("\r\n"))

	tpr := textproto.NewReader(bufio.NewReader(c.Conn))
	_, code, _, ok := parseFirstLine(tpr)
	if !ok || code != "101" {
		return errors.New("error in ws handshake")
	}

	respHeader, err := tpr.ReadMIMEHeader()
	if err != nil {
		return err
	}

	// TODO: verify this
	respHeader.Get("Sec-WebSocket-Accept")
	// fmt.Printf("respHeader: %+v\n", respHeader)

	return nil
}
func (c *Conn) Write(b []byte) (n int, err error) {
	if c.writer == nil {
		c.writer = FrameWriter(c.Conn)
	}

	return c.writer.Write(b)
}
func (c *Conn) Read(b []byte) (n int, err error) {
	if c.reader == nil {
		c.reader = FrameReader(c.Conn)
	}

	return c.reader.Read(b)
}

// parseFirstLine parses "GET /foo HTTP/1.1" OR "HTTP/1.1 200 OK" into its three parts.
// TODO: move to seperate http lib package for reuse(also for http proxy module)
func parseFirstLine(tp *textproto.Reader) (r1, r2, r3 string, ok bool) {
	line, err := tp.ReadLine()
	// log.F("first line: %s", line)
	if err != nil {
		log.F("[http] read first line error:%s", err)
		return
	}

	s1 := strings.Index(line, " ")
	s2 := strings.Index(line[s1+1:], " ")
	if s1 < 0 || s2 < 0 {
		return
	}
	s2 += s1 + 1
	return line[:s1], line[s1+1 : s2], line[s2+1:], true
}
