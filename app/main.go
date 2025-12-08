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
		"pwd":  pwdFunc,
		"cd":   cdFunc,
	}
}

func echoFunc(args []string) error {
	merged := strings.Join(args, " ")
	var result strings.Builder
	inQuote := false
	for _, ch := range merged {
		if ch == '\'' {
			inQuote = !inQuote
			continue
		}
		if ch != ' ' || inQuote {
			result.WriteRune(ch)
		} else {
			result.WriteRune(' ')
		}
	}
	fmt.Println(result.String())
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

func pwdFunc(args []string) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	fmt.Println(dir)
	return nil
}

func cdFunc(args []string) error {
	if len(args) == 0 {
		dir, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		err = os.Chdir(dir)
		if err != nil {
			fmt.Printf("cd: %s: No such file or directory\n", dir)
		}
		return nil
	} else if len(args) > 1 {
		fmt.Println("too many arguments")
		return nil
	} else if len(args) == 1 {
		if args[0] == "~" {
			dir, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			err = os.Chdir(dir)
			if err != nil {
				fmt.Printf("cd: %s: No such file or directory\n", dir)
			}
			return nil
		}
		err := os.Chdir(args[0])
		if err != nil {
			fmt.Printf("cd: %s: No such file or directory\n", args[0])
		}
		return nil
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
