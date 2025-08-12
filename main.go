package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/joho/godotenv"
	"github.com/sergi/go-diff/diffmatchpatch"
)

var addedFiles []string
var chatHistory []ChatMessage
var isDiffOn = true
var undoHistory = make(map[string]string)

var fileTemplates = map[string]string{
	"python":     "def main():\n    pass\n\nif __name__ == \"__main__\":\n    main()",
	"html":       "<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n    <meta charset=\"UTF-8\">\n    <meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n    <title>Document</title>\n</head>\n<body>\n    \n</body>\n</html>",
	"javascript": "// Your JavaScript code here",
}

const (
	SYSTEM_PROMPT = "You are an incredible developer assistant. You have the following traits:\n- You write clean, efficient code\n- You explain concepts with clarity\n- You think through problems step-by-step\n- You're passionate about helping developers improve"
)

var (
	OPENROUTER_API_KEY string
	DEFAULT_MODEL      = "anthropic/claude-3.7-sonnet:thinking"
)

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type APIRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

type APIResponse struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
}

func init() {
	godotenv.Load()
	OPENROUTER_API_KEY = os.Getenv("OPENROUTER_API_KEY")
	chatHistory = append(chatHistory, ChatMessage{Role: "system", Content: SYSTEM_PROMPT})
}

func handleShowCommand(filepath string) {
	content, err := os.ReadFile(filepath)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return
	}
	fmt.Println(string(content))
}

func handleNewCommand(filepath string) {
	if _, err := os.Stat(filepath); err == nil {
		fmt.Printf("File already exists: %s\n", filepath)
		return
	}

	ext := strings.Split(filepath, ".")
	template := ""
	if len(ext) > 1 {
		template = fileTemplates[ext[len(ext)-1]]
	}

	err := os.WriteFile(filepath, []byte(template), 0644)
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		return
	}
	fmt.Printf("File created: %s\n", filepath)
}

func isTextFile(filepath string) (bool, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err.Error() != "EOF" {
		return false, err
	}
	buffer = buffer[:n]

	for _, b := range buffer {
		if b == 0 {
			return false, nil
		}
	}

	return true, nil
}

func handleAddCommand(path string) {
	info, err := os.Stat(path)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if info.IsDir() {
		files, err := os.ReadDir(path)
		if err != nil {
			fmt.Printf("Error reading directory: %v\n", err)
			return
		}

		for _, file := range files {
			filepath := fmt.Sprintf("%s/%s", path, file.Name())
			if !file.IsDir() {
				isText, err := isTextFile(filepath)
				if err == nil && isText {
					addedFiles = append(addedFiles, filepath)
					fmt.Printf("Added file: %s\n", filepath)
				}
			}
		}
	} else {
		isText, err := isTextFile(path)
		if err == nil && isText {
			addedFiles = append(addedFiles, path)
			fmt.Printf("Added file: %s\n", path)
		}
	}
}

func handleEditCommand(filepath string) {
	originalContent, err := os.ReadFile(filepath)
	if err != nil {
		fmt.Printf("Error reading file for undo history: %v\n", err)
	} else {
		undoHistory[filepath] = string(originalContent)
	}
	fmt.Printf("Editing file (placeholder): %s\n", filepath)
}

func handleDiffCommand() {
	isDiffOn = !isDiffOn
	if isDiffOn {
		fmt.Println("Diff is now on")
	} else {
		fmt.Println("Diff is now off")
	}
}

func displayDiff(text1, text2 string) {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(text1, text2, false)
	fmt.Println(dmp.DiffPrettyText(diffs))
}

func handleHistoryCommand() {
	for _, msg := range chatHistory {
		fmt.Printf("[%s]: %s\n", msg.Role, msg.Content)
	}
}

func handleSaveCommand(filename string) {
	data, err := json.MarshalIndent(chatHistory, "", "  ")
	if err != nil {
		fmt.Printf("Error marshalling history: %v\n", err)
		return
	}
	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		fmt.Printf("Error saving history: %v\n", err)
		return
	}
	fmt.Printf("History saved to %s\n", filename)
}

func handleLoadCommand(filename string) {
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error loading history: %v\n", err)
		return
	}
	err = json.Unmarshal(data, &chatHistory)
	if err != nil {
		fmt.Printf("Error unmarshalling history: %v\n", err)
		return
	}
	fmt.Printf("History loaded from %s\n", filename)
}

