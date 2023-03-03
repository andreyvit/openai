package openai

// FitChatContext returns messages that fit into maxTokenCount, skipping those that don't
// fit. This is meant for including knowledge base context into ChatGPT prompts.
func FitChatContext(candidates []Msg, maxTokenCount int, model string) ([]Msg, int) {
	const minReasonableLen = 20
	used := 0
	remaining := maxTokenCount
	var result []Msg
	for _, msg := range candidates {
		n := MsgTokenCount(msg, model)
		if remaining >= n {
			result = append(result, msg)
			remaining -= n
			used += n
		}
		if remaining < minReasonableLen {
			break // don't waste time trying to fill a tiny hole
		}
	}
	return result, used
}

func DropChatHistoryIfNeeded(chat []Msg, fixedSuffixLen int, maxTokens int, model string) ([]Msg, int) {
	msgTokens := make([]int, len(chat))
	usedTokens := chatTokenOverhead
	for i, msg := range chat {
		c := MsgTokenCount(msg, model) // this is by far the slowest op here; cache result to avoid calling twice
		msgTokens[i] = c
		usedTokens += c
	}

	var dropCount int
	maxDropCount := len(chat) - fixedSuffixLen
	for usedTokens > maxTokens && dropCount < maxDropCount {
		i := fixedSuffixLen + dropCount // dropping message at this index
		usedTokens -= msgTokens[i]
		dropCount++
	}

	if dropCount > 0 {
		copy(chat[fixedSuffixLen:], chat[fixedSuffixLen+dropCount:])
		return chat[:len(chat)-dropCount], usedTokens
	} else {
		return chat, usedTokens
	}
}
