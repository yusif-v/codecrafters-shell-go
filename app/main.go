package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"unicode"
)

type HandlerFunc func(args []string) error

var handlers map[string]HandlerFunc

// The Builtins
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
	fmt.Println(strings.Join(args, " "))
	return nil
}

func exitFunc(args []string) error {
	code := 0
	if len(args) > 0 {
		fmt.Sscanf(args[0], "%d", &code)
	}
	os.Exit(code)
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
	}

	target := args[0]
	if target == "~" {
		dir, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		target = dir
	}

	err := os.Chdir(target)
	if err != nil {
		fmt.Printf("cd: %s: No such file or directory\n", args[0])
	}
	return nil
}

// Parse the input with backslash escaping support
func parseCommand(line string) []string {
	var args []string
	var current strings.Builder
	inSingle := false
	inDouble := false
	escaped := false

	for _, ch := range line {
		if escaped {
			// Previous char was backslash outside quotes,
			// this char is literal regardless of what it is
			current.WriteRune(ch)
			escaped = false
			continue
		}

		switch ch {
		case '\\':
			if !inSingle && !inDouble {
				// Backslash outside quotes: escape next character
				escaped = true
			} else {
				// Inside quotes, backslash is literal (for now)
				current.WriteRune(ch)
			}
		case '"':
			if !inSingle {
				if !inDouble {
					inDouble = true
				} else {
					inDouble = false
				}
			} else {
				current.WriteRune(ch)
			}
		case '\'':
			if !inDouble {
				if !inSingle {
					inSingle = true
				} else {
					inSingle = false
				}
			} else {
				current.WriteRune(ch)
			}
		default:
			if unicode.IsSpace(ch) && !(inSingle || inDouble) {
				if current.Len() > 0 {
					args = append(args, current.String())
					current.Reset()
				}
				// If current is empty, skip whitespace (handles multiple spaces between args)
			} else {
				current.WriteRune(ch)
			}
		}
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}
	return args
}

// Main Loop
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

		parts := parseCommand(line)
		if len(parts) == 0 {
			continue
		}

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