func handleUndoCommand(filepath string) {
	originalContent, ok := undoHistory[filepath]
	if !ok {
		fmt.Printf("No undo history for %s\n", filepath)
		return
	}
	err := os.WriteFile(filepath, []byte(originalContent), 0644)
	if err != nil {
		fmt.Printf("Error writing file for undo: %v\n", err)
		return
	}
	fmt.Printf("Undone changes for %s\n", filepath)
	delete(undoHistory, filepath)
}

type DDGResult struct {
	AbstractURL string `json:"AbstractURL"`
	Answer      string `json:"Answer"`
	AnswerType  string `json:"AnswerType"`
	Definition  string `json:"Definition"`
	Heading     string `json:"Heading"`
	Image       string `json:"Image"`
	Redirect    string `json:"Redirect"`
	RelatedTopics []struct {
		FirstURL string `json:"FirstURL"`
		Icon     struct {
			Height string `json:"Height"`
			URL    string `json:"URL"`
			Width  string `json:"Width"`
		} `json:"Icon"`
		Result string `json:"Result"`
		Text   string `json:"Text"`
	} `json:"RelatedTopics"`
	Results []struct {
		FirstURL string `json:"FirstURL"`
		Icon     struct {
			Height string `json:"Height"`
			URL    string `json:"URL"`
			Width  string `json:"Width"`
		} `json:"Icon"`
		Result string `json:"Result"`
		Text   string `json:"Text"`
	} `json:"Results"`
	Type string `json:"Type"`
}

func handleSearchCommand(query string) {
	resp, err := http.Get(fmt.Sprintf("https://api.duckduckgo.com/?q=%s&format=json", query))
	if err != nil {
		fmt.Printf("Error making search request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var result DDGResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("Error decoding search results: %v\n", err)
		return
	}

	if result.AbstractURL != "" {
		fmt.Printf("Abstract: %s\n", result.AbstractURL)
	}
	for i, topic := range result.RelatedTopics {
		if i > 5 {
			break
		}
		fmt.Printf("Result %d: %s\n", i+1, topic.Result)
	}
}

func isURL(str string) bool {
	return strings.HasPrefix(str, "http://") || strings.HasPrefix(str, "https://")
}

func encodeImage(filepath string) (string, error) {
	bytes, err := os.ReadFile(filepath)
	if err != nil {
		return "", err
	}
	return "data:image/jpeg;base64," + string(bytes), nil
}

func handleImageCommand(paths []string) {
	for _, path := range paths {
		if isURL(path) {
			// For now, we just print that we would handle the URL
			fmt.Printf("Handling image URL (placeholder): %s\n", path)
		} else {
			_, err := encodeImage(path)
			if err != nil {
				fmt.Printf("Error encoding image: %v\n", err)
				continue
			}
			// For now, we just print that we would handle the local image
			fmt.Printf("Handling local image (placeholder): %s\n", path)
		}
	}
}

func handleModelCommand() {
	fmt.Printf("Current model: %s\n", DEFAULT_MODEL)
}

func handleChangeModelCommand(newModel string) {
	DEFAULT_MODEL = newModel
	fmt.Printf("Model changed to: %s\n", DEFAULT_MODEL)
}

func getStreamingResponse() {
	apiRequest := APIRequest{
		Model:    DEFAULT_MODEL,
		Messages: chatHistory,
		Stream:   true,
	}

	jsonData, err := json.Marshal(apiRequest)
	if err != nil {
		fmt.Printf("Error marshalling request: %v\n", err)
		return
	}

	req, err := http.NewRequest("POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+OPENROUTER_API_KEY)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	reader := bufio.NewReader(resp.Body)
	fullResponse := ""
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break // EOF
		}

		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if strings.TrimSpace(data) == "[DONE]" {
				break
			}

			var apiResp APIResponse
			err := json.Unmarshal([]byte(data), &apiResp)
			if err == nil && len(apiResp.Choices) > 0 {
				content := apiResp.Choices[0].Delta.Content
				fmt.Print(content)
				fullResponse += content
			}
		}
	}
	fmt.Println()
	chatHistory = append(chatHistory, ChatMessage{Role: "assistant", Content: fullResponse})
}

