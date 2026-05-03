package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

func runExternal(cmd string, args []string, stdout, stderr io.Writer) error {
	path, err := exec.LookPath(cmd)
	if err != nil {
		return fmt.Errorf("%s: command not found", cmd)
	}

	execCmd := exec.Command(path, args...)
	execCmd.Args = append([]string{cmd}, args...)
	execCmd.Stdout = stdout
	execCmd.Stderr = stderr
	execCmd.Stdin = os.Stdin

	if err := execCmd.Run(); err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			return err
		}
	}
	return nil
}
