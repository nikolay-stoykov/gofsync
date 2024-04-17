package transport

import (
	"context"
	"fmt"
	"net"
)

type Client struct {
	conn net.Conn
}

func (c *Client) Connect(ctx context.Context, address string) error {
	dialer := net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return fmt.Errorf("could not dial server: %w", err)
	}

	c.conn = conn
	return nil
}

func (c *Client) Send(ctx context.Context) error {
	_, err := c.conn.Write([]byte("hello"))
	if err != nil {
		return err
	}

	return nil
}
