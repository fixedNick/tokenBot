package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	TelegramData string `yaml:"telegram_data"`
	Buy          struct {
		Common   float64 `yaml:"common"`
		Uncommon float64 `yaml:"uncommon"`
		Rare     float64 `yaml:"rare"`
		Epic     float64 `yaml:"epic"`
	} `yaml:"buy"`
	Sell struct {
		Common   float64 `yaml:"common"`
		Uncommon float64 `yaml:"uncommon"`
		Rare     float64 `yaml:"rare"`
		Epic     float64 `yaml:"epic"`
	} `yaml:"sell"`
}

func NewConfig(filePath string) *Config {
	yamlFile, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatalf("Ошибка при чтении файла: %v", err)
	}

	// Создание объекта конфигурации
	var config Config

	// Преобразование YAML в объект структуры
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatalf("Ошибка при разборе YAML: %v", err)
	}
	return &config
}
