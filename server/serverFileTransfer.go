package server

import (
	"encoding/binary"
	"errors"
	"github.com/lishengye/sfs"
	"github.com/lishengye/sfs/log"
	"io"
	"os"
	"path/filepath"
)

type FileTransfer struct {
	connection   *sfs.Connection
	DownloadPath string
	FileSize     uint64
	FileName     string
}

/*

req:
	| 8 | 8 | 8 |
	| downing/ downloaded | start | end |
res:
	| 0 | xxx|
*/

func (fileTransfer *FileTransfer) SendFile() error {
	File, err := os.Open(filepath.Join(fileTransfer.DownloadPath, fileTransfer.FileName));
	if err != nil {
		return err
	}

	log.Info("SendFile: %s starting", fileTransfer.FileName)

	for {
		req := make([]byte, 0)
		if err := fileTransfer.connection.ReceiveMsg(req); err != nil {
			log.Error("SendFile receivemsg error: %s", err)
			return errors.New("SendFile receivemsg error")
		}

		method := string(req[0:8])
		if method == sfs.MethodDownloading {
			fileName := string(req[0:8])
			start, end := binary.BigEndian.Uint64(req[8:16]), binary.BigEndian.Uint64(req[16:24])

			if fileName != fileTransfer.FileName || end > uint64(fileTransfer.FileSize) {
				log.Error("invalid argument fileName: %s, start: %d, end: %d", fileName, start, end)
				return errors.New("invalid argument")
			}

			if _, err := File.Seek(int64(start), 0); err != nil {
				log.Error("Seek error: %s", err.Error())
				return err
			}

			content := make([]byte, end-start)
			if _, err := io.ReadFull(File, content); err != nil {
				log.Error(err.Error())
				return err
			}

			res := append([]byte{0}, content...)
			if err := fileTransfer.connection.SendMsg(res); err != nil {
				return errors.New("error")
			}
		} else if method == sfs.MethodDownloadCompleted {
			res := append([]byte{0}, []byte("ok")...)
			if err := fileTransfer.connection.SendMsg(res); err != nil {
				log.Error("SendFile Sendmsg error")
				return errors.New("Sendmsg error")
			}
			break
		} else {
			errMsg := "Download Command invalid"
			log.Error(errMsg)
			return errors.New(errMsg)
		}

	}
	log.Info("SendFile successfully: %s", fileTransfer.FileName)
	return nil
}

/*
req
	| 8 | 8 | 8 | xxx |
	| uping | start | end | bytes |
res
	|	1   |	xxx	|
	|   0	|  "ok"	|
*/
func (fileTransfer *FileTransfer) ReceiveFile() error {
	log.Info("ReceiveFile: %s starting", fileTransfer.FileName)
	tempFile := fileTransfer.FileName + ".temp"

	for {
		req := make([]byte, 0)
		if err := fileTransfer.connection.ReceiveMsg(req); err != nil {
			log.Error("ReceiveFile receivemsg error: %s", err.Error())
			return errors.New("ReceiveFile receivemsg error")
		}
		method := string(req[0:8])

		File, err := os.OpenFile(filepath.Join(fileTransfer.DownloadPath, tempFile),
			os.O_APPEND | os.O_WRONLY | os.O_CREATE, 0644)
		if err != nil {
			log.Error("Open tempfile error: %s", err.Error())
			return errors.New("open tempfile error")
		}
		v, _ := File.Stat()

		if method == sfs.MethodUploading {
			start, end := binary.BigEndian.Uint64(req[8:16]), binary.BigEndian.Uint64(req[16:24])
			if v.Size() != int64(start) {
				log.Error("FileSize: %d, writing start: %d", v.Size(), start)
				return errors.New("Write start boundary not end")
			}

			log.Info("ReceiveFile write %d:%d chunk to %s", start, end, fileTransfer.FileName+".temp")

			content := req[24:]
			_, err := File.Write(content);
			if  err != nil {
				log.Error("ReceiveFile write to tempfile error: %s", err.Error())
				return errors.New("ReceiveFile write to tempfile error")
			}

			res := append([]byte{0}, []byte("ok")...)
			if err := fileTransfer.connection.SendMsg(res); err != nil {
				log.Error("error")
				return errors.New("error")
			}

			if err := File.Close(); err != nil {
				log.Error("CloseFile error")
				return errors.New("closefile error")
			}
		} else if method == sfs.MethodUploadCompleted {
			res := append([]byte{0}, []byte("ok")...)
			if err := fileTransfer.connection.SendMsg(res); err != nil {
				return errors.New("error")
			}

			if err := File.Close(); err != nil {
				log.Error("CloseFile error")
				return errors.New("closefile error")
			}

			if err := os.Rename(filepath.Join(fileTransfer.DownloadPath, tempFile),
				filepath.Join(fileTransfer.DownloadPath, fileTransfer.FileName)); err != nil {
					log.Error("Receive Rename error")
					return errors.New("Rename error")
			}
			break
		} else {
			log.Error("Receive File invalid command: %s", method)
			return errors.New("Receive File invalid command")
		}

	}
	log.Info("ReceiveFile: %s successfully", fileTransfer.FileName)
	return nil
}