func executor(in string) {
	in = strings.TrimSpace(in)
	if in == "" {
		return
	}

	parts := strings.Fields(in)
	command := parts[0]
	args := parts[1:]

	switch command {
	case "/help":
		printHelp()
	case "/exit":
		fmt.Println("Goodbye!")
		os.Exit(0)
	case "/clear":
		// This is a placeholder. In a real terminal, you'd clear the screen.
		// For now, we'll just print a message.
		fmt.Println("Screen cleared.")
	case "/show":
		if len(args) == 0 {
			fmt.Println("Usage: /show <filepath>")
			return
		}
		handleShowCommand(args[0])
	case "/new":
		if len(args) == 0 {
			fmt.Println("Usage: /new <filepath>")
			return
		}
		handleNewCommand(args[0])
	case "/add":
		if len(args) == 0 {
			fmt.Println("Usage: /add <path>")
			return
		}
		handleAddCommand(args[0])
	case "/edit":
		if len(args) == 0 {
			fmt.Println("Usage: /edit <filepath>")
			return
		}
		handleEditCommand(args[0])
	case "/diff":
		handleDiffCommand()
	case "/history":
		handleHistoryCommand()
	case "/save":
		if len(args) == 0 {
			fmt.Println("Usage: /save <filename>")
			return
		}
		handleSaveCommand(args[0])
	case "/load":
		if len(args) == 0 {
			fmt.Println("Usage: /load <filename>")
			return
		}
		handleLoadCommand(args[0])
	case "/undo":
		if len(args) == 0 {
			fmt.Println("Usage: /undo <filepath>")
			return
		}
		handleUndoCommand(args[0])
	case "/search":
		if len(args) == 0 {
			fmt.Println("Usage: /search <query>")
			return
		}
		handleSearchCommand(strings.Join(args, " "))
	case "/image":
		if len(args) == 0 {
			fmt.Println("Usage: /image <path/url...>")
			return
		}
		handleImageCommand(args)
	case "/model":
		handleModelCommand()
	case "/change_model":
		if len(args) == 0 {
			fmt.Println("Usage: /change_model <model_name>")
			return
		}
		handleChangeModelCommand(args[0])
	default:
		chatHistory = append(chatHistory, ChatMessage{Role: "user", Content: in})
		getStreamingResponse()
	}
}

func completer(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "/help", Description: "Show help message"},
		{Text: "/exit", Description: "Exit the application"},
		{Text: "/clear", Description: "Clear the screen"},
		{Text: "/add", Description: "Add files to AI's knowledge base"},
		{Text: "/edit", Description: "Edit existing files"},
		{Text: "/new", Description: "Create new files"},
		{Text: "/search", Description: "Perform a DuckDuckGo search"},
		{Text: "/image", Description: "Add image(s) to AI's knowledge base"},
		{Text: "/reset", Description: "Reset entire chat and file memory"},
		{Text: "/diff", Description: "Toggle display of diffs"},
		{Text: "/history", Description: "View chat history"},
		{Text: "/save", Description: "Save chat history to a file"},
		{Text: "/load", Description: "Load chat history from a file"},
		{Text: "/undo", Description: "Undo last edit for a specific file"},
		{Text: "/model", Description: "Show current AI model"},
		{Text: "/change_model", Description: "Change the AI model"},
		{Text: "/show", Description: "Show content of a file"},
	}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

func printHelp() {
	fmt.Println(`
Commands:
  /help          Show this help message
  /exit          Exit the application
  /clear         Clear the screen (placeholder)
  /add           Add files to AI's knowledge base
  /edit          Edit existing files
  /new           Create new files
  /search        Perform a DuckDuckGo search
  /image         Add image(s) to AI's knowledge base
  /reset         Reset entire chat and file memory
  /diff          Toggle display of diffs
  /history       View chat history
  /save          Save chat history to a file
  /load          Load chat history from a file
  /undo          Undo last edit for a specific file
  /model         Show current AI model
  /change_model  Change the AI model
  /show          Show content of a file
	`)
}

func main() {
	fmt.Println("Welcome to Omni Engineer (Go version)!")
	fmt.Println("------------------------------------")
	p := prompt.New(
		executor,
		completer,
		prompt.OptionPrefix(">>> "),
		prompt.OptionTitle("omni-engineer-go"),
	)
	p.Run()
}
