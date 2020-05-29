package sfs

const (
	BUF_SIZE = 4096
)

const (
	MethodConnect  = "CONNECT"
	MethodList     = "LIST"
	MethodDownload = "DOWNLOAD"
	MethodUpload   = "UPLOAD"
	MethodExit     = "EXIT"

	MethodDownloading       = "DOWNING"
	MethodDownloadCompleted = "DOWNED"
	MethodUploading         = "UPING"
	MethodUploadCompleted   = "UPED"
)

const WhiteSpace = " "

const ChunkSize10MB = 10 * 1024 * 1024
