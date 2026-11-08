Pragmatic OpenAI SDK for Go focused on ChatGPT
==============================================

Go client for Chat Completions, legacy Completions and Embeddings, with a tokenizer and streaming support.

Install:

    go get github.com/andreyvit/openai@latest

Run the example:

    export OPENAI_API_KEY=...
    go run github.com/andreyvit/openai/cmd/openai-example-bot

Batteries included:

* Call the implemented Chat Completions, Completions and Embeddings endpoints
* Use Embeddings API to add knowledge base excerpts (“context”) to your prompts
* Stream chat completions
* Estimate token counts with the bundled tokenizer (encoding/decoding included)
* Estimate costs using bundled model prices
* Utilities to trim history

Pragmatic:

* No dependencies
* No abstractions
* Sensible error handling


Status
------

The API surface and bundled model tables reflect the SDK’s original 2023-era scope. There is no Responses API implementation. Tokenization, context limits and prices are model-specific; verify the bundled tables before using these estimates for a newer model or billing. Model constants do not guarantee that an endpoint still serves that model.

Maintenance: this repository does not promise coverage of new OpenAI APIs or models. Check implementation and recent changes for your integration.

Note: if you're looking to hire someone to build an AI bot, ping me at [andrey@tarantsov.com](mailto:andrey+chatgptbots@tarantsov.com).


One request
-----------

This sends a paid API request.

```go
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/andreyvit/openai"
)

func main() {
	creds := openai.Credentials{APIKey: os.Getenv("OPENAI_API_KEY")}
	opt := openai.DefaultChatOptions()
	// Set OPENAI_MODEL to an available Chat Completions model supported by this SDK.
	opt.Model = os.Getenv("OPENAI_MODEL")
	if creds.APIKey == "" || opt.Model == "" {
		panic("set OPENAI_API_KEY and OPENAI_MODEL")
	}
	opt.MaxTokens = 128
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	msgs, _, err := openai.Chat(ctx, []openai.Msg{openai.UserMsg("Hello!")}, opt,
		&http.Client{Timeout: time.Minute}, creds)
	if err != nil {
		panic(err)
	}
	if len(msgs) == 0 {
		panic("no response messages")
	}
	fmt.Println(msgs[0].Content)
}
```

See [`cmd/openai-example-bot/example.go`](cmd/openai-example-bot/example.go) for the full runnable code.


Contributing
------------

Contributions are welcome, but keep in mind that I want to keep this library short and focused.

Auto-testing via modd (`go install github.com/cortesi/modd/cmd/modd@latest`):

    modd

Updating binary files:

    src=~/Downloads/GPT-3-Encoder-master
    cp $src/vocab.bpe tokenizer-bpe.bin
    go run ./_convert-encoder-json-to-tokens-bin.go $src/encoder.json tokenizer-tokens.bin

TODO:

- [ ] Rewrite tokenizer based on code and binary data from [tiktoken](https://github.com/openai/tiktoken/tree/main/tiktoken)


MIT License
-----------

Copyright © 2023, Andrey Tarantsov. Distributed under the [MIT license](LICENSE).
