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

type HandlerFunc func(args []string, stdout, stderr io.Writer) error

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

func echoFunc(args []string, stdout, stderr io.Writer) error {
	fmt.Fprintln(stdout, strings.Join(args, " "))
	return nil
}

func exitFunc(args []string, stdout, stderr io.Writer) error {
	code := 0
	if len(args) > 0 {
		fmt.Sscanf(args[0], "%d", &code)
	}
	os.Exit(code)
	return nil
}

func typeFunc(args []string, stdout, stderr io.Writer) error {
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

func pwdFunc(args []string, stdout, stderr io.Writer) error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	fmt.Fprintln(stdout, dir)
	return nil
}

func cdFunc(args []string, stdout, stderr io.Writer) error {
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

// parseRedirection separates command args from stdout (>) and stderr (2>) redirections
func parseRedirection(args []string) ([]string, string, string, error) {
	var cmdArgs []string
	var stdoutFile, stderrFile string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == ">" || arg == "1>" {
			if stdoutFile != "" {
				return nil, "", "", fmt.Errorf("multiple stdout redirections")
			}
			if i+1 >= len(args) {
				return nil, "", "", fmt.Errorf("no filename for stdout redirection")
			}
			stdoutFile = args[i+1]
			i++
		} else if arg == "2>" {
			if stderrFile != "" {
				return nil, "", "", fmt.Errorf("multiple stderr redirections")
			}
			if i+1 >= len(args) {
				return nil, "", "", fmt.Errorf("no filename for stderr redirection")
			}
			stderrFile = args[i+1]
			i++
		} else {
			cmdArgs = append(cmdArgs, arg)
		}
	}

	return cmdArgs, stdoutFile, stderrFile, nil
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

		parts := parseCommand(line)
		if len(parts) == 0 {
			continue
		}

		cmd := parts[0]
		args := parts[1:]

		cmdArgs, stdoutFile, stderrFile, err := parseRedirection(args)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			continue
		}

		var stdout io.Writer = os.Stdout
		var stderr io.Writer = os.Stderr
		var outFile, errFile *os.File

		if stdoutFile != "" {
			outFile, err = os.Create(stdoutFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				continue
			}
			stdout = outFile
		}

		if stderrFile != "" {
			errFile, err = os.Create(stderrFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				if outFile != nil {
					outFile.Close()
				}
				continue
			}
			stderr = errFile
		}

		handler, ok := handlers[cmd]

		if ok {
			if err := handler(cmdArgs, stdout, stderr); err != nil {
				fmt.Fprintln(os.Stderr, "error:", err)
			}
		} else {
			path, err := exec.LookPath(cmd)
			if err != nil {
				fmt.Println(cmd + ": command not found")
				if outFile != nil {
					outFile.Close()
				}
				if errFile != nil {
					errFile.Close()
				}
				continue
			}

			execCmd := exec.Command(path, cmdArgs...)
			execCmd.Args = append([]string{cmd}, cmdArgs...)
			execCmd.Stdout = stdout
			execCmd.Stderr = stderr
			execCmd.Stdin = os.Stdin

			if err := execCmd.Run(); err != nil {
				if _, ok := err.(*exec.ExitError); !ok {
					fmt.Fprintln(os.Stderr, "error:", err)
				}
			}
		}

		if outFile != nil {
			outFile.Close()
		}
		if errFile != nil {
			errFile.Close()
		}
	}
}
