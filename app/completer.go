package main

import (
	"os"
	"path/filepath"
	"strings"
)

type ShellCompleter struct{}

func (s *ShellCompleter) Do(line []rune, pos int) ([][]rune, int) {
	prefix := string(line[:pos])

	// Find start of current word
	wordStart := pos
	for wordStart > 0 && prefix[wordStart-1] != ' ' {
		wordStart--
	}
	word := prefix[wordStart:]

	// Are we on the first word? (only spaces before wordStart)
	isFirstWord := true
	for i := 0; i < wordStart; i++ {
		if prefix[i] != ' ' {
			isFirstWord = false
			break
		}
	}

	var matches []string
	if isFirstWord {
		matches = s.completeCommand(word)
	} else {
		matches = s.completeFile(word)
	}

	if len(matches) == 0 {
		return nil, 0
	}

	runes := make([][]rune, len(matches))
	for i, m := range matches {
		runes[i] = []rune(m)
	}
	return runes, len(word)
}

func (s *ShellCompleter) completeCommand(word string) []string {
	seen := make(map[string]bool)
	var matches []string

	// Builtins
	for name := range handlers {
		if strings.HasPrefix(name, word) && !seen[name] {
			seen[name] = true
			matches = append(matches, name)
		}
	}

	// PATH executables
	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			name := e.Name()
			if strings.HasPrefix(name, word) && !seen[name] {
				info, err := e.Info()
				if err != nil {
					continue
				}
				if info.Mode()&0111 != 0 {
					seen[name] = true
					matches = append(matches, name)
				}
			}
		}
	}

	// KEY FIX: if exactly one match, append trailing space
	if len(matches) == 1 {
		matches[0] += " "
	}

	return matches
}

func (s *ShellCompleter) completeFile(word string) []string {
	dir := "."
	prefix := word

	if idx := strings.LastIndex(word, "/"); idx >= 0 {
		dir = word[:idx]
		prefix = word[idx+1:]
		if dir == "" {
			dir = "/"
		}
	}

	if strings.HasPrefix(dir, "~") {
		home, err := os.UserHomeDir()
		if err == nil {
			dir = home + dir[1:]
		}
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var matches []string
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, prefix) {
			if e.IsDir() {
				name += "/"
			}
			if idx := strings.LastIndex(word, "/"); idx >= 0 {
				name = word[:idx+1] + name
			}
			matches = append(matches, name)
		}
	}

	return matches
}
