package main

import (
	"fmt"
	"strings"
	"unicode"
)

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

// parseRedirection separates command args from redirections.
// Returns: cmdArgs, stdoutFile, stdoutAppend, stderrFile, stderrAppend, error
func parseRedirection(args []string) ([]string, string, bool, string, bool, error) {
	var cmdArgs []string
	var stdoutFile, stderrFile string
	var stdoutAppend, stderrAppend bool

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case ">", "1>":
			if stdoutFile != "" {
				return nil, "", false, "", false, fmt.Errorf("multiple stdout redirections")
			}
			if i+1 >= len(args) {
				return nil, "", false, "", false, fmt.Errorf("no filename for stdout redirection")
			}
			stdoutFile = args[i+1]
			stdoutAppend = false
			i++
		case ">>", "1>>":
			if stdoutFile != "" {
				return nil, "", false, "", false, fmt.Errorf("multiple stdout redirections")
			}
			if i+1 >= len(args) {
				return nil, "", false, "", false, fmt.Errorf("no filename for stdout redirection")
			}
			stdoutFile = args[i+1]
			stdoutAppend = true
			i++
		case "2>":
			if stderrFile != "" {
				return nil, "", false, "", false, fmt.Errorf("multiple stderr redirections")
			}
			if i+1 >= len(args) {
				return nil, "", false, "", false, fmt.Errorf("no filename for stderr redirection")
			}
			stderrFile = args[i+1]
			stderrAppend = false
			i++
		case "2>>":
			if stderrFile != "" {
				return nil, "", false, "", false, fmt.Errorf("multiple stderr redirections")
			}
			if i+1 >= len(args) {
				return nil, "", false, "", false, fmt.Errorf("no filename for stderr redirection")
			}
			stderrFile = args[i+1]
			stderrAppend = true
			i++
		default:
			cmdArgs = append(cmdArgs, arg)
		}
	}

	return cmdArgs, stdoutFile, stdoutAppend, stderrFile, stderrAppend, nil
}
