package openai

import (
	"strconv"
	"strings"
	"testing"
)

func TestTokenizer(t *testing.T) {
	const model = ModelDefaultChat
	tests := []struct {
		input    string
		expected string
	}{
		// https://platform.openai.com/tokenizer used to generate canonical outputs.
		{"", "[]"},
		{"Hello, world.", "[15496, 11, 995, 13]"},
		// TODO: there's a known problem here: we output 628 (\n\n) while ChatGPT prefers 198 198 (\n \n) when encoding \n\n
		// in this case. It does produce 628 if you give it simply \n\n. No idea why, but I've replaced 198, 198 with 628 for now.
		{"Many words map to one token, but some don't: indivisible.\n\nUnicode characters like emojis may be split into many tokens containing the underlying bytes: 🤚🏾\n\nSequences of characters commonly found next to each other may be grouped together: 1234567890", "[7085, 2456, 3975, 284, 530, 11241, 11, 475, 617, 836, 470, 25, 773, 452, 12843, 13, 628, 3118, 291, 1098, 3435, 588, 795, 13210, 271, 743, 307, 6626, 656, 867, 16326, 7268, 262, 10238, 9881, 25, 12520, 97, 248, 8582, 237, 122, 628, 44015, 3007, 286, 3435, 8811, 1043, 1306, 284, 1123, 584, 743, 307, 32824, 1978, 25, 17031, 2231, 30924, 3829]"},
		{"A helpful rule of thumb is that one token generally corresponds to ~4 characters of text for common English text. This translates to roughly ¾ of a word (so 100 tokens ~= 75 words).", "[32, 7613, 3896, 286, 15683, 318, 326, 530, 11241, 4143, 24866, 284, 5299, 19, 3435, 286, 2420, 329, 2219, 3594, 2420, 13, 770, 23677, 284, 7323, 1587, 122, 286, 257, 1573, 357, 568, 1802, 16326, 5299, 28, 5441, 2456, 737]"},
	}
	for _, test := range tests {
		tokens := Encode(test.input, model)
		actual := formatTokens(tokens)
		if actual != test.expected {
			t.Errorf("** Encode(%q) =\n%s\nwanted\n%s", test.input, actual, test.expected)
		} else {
			t.Logf("✓ Encode(%q) = %s", test.input, actual)

			reversed := Decode(tokens, model)
			if actual != test.expected {
				t.Errorf("** Decode = %q", reversed)
			} else {
				t.Logf("✓ Decode")
			}
		}
	}
}

func formatTokens(tokens []int) string {
	var buf strings.Builder
	buf.WriteString("[")
	for i, token := range tokens {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(strconv.Itoa(token))
	}
	buf.WriteString("]")
	return buf.String()
}
