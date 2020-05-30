package server

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/lishengye/sfs"
	"github.com/lishengye/sfs/log"
	"io/ioutil"
	"os"
	"path/filepath"
)

type ClientHandler struct {
	connection *sfs.Connection
	Config     Config
	token      string
}

func (clientHandler *ClientHandler) Handle() {
	defer clientHandler.connection.Conn.Close()

	if err := clientHandler.HandShake(); err != nil {
		log.Error("Handshake error: %s", err.Error())
		return
	}
	log.Info("Handshake successfully")

	for {
		data, err := clientHandler.connection.ReceiveMsg();
		if  err != nil {
			log.Error("Receive command error")
			return
		}

		switch string(data[:8]) {
		case sfs.MethodList:
			err = clientHandler.List(data)
		case sfs.MethodDownload:
			err = clientHandler.Download(data)
		case sfs.MethodUpload:
			err = clientHandler.Upload(data)
		case sfs.MethodExit:
			err = clientHandler.Exit()
			return
		}
		if err != nil {
			log.Error(err.Error())
			return
		}
	}
}

/*
req:
	| 8 	 | 		4 	 | xxxx 	| 4  	   | xxxx |
	|CONNECT | user_len  | username | pass_len | pass |
res:
	| 1   | 	 xxx 		|
	|0/1 |  "token"/err    |

*/
func (clientHandler *ClientHandler) HandShake() error {
	req, err := clientHandler.connection.ReceiveMsg();
	if err != nil {
		log.Error("Handshake ReceiveMsg error: %s", err.Error())
		return errors.New("ReceiveMsg error")
	}

	method := string(req[:8])
	if method != sfs.MethodConnect {
		log.Error("Handshake not connect command: %s, %s", method, sfs.MethodConnect)
		return errors.New("invalid command")
	}

	userLen := binary.BigEndian.Uint32(req[8:12])
	if 8+4+userLen > uint32(len(req)) {
		log.Error("Handshake len error: %v", req)
		return errors.New("invalid data protocal")
	}

	userName := string(req[12 : 12+userLen])

	passLen := binary.BigEndian.Uint32(req[8+4+userLen : 8+4+userLen+4])
	// 校验消息长度
	if int(8+4+userLen+4+passLen) != len(req) {
		log.Error("Handshake len error: %v", req)
		return errors.New("invalid data protocal")
	}
	password := string(req[16+userLen : 16+userLen+passLen])

	if sfs.CheckUser(userName, password) {
		errMsg := "Uncorrect User/Pass"
		res := append([]byte{1}, []byte(errMsg)...)
		if err := clientHandler.connection.SendMsg(res); err != nil {
			log.Error("Handshake SendMsg to client error")
			return errors.New("SendMsg error")
		}
	}
	log.Info("Handshake User/Pass ok")

	clientHandler.token = sfs.GenToken()
	res := append([]byte{0}, []byte(clientHandler.token)...)
	if err := clientHandler.connection.SendMsg(res); err != nil {
		log.Error("Handshake SendMsg to client error")
		return errors.New("SendMsg error")
	}
	return nil
}

/*
8		8
list 	token

1		xxx
0/1	res/errMsg

*/
func (clientHandler *ClientHandler) List(req []byte) error {
	// check req len
	if len(req) != 16 {
		log.Error("List data lenth not 16, len: %v", len(req))
		return errors.New("List data lenth error")
	}

	// check token
	if token := string(req[8:]); token != clientHandler.token {
		log.Error("List invalid token: %s, %s", clientHandler.token, token)
		return errors.New("List invalid token")
	}

	// list dir
	files, _ := ioutil.ReadDir(clientHandler.Config.Directory)
	result := ""
	for _, v := range files {
		if v.IsDir() {
			continue
		}
		result += fmt.Sprintf("%s %s", string(v.Size())+"B", v.Name())
	}

	// response
	res := []byte{0}
	res = append(res, []byte(result)...)
	if err := clientHandler.connection.SendMsg(res); err != nil {
		log.Error("List SendMsg error: %s", err.Error())
		return errors.New("List SendMsg error")
	}
	return nil
}

/*
1:
	|	8	|	8		|   xxxx  |		xxxx    |
	|	download	|  token 	|	file_name	|

	|	1		|	xxx		|
	|	1		| 	errMsg 	|
	|	0		|   fileSize|

2:
	tranfor file

*/
func (clientHandler *ClientHandler) Download(req []byte) error {

	if token := string(req[8:]); token != clientHandler.token {
		log.Error("Download invalid token: %s, %s", clientHandler.token, token)
		return errors.New("download invalid token")
	}

	fileName := string(req[16:])
	v, err := os.Stat(fileName)
	if err != nil || v.IsDir() {
		errMsg := "File not found or download directory"
		res := append([]byte{0}, []byte(errMsg)...)
		if err := clientHandler.connection.SendMsg(res); err != nil {
			log.Error("Download SendMsg error: %s", err.Error())
			return errors.New("Download SendMsg error")
		}
	}

	temp := make([]byte, 8)
	binary.BigEndian.PutUint64(temp, uint64(v.Size()))
	res := append([]byte{0}, temp...)
	if err := clientHandler.connection.SendMsg(res); err != nil {
		log.Error("Download SendMsg error: %s", err.Error())
		return errors.New("Download SendMsg error")
	}

	fileTransfer := FileTransfer{
		connection:   clientHandler.connection,
		FileSize:     uint64(v.Size()),
		FileName:     fileName,
		DownloadPath: clientHandler.Config.Directory,
	}

	err = fileTransfer.SendFile()
	if err != nil {
		log.Error("Download send file error: %s", err.Error())
		return errors.New("Download send file error")
	}
	return nil
}

/*
1:
	|	8	|	8		|   8    |  xxxx  |
	|	upload	|  token 	|	fileSize 	| file_name	|

	|	1		|	xxx		|
	|	0/1	| 	"ok"/msg 	|

2:
	tranfor file

*/
func (clientHandler *ClientHandler) Upload(req []byte) error {
	if string(req[8:]) != clientHandler.token {
		errMsg := "Invalid token"
		fmt.Println(errMsg)
		return errors.New(errMsg)
	}

	fileSize := binary.BigEndian.Uint64(req[16:24])
	fileName := string(req[24:])
	_, err := os.Stat(filepath.Join(clientHandler.Config.Directory, fileName))
	if err == nil {
		errMsg := "Upload File exist"
		res := append([]byte{1}, []byte(errMsg)...)
		if err := clientHandler.connection.SendMsg(res); err != nil {
			log.Error("Download SendMsg error: %s", err.Error())
			return errors.New("Download SendMsg error")
		}
	}
	if err := clientHandler.connection.SendMsg(append([]byte{0}, []byte("ok")...)); err != nil {
		log.Error("Download SendMsg error: %s", err.Error())
		return errors.New("Download SendMsg error")
	}

	fileTransfer := FileTransfer{
		connection:   clientHandler.connection,
		FileName:     fileName,
		FileSize:     fileSize,
		DownloadPath: clientHandler.Config.Directory,
	}

	err = fileTransfer.ReceiveFile()
	if err != nil {
		log.Error("Download send file error: %s", err.Error())
		return errors.New("Download send file error")
	}
	return nil
}

/*
	|	8	 |
	|   exit |

*/
func (clientHandler *ClientHandler) Exit() error {
	res := append([]byte{0}, []byte("ok")...)
	if err := clientHandler.connection.SendMsg(res); err != nil {
		log.Error("Exit SendMsg error: %s", err.Error())
		return errors.New("Exit SendMsg error")
	}
	log.Info("Clienthandler exit")
	return nil
}
