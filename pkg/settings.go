package pkg

import (
	"fmt"
	"os"
	"sync"

	"github.com/joho/godotenv"
)

type Settings struct {
	Endpoint string `json:"endpoint"`
	APIKey   string `json:"api_key"`
	Model    string `json:"model"`
	Type     string `json:"type"` // Type of the model, e.g., "gpt-3.5-turbo"
}

var (
	settingsInstance *Settings
	once             sync.Once
)

func GetSettings() *Settings {
	once.Do(func() {
		godotenv.Load()
		settingsInstance = &Settings{}
		settingsInstance.Endpoint = os.Getenv("CAI_ENDPOINT")
		settingsInstance.APIKey = os.Getenv("CAI_KEY")
		settingsInstance.Model = os.Getenv("CAI_MODEL")
		settingsInstance.Type = os.Getenv("CAI_TYPE")

		if settingsInstance.Endpoint == "" {
			settingsInstance.Endpoint = "azure"
		}
		if settingsInstance.Endpoint == "" || settingsInstance.APIKey == "" || settingsInstance.Model == "" {
			fmt.Println("OpenAI settings are not properly configured. Please set ENDPOINT, KEY, and MODEL environment variables.")
			os.Exit(1)
		}
	})
	return settingsInstance
}
