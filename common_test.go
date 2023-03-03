package openai

import "testing"

func TestCost(t *testing.T) {
	const fine = "davinci:ft-12345"
	tests := []struct {
		input   int
		expChat string
		expText string
		expFine string
	}{
		{0, "$0.00", "$0.00", "$0.00"},
		{1000, "$0.00", "$0.02", "$0.12"},
		{1000000, "$2.00", "$20.00", "$120.00"},
	}
	for _, test := range tests {
		if a := Cost(test.input, ModelBestChat).String(); a != test.expChat {
			t.Errorf("** Cost(%d, %s) = %s, wanted %s", test.input, ModelBestChat, a, test.expChat)
		}
		if a := Cost(test.input, ModelBestTextInstruction).String(); a != test.expText {
			t.Errorf("** Cost(%d, %s) = %s, wanted %s", test.input, ModelBestTextInstruction, a, test.expText)
		}
		if a := Cost(test.input, fine).String(); a != test.expFine {
			t.Errorf("** Cost(%d, %s) = %s, wanted %s", test.input, fine, a, test.expFine)
		}
	}
}
