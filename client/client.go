package client

import (
	"github.com/lishengye/sfs"
)

type Client struct {
	Connection sfs.Connection
	token  	   string
}


func NewClient() *Client {
	return &Client{

	}
}

func (client *Client) Handshake() error {
	return nil
}

func (client *Client) Download() error {
	return nil
}

func (client *Client) Upload() error {

}

func (client *Client) Exit() error {

}