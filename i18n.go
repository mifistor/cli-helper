package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Локализатор для работы с переводами
type Localizer struct {
	translations map[string]map[string]string
	lang         string
}

// Создает новый локализатор
func NewLocalizer() (*Localizer, error) {
	// Определяем язык системы
	lang := getSystemLanguage()
	
	// Создаем локализатор
	l := &Localizer{
		translations: make(map[string]map[string]string),
		lang:         lang,
	}

	// Загружаем переводы
	err := l.loadTranslations()
	if err != nil {
		return nil, err
	}

	return l, nil
}

// Получает перевод по ключу
func (l *Localizer) Get(key string) string {
	// Проверяем наличие перевода для текущего языка
	if translations, ok := l.translations[l.lang]; ok {
		if translation, ok := translations[key]; ok {
			return translation
		}
	}

	// Если перевод не найден, пробуем английский
	if l.lang != "en" {
		if translations, ok := l.translations["en"]; ok {
			if translation, ok := translations[key]; ok {
				return translation
			}
		}
	}

	// Если перевод не найден, возвращаем ключ
	return key
}

// Получает форматированный перевод по ключу
func (l *Localizer) GetF(key string, args ...interface{}) string {
	return fmt.Sprintf(l.Get(key), args...)
}

// Загружает переводы из файлов
func (l *Localizer) loadTranslations() error {
	// Пути для поиска директории с локализациями
	var localesDirs []string

	// 1. Директория рядом с исполняемым файлом
	if execDir := getExecutableDir(); execDir != "" {
		localesDirs = append(localesDirs, filepath.Join(execDir, "locales"))
	}

	// 2. Директория в домашней директории пользователя
	if homeDir, err := os.UserHomeDir(); err == nil {
		localesDirs = append(localesDirs, filepath.Join(homeDir, ".cli-helper", "locales"))
	}

	// 3. Системная директория
	localesDirs = append(localesDirs, "/usr/local/share/cli-helper/locales")

	// 4. Текущая директория
	localesDirs = append(localesDirs, "locales")

	// Ищем директорию с локализациями
	var localesDir string
	for _, dir := range localesDirs {
		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			localesDir = dir
			break
		}
	}

	// Если директория не найдена
	if localesDir == "" {
		return fmt.Errorf("locales directory not found in any of the search paths")
	}

	// Получаем список файлов в директории
	files, err := os.ReadDir(localesDir)
	if err != nil {
		return err
	}

	// Загружаем переводы из каждого файла
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		// Получаем код языка из имени файла
		lang := strings.TrimSuffix(file.Name(), ".json")

		// Загружаем файл
		data, err := os.ReadFile(filepath.Join(localesDir, file.Name()))
		if err != nil {
			return err
		}

		// Парсим JSON
		var translations map[string]string
		err = json.Unmarshal(data, &translations)
		if err != nil {
			return err
		}

		// Сохраняем переводы
		l.translations[lang] = translations
	}

	return nil
}

// Получает язык системы
func getSystemLanguage() string {
	// Проверяем переменные окружения
	for _, envVar := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if value := os.Getenv(envVar); value != "" {
			// Формат: ru_RU.UTF-8
			parts := strings.Split(value, ".")
			if len(parts) > 0 {
				langParts := strings.Split(parts[0], "_")
				if len(langParts) > 0 {
					return strings.ToLower(langParts[0])
				}
			}
		}
	}

	// По умолчанию - английский
	return "en"
}

// Получает директорию исполняемого файла
func getExecutableDir() string {
	// Получаем путь к исполняемому файлу
	execPath, err := os.Executable()
	if err != nil {
		return ""
	}

	// Получаем директорию
	return filepath.Dir(execPath)
}
