package openai

import (
	"bytes"
	_ "embed"
	"log"
	"math"
	"strings"
	"sync"
	"unicode"
)

const (
	chatTokenOverhead       = 2
	chatTokenOverheadPerMsg = 5
)

// TokenCount counts GPT-3 tokens in the given text for the given model.
func TokenCount(text, model string) int {
	var result int
	EncodeEnum(text, model, func(token int) {
		result++
	})
	return result
}

func MsgTokenCount(msg Msg, model string) int {
	// We don't know the actual metaencoding, but it must be something similar.
	// Add a bit just to be sure.
	return TokenCount(msg.Content, model) + chatTokenOverheadPerMsg
}

func ChatTokenCount(msgs []Msg, model string) int {
	result := chatTokenOverhead
	for _, msg := range msgs {
		result += MsgTokenCount(msg, model)
	}
	return result
}

func Encode(text, model string) []int {
	var result []int
	EncodeEnum(text, model, func(token int) {
		result = append(result, token)
	})
	return result
}

func EncodeEnum(text, model string, f func(int)) {
	initEncoder()
	split(text, func(chunk string) {
		var tokens []string
		for _, b := range []byte(chunk) {
			tokens = append(tokens, string(byteEncoding[b]))
		}
		tokens = bpe(tokens)
		for _, token := range tokens {
			v, ok := tokenEncodings[token]
			if ok {
				f(v)
			} else {
				log.Printf("no encoding found for token %q", token)
			}
		}
	})
}

func Decode(tokens []int, model string) string {
	decoderOnce.Do(func() {
		initEncoder()
		tokenDecodings = make(map[int][]rune)
		for k, v := range tokenEncodings {
			tokenDecodings[v] = []rune(k)
		}
		byteDecoding = make(map[rune]byte)
		for b, s := range byteEncoding {
			byteDecoding[s] = byte(b)
		}
	})
	var buf strings.Builder
	for _, token := range tokens {
		runes, ok := tokenDecodings[token]
		if !ok {
			panic("no decoding found for token")
		}
		for _, r := range runes {
			buf.WriteByte(byteDecoding[r])
		}
	}
	return buf.String()
}

//go:embed tokenizer-tokens.bin
var rawTokens []byte

//go:embed tokenizer-bpe.bin
var rawMerges string

var (
	encoderOnce    sync.Once
	bpeMerges      []*bpeMerge
	bpeMergeIndex  map[bpePair]int
	tokenEncodings map[string]int
	byteEncoding   [256]rune
)

var (
	decoderOnce    sync.Once
	tokenDecodings map[int][]rune
	byteDecoding   map[rune]byte
)

func initEncoder() {
	encoderOnce.Do(func() {
		tokenEncodings = make(map[string]int)
		for i, token := range bytes.Split(rawTokens, []byte{0}) {
			if len(token) == 0 {
				continue
			}
			tokenEncodings[string(token)] = i
		}

		for _, line := range strings.FieldsFunc(rawMerges, isNewLine)[1:] {
			first, second, ok := strings.Cut(line, " ")
			if !ok {
				panic("invalid bpe line")
			}
			bpeMerges = append(bpeMerges, &bpeMerge{first, second, first + second})
		}
		bpeMergeIndex = make(map[bpePair]int)
		for i, merge := range bpeMerges {
			bpeMergeIndex[bpePair{merge.First, merge.Second}] = i
		}

		// TODO: this “UTF8 bytes to code points” encoding seems entirely pointless.
		// If that's correct, preprocess data files to undo this encoding, and then get rid of the code.
		for b := '!'; b <= '~'; b++ {
			byteEncoding[b] = rune(b)
		}
		for b := '¡'; b <= '¬'; b++ {
			byteEncoding[b] = rune(b)
		}
		for b := '®'; b <= 'ÿ'; b++ {
			byteEncoding[b] = rune(b)
		}
		var next rune = 256
		for b := 0; b <= 255; b++ {
			if byteEncoding[b] == 0 {
				byteEncoding[b] = next
				next++
			}
		}
	})
}

type bpeMerge struct {
	First  string
	Second string
	Result string
}

type bpePair struct {
	First  string
	Second string
}

// bpe merges consecutive pairs of tokens according to bpeMerges
// until no further merging is possible.
//
// Original code caches the results of this method, but is that really needed?
func bpe(tokens []string) []string {
	for {
		// log.Printf("bpe: % v", tokens)
		merge := findBestMerge(tokens)
		if merge == nil {
			break
		}
		// log.Printf("bpe merge %q + %q", merge.First, merge.Second)
		tokens = mergeAll(tokens, merge)
	}
	return tokens
}

