Pragmatic OpenAI SDK for Go focused on ChatGPT
==============================================

Zero Dependencies • Tokenizer • Simple Code • Best way to use ChatGPT from Go

Install:

    go get github.com/andreyvit/openai

Run the example:

    export OPENAI_API_KEY=...
    go run github.com/andreyvit/openai/cmd/openai-example-bot

Batteries included:

* Use ChatGPT 3 & 4, `text-davinci-003` and fine-tuned models
* Use Embeddings API to add knowledge base excerpts (“context”) to your prompts
* Compute token count (plus a full tokenizer with encoding/decoding)
* Compute costs
* Utilities to trim history

Pragmatic:

* No dependencies
* No abstractions
* Under 1000 lines of code (you should read and understand the code of all your dependencies)
* Sensible error handling


Status
------

Used in production. Note: if you're looking to hire someone to build an AI bot, ping me at [andrey@tarantsov.com](mailto:andrey+chatgptbots@tarantsov.com).


Example Bot
-----------

```go
func main() {
    opt := openai.DefaultChatOptions()
    opt.MaxTokens = 768 // limit response length
    chat := []openai.Msg{
        openai.SystemMsg("You are a helpful assistant. Many would say you are way too cheerful and over the top. Answer concisely, adding jokes and exclamantions."),
    }
    client := &http.Client{
        Timeout: 2 * time.Minute, // sometimes this stuff takes a long time to respond
    }

    scanner := bufio.NewScanner(bufio.NewReader(os.Stdin))
    for scanner.Scan() {
        input := scanner.Text()

        // Limit input message length to 768 tokens
        if openai.TokenCount(input, opt.Model) > 768 {
            fmt.Printf("** message too long\n")
            continue
        }

        // Add message to history and truncate oldest messages if history no longer fits
        chat = append(chat, openai.UserMsg(input))
        chat, _ = openai.DropChatHistoryIfNeeded(chat, 1, openai.MaxTokens(opt.Model), opt.Model)

        msgs, usage, err := openai.Chat(context.Background(), chat, opt, client, creds)
        if err != nil {
            fmt.Printf("** %v\n", err)
            continue
        }

        fmt.Printf("\nChatGPT: %s\n\n", strings.TrimSpace(msgs[0].Content))
        chat = append(chat, msgs[0])
    }
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

- [ ] Rewrite based on code and binary data from [tiktoken](https://github.com/openai/tiktoken/tree/main/tiktoken)


MIT License
-----------

Copyright © 2023, Andrey Tarantsov. Distributed under the [MIT license](LICENSE).
