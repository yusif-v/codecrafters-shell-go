package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"unicode"
)

type HandlerFunc func(args []string, stdout io.Writer) error

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

func echoFunc(args []string, stdout io.Writer) error {
	fmt.Fprintln(stdout, strings.Join(args, " "))
	return nil
}

func exitFunc(args []string, stdout io.Writer) error {
	code := 0
	if len(args) > 0 {
		fmt.Sscanf(args[0], "%d", &code)
	}
	os.Exit(code)
	return nil
}

func typeFunc(args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return nil
	}
	for _, cmd := range args {
		if _, ok := handlers[cmd]; ok {
			fmt.Fprintln(stdout, cmd+" is a shell builtin")
		} else {
			path, err := exec.LookPath(cmd)
			if err == nil {
				fmt.Fprintln(stdout, cmd+" is "+path)
			} else {
				fmt.Fprintln(stdout, cmd+": not found")
			}
		}
	}
	return nil
}

func pwdFunc(args []string, stdout io.Writer) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	fmt.Fprintln(stdout, dir)
	return nil
}

func cdFunc(args []string, stdout io.Writer) error {
	if len(args) == 0 {
		dir, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		err = os.Chdir(dir)
		if err != nil {
			fmt.Fprintf(stdout, "cd: %s: No such file or directory\n", dir)
		}
		return nil
	} else if len(args) > 1 {
		fmt.Fprintln(stdout, "too many arguments")
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
		fmt.Fprintf(stdout, "cd: %s: No such file or directory\n", args[0])
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

	for i, ch := range line {
		if escaped {
			current.WriteRune(ch)
			escaped = false
			continue
		}

		switch ch {
		case '\\':
			if inSingle {
				current.WriteRune(ch)
			} else if inDouble {
				nextIdx := i + 1
				if nextIdx < len(line) {
					nextCh := rune(line[nextIdx])
					if nextCh == '"' || nextCh == '\\' || nextCh == '$' || nextCh == '`' || nextCh == '\n' {
						escaped = true
					} else {
						current.WriteRune(ch)
					}
				} else {
					current.WriteRune(ch)
				}
			} else {
				escaped = true
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

// parseRedirection separates command args from output redirection
func parseRedirection(args []string) ([]string, string, error) {
	var cmdArgs []string
	var outputFile string
	foundRedirect := false

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == ">" || arg == "1>" {
			if foundRedirect {
				return nil, "", fmt.Errorf("multiple redirections not supported")
			}
			foundRedirect = true
			if i+1 >= len(args) {
				return nil, "", fmt.Errorf("no filename specified for redirection")
			}
			outputFile = args[i+1]
			i++ // skip the filename
		} else {
			cmdArgs = append(cmdArgs, arg)
		}
	}

	return cmdArgs, outputFile, nil
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

		// Check for redirection
		cmdArgs, outputFile, err := parseRedirection(args)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			continue
		}

		var stdout io.Writer = os.Stdout
		var file *os.File
		if outputFile != "" {
			var err error
			file, err = os.Create(outputFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				continue
			}
			stdout = file
		}

		handler, ok := handlers[cmd]

		if ok {
			if err := handler(cmdArgs, stdout); err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
			}
		} else {
			path, err := exec.LookPath(cmd)
			if err != nil {
				fmt.Println(cmd + ": command not found")
				if file != nil {
					file.Close()
				}
				continue
			}

			execCmd := exec.Command(path, cmdArgs...)
			execCmd.Args = append([]string{cmd}, cmdArgs...)
			execCmd.Stdout = stdout
			execCmd.Stderr = os.Stderr
			execCmd.Stdin = os.Stdin

			if err := execCmd.Run(); err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
			}
		}

		if file != nil {
			file.Close()
		}
	}
}
