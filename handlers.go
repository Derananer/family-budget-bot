package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func handleDocument(bot *tgbotapi.BotAPI, message *tgbotapi.Message, tmpDir string) {
	debug := os.Getenv("DEBUG") == "true"
	log.Printf("INFO: Processing document from user %s (ID: %d)", message.From.UserName, message.From.ID)

	// Send processing message
	reply := tgbotapi.NewMessage(message.Chat.ID, "Processing your bank statement...")
	bot.Send(reply)

	// Download the file
	fileID := message.Document.FileID
	if debug {
		log.Printf("DEBUG: Downloading file with ID: %s", fileID)
	}

	file, err := bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		sendErrorMessage(bot, message.Chat.ID, "Error downloading file", err)
		return
	}

	// Create unique directory for this processing
	processDir := filepath.Join(tmpDir, fileID)
	if debug {
		log.Printf("DEBUG: Creating process directory: %s", processDir)
	}

	err = os.MkdirAll(processDir, 0755)
	if err != nil {
		sendErrorMessage(bot, message.Chat.ID, "Error creating processing directory", err)
		return
	}

	if !debug {
		defer os.RemoveAll(processDir)
	}

	// Download PDF
	pdfPath := filepath.Join(processDir, "statement.pdf")
	log.Printf("INFO: Downloading PDF to: %s", pdfPath)

	err = downloadFile(file.Link(bot.Token), pdfPath)
	if err != nil {
		sendErrorMessage(bot, message.Chat.ID, "Error saving PDF", err)
		return
	}

	// Convert PDF to images
	log.Println("INFO: Converting PDF to images")
	images, err := convertPDFToImages(pdfPath, processDir)
	if err != nil {
		sendErrorMessage(bot, message.Chat.ID, "Error converting PDF to images", err)
		return
	}
	if debug {
		log.Printf("DEBUG: Converted %d pages to images", len(images))
	}

	// Process images with Azure OpenAI
	log.Println("INFO: Processing images with Azure OpenAI")
	bot.Send(tgbotapi.NewMessage(message.Chat.ID, "Extracting transactions from images..."))

	transactions, err := processImagesWithAI(images)
	if err != nil {
		sendErrorMessage(bot, message.Chat.ID, "Error extracting transactions", err)
		if debug {
			log.Printf("DEBUG: AI Processing error: %v", err)
		}
		return
	}

	bot.Send(tgbotapi.NewMessage(message.Chat.ID,
		fmt.Sprintf("Found %d transactions. Generating report...", len(transactions))))

	// Generate markdown file
	mdPath := filepath.Join(processDir, "transactions.md")
	log.Printf("INFO: Generating markdown file: %s", mdPath)
	err = generateMarkdownFile(transactions, mdPath)
	if err != nil {
		sendErrorMessage(bot, message.Chat.ID, "Error generating markdown file", err)
		return
	}

	// Send markdown file
	log.Println("INFO: Sending markdown file to user")
	doc := tgbotapi.NewDocument(message.Chat.ID, tgbotapi.FilePath(mdPath))
	_, err = bot.Send(doc)
	if err != nil {
		sendErrorMessage(bot, message.Chat.ID, "Error sending markdown file", err)
		return
	}
	log.Printf("INFO: Successfully processed document for user %s", message.From.UserName)
}

func sendErrorMessage(bot *tgbotapi.BotAPI, chatID int64, message string, err error) {
	errorMsg := fmt.Sprintf("%s: %v", message, err)
	log.Printf("Error: %s", errorMsg)
	bot.Send(tgbotapi.NewMessage(chatID, errorMsg))
}
