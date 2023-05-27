package openai

import (
	"context"
	"net/http"
)

func ComputeEmbedding(ctx context.Context, input string, client *http.Client, creds Credentials) ([]float64, Usage, error) {
	const callID = "ComputeEmbedding"

	req := &embeddingsRequest{
		Model: ModelEmbeddingAda002,
		Input: input,
	}

	var resp embeddingsResponse
	err := post(ctx, callID, "https://api.openai.com/v1/embeddings", client, creds, req, &resp)
	if err != nil {
		return nil, Usage{}, err
	}
	if len(resp.Data) == 0 {
		return nil, Usage{}, &Error{
			CallID:  callID,
			Message: "no results",
		}
	}

	return resp.Data[0].Embedding, resp.Usage, nil
}

type embeddingsRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type embeddingsResponse struct {
	Data  []embeddingsData `json:"data"`
	Usage Usage            `json:"usage"`
}

type embeddingsData struct {
	Embedding []float64 `json:"embedding"`
}
