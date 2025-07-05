package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aliphe/skipery/agent"
	"github.com/aliphe/skipery/llm"
	"github.com/aliphe/skipery/tool"
	"google.golang.org/genai"
)

func main() {
	geminiClient, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey: os.Getenv("GEMINI_API_KEY"),
	})
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	toolBelt := tool.NewToolBelt(tool.NewName())
	agent := agent.NewAgent(toolBelt, llm.NewGemini(geminiClient))

	ctx := context.Background()

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Skipery started. Type 'exit' to quit.")

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "exit" {
			break
		}

		if input == "" {
			continue
		}

		response, err := agent.SendMessage(ctx, input)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		fmt.Println(response)
	}
}
