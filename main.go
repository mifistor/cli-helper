package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/joho/godotenv"
	openai "github.com/sashabaranov/go-openai"
)

// Глобальный локализатор
var i18n *Localizer

func getClient() (*openai.Client, error) {
	// Проверяем наличие API ключа
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf(i18n.Get("api_key_error"))
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
	// Получаем шаблон промта из переменной окружения или используем значение по умолчанию
	promptTemplate := os.Getenv("PROMPT_TEMPLATE")
	if promptTemplate == "" {
		promptTemplate = i18n.Get("default_prompt_template")
	}

	// Определяем текущую операционную систему
	currentOS := runtime.GOOS
	switch currentOS {
	case "darwin":
		currentOS = "macOS"
	case "windows":
		currentOS = "Windows"
	case "linux":
		currentOS = "Linux"
	}

	// Форматируем промт с текущей ОС и запросом
	prompt := fmt.Sprintf(promptTemplate, currentOS, query)

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: i18n.Get("system_prompt"),
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

// Копирует текст в буфер обмена
func copyToClipboard(text string) error {
	// Удаляем символы ``` из команды, если они есть
	cleanText := text
	if strings.HasPrefix(cleanText, "```") && strings.HasSuffix(cleanText, "```") {
		cleanText = cleanText[3 : len(cleanText)-3]
	} else {
		// Удаляем только в начале или в конце, если есть
		if strings.HasPrefix(cleanText, "```") {
			cleanText = cleanText[3:]
		}
		if strings.HasSuffix(cleanText, "```") {
			cleanText = cleanText[:len(cleanText)-3]
		}
	}
	
	// Удаляем лишние пробелы в начале и конце
	cleanText = strings.TrimSpace(cleanText)

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin": // macOS
		cmd = exec.Command("pbcopy")
	case "linux":
		// Проверяем наличие xclip
		if _, err := exec.LookPath("xclip"); err == nil {
			cmd = exec.Command("xclip", "-selection", "clipboard")
		} else if _, err := exec.LookPath("xsel"); err == nil {
			// Если xclip не найден, пробуем xsel
			cmd = exec.Command("xsel", "--clipboard", "--input")
		} else {
			return fmt.Errorf(i18n.Get("clipboard_tools_not_found"))
		}
	default:
		return fmt.Errorf(i18n.GetF("clipboard_not_supported", runtime.GOOS))
	}

	cmd.Stdin = strings.NewReader(cleanText)
	return cmd.Run()
}

func main() {
	// Инициализируем локализатор
	var err error
	i18n, err = NewLocalizer()
	if err != nil {
		// Если не удалось инициализировать локализатор, используем простые сообщения на английском
		log.Printf("Warning: Failed to initialize localizer: %v", err)
		i18n = &Localizer{
			translations: make(map[string]map[string]string),
			lang:         "en",
		}
	}

	// Загружаем .env файл
	if err := godotenv.Load(); err != nil {
		log.Println(i18n.Get("env_warning"))
	}

	// Парсим аргументы командной строки
	modelFlag := flag.String("model", "", "Override the default model")
	clipboardFlag := flag.Bool("copy", false, "Copy the result to clipboard")
	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, i18n.Get("usage_error"))
		fmt.Fprintln(os.Stderr, i18n.Get("usage_help"))
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

	// Копируем в буфер обмена, если указан флаг
	if *clipboardFlag {
		if err := copyToClipboard(command); err != nil {
			fmt.Fprintf(os.Stderr, i18n.GetF("clipboard_error", err))
		} else {
			fmt.Println(i18n.Get("clipboard_success"))
		}
	}
}
