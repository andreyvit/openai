package openai

import (
	"bufio"
	"bytes"
	"errors"
	"io"
)

var (
	dataPrefixBytes     = []byte("data:")
	eventPrefixBytes    = []byte("event:")
	idPrefixBytes       = []byte("id:")
	bomBytes            = []byte{0xEF, 0xBB, 0xBF}
	errCloseEventStream = errors.New("close event stream")
)

// Note: this version doesn't handle LF line terminators because the chance
// of OpenAI using those is nil. Retry fields have also been removed.
func parseEventStream(r io.Reader, maxSize int, f func(id, event string, data []byte) error, retryf func(ms uint64)) error {
	scanner := bufio.NewScanner(r)

	// Use a stack-allocated buffer if we can fit lines into it
	var data [512]byte
	scanner.Buffer(data[:], maxSize)

	var id, event string
	var dataBuf bytes.Buffer
	for scanner.Scan() {
		line := scanner.Bytes()
		line = bytes.TrimPrefix(line, bomBytes)

		if len(line) == 0 {
			if dataBuf.Len() > 0 {
				data := stripTrailingLF(dataBuf.Bytes())
				if err := f(id, event, data); err != nil {
					if err == errCloseEventStream {
						err = nil
					}
					return err
				}
				id = ""
				event = ""
				dataBuf.Reset()
			}

		} else if data, ok := bytes.CutPrefix(line, dataPrefixBytes); ok {
			dataBuf.Write(stripLeadingSpace(data))
			dataBuf.WriteByte('\n')

		} else if data, ok := bytes.CutPrefix(line, eventPrefixBytes); ok {
			event = string(stripLeadingSpace(data))

		} else if data, ok := bytes.CutPrefix(line, idPrefixBytes); ok {
			id = string(stripLeadingSpace(data))
		} // ignore unknown lines
	}
	// incomplete events are discarded
	return scanner.Err()
}

func stripLeadingSpace(data []byte) []byte {
	if len(data) > 0 && data[0] == ' ' {
		return data[1:]
	} else {
		return data
	}
}

func stripTrailingLF(data []byte) []byte {
	n := len(data)
	if n > 0 && data[n-1] == '\n' {
		return data[:n-1]
	} else {
		return data
	}
}
