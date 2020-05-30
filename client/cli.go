package client

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Cli struct {
}


func (cli *Cli) Promt(s string) {
	fmt.Println(">>> %s", s)
}

func (cli *Cli) Info(s string) {
	fmt.Println("[Info] %s", s)
}

func (cli *Cli) Error(s string) {
	fmt.Println("[Error] %s", s)
}

func (cli *Cli) Print(s string) {
	fmt.Printf("Enter password:")
}

func (cli *Cli) GetComand() []string {
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	return strings.Fields(text)
}
