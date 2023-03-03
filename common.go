package openai

import (
	"fmt"
	"strings"
)

const (
	// ModelBestChat is the current best chat model, and should generally be the default one to use.
	ModelBestChat = ModelChatGPT35TurboCurrent

	// ModelBestTextInstruction is the current best instruction-following model for text.
	// Not recommended for basically anything any more.
	ModelBestTextInstruction = ModelTextDavinci003

	// ModelBestBase is the current best and largest fine-tunable model. Should never be used as is,
	// only meant for fine-tuning.
	ModelBestBase = ModelBaseDavinci

	ModelBaseDavinci           = "davinci"
	ModelTextDavinci003        = "text-davinci-003"
	ModelChatGPT35TurboCurrent = "gpt-3.5-turbo"
)

func MaxTokens(model string) int {
	switch model {
	case "ada", "babbage", "curie", ModelBaseDavinci, "text-ada-001", "text-babbage-001", "text-curie-001":
		return 2048
	case "code-davinci-002", "text-davinci-002":
		return 4000 // from docs: https://platform.openai.com/docs/models/gpt-3-5
	case "text-davinci-003":
		return 4097
	case ModelChatGPT35TurboCurrent, "gpt-3.5-turbo-0301":
		return 4096
	case "text-embedding-ada-002":
		return 8192
	default:
		if base, _, ok := strings.Cut(model, ":ft-"); ok {
			return MaxTokens(base)
		}
		panic(fmt.Errorf("unknown model name %q", model))
	}
}

// Price is an amount in 1/1_000_000 of a cent. I.e. $0.002 per 1K tokens = Price(200) per token.
type Price int64

// String formats the price as dollars and cents, e.g. $3.14.
func (p Price) String() string {
	return fmt.Sprintf("$%0.2f", float64(p)/100_000_000)
}

// Cost estimates the cost of processing the given number of tokens with the given model.
func Cost(tokens int, model string) Price {
	switch model {
	case ModelChatGPT35TurboCurrent, "gpt-3.5-turbo-0301":
		return Price(tokens) * 200
	case "davinci", "text-davinci-003":
		return Price(tokens) * 2000
	case "curie", "text-curie-001":
		return Price(tokens) * 200
	case "babbage", "text-babbage-001":
		return Price(tokens) * 50
	case "ada", "text-ada-001":
		return Price(tokens) * 40
	case "text-embedding-ada-002":
		return Price(tokens) * 40
	case "code-davinci-002", "text-davinci-002":
		return Price(tokens) * 2000 // just a guess; https://openai.com/pricing doesn't say anything
	default:
		if base, _, ok := strings.Cut(model, ":ft-"); ok {
			switch base {
			case "davinci":
				return Price(tokens) * 12000
			case "curie":
				return Price(tokens) * 1200
			case "babbage":
				return Price(tokens) * 240
			case "ada":
				return Price(tokens) * 160
			default:
				panic(fmt.Errorf("unknown base model name %q in %q", base, model))
			}
		}
		panic(fmt.Errorf("unknown model name %q", model))
	}
}

// FineTuningCost estimates the cost of fine-tuning the given model using the given number of tokens of sample data.
func FineTuningCost(tokens int, model string) Price {
	switch model {
	case "davinci":
		return Price(tokens) * 3000
	case "curie":
		return Price(tokens) * 300
	case "babbage":
		return Price(tokens) * 60
	case "ada":
		return Price(tokens) * 40
	default:
		if base, _, ok := strings.Cut(model, ":ft-"); ok {
			return FineTuningCost(tokens, base)
		}
		panic(fmt.Errorf("unknown model name %q", model))
	}
}

// Credentials are used to authenticate with OpenAI.
type Credentials struct {
	APIKey         string
	OrganizationID string
}

// Options adjust details of how Chat and Complete calls behave.
type Options struct {
	Model string `json:"model"`

	// MaxTokens is upper limit on completion length. In chat API, use 0 to allow the maximum possible length (4096 minus prompt length).
	MaxTokens int `json:"max_tokens,omitempty"`

	Temperature float64 `json:"temperature"`

	TopP float64 `json:"top_p"`

	// N determines how many choices to return for each prompt. Defaults to 1. Must be less or equal to BestOf if both are specified.
	N int `json:"n,omitempty"`

	// BestOf determines how many choices to create for each prompt. Defaults to 1. Must be greater or equal to N if both are specified.
	BestOf int `json:"best_of,omitempty"`

	// Stop is up to 4 sequences where the API will stop generating tokens. Response will not contain the stop sequence.
	Stop []string `json:"stop,omitempty"`

	// PresencePenalty number between 0 and 1 that penalizes tokens that have already appeared in the text so far.
	PresencePenalty float64 `json:"presence_penalty"`

	// FrequencyPenalty number between 0 and 1 that penalizes tokens on existing frequency in the text so far.
	FrequencyPenalty float64 `json:"frequency_penalty"`
}

type FinishReason string

const (
	FinishReasonStop   FinishReason = "stop"
	FinishReasonLength FinishReason = "length"
)

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
