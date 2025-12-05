package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type HandlerFunc func(args []string) error

var handlers = map[string]HandlerFunc{
	"echo": echoFunc,
	"exit": exitFunc,
	"type": typeFunc,
}

func echoFunc(args []string) error {
	fmt.Println(strings.Join(args, " "))
	return nil
}

func exitFunc(args []string) error {
	os.Exit(0)
	return nil
}

func typeFunc(args []string) error {
	return nil
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("$ ")

		command, err := reader.ReadString('\n')

		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			os.Exit(1)
		}

		line := strings.TrimSpace(command)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		cmd := parts[0]
		args := parts[1:]

		handler, ok := handlers[cmd]
		if !ok {
			fmt.Println(cmd + ": command not found")
			continue
		}

		if err := handler(args); err != nil {
			fmt.Println("error:", err)
		}
	}
}
