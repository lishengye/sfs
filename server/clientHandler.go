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
		method := string(data[:8])
		if method[:len(sfs.MethodList)] == sfs.MethodList {
			err = clientHandler.List(data)
		} else if method[:len(sfs.MethodDownload)] == sfs.MethodDownload {
			err = clientHandler.Download(data)
		} else if method[:len(sfs.MethodUpload)] == sfs.MethodUpload {
			err = clientHandler.Upload(data)
		} else if method[:len(sfs.MethodExit)] == sfs.MethodExit {
			err = clientHandler.Exit()
			return
		} else {
			err = errors.New("invalid command")
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
	// 坑
	if method[:len(sfs.MethodConnect)] != sfs.MethodConnect {
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

	if !sfs.CheckUser(userName, password) {
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
	log.Info("List handling")
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
		result += fmt.Sprintf("%s  %dB\n", v.Name(), v.Size())
	}

	// response
	res := append([]byte{0}, []byte(result)...)
	if err := clientHandler.connection.SendMsg(res); err != nil {
		log.Error("List SendMsg error: %s", err.Error())
		return errors.New("List SendMsg error")
	}
	log.Info("List successfully: %s", clientHandler.Config.Directory)
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
	if token := string(req[8:16]); token != clientHandler.token {
		log.Error("Download invalid token: %s, %s", clientHandler.token, token)
		return errors.New("download invalid token")
	}

	fileName := string(req[16:])
	log.Info("Download handling: %s", fileName)
	v, err := os.Stat(filepath.Join(clientHandler.Config.Directory, fileName))
	if err != nil || v.IsDir() {
		errMsg := "File not found or downloading a directory"
		res := append([]byte{2}, []byte(errMsg)...)
		if err := clientHandler.connection.SendMsg(res); err != nil {
			log.Error("Download SendMsg error: %s", err.Error())
			return errors.New("Download SendMsg error")
		}
		log.Warn("Download handling Warning %s : %s", errMsg, fileName)
		return nil
	}
	if v.Size() == 0 {
		errMsg := "Target file empty"
		res := append([]byte{2}, []byte(errMsg)...)
		if err := clientHandler.connection.SendMsg(res); err != nil {
			log.Error("Download SendMsg error: %s", err.Error())
			return errors.New("Download SendMsg error")
		}
		log.Warn("Download handling Warning %s : %s", errMsg, fileName)
		return nil
	}

	res := append([]byte{0}, sfs.PutUint64(uint64(v.Size()))...)
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
	log.Info("Download handle successful: %s", fileName)
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
	if string(req[8:16]) != clientHandler.token {
		errMsg := "Invalid token"
		fmt.Println(errMsg)
		return errors.New(errMsg)
	}

	fileSize := binary.BigEndian.Uint64(req[16:24])
	fileName := string(req[24:])
	log.Info("Upload handling: %s", fileName)
	_, err := os.Stat(filepath.Join(clientHandler.Config.Directory, fileName))
	if err == nil {
		errMsg := "Overwriting file with the same name in server"
		log.Warn(errMsg + ":" + fileName)
		res := append([]byte{3}, []byte(errMsg)...)
		if err := clientHandler.connection.SendMsg(res); err != nil {
			log.Error("Download SendMsg error: %s", err.Error())
			return errors.New("Download SendMsg error")
		}
	} else {
		if err := clientHandler.connection.SendMsg(append([]byte{0}, []byte("ok")...)); err != nil {
			log.Error("Download SendMsg error: %s", err.Error())
			return errors.New("Download SendMsg error")
		}
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
	log.Info("Upload handle successful: %s", fileName)
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
