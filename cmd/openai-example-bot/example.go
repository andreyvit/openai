package main

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/andreyvit/openai"
)

func main() {
	creds := openai.Credentials{
		APIKey: os.Getenv("OPENAI_API_KEY"),
	}
	opt := openai.DefaultChatOptions()
	opt.MaxTokens = 768 // limit response length
	chat := []openai.Msg{
		openai.SystemMsg("You are a helpful assistant. Many would say you are way too cheerful and over the top. Answer concisely, adding jokes and exclamantions."),
	}
	client := &http.Client{
		Timeout: 2 * time.Minute, // sometimes this stuff takes a long time to respond
	}

	var cost openai.Price
	scanner := bufio.NewScanner(bufio.NewReader(os.Stdin))
	fmt.Printf("User: ")
	for scanner.Scan() {
		input := scanner.Text()
		if openai.TokenCount(input, opt.Model) > 768 {
			fmt.Printf("** message too long\n")
			continue
		}

		chat = append(chat, openai.UserMsg(input))
		chat, _ = openai.DropChatHistoryIfNeeded(chat, 1, openai.MaxTokens(opt.Model), opt.Model)

		msgs, usage, err := openai.Chat(context.Background(), chat, opt, client, creds)
		if err != nil {
			fmt.Printf("** %v\n", err)
			continue
		}
		cost += openai.Cost(usage.TotalTokens, opt.Model)

		fmt.Printf("\nChatGPT: %s\n\n", strings.TrimSpace(msgs[0].Content))

		chat = append(chat, msgs[0])
		fmt.Printf("[%v spent, history has %d tokens in %d messages]\nUser: ", cost, openai.ChatTokenCount(chat, opt.Model), len(chat))
	}
}
