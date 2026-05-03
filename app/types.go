package main

import "io"

type HandlerFunc func(args []string, stdout, stderr io.Writer) error
