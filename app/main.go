package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

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

		cmdArgs, stdoutFile, stdoutAppend, stderrFile, stderrAppend, err := parseRedirection(args)
		if err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			continue
		}

		var stdout io.Writer = os.Stdout
		var stderr io.Writer = os.Stderr
		var outFile, errFile *os.File

		if stdoutFile != "" {
			flags := os.O_CREATE | os.O_WRONLY
			if stdoutAppend {
				flags |= os.O_APPEND
			} else {
				flags |= os.O_TRUNC
			}
			outFile, err = os.OpenFile(stdoutFile, flags, 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				continue
			}
			stdout = outFile
		}

		if stderrFile != "" {
			flags := os.O_CREATE | os.O_WRONLY
			if stderrAppend {
				flags |= os.O_APPEND
			} else {
				flags |= os.O_TRUNC
			}
			errFile, err = os.OpenFile(stderrFile, flags, 0644)
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
			if err := runExternal(cmd, cmdArgs, stdout, stderr); err != nil {
				fmt.Fprintln(os.Stderr, err)
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
