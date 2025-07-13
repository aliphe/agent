package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
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
	ctx := context.Background()
	config, err := agent.ParseConfig(ctx, "agent.json")
	if err != nil {
		slog.Info("parse config", "error", err)
	}
	geminiClient, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: os.Getenv("GEMINI_API_KEY"),
	})
	if err != nil {
		log.Panicf("load gemini client: %v", err)
	}
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./agent.db"
	}
	db, err := sqlx.Open("sqlite3", dbPath)
	if err != nil {
		log.Panicf("load database: %v", err)
	}
	defer db.Close()

	tools := []tool.Tool{
		tool.NewUserName(),
		tool.NewMath(),
		tool.NewSQL(db),
	}
	if config != nil {
		tools = append(tools, config.MCP.Tools()...)
	}

	toolBelt := tool.NewToolBelt(tools...)
	chatStore := store.NewChatStore(db)
	agent := agent.NewAgent(config, toolBelt, chatStore, llm.NewGemini(geminiClient))

	scanner := bufio.NewScanner(os.Stdin)
	slog.Info("Agent started. Type 'exit' to quit.")

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
