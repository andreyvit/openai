package openai

import "testing"

func TestCost(t *testing.T) {
	const fine = "davinci:ft-12345"
	tests := []struct {
		prompt   int
		compl    int
		model    string
		expected string
	}{
		{0, 0, ModelChatGPT35Turbo, "$0.00"},
		{0, 0, ModelChatGPT4, "$0.00"},
		{0, 0, ModelTextDavinci003, "$0.00"},
		{0, 0, fine, "$0.00"},

		{1000, 0, ModelChatGPT4, "$0.03"},
		{0, 1000, ModelChatGPT4, "$0.06"},
		{0, 1000, ModelTextDavinci003, "$0.02"},
		{0, 1000, fine, "$0.12"},
		{1000, 0, ModelTextDavinci003, "$0.02"},
		{1000, 0, fine, "$0.12"},

		// This isn't just a test, but also a useful reference table for these prices.

		// max GPT-4 32k context
		{32000, 768, ModelChatGPT4With32k, "$2.01"},
		{32512, 256, ModelChatGPT4With32k, "$1.98"},

		// max GPT-4 8k context
		{7423, 768, ModelChatGPT4, "$0.27"},
		{7936, 256, ModelChatGPT4, "$0.25"},
		{7423, 768, ModelChatGPT4With32k, "$0.54"},
		{7936, 256, ModelChatGPT4With32k, "$0.51"},

		// max GPT-3.5 context
		{3328, 768, ModelChatGPT4, "$0.15"},
		{3840, 256, ModelChatGPT4, "$0.13"},
		{3328, 768, ModelChatGPT35Turbo, "$0.01"},
		{3840, 256, ModelChatGPT35Turbo, "$0.01"},

		// max GPT-3.5 context x 100 messages
		{3328_00, 768_00, ModelChatGPT4, "$14.59"},
		{3840_00, 256_00, ModelChatGPT4, "$13.06"},
		{3328_00, 768_00, ModelChatGPT35Turbo, "$0.82"},
		{3840_00, 256_00, ModelChatGPT35Turbo, "$0.82"},

		// random large example
		{1_000_000, 100_000, ModelChatGPT4, "$36.00"},
		{1_000_000, 100_000, ModelChatGPT4With32k, "$72.00"},
		{1_000_000, 100_000, ModelChatGPT35Turbo, "$2.20"},
		{1_000_000, 100_000, fine, "$132.00"},

		// {0, 1000000, "$2.00", "$60.00", "$20.00", "$120.00"},
	}
	for _, test := range tests {
		if a := Cost(test.prompt, test.compl, test.model).String(); a != test.expected {
			t.Errorf("** Cost(%d, %d, %s) = %s, wanted %s", test.prompt, test.compl, test.model, a, test.expected)
		}
	}
}
