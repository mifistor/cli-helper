.PHONY: build install clean test run help example

# Имя бинарного файла
BINARY_NAME=cli-helper

# Директории
BUILD_DIR=build
EXAMPLE_DIR=example

# Go команды
GO=go
GOBUILD=$(GO) build
GOTEST=$(GO) test
GOMOD=$(GO) mod
GORUN=$(GO) run

# Цвета для вывода
GREEN=\033[0;32m
NC=\033[0m # No Color

# Основная цель - сборка
build:
	@echo "${GREEN}Сборка $(BINARY_NAME)...${NC}"
	@mkdir -p $(BUILD_DIR)/locales
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) -v
	@cp -r locales/* $(BUILD_DIR)/locales/
	@echo "${GREEN}Файлы локализации скопированы в $(BUILD_DIR)/locales/${NC}"

# Установка зависимостей
deps:
	@echo "${GREEN}Установка зависимостей...${NC}"
	$(GOMOD) tidy

# Установка бинарного файла в домашнюю директорию
install: build
	@echo "${GREEN}Установка $(BINARY_NAME) в ~/.$(BINARY_NAME)/...${NC}"
	@mkdir -p ~/.$(BINARY_NAME)/locales
	cp $(BUILD_DIR)/$(BINARY_NAME) ~/.$(BINARY_NAME)/
	cp -r locales/* ~/.$(BINARY_NAME)/locales/
	@if [ -f .env ]; then \
		cp .env ~/.$(BINARY_NAME)/; \
		echo "${GREEN}Файл .env скопирован в ~/.$(BINARY_NAME)/${NC}"; \
	else \
		echo "${GREEN}Файл .env не найден, создайте его в ~/.$(BINARY_NAME)/${NC}"; \
	fi
	@echo "${GREEN}Файлы локализации установлены в ~/.$(BINARY_NAME)/locales/${NC}"
	@echo "${GREEN}Добавьте ~/.$(BINARY_NAME) в PATH для использования $(BINARY_NAME) из любой директории${NC}"

# Установка бинарного файла в систему (требует sudo)
install-system: build
	@echo "${GREEN}Установка $(BINARY_NAME) в /usr/local/bin...${NC}"
	@mkdir -p /usr/local/share/$(BINARY_NAME)/locales
	cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	cp -r locales/* /usr/local/share/$(BINARY_NAME)/locales/
	@if [ -f .env ]; then \
		cp .env /usr/local/share/$(BINARY_NAME)/; \
		echo "${GREEN}Файл .env скопирован в /usr/local/share/$(BINARY_NAME)/${NC}"; \
	else \
		echo "${GREEN}Файл .env не найден, создайте его в /usr/local/share/$(BINARY_NAME)/${NC}"; \
	fi
	@echo "${GREEN}Файлы локализации установлены в /usr/local/share/$(BINARY_NAME)/locales/${NC}"

# Очистка сборки
clean:
	@echo "${GREEN}Очистка...${NC}"
	@rm -rf $(BUILD_DIR)

# Запуск тестов
test:
	@echo "${GREEN}Запуск тестов...${NC}"
	$(GOTEST) -v ./...

# Запуск примера
example:
	@echo "${GREEN}Запуск примера...${NC}"
	cd $(EXAMPLE_DIR) && $(GORUN) help.go

# Запуск основной программы
run:
	@echo "${GREEN}Запуск $(BINARY_NAME)...${NC}"
	$(GORUN) . $(COPY) $(ARGS)

# Справка
help:
	@echo "Доступные команды:"
	@echo "  make build         - Сборка бинарного файла"
	@echo "  make deps          - Установка зависимостей"
	@echo "  make install       - Установка бинарного файла в ~/.$(BINARY_NAME)/"
	@echo "  make install-system - Установка бинарного файла в /usr/local/bin (требует sudo)"
	@echo "  make clean         - Очистка сборки"
	@echo "  make test          - Запуск тестов"
	@echo "  make example       - Запуск примера"
	@echo "  make run ARGS=\"ваш запрос\" - Запуск основной программы с аргументами"
	@echo "  make run COPY=\"--no-copy\" ARGS=\"ваш запрос\" - Запуск с отключением копирования в буфер обмена"
	@echo "  make help          - Показать эту справку"

# По умолчанию
default: help
