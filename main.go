package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/joho/godotenv"
	openai "github.com/sashabaranov/go-openai"
)

func getClient() (*openai.Client, error) {
	// Проверяем наличие API ключа
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY not found in .env file or environment variables")
	}

	// Конфигурация клиента
	config := openai.DefaultConfig(apiKey)

	// Добавляем кастомный endpoint если указан
	if baseURL := os.Getenv("OPENAI_API_BASE"); baseURL != "" {
		config.BaseURL = baseURL
	}

	return openai.NewClientWithConfig(config), nil
}

func getCommand(client *openai.Client, query, model string) (string, error) {
	prompt := fmt.Sprintf(`
    Я хочу выполнить действие в терминале. Ответь ТОЛЬКО командой без пояснений.
    Если команда опасная (может удалить данные или навредить системе), добавь предупреждение 'WARNING: ' перед командой.
    
    Запрос: %s
    Команда:`, query)

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are a helpful assistant that provides terminal commands.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			Temperature: 0.1,
		},
	)

	if err != nil {
		return "", err
	}

	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}

func main() {
	// Загружаем .env файл
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found or error loading it")
	}

	// Парсим аргументы командной строки
	modelFlag := flag.String("model", "", "Override the default model")
	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "Error: Query is required")
		fmt.Fprintln(os.Stderr, "Usage: cli-helper [--model MODEL] QUERY")
		os.Exit(1)
	}

	// Получаем полный запрос
	fullQuery := strings.Join(flag.Args(), " ")

	// Получаем клиент
	client, err := getClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Определяем модель
	model := os.Getenv("OPENAI_MODEL")
	if model == "" {
		model = "gpt-3.5-turbo"
	}
	if *modelFlag != "" {
		model = *modelFlag
	}

	// Получаем команду
	command, err := getCommand(client, fullQuery, model)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Выводим команду
	fmt.Println(command)
}
