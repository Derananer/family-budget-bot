package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/ai/azopenai"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
)

type Transaction struct {
	Date        string
	Description string
	Amount      string
}

func processImagesWithAI(imagePaths []string) ([]Transaction, error) {
	debug := os.Getenv("DEBUG") == "true"
	var transactions []Transaction

	for _, imagePath := range imagePaths {
		if debug {
			log.Printf("DEBUG: Processing image: %s", imagePath)
		}

		// Read and encode image
		imageData, err := os.ReadFile(imagePath)
		if err != nil {
			return nil, fmt.Errorf("error reading image: %v", err)
		}

		base64Image := base64.StdEncoding.EncodeToString(imageData)
		log.Println("INFO: Sending image to Azure OpenAI for processing")

		// Process with Azure OpenAI
		text, err := sendImageToAzureOpenAI(base64Image)
		if err != nil {
			return nil, fmt.Errorf("error processing with Azure OpenAI: %v", err)
		}

		if debug {
			log.Printf("DEBUG: Received response from Azure OpenAI: %s", text)
		}

		// Parse transactions from text
		pageTransactions, err := parseTransactions(text)
		if err != nil {
			return nil, fmt.Errorf("error parsing transactions: %v", err)
		}

		if debug {
			log.Printf("DEBUG: Parsed %d transactions from page", len(pageTransactions))
		}

		transactions = append(transactions, pageTransactions...)
	}

	return transactions, nil
}

func sendImageToAzureOpenAI(base64Image string) (string, error) {
	debug := os.Getenv("DEBUG") == "true"
	key := os.Getenv("AZURE_OPENAI_KEY")
	endpoint := os.Getenv("AZURE_OPENAI_ENDPOINT")
	deployment := os.Getenv("AZURE_OPENAI_DEPLOYMENT")

	keyCredential := azcore.NewKeyCredential(key)
	client, err := azopenai.NewClientWithKeyCredential(endpoint, keyCredential, nil)
	if err != nil {
		return "", err
	}

	systemPrompt := `You are a bank statement analyzer. Extract transaction details from the bank statement image.
Format each transaction exactly as follows, one per line:
{DATE}|{DESCRIPTION}|{AMOUNT}

Rules:
1. DATE format: DD.MM.YYYY
2. DESCRIPTION: Keep original text, remove any extra spaces
3. AMOUNT: Include currency symbol if present
4. Do not include headers or any other text
5. Do not include table formatting or markdown
6. Each field must be separated by | character
7. Each transaction must be on a new line

Example output:
01.03.2024|PAYMENT TO SHOP|₸50.00
02.03.2024|ATM WITHDRAWAL|₸100.00
03.03.2024|ONLINE TRANSFER|₽500.00
04.03.2024|GROCERY STORE|₽1,500.00`

	if debug {
		log.Printf("DEBUG: Using system prompt: %s", systemPrompt)
	}

	messages := []azopenai.ChatRequestMessageClassification{
		&azopenai.ChatRequestSystemMessage{
			Content: azopenai.NewChatRequestSystemMessageContent(systemPrompt),
		},
		&azopenai.ChatRequestUserMessage{
			Content: azopenai.NewChatRequestUserMessageContent([]azopenai.ChatCompletionRequestMessageContentPartClassification{
				&azopenai.ChatCompletionRequestMessageContentPartImage{
					ImageURL: &azopenai.ChatCompletionRequestMessageContentPartImageURL{
						URL: to.Ptr("data:image/jpeg;base64," + base64Image),
					},
				},
			}),
		},
	}

	resp, err := client.GetChatCompletions(context.TODO(), azopenai.ChatCompletionsOptions{
		Messages:       messages,
		DeploymentName: &deployment,
		MaxTokens:      to.Ptr[int32](2048),
		Temperature:    to.Ptr[float32](0.0),
	}, nil)

	if err != nil {
		return "", err
	}

	if len(resp.Choices) > 0 && resp.Choices[0].Message != nil {
		return *resp.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("no response from AI")
}
