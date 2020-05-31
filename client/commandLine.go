package client

import (
	"bufio"
	"fmt"
	"github.com/fatih/color"
	"os"
	"strings"
)

type CommandLine struct {
	cyan	func(a ...interface{})
	info    func(a ...interface{})
	error   func(a ...interface{})
	warn   func(a ...interface{})
}

func NewCommandLine() CommandLine {
	return CommandLine{
		cyan:	color.New(color.FgCyan).PrintFunc(),
		info:	color.New(color.FgBlue).PrintFunc(),
		error:	color.New(color.FgRed).PrintFunc(),
		warn:	color.New(color.FgYellow).PrintFunc(),
	}
}

func (cli *CommandLine) Promt(s string) {
	cli.cyan(">>>")
	fmt.Printf(" %s", s)
}

func (cli *CommandLine) Info(s string) {
	fmt.Printf("[Info] %s\n", s)
}

func (cli *CommandLine) Error(s string) {
	cli.error("[Error]")
	fmt.Printf(" %s\n", s)
}

func (cli *CommandLine) Warn(s string) {
	cli.warn("[Warn]")
	fmt.Printf(" %s\n", s)
}

func (cli *CommandLine) Print(s string) {
	fmt.Printf(s)
}

func (cli *CommandLine) PrintLn(s string) {
	fmt.Println(s)
}

func (cli *CommandLine) GetComand() []string {
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	return strings.Fields(text)
}
