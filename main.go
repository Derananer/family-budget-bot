package main

import (
	"log"
	"os"
	"path/filepath"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func main() {
	// Configure logging
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found or error loading it: %v", err)
	}

	// Check if debug mode is enabled
	debug := os.Getenv("DEBUG") == "true"
	if debug {
		log.Println("Debug mode enabled")
	}
	// Initialize bot
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_BOT_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = debug
	log.Printf("INFO: Bot initialized successfully. Authorized on account %s", bot.Self.UserName)

	// Create temporary directory for processing files
	var tmpDir string
	if debug {
		// Create directory in current project directory
		cwd, err := os.Getwd()
		if err != nil {
			log.Panic(err)
		}
		tmpDir = "tmp"
		tmpDir = filepath.Join(cwd, tmpDir)
		err = os.MkdirAll(tmpDir, 0755)
	} else {
		tmpDir, err = os.MkdirTemp(os.TempDir(), "bank_statements")
	}
	if err != nil {
		log.Panic(err)
	}
	log.Printf("INFO: Created temporary directory: %s", tmpDir)
	if !debug {
		defer os.RemoveAll(tmpDir)
	}

	// Initialize update configuration
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60
	log.Println("INFO: Starting bot update listener")

	// Start receiving updates
	updates := bot.GetUpdatesChan(updateConfig)

	// Handle updates
	for update := range updates {
		if update.Message == nil {
			continue
		}

		if update.Message.Document != nil {
			if debug {
				log.Printf("DEBUG: Received document from user %s (ID: %d)",
					update.Message.From.UserName,
					update.Message.From.ID)
			}
			go handleDocument(bot, update.Message, tmpDir)
		}
	}
}
