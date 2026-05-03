package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

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
