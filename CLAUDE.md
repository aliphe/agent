# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Skipery is a command-line AI agent built in Go that provides an interactive chat interface with tool calling capabilities. The agent uses Google's Gemini AI model and persists conversations to a SQLite database.

## Key Architecture

- **Entry Point**: `cmd/term/main.go` - Terminal interface that handles user input/output
- **Agent Layer**: `agent/` - Core agent logic that orchestrates chat flow and tool execution
- **LLM Integration**: `llm/gemini.go` - Gemini AI model integration with function calling
- **Tool System**: `tool/` - Extensible tool framework for agent capabilities
- **Storage**: `db/` - SQLite-based chat persistence with migrations
- **Chat Types**: `agent/chat/` - Message and conversation data structures

## Core Components

### Agent Flow
The agent follows a recursive pattern in `agent/agent.go:90-116`:
1. Send message to LLM with available tools
2. If LLM returns function calls, execute them and recurse
3. Return final response to user

### Tool System
Tools implement the `Tool` interface with:
- `Functions()` - Returns available function definitions
- `Call()` - Executes function with parameters

Current tools:
- `Math` - Basic arithmetic operations (sum, subtract)
- `UserName` - Returns current system username

### Database Schema
SQLite tables in `db/migrations/000001_chat_tables.up.sql`:
- `chats` - Chat metadata (id, title, created_at)
- `messages` - Chat messages with JSON-serialized function calls/responses

## Common Commands

### Running the Application
```bash
go run cmd/term/main.go
```

### Database Migrations
```bash
# Apply migrations
make migrate-up

# Rollback migrations  
make migrate-down
```

### Environment Setup
Requires `GEMINI_API_KEY` environment variable for Google Gemini API access.

## Development Notes

### Adding New Tools
1. Create new tool in `tool/` directory implementing `Tool` interface
2. Define functions with JSON schema parameters
3. Add to `NewToolBelt()` in `cmd/term/main.go:36`

### Database Path
Currently hardcoded to `/Users/matthias/work/skipr/skipery/skipery.db` in `cmd/term/main.go:30`

### LLM Model
Uses `gemini-2.5-flash-lite-preview-06-17` model (see `llm/gemini.go:109`)