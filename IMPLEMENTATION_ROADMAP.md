# Omni Engineer - Go Rewrite Implementation Roadmap

This document outlines the plan for rewriting the Omni Engineer Python application into Go. The goal is to create a more performant, concurrent, and maintainable version of the application while retaining all of its core features.

## 1. Analysis of Current Python Application

The current application is a single-file Python script (`main.py`) that provides a command-line interface for interacting with an AI assistant. It supports a variety of features, including:

- **AI-Powered Responses:** Streaming output from various AI models via OpenRouter.
- **File Management:** Adding, editing, creating, and showing file content.
- **Multi-File Editing:** Support for editing multiple files at once.
- **Web Searching:** Integration with DuckDuckGo Search.
- **Image Processing:** Handling of local and URL-based images.
- **Session Management:** Saving and loading chat history.
- **Undo Functionality:** Reverting file edits.
- **Syntax Highlighting:** For code display.
- **Diff Display:** Showing changes made to files.
- **Model Switching:** Changing the AI model on the fly.

## 2. Proposed Go Project Structure

To ensure maintainability and scalability, the Go application will be structured into multiple packages and files:

```
omni-engineer-go/
├── main.go                 # Entry point of the application
├── go.mod                  # Go module definition
├── go.sum                  # Go module checksums
├── cmd/                    # Command-line interface and command handling
│   ├── root.go             # Root command and CLI setup
│   ├── add.go
│   ├── edit.go
│   ├── new.go
│   └── ...                 # Other command files
├── internal/               # Internal application logic
│   ├── api/                # API clients (OpenRouter, DuckDuckGo)
│   │   └── openrouter.go
│   ├── fileutil/           # File handling utilities
│   │   └── file.go
│   ├── imageutil/          # Image processing utilities
│   │   └── image.go
│   └── session/            # Session management
│       └── history.go
└── pkg/                    # Reusable packages (if any)
    └── ...
```

## 3. Phased Implementation Plan

The rewrite will be done in phases to ensure a smooth transition and incremental progress.

### Phase 1: Core Functionality (Current Goal)

This phase focuses on implementing the essential features of the application.

- **CLI Setup:**
  - Implement a command-line interface using a library like `github.com/c-bata/go-prompt`.
  - Set up command handling for all the commands available in the Python version.
- **File Handling:**
  - Implement `/add`, `/edit`, `/new`, and `/show` commands.
  - Support for both single files and directories.
- **AI Integration:**
  - Integrate with the OpenRouter API using `net/http` or a suitable client library.
  - Implement streaming responses.
  - Implement `/model` and `/change_model` commands.
- **Search and Image Handling:**
  - Implement `/search` using the DuckDuckGo Search API.
  - Implement `/image` for local and remote images.
- **Advanced Features:**
  - Implement `/diff` to show file changes.
  - Implement `/history`, `/save`, `/load`, and `/undo` for session management.

### Phase 2: Concurrency and Performance Improvements

- **Concurrent Operations:**
  - Implement concurrent file processing for `/add` and `/edit` commands.
  - Use goroutines and channels to handle multiple API requests concurrently.
- **Performance Optimization:**
  - Profile the application and identify performance bottlenecks.
  - Optimize file I/O and API interactions.

### Phase 3: Testing and Refinement

- **Unit Tests:**
  - Write comprehensive unit tests for all packages.
  - Aim for high test coverage.
- **Integration Tests:**
  - Write integration tests to ensure all components work together correctly.
- **Refinement:**
  - Refine the user interface and user experience based on feedback.
  - Improve error handling and logging.

## 4. Go Libraries and Dependencies

The following Go libraries will be considered for the rewrite:

- **CLI:** `github.com/c-bata/go-prompt`
- **Color Output:** `github.com/fatih/color`
- **Tables:** `github.com/olekukonko/tablewriter`
- **Syntax Highlighting:** `github.com/alecthomas/chroma`
- **OpenAI/OpenRouter Client:** `github.com/sashabaranov/go-openai` or custom `net/http` implementation.
- **Dotenv:** `github.com/joho/godotenv`
- **Image Processing:** Standard `image` package.
- **Diff:** Standard library or a third-party library for better unified diffs.
- **DuckDuckGo Search:** Custom `net/http` implementation or a third-party library.
