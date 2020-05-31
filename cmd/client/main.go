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

	cli := client.Cli{}
	client := client.Client{}

	if err := client.Connect(*ip, *port); err != nil {
		cli.Error("Connect error")
		return
	}

	cli.Print("Enter password: ")
	pass := cli.GetComand()[0]


	if err := client.Handshake(*user, pass); err != nil {
		cli.Error("Hanshake failed: " + err.Error())
		return
	}

	for {
		cli.Promt("")
		command := cli.GetComand()
		n := len(command)
		if command[0] == "help" {
			cli.Info("invalid command:\n" +
				"list\n" +
				"download file\n" +
				"upload file\n" +
				"help\n")
		} else if command[0] == "list"{
			if n != 1 {
				cli.Error("Command too long")
			}
			if res, err := client.List(); err != nil {
				cli.Error("List error "+err.Error())
				return
			} else {
				cli.PrintLn(res)
			}
		} else if command[0] == "download"{
			if n != 2 {
				cli.Error("Please specific a file")
				continue
			}
			fileName := command[1]
			if err := client.Download(fileName); err != nil {
				cli.Error("Please specific a file")
				return
			}
		} else if command[0] == "upload"{
			if n != 2 {
				cli.Error("Please specific a file")
				continue
			}
			filePath := command[1]
			file, err := os.Stat(filePath)
			if err != nil {
				cli.Error("File not exist")
				continue
			}
			if err := client.Upload(file.Name(), uint64(file.Size())); err != nil {
				cli.Error("Please specific a file")
				return
			}
		} else {
			cli.Info("invalid command: list, download, upload, help\n")
		}
	}

}
