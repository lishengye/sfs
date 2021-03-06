package client

import (
	"encoding/binary"
	"errors"
	"github.com/lishengye/sfs"
	"net"
	"path/filepath"
)

type Client struct {
	Connection *sfs.Connection
	token      string
	commandLine        CommandLine
}

func NewClient(commandLine CommandLine) Client {
	return Client{
		commandLine: commandLine,
	}
}

func (client *Client) Connect(ip, port string) error {
	conn, err := net.Dial("tcp", ip+":"+port)
	if err != nil {
		return err
	}

	client.Connection = sfs.NewConnection(conn)
	return nil
}

func (client *Client) Handshake(username, password string) error {

	user := []byte(username)
	pass := []byte(password)
	req := make([]byte, 8+4+len(user)+4+len(pass))

	copy(req[0:8], []byte(sfs.MethodConnect))

	copy(req[8:12], sfs.PutUint32(uint32(len(user))))

	copy(req[12:12+len(user)], []byte(user))

	copy(req[12+len(user):12+len(user)+4], sfs.PutUint32(uint32(len(pass))))

	copy(req[16+len(user):16+len(user)+len(pass)], []byte(pass))

	err := client.Connection.SendMsg(req)
	if err != nil {
		return err
	}

	res, err := client.Connection.ReceiveMsg()
	if err != nil {
		return err
	}

	if client.checkOk(res) != nil{
		return errors.New(string(res[1:]))
	}

	client.token = string(res[1:])
	return nil
}

func (client *Client) List() (string, error) {
	req := make([]byte, 8+8)
	copy(req[0:8], []byte(sfs.MethodList))

	copy(req[8:16], []byte(client.token))

	if err := client.Connection.SendMsg(req); err != nil {
		return "", err
	}

	res, err := client.Connection.ReceiveMsg()
	if err != nil {
		return "", err
	}

	if res[0] != 0 {
		return "", errors.New(string(res[1:]))
	}
	return string(res[1:]), nil
}

func (client *Client) Download(fileName string) error {

	req := make([]byte, 8+8+len(fileName))

	copy(req[0:8], []byte(sfs.MethodDownload))

	copy(req[8:16], []byte(client.token))

	copy(req[16:], []byte(fileName))

	if err := client.Connection.SendMsg(req); err != nil {
		return err
	}

	res, err := client.Connection.ReceiveMsg()
	if err != nil {
		return err
	}

	if err := client.checkOk(res); err != nil {
		if res[0] == 2 {
			client.commandLine.Error(err.Error())
			return nil
		}
		return err
	}

	fileSize := binary.BigEndian.Uint64(res[1:9])

	fileTransfer := FileTransfer{
		connection: client.Connection,
		CommandLine:		client.commandLine,
		FileSize:   fileSize,
		FileName:   fileName,
	}

	err = fileTransfer.ReceiveFile()

	return err
}

func (client *Client) Upload(filePath string, fileSize uint64) error {
	dir, fileName := filepath.Dir(filePath), filepath.Base(filePath)
	req := make([]byte, 8+8+8+len(fileName))

	copy(req[0:8], []byte(sfs.MethodUpload))

	copy(req[8:16], []byte(client.token))

	copy(req[8:16], []byte(client.token))

	copy(req[16:24], sfs.PutUint64(fileSize))

	copy(req[24:], []byte(fileName))

	if err := client.Connection.SendMsg(req); err != nil {
		return err
	}

	res, err := client.Connection.ReceiveMsg()
	if err != nil {
		return err
	}

	if err := client.checkOk(res); err != nil {
		if res[0] == 3 {
			client.commandLine.Warn(err.Error())
		} else {
			return err
		}
	}

	fileTransfer := FileTransfer{
		connection: client.Connection,
		CommandLine:  		client.commandLine,
		DownloadPath: dir,
		FileName:   fileName,
		FileSize:   fileSize,
	}

	err = fileTransfer.SendFile()

	return err
}

func (client *Client) Exit() error {
	req := make([]byte, 8)

	copy(req[0:8], []byte(sfs.MethodExit))

	if err := client.Connection.SendMsg(req); err != nil {
		return err
	}

	res, err := client.Connection.ReceiveMsg()
	if err != nil {
		return err
	}

	if err := client.checkOk(res); err != nil {
		return err
	}

	return nil
}

func (client *Client) checkOk(res []byte) error {
	if res[0] != 0 {
		return errors.New(string(res[1:]))
	}
	return nil
}
