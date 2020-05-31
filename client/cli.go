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
	fmt.Printf(">>> %s", s)
}

func (cli *Cli) Info(s string) {
	fmt.Printf("[Info] %s\n", s)
}

func (cli *Cli) Error(s string) {
	fmt.Printf("[Error] %s\n", s)
}

func (cli *Cli) Print(s string) {
	fmt.Printf(s)
}

func (cli *Cli) PrintLn(s string) {
	fmt.Printf(s)
}

func (cli *Cli) GetComand() []string {
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	return strings.Fields(text)
}
