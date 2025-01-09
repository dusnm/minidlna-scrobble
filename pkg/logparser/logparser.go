package logparser

import (
	"bufio"
	"strings"
)

const (
	stateSource = iota
	stateLineNumber
	stateLogLevel
	stateMessage
	stateMessageID
	stateFilepath
)

type (
	Data struct {
		SourceFile string
		LineNumber string
		LogLevel   string
		Message    string
		MessageID  string
		Filepath   string
	}
)

func ParseLine(line string) (Data, error) {
	scanner := bufio.NewScanner(strings.NewReader(line))
	scanner.Split(bufio.ScanRunes)

	// This is the end delimeter
	// that needs to be handled like
	// a special case, if it appears in the filepath
	special := "]"
	specialCount := 0
	totalSpecialCount := strings.Count(line, special)
	state := stateSource
	data := Data{}
	for scanner.Scan() {
		c := scanner.Text()
		switch state {
		case stateSource:
			if c != ":" {
				data.SourceFile += c
				continue
			}

			state = stateLineNumber
		case stateLineNumber:
			if c != ":" {
				data.LineNumber += c
				continue
			}

			state = stateLogLevel
		case stateLogLevel:
			if c != ":" {
				data.LogLevel += strings.TrimSpace(c)
				continue
			}

			state = stateMessage
		case stateMessage:
			if c != ":" {
				data.Message += strings.TrimSpace(c)
				continue
			}

			state = stateMessageID
		case stateMessageID:
			if c != "[" {
				data.MessageID += strings.TrimSpace(c)
				continue
			}

			state = stateFilepath
		case stateFilepath:
			if c == special {
				// handle the case where this is not the last occurrence
				specialCount++
				if specialCount < totalSpecialCount {
					data.Filepath += c
				}
			} else {
				data.Filepath += c
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return Data{}, err
	}

	return data, nil
}
