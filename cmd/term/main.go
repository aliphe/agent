package main

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/aliphe/skipery/agent"
	store "github.com/aliphe/skipery/db"
	"github.com/aliphe/skipery/llm"
	"github.com/aliphe/skipery/tool"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
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

	db, err := sqlx.Open("sqlite3", "/Users/matthias/work/skipr/skipery/skipery.db")
	if err != nil {
		slog.Error("error initializing database connection", "error", err)
	}
	defer db.Close()

	toolBelt := tool.NewToolBelt(tool.NewUserName(), tool.NewMath())
	chatStore := store.NewChatStore(db)
	agent := agent.NewAgent(toolBelt, chatStore, llm.NewGemini(geminiClient))

	ctx := context.Background()

	scanner := bufio.NewScanner(os.Stdin)
	slog.Info("Skipery started. Type 'exit' to quit.")

	chatID := uuid.New().String()

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

		response, err := agent.SendMessage(ctx, chatID, input)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		for _, msg := range response {
			switch {
			case msg.FunctionCalls != nil:
				for _, fc := range msg.FunctionCalls {
					fmt.Printf("[author: %s]: %s %+v\n", msg.Author, fc.Name, fc.Args)
				}
			case msg.FunctionResponses != nil:
				for name, fr := range msg.FunctionResponses {
					fmt.Printf("[author: %s]: %s: %+v\n", msg.Author, name, fr)
				}
			default:
				fmt.Printf("[author: %s]: %s\n", msg.Author, msg.Text)
			}
		}
	}
}
