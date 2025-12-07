package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type HandlerFunc func(args []string) error

var handlers map[string]HandlerFunc

func init() {
	handlers = map[string]HandlerFunc{
		"echo": echoFunc,
		"exit": exitFunc,
		"type": typeFunc,
	}
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
	if len(args) == 0 {
		return nil
	}
	for _, cmd := range args {
		if _, ok := handlers[cmd]; ok {
			fmt.Println(cmd + " is a shell builtin")
		} else {
			path, err := exec.LookPath(cmd)
			if err == nil {
				fmt.Println(cmd + " is " + path)
			} else {
				fmt.Println(cmd + ": not found")
			}
		}
	}
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
		if ok {
			if err := handler(args); err != nil {
				fmt.Println("error:", err)
			}
		} else {
			path, err := exec.LookPath(cmd)
			if err != nil {
				fmt.Println(cmd + ": command not found")
				continue
			}
			execCmd := exec.Command(path, args...)
			execCmd.Args = append([]string{cmd}, args...)
			execCmd.Stdout = os.Stdout
			execCmd.Stderr = os.Stderr
			execCmd.Stdin = os.Stdin
			if err := execCmd.Run(); err != nil {
				fmt.Println("error:", err)
			}
		}
	}
}
