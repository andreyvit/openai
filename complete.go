package openai

import (
	"context"
	"net/http"
)

// DefaultCompleteOptions provides a safe and conservative starting point for Complete call options.
// Note that it sets Temperature to 0 and MaxTokens to 256.
func DefaultCompleteOptions() Options {
	return Options{
		Model:       ModelBestTextInstruction,
		MaxTokens:   256,
		Temperature: 0,
		TopP:        1.0,
		N:           0,
	}
}

// Complete generates a completion for the given prompt using a non-chat model.
// This is mainly of interest when using fine-tuned models now.
// When successful, always returns at least one Completion; more if you set opt.N.
// Options should originate from DefaultCompleteOptions, not DefaultChatOptions.
func Complete(ctx context.Context, prompt string, opt Options, client *http.Client, creds Credentials) ([]Completion, Usage, error) {
	const callID = "Complete"

	req := &completionRequest{
		Prompt:  []string{prompt},
		Options: opt,
	}

	var resp completionResponse
	err := post(ctx, callID, "https://api.openai.com/v1/completions", client, creds, req, &resp)
	if err != nil {
		return nil, Usage{}, err
	}
	if len(resp.Choices) == 0 {
		return nil, resp.Usage, &Error{
			CallID:  callID,
			Message: "no results",
		}
	}

	result := make([]Completion, 0, len(resp.Choices))
	for _, choice := range resp.Choices {
		result = append(result, Completion{
			Text:         choice.Text,
			FinishReason: choice.FinishReason,
		})
	}
	return result, resp.Usage, nil
}

type Completion struct {
	Text         string       `json:"text"`
	FinishReason FinishReason `json:"finish_reason"`
}

type completionRequest struct {
	Prompt []string `json:"prompt"`
	Options
	Stream bool `json:"stream,omitempty"`
}

type completionResponse struct {
	ID      string             `json:"id"`
	Object  string             `json:"object"`
	Created int                `json:"created"`
	Model   string             `json:"model"`
	Choices []completionChoice `json:"choices"`
	Usage   Usage              `json:"usage"`
}

type completionChoice struct {
	Text         string         `json:"text"`
	Index        int            `json:"index"`
	LogProbs     *logprobResult `json:"logprobs"`
	FinishReason FinishReason   `json:"finish_reason"`
}

type logprobResult struct {
	Tokens        []string             `json:"tokens"`
	TokenLogprobs []float64            `json:"token_logprobs"`
	TopLogprobs   []map[string]float64 `json:"top_logprobs"`
	TextOffset    []int                `json:"text_offset"`
}
