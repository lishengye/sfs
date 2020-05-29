package client

import (
	"errors"
	"github.com/lishengye/sfs"
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
req
	| 8 | 8 | 8 | xxx |
	| uping | start | end | bytes |
res
	|	1   |	xxx	|
	|   0	|  "ok"	|
*/
func (fileTransfer *FileTransfer) SendFile() error {
	File, err := os.Open(filepath.Join(fileTransfer.DownloadPath, fileTransfer.FileName))
	if err != nil {
		return err
	}

	start := uint64(0)
	for {
		if start > fileTransfer.FileSize {
			break
		}

		end := start + sfs.ChunkSize10MB
		if end > fileTransfer.FileSize {
			end = fileTransfer.FileSize
		}

		req := make([]byte, 24+end-start)
		copy(req[0:8], sfs.MethodUploading)

		copy(req[8:16], sfs.PutUint64(start))

		copy(req[16:24], sfs.PutUint64(end))

		if _, err := File.Seek(int64(start), 0); err != nil {
			return err
		}

		content := make([]byte, end-start)
		if _, err := io.ReadFull(File, content); err != nil {
			return err
		}

		copy(req[24:], content)

		if err := fileTransfer.connection.SendMsg(req); err != nil {
			return err
		}

		res := make([]byte, 0)
		if err := fileTransfer.connection.ReceiveMsg(req); err != nil {
			return err
		}

		if res[0] != 0 {
			// todo retry
			return errors.New("error")
		}
		start = end

	}

	req := make([]byte, 8)
	copy(req[:], sfs.MethodUploadCompleted)

	if err := fileTransfer.connection.SendMsg(req); err != nil {
		return err
	}

	res := make([]byte, 0)
	if err := fileTransfer.connection.ReceiveMsg(req); err != nil {
		return err
	}
	if res[0] != 0 {
		return errors.New("error")
	}

	return nil
}

/*

req:
	| 8 | 8 | 8 |  xxx |
	| downing/ downloaded | start | end | file_name|
res:
	| 0 | xxx|
*/
func (fileTransfer *FileTransfer) ReceiveFile() error {
	tempFile := fileTransfer.FileName + ".temp"

	start := uint64(0)
	for {
		if start > fileTransfer.FileSize {
			break
		}

		end := start + sfs.ChunkSize10MB
		if end > fileTransfer.FileSize {
			end = fileTransfer.FileSize
		}

		req := make([]byte, 0)
		copy(req[0:8], sfs.MethodDownloading)

		copy(req[8:16], sfs.PutUint64(start))

		copy(req[16:24], sfs.PutUint64(end))

		if err := fileTransfer.connection.SendMsg(req); err != nil {
			return err
		}

		res := make([]byte, 0)
		if err := fileTransfer.connection.ReceiveMsg(res); err != nil {
			return err
		}
		if res[0] != 0 {
			return errors.New("error")
		}

		content := res[1:]

		File, err := os.OpenFile(filepath.Join(fileTransfer.DownloadPath, tempFile),
			os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			return err
		}

		v, _ := File.Stat()
		if start != uint64(v.Size()) {
			return errors.New("Size unmatch")
		}

		_, err = File.Write(content)
		if err != nil {
			return err
		}

		if err := File.Close(); err != nil {
			return err
		}

		start = end
	}

	req := make([]byte, 8)
	copy(req[:], sfs.MethodDownloadCompleted)

	if err := fileTransfer.connection.SendMsg(req); err != nil {
		return err
	}

	res := make([]byte, 0)
	if err := fileTransfer.connection.ReceiveMsg(req); err != nil {
		return err
	}
	if res[0] != 0 {
		return errors.New("error")
	}

	return nil
}
