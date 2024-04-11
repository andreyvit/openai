package openai

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	// ModelChatGPT4Turbo is the current best chat model, gpt-4-turbo, with 128k context.
	ModelChatGPT4Turbo = "gpt-4-turbo"

	// ModelChatGPT4TurboPreview is the preview of GPT-4 Turbo, with 128k context.
	ModelChatGPT4TurboPreview = "gpt-4-turbo-preview"

	// ModelChatGPT4 is the current best chat model, gpt-4, with 8k context.
	ModelChatGPT4 = "gpt-4"

	// ModelChatGPT4With32k is a version of ModelChatGPT4 with a 32k context.
	ModelChatGPT4With32k = "gpt-4-32k"

	// ModelChatGPT35Turbo is the current best, cheapest and universally available ChatGPT 3.5 model.
	ModelChatGPT35Turbo = "gpt-3.5-turbo"

	// ModelDefaultChat is a chat model used by default. This is going to be set to whatever
	// default choice the author of this library feels appropriate going forward, but really,
	// you should be specifying a specific model like ModelChatGPT4 or ModelChatGPT35Turbo.
	ModelDefaultChat = ModelChatGPT4Turbo

	// ModelDefaultCompletion is the current best instruction-following model for text completion.
	// Not recommended for basically anything any more because gpt-3.5-turbo is 10x cheaper and just as good.
	ModelDefaultCompletion = ModelTextDavinci003

	// ModelTextDavinci003 is the current best instruction-following model.
	// Not recommended for basically anything any more because gpt-3.5-turbo is 10x cheaper and just as good.
	ModelTextDavinci003 = "text-davinci-003"

	// ModelBaseDavinci is an older GPT 3 (not GPT 3.5) family base model. Only useful for fine-tuning.
	ModelBaseDavinci = "davinci"

	// ModelEmbedding3Large is the best embedding model so far.
	ModelEmbedding3Large = "text-embedding-3-large"

	// ModelEmbedding3Small is a better version of ModelEmbeddingAda002.
	ModelEmbedding3Small = "text-embedding-3-small"

	// ModelEmbeddingAda002 is the original embedding model, its use is no longer recommended.
	ModelEmbeddingAda002 = "text-embedding-ada-002"

	modelChatGPT35TurboPrefix  = "gpt-3.5-turbo-"
	modelChatGPT4Prefix        = "gpt-4-"
	modelChatGPT4With32kPrefix = "gpt-4-32k-"
)

var (
	gpt4TurboSnapshotRe   = regexp.MustCompile(`^gpt-4-turbo-\d{4}-\d{2}-\d{2}$`)
	gpt35TurboSnapshotRe  = regexp.MustCompile(`^gpt-3.5-turbo-\d{4}$`)
	gpt4With32kSnapshotRe = regexp.MustCompile(`^gpt-4-32k-\d{4}$`)
	gpt4SnapshotRe        = regexp.MustCompile(`^gpt-4-\d{4}$`)
)

// MaxTokens returns the maximum number of tokens the given model supports. This is a sum of
// prompt and completion tokens.
func MaxTokens(model string) int {
	switch model {
	case "ada", "babbage", "curie", ModelBaseDavinci, "text-ada-001", "text-babbage-001", "text-curie-001":
		return 2048
	case "code-davinci-002", "text-davinci-002":
		return 4000 // from docs: https://platform.openai.com/docs/models/gpt-3-5
	case "text-davinci-003":
		return 4097
	case ModelChatGPT35Turbo:
		return 4096
	case ModelChatGPT4:
		return 8192
	case ModelChatGPT4With32k:
		return 32768
	case ModelChatGPT4Turbo, ModelChatGPT4TurboPreview, "gpt-4-1106-preview", "gpt-4-0125-preview":
		return 128000
	case ModelEmbeddingAda002, ModelEmbedding3Small, ModelEmbedding3Large:
		return 8192
	default:
		if base, _, ok := strings.Cut(model, ":ft-"); ok {
			return MaxTokens(base)
		}
		if generic := snapshotToGeneric(model); generic != "" {
			return MaxTokens(generic)
		}
		panic(fmt.Errorf("unknown model name %q", model))
	}
}

