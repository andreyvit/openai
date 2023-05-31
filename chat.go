package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Role string

const (
	System    Role = "system"
	User      Role = "user"
	Assistant Role = "assistant"
)

// Msg is a single chat message.
type Msg struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

// SystemMsg makes an Msg with a System role.
func SystemMsg(content string) Msg {
	return Msg{System, content}
}

// UserMsg makes an Msg with a User role.
func UserMsg(content string) Msg {
	return Msg{User, content}
}

// AssistantMsg makes an Msg with an Assistant role.
func AssistantMsg(content string) Msg {
	return Msg{Assistant, content}
}

// DefaultChatOptions provides a safe and conservative starting point for Chat call options.
// Note that it sets Temperature to 0 and enables unlimited MaxTokens.
func DefaultChatOptions() Options {
	return Options{
		Model:       ModelDefaultChat,
		MaxTokens:   0,
		Temperature: 0,
		TopP:        1.0,
		N:           0,
	}
}

// Chat suggests the next assistant's message for the given prompt via ChatGPT..
// When successful, always returns at least one Msg, more if you set opt.N
// (these are multiple choices for the next message, not multiple messages).
// Options should originate from DefaultChatOptions, not DefaultCompleteOptions.
func Chat(ctx context.Context, messages []Msg, opt Options, client *http.Client, creds Credentials) ([]Msg, Usage, error) {
	const callID = "Chat"

	req := &chatRequest{
		Msgs:    messages,
		Options: opt,
	}

	var resp chatResponse
	err := post(ctx, callID, "https://api.openai.com/v1/chat/completions", client, creds, req, &resp)
	if err != nil {
		return nil, Usage{}, err
	}
	if len(resp.Choices) == 0 {
		return nil, Usage{}, &Error{
			CallID:  callID,
			Message: "no results",
		}
	}

	result := make([]Msg, 0, len(resp.Choices))
	for _, choice := range resp.Choices {
		result = append(result, choice.Msg)
	}
	return result, resp.Usage, nil
}

// StreamChat suggests the next assistant's message for the given prompt
// via ChatGPT, streaming the response.
// Options should originate from DefaultChatOptions, not DefaultCompleteOptions.
// Options.N must be 0 or 1.
func StreamChat(ctx context.Context, messages []Msg, opt Options, client *http.Client, creds Credentials, f func(msg *Msg, delta string) error) (Msg, error) {
	const callID = "StreamChat"

	req := &chatRequest{
		Msgs:    messages,
		Options: opt,
		Stream:  true,
	}

	var msg Msg
	var buf strings.Builder
	err := post(ctx, callID, "https://api.openai.com/v1/chat/completions", client, creds, req, func(data []byte) error {
		var resp chatStreamingResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			return err
		}
		if len(resp.Choices) != 1 {
			return fmt.Errorf("expected exactly one choice")
		}
		delta := resp.Choices[0].Delta
		if delta.Role != "" {
			msg.Role = delta.Role
		}
		if delta.Content != "" {
			buf.WriteString(delta.Content)
			msg.Content = buf.String()
		}
		return f(&msg, delta.Content)
	})
	return msg, err
}

type chatRequest struct {
	Msgs []Msg `json:"messages"`
	Options
	Stream bool `json:"stream,omitempty"`
}

type message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	// ID      string       `json:"id"`
	// Object  string       `json:"object"`
	// Created int          `json:"created"`
	// Model   string       `json:"model"`
	Choices []chatChoice `json:"choices"`
	Usage   Usage        `json:"usage"`
}

type chatChoice struct {
	Msg          Msg    `json:"message"`
	Index        int    `json:"index"`
	FinishReason string `json:"finish_reason"`
}

type chatStreamingResponse struct {
	Choices []chatStreamingChoice `json:"choices"`
}

type chatStreamingChoice struct {
	Delta        Msg    `json:"delta"`
	FinishReason string `json:"finish_reason"`
}
