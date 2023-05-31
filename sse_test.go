package openai

import (
	"fmt"
	"io"
	"strings"
	"testing"
	"testing/iotest"
)

func TestParseEventStream(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			"empty stream",
			"",
			"",
		},
		{
			"empty message",
			"\n\n",
			"",
		},
		{
			"empty data, but sent because LF stripping is after empty check",
			"data: \n\n",
			"# <> []",
		},
		{
			"simple message",
			"data: foo\n\n",
			"# <> [foo]",
		},
		{
			"multiline message",
			"data: foo\ndata: bar\n\n",
			"# <> [foo\nbar]",
		},
		{
			"event, multiline message",
			"event: test\ndata: foo\ndata: bar\n\n",
			"# <test> [foo\nbar]",
		},
		{
			"id, event, multiline message",
			"id: 1234\nevent: test\ndata: foo\ndata: bar\n\n",
			"#1234 <test> [foo\nbar]",
		},
		{
			"no spaces after fields",
			"id:1234\nevent:test\ndata:foo\ndata:bar\n\n",
			"#1234 <test> [foo\nbar]",
		},
		{
			"three messages",
			"id: 1234\nevent: test\ndata: foo\ndata: bar\n\nid:5\ndata:boz\n\ndata:fubar\nevent:boo\n\n",
			"#1234 <test> [foo\nbar] | #5 <> [boz] | # <boo> [fubar]",
		},
		{
			"retry",
			"id: 1234\nevent: test\ndata: foo\ndata: bar\n\nid:5\nretry: 500\ndata:boz\n\ndata:fubar\nevent:boo\n\n",
			"#1234 <test> [foo\nbar] | retry 500 | #5 <> [boz] | # <boo> [fubar]",
		},

		{
			"unfinished message",
			"data: foo\n",
			"",
		},
		{
			"unfinished message at the end",
			"id: 1234\nevent: test\ndata: foo\ndata: bar\n\nid:5\ndata:boz\n\ndata:fubar\nevent:boo\n\ndata: foo\n",
			"#1234 <test> [foo\nbar] | #5 <> [boz] | # <boo> [fubar]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := parseEventStreamStr(tt.input)
			if actual != tt.expected {
				t.Errorf("** ParseStream(%q) == %q, expected %q", tt.input, actual, tt.expected)
			}
		})
	}
}

func parseEventStreamStr(input string) string {
	var events []string
	parseEventStream(io.NopCloser(iotest.OneByteReader(strings.NewReader(input))), 1024, func(id, event string, data []byte) error {
		events = append(events, fmt.Sprintf("#%s <%s> [%s]", id, event, data))
		return nil
	}, func(ms uint64) {
		events = append(events, fmt.Sprintf("retry %d", ms))
	})
	return strings.Join(events, " | ")
}
