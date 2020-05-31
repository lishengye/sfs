package main

import (
	"flag"
	"os"
)
import "github.com/lishengye/sfs/client"

func main() {
	ip := flag.String("s", "", "remote ip")
	port := flag.String("p", "6679", "remote port")
	user := flag.String("u",  "admin", "username")
	flag.Parse()

	commandLine := client.NewCommandLine()
	client := client.NewClient(commandLine)

	if err := client.Connect(*ip, *port); err != nil {
		commandLine.Error("Connect error")
		return
	}

	commandLine.Print("Enter password: ")
	text := commandLine.GetComand()
	if len(text) == 0 {
		commandLine.Error("Password cannot be empty")
		return
	}
	pass := text[0]


	if err := client.Handshake(*user, pass); err != nil {
		commandLine.Error("Hanshake failed: " + err.Error())
		return
	}

	for {
		commandLine.Promt("")
		command := commandLine.GetComand()
		n := len(command)
		if n == 0 {
			continue
		}
		if command[0] == "help" {
			commandLine.PrintLn("list: list file\n" +
				"download filename: download file from server\n" +
				"upload filename:  upload file to server\n" +
				"help: help")
		} else if command[0] == "list"{
			if n != 1 {
				commandLine.Warn("Command too long")
			}
			if res, err := client.List(); err != nil {
				commandLine.Error("List error "+err.Error())
				return
			} else {
				commandLine.Print(res)
			}
		} else if command[0] == "download"{
			if n != 2 {
				commandLine.Error("Please specific a file")
				continue
			}
			fileName := command[1]
			if err := client.Download(fileName); err != nil {
				commandLine.Error(err.Error())
				return
			}
		} else if command[0] == "upload"{
			if n != 2 {
				commandLine.Error("Please specific a file")
				continue
			}
			filePath := command[1]
			file, err := os.Stat(filePath)
			if err != nil {
				commandLine.Error("File not exist")
				continue
			}
			if err := client.Upload(filePath, uint64(file.Size())); err != nil {
				commandLine.Error(err.Error())
				return
			}
		} else if command[0] == "exit"{
			if n != 1 {
				commandLine.Warn("Command too long")
			}
			if err := client.Exit(); err != nil {
				commandLine.Error("List error "+err.Error())
			}
			return
		} else {
			commandLine.Error("Accepted command: list, download, upload, help")
		}
	}

}
