package main

import (
	"context"
	"fmt"
	"os"

	openai "github.com/sashabaranov/go-openai"
)

func main() {
	// Создаем клиент OpenAI
	client := openai.NewClient(os.Getenv("SOY_TOKEN"))
	client.BaseURL = "http://api.eliza.yandex.net/raw/openai/v1"

	// Отправляем запрос
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: "gpt-4o",
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "Hello, world!",
				},
			},
		},
	)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Выводим ответ
	fmt.Println(resp.Choices[0].Message)
}
