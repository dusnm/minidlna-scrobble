package logparser

import (
	"bufio"
	"io"
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

func ParseLine(line io.Reader) (Data, error) {
	scanner := bufio.NewScanner(line)
	scanner.Split(bufio.ScanRunes)

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
			if c != "]" {
				data.Filepath += c
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return Data{}, err
	}

	return data, nil
}