// findBestMerge finds lowest-index bpeMerge where [..., merge.First, merge.Second, ...]
// occurs somewhere within the tokens.
func findBestMerge(tokens []string) *bpeMerge {
	n := len(tokens)
	var best = math.MaxInt
	for i := 1; i < n; i++ {
		i, ok := bpeMergeIndex[bpePair{tokens[i-1], tokens[i]}]
		if ok && i < best {
			best = i
		}
	}
	if best < math.MaxInt {
		return bpeMerges[best]
	}
	return nil
}

// mergeAll replaces all occurences of [..., merge.First, merge.Second, ...]
// with [..., merge.Result, ...]
func mergeAll(tokens []string, merge *bpeMerge) []string {
	n := len(tokens)
	dst := 0
	for src := 0; src < n; src++ {
		if tokens[src] == merge.First && src+1 < n && tokens[src+1] == merge.Second {
			tokens[dst] = merge.Result
			src++
		} else {
			tokens[dst] = tokens[src]
		}
		dst++
	}
	return tokens[:dst]
}

// split enumerates consecutive token candidates (chunks) in a string
// token candidates look like this:
//
//    - 's
//    - 't
//    - 're
//    - 've
//    - 'm
//    - 'll
//    - 'd
//    - space? letter+
//    - space? number+
//    - space? other+
//    - whitespace+
//
// where other is not a space, number or letter.
//
// In a run of consecutive whitespace, the trailing space (if any)
// becomes part of the next token if possible.

// 's|'t|'re|'ve|'m|'ll|'d| ?\pL+| ?\pN+| ?[^\s\pL\pN]+|\s+(?!\S)|\s+
func split(text string, f func(chunk string)) {
	var state splitState
	var start int

	flush := func(end int) {
		if end > start {
			f(text[start:end])
			start = end
		}
	}

	for pos, r := range text {
		isS, isL, isN := unicode.IsSpace(r), unicode.IsLetter(r), unicode.IsNumber(r)
		again := true
		for again {
			again = false
			switch state {
			case initial:
				if r == '\'' {
					state = afterApostrophe
				} else if r == ' ' {
					state = afterSpace
				} else if isS {
					state = inWhitespaceTokenAfterOtherWhitespace
				} else if isL {
					state = inLetterToken
				} else if isN {
					state = inNumberToken
				} else {
					state = inOtherToken
				}
			case afterApostrophe:
				if r == 's' || r == 't' || r == 'm' || r == 'd' {
					flush(pos + 1)
					state = initial
				} else if r == 'r' {
					state = afterApostropheNeedE
				} else if r == 'v' {
					state = afterApostropheNeedE
				} else if r == 'l' {
					state = afterApostropheNeedL
				} else {
					state, again = inOtherToken, true
				}
			case afterApostropheNeedE:
				if r == 'e' {
					flush(pos + 1)
					state = initial
				} else {
					state, again = inOtherToken, true
				}
			case afterApostropheNeedL:
				if r == 'l' {
					flush(pos + 1)
					state = initial
				} else {
					state, again = inOtherToken, true
				}
			case afterSpace:
				if r == ' ' {
					state = inWhitespaceTokenAfterSpace
				} else if isS {
					state = inWhitespaceTokenAfterOtherWhitespace
				} else if isL {
					state = inLetterToken
				} else if isN {
					state = inNumberToken
				} else {
					state = inOtherToken
				}
			case inWhitespaceTokenAfterOtherWhitespace:
				if r == ' ' {
					state = inWhitespaceTokenAfterSpace
				} else if !isS {
					flush(pos)
					state, again = initial, true
				}
			case inWhitespaceTokenAfterSpace:
				if r == ' ' {
					// nop
				} else if isS {
					state = inWhitespaceTokenAfterOtherWhitespace
				} else {
					flush(pos - 1) // the final space needs to be attributed to the next token
					state, again = initial, true
					again = true
				}
			case inLetterToken:
				if !isL {
					flush(pos)
					state, again = initial, true
				}
			case inNumberToken:
				if !isN {
					flush(pos)
					state, again = initial, true
				}
			case inOtherToken:
				if isS || isL || isN {
					flush(pos)
					state, again = initial, true
				}
			}
		}
	}
	flush(len(text))
}

type splitState int

const (
	initial = splitState(iota)
	afterApostrophe
	afterApostropheNeedE
	afterApostropheNeedL
	afterSpace
	inWhitespaceTokenAfterSpace
	inWhitespaceTokenAfterOtherWhitespace
	inLetterToken
	inNumberToken
	inOtherToken
)

func isNewLine(r rune) bool {
	return r == '\n'
}
