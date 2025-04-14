package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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
	
	// Удаляем префикс "bash" если он есть
	if strings.HasPrefix(cleanText, "bash") {
		cleanText = strings.TrimPrefix(cleanText, "bash")
	}
	
	// Удаляем префикс "WARNING: " если он есть
	if strings.HasPrefix(cleanText, "WARNING: ") {
		cleanText = strings.TrimPrefix(cleanText, "WARNING: ")
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

// Загружает файл .env из различных мест
func loadEnvFile() error {
	// Пути для поиска файла .env
	var envPaths []string

	// 1. Текущая директория
	envPaths = append(envPaths, ".env")

	// 2. Директория рядом с исполняемым файлом
	if execDir := getExecutableDir(); execDir != "" {
		envPaths = append(envPaths, filepath.Join(execDir, ".env"))
	}

	// 3. Домашняя директория пользователя
	if homeDir, err := os.UserHomeDir(); err == nil {
		envPaths = append(envPaths, filepath.Join(homeDir, ".cli-helper", ".env"))
	}

	// Пробуем загрузить файл .env из каждого пути
	for _, path := range envPaths {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			return godotenv.Load(path)
		}
	}

	// Если файл не найден, возвращаем ошибку
	return fmt.Errorf(".env file not found in any of the search paths")
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
	if err := loadEnvFile(); err != nil {
		log.Println("Warning:", err)
	}

	// Парсим аргументы командной строки
	modelFlag := flag.String("model", "", "Override the default model")
	noClipboardFlag := flag.Bool("no-copy", false, "Disable automatic copying to clipboard")
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

	// Удаляем символы ``` из команды
	cleanCommand := command
	if strings.HasPrefix(cleanCommand, "```") && strings.HasSuffix(cleanCommand, "```") {
		cleanCommand = cleanCommand[3 : len(cleanCommand)-3]
	} else {
		// Удаляем только в начале или в конце, если есть
		if strings.HasPrefix(cleanCommand, "```") {
			cleanCommand = cleanCommand[3:]
		}
		if strings.HasSuffix(cleanCommand, "```") {
			cleanCommand = cleanCommand[:len(cleanCommand)-3]
		}
	}
	
	// Удаляем префикс "bash" если он есть
	if strings.HasPrefix(cleanCommand, "bash") {
		cleanCommand = strings.TrimPrefix(cleanCommand, "bash")
	}
	
	// Удаляем префикс "WARNING: " если он есть
	if strings.HasPrefix(cleanCommand, "WARNING: ") {
		cleanCommand = strings.TrimPrefix(cleanCommand, "WARNING: ")
	}
	
	// Удаляем лишние пробелы в начале и конце
	cleanCommand = strings.TrimSpace(cleanCommand)

	// Выводим команду
	fmt.Println(cleanCommand)

	// Копируем в буфер обмена, если не указан флаг --no-copy
	if !*noClipboardFlag {
		if err := copyToClipboard(cleanCommand); err != nil {
			fmt.Fprintf(os.Stderr, i18n.GetF("clipboard_error", err))
		} else {
			fmt.Println(i18n.Get("clipboard_success"))
		}
	}
}