// Price is an amount in 1/1_000_000 of a cent. I.e. $2 per 1M tokens = $0.002 per 1K tokens = Price(200) per token.
type Price int64

// String formats the price as dollars and cents, e.g. $3.14.
func (p Price) String() string {
	return fmt.Sprintf("$%0.2f", float64(p)/100_000_000)
}

// Cost estimates the cost of processing the given number of prompt & completion
// tokens with the given model.
func Cost(promptTokens, completionTokens int, model string) Price {
	switch model {
	case ModelChatGPT4TurboPreview, "gpt-4-1106-preview", "gpt-4-0125-preview":
		return Price(promptTokens)*1000 + Price(completionTokens)*3000
	case ModelChatGPT4:
		return Price(promptTokens)*3000 + Price(completionTokens)*6000
	case ModelChatGPT4With32k:
		return Price(promptTokens)*6000 + Price(completionTokens)*12000
	case ModelChatGPT35Turbo:
		return Price(promptTokens)*50 + Price(completionTokens)*150
	case "davinci", "text-davinci-003":
		return Price(promptTokens+completionTokens) * 2000
	case "curie", "text-curie-001":
		return Price(promptTokens+completionTokens) * 200
	case "babbage", "text-babbage-001":
		return Price(promptTokens+completionTokens) * 50
	case "ada", "text-ada-001":
		return Price(promptTokens+completionTokens) * 40
	case ModelEmbeddingAda002:
		return Price(promptTokens+completionTokens) * 10
	case ModelEmbedding3Small:
		return Price(promptTokens+completionTokens) * 2
	case ModelEmbedding3Large:
		return Price(promptTokens+completionTokens) * 13
	case "code-davinci-002", "text-davinci-002":
		return Price(promptTokens+completionTokens) * 2000 // just a guess; https://openai.com/pricing doesn't say anything
	default:
		if base, _, ok := strings.Cut(model, ":ft-"); ok {
			switch base {
			case "davinci":
				return Price(promptTokens+completionTokens) * 12000
			case "curie":
				return Price(promptTokens+completionTokens) * 1200
			case "babbage":
				return Price(promptTokens+completionTokens) * 240
			case "ada":
				return Price(promptTokens+completionTokens) * 160
			default:
				panic(fmt.Errorf("unknown base model name %q in %q", base, model))
			}
		}
		if generic := snapshotToGeneric(model); generic != "" {
			return Cost(promptTokens, completionTokens, model)
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

func snapshotToGeneric(model string) string {
	if gpt4TurboSnapshotRe.MatchString(model) {
		return ModelChatGPT4Turbo
	}
	if gpt35TurboSnapshotRe.MatchString(model) {
		return ModelChatGPT35Turbo
	}
	if gpt4With32kSnapshotRe.MatchString(model) {
		return ModelChatGPT4With32k
	}
	if gpt4SnapshotRe.MatchString(model) {
		return ModelChatGPT4
	}
	return ""
}

// Credentials are used to authenticate with OpenAI.
type Credentials struct {
	APIKey         string
	OrganizationID string
}

// Options adjust details of how Chat and Complete calls behave.
type Options struct {
	// Model is the OpenAI model to use, see https://platform.openai.com/docs/models/.
	//
	Model string `json:"model"`

	// MaxTokens is upper limit on completion length. In chat API, use 0 to allow the maximum possible length (4096 minus prompt length).
	MaxTokens int `json:"max_tokens,omitempty"`

	Functions        []any `json:"functions,omitempty"`
	FunctionCallMode any   `json:"function_call,omitempty"`
	Tools            []any `json:"tools,omitempty"`
	ToolChoice       any   `json:"tool_choice,omitempty"`

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

// ForceFunctionCall is a value to use in Options.FunctionCallMode.
type ForceFunctionCall struct {
	Name string `json:"name"`
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
