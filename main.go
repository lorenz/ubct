package ubct

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
)

var eofMarker = []byte{0x04, 0x0a}

// Client represents a client to an Unbound server
type Client struct {
	Version   int32
	TLSConfig *tls.Config
	Address   string
}

func (c *Client) dial() (*tls.Conn, error) {
	return tls.Dial("tcp", c.Address, c.TLSConfig)
}

func readResponse(conn *tls.Conn) (string, error) {
	rawRes, err := ioutil.ReadAll(conn)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}
	res := string(rawRes)
	if strings.HasPrefix(res, "error") {
		return "", errors.New(res)
	}
	return res, nil
}

// RunCommand runs a command on the Unbound server and gives back the response
func (c *Client) RunCommand(cmd string, args ...string) (string, error) {
	conn, err := c.dial()
	if err != nil {
		return "", fmt.Errorf("failed to connect: %v", err)
	}
	_, err = conn.Write([]byte(fmt.Sprintf("UBCT%v %v %v\n", c.Version, cmd, strings.Join(args, " "))))
	if err != nil {
		return "", fmt.Errorf("failed to send command: %v", err)
	}
	return readResponse(conn)
}

// RunFileCommand runs a command with a data file on the Unbound server and gives back the response
func (c *Client) RunFileCommand(cmd string, file io.Reader, args ...string) (string, error) {
	conn, err := c.dial()
	if err != nil {
		return "", fmt.Errorf("failed to connect: %v", err)
	}
	_, err = conn.Write([]byte(fmt.Sprintf("UBCT%v %v %v\n", c.Version, cmd, strings.Join(args, " "))))
	if err != nil {
		return "", fmt.Errorf("failed to send command: %v", err)
	}
	_, err = io.Copy(conn, file)
	if err != nil {
		return "", fmt.Errorf("failed to send data file: %v", err)
	}
	_, err = conn.Write(eofMarker)
	if err != nil {
		return "", fmt.Errorf("failed to send EOF: %v", err)
	}
	return readResponse(conn)
}
