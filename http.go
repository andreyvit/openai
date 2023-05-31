package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"unicode/utf8"
)

const (
	eventStreamContentType = "text/event-stream"
)

var (
	streamEndMarker = []byte("[DONE]")
)

type Error struct {
	CallID            string
	IsNetwork         bool
	StatusCode        int
	Type              string
	Message           string
	RawResponseBody   []byte
	PrintResponseBody bool
	Cause             error
}

func (e *Error) Error() string {
	var buf strings.Builder
	fmt.Fprintf(&buf, "%s: HTTP %d", e.CallID, e.StatusCode)
	if e.IsNetwork {
		buf.WriteString("network: ")
	}
	if e.Type != "" {
		buf.WriteString(": ")
		buf.WriteString(e.Type)
	}
	if e.Message != "" {
		buf.WriteString(": ")
		buf.WriteString(e.Message)
	}
	if e.Cause != nil {
		buf.WriteString(": ")
		buf.WriteString(e.Cause.Error())
	}
	if e.PrintResponseBody {
		buf.WriteString("  // response: ")
		if len(e.RawResponseBody) == 0 {
			buf.WriteString("<empty>")
		} else if utf8.Valid(e.RawResponseBody) {
			buf.Write(e.RawResponseBody)
		} else {
			return fmt.Sprintf("<binary %d bytes>", len(e.RawResponseBody))
		}
	}
	return buf.String()
}

func (e *Error) Unwrap() error {
	return e.Cause
}

type streamSync = func(data []byte) error

func post(ctx context.Context, callID, endpoint string, client *http.Client, creds Credentials, input any, outputPtr any) error {
	inputRaw := must(json.Marshal(input))
	r := must(http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(inputRaw)))

	h := r.Header
	h.Set("Authorization", "Bearer "+creds.APIKey)
	h.Set("Content-Type", "application/json")
	if creds.OrganizationID != "" {
		h.Set("OpenAI-Organization", creds.OrganizationID)
	}

	// log.Printf("%s: %s", callID, curl(r.Method, r.URL.String(), r.Header, inputRaw))

	resp, err := client.Do(r)
	if err != nil {
		return &Error{
			CallID:    callID,
			IsNetwork: true,
			Cause:     err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		ctype := resp.Header.Get("Content-Type")

		if f, ok := outputPtr.(streamSync); ok {
			if ctype != eventStreamContentType {
				outputRaw, err := io.ReadAll(resp.Body)
				if err != nil {
					return &Error{
						CallID:    callID,
						IsNetwork: true,
						Cause:     err,
					}
				}
				return &Error{
					CallID:            callID,
					StatusCode:        resp.StatusCode,
					RawResponseBody:   outputRaw,
					PrintResponseBody: true,
				}
			}

			err = parseEventStream(resp.Body, 1024*1024, func(id, event string, data []byte) error {
				if bytes.Equal(data, streamEndMarker) {
					return errCloseEventStream
				}
				return f(data)
			}, nil)
			if err != nil {
				return &Error{
					CallID:     callID,
					IsNetwork:  false,
					StatusCode: resp.StatusCode,
					Message:    "error processing chunk of streaming body",
					Cause:      err,
				}
			}
		} else {
			outputRaw, err := io.ReadAll(resp.Body)
			if err != nil {
				return &Error{
					CallID:    callID,
					IsNetwork: true,
					Cause:     err,
				}
			}

			err = json.Unmarshal(outputRaw, outputPtr)
			if err != nil {
				return &Error{
					CallID:            callID,
					IsNetwork:         len(outputRaw) == 0 || outputRaw[0] != '{',
					StatusCode:        resp.StatusCode,
					Message:           "error unmashalling body",
					RawResponseBody:   outputRaw,
					PrintResponseBody: true,
					Cause:             err,
				}
			}
		}
		return nil
	} else {
		outputRaw, err := io.ReadAll(resp.Body)
		if err != nil {
			return &Error{
				CallID:    callID,
				IsNetwork: true,
				Cause:     err,
			}
		}

		errResult := &Error{
			CallID:            callID,
			StatusCode:        resp.StatusCode,
			RawResponseBody:   outputRaw,
			PrintResponseBody: true,
		}
		var errResp errorResponse
		err = json.Unmarshal(outputRaw, &errResp)
		if err == nil && errResp.Error != nil {
			if s, ok := errResp.Error.Type.(string); ok {
				errResult.Type = s
			}
			if s, ok := errResp.Error.Message.(string); ok {
				s = strings.TrimSpace(s) // unlikely, but just in case
				if s != "" {
					errResult.Message = s
					errResult.PrintResponseBody = false
				}
			}
		}
		return errResult
	}
}

type errorResponse struct {
	Error *struct {
		Message any `json:"message"`
		Type    any `json:"type"` // "invalid_request_error", "server_error"
	} `json:"error"`
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func curl(method, path string, headers http.Header, body []byte) string {
	var buf strings.Builder
	buf.WriteString("curl")
	buf.WriteString(" -i")
	if method != "GET" {
		buf.WriteString(" -X")
		buf.WriteString(method)
	} else {
		body = nil
	}
	for k, vv := range headers {
		for _, v := range vv {
			buf.WriteString(" -H '")
			buf.WriteString(k)
			buf.WriteString(": ")
			buf.WriteString(v)
			buf.WriteString("'")
		}
	}
	if body != nil {
		buf.WriteString(" -d ")
		buf.WriteString(shellQuote(string(body)))
	}
	buf.WriteString(" '")
	buf.WriteString(path)
	buf.WriteString("'")
	return buf.String()
}

func shellQuote(source string) string {
	const specialChars = "\\'\"`${[|&;<>()*?! \t\n~"
	const specialInDouble = "$\\\""

	var buf strings.Builder
	buf.Grow(len(source) + 10)

	// pick quotation style, preferring single quotes
	if !strings.ContainsAny(source, specialChars) {
		buf.WriteString(source)
	} else if !strings.ContainsRune(source, '\'') {
		buf.WriteByte('\'')
		buf.WriteString(source)
		buf.WriteByte('\'')
	} else {
		buf.WriteByte('"')
		for {
			i := strings.IndexAny(source, specialInDouble)
			if i < 0 {
				break
			}
			buf.WriteString(source[:i])
			buf.WriteByte('\\')
			buf.WriteByte(source[i])
			source = source[i+1:]
		}
		buf.WriteString(source)
		buf.WriteByte('"')
	}
	return buf.String()
}
