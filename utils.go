package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

func downloadFile(url, filepath string) error {
	debug := os.Getenv("DEBUG") == "true"
	if debug {
		log.Printf("DEBUG: Downloading file from %s to %s", url, filepath)
	}

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err == nil {
		log.Printf("INFO: Successfully downloaded file to %s", filepath)
	}
	return err
}

func generateMarkdownFile(transactions []Transaction, filepath string) error {
	debug := os.Getenv("DEBUG") == "true"
	if debug {
		log.Printf("DEBUG: Generating markdown file with %d transactions", len(transactions))
	}

	var content strings.Builder

	content.WriteString("# Bank Statement Transactions\n\n")
	content.WriteString("| Date | Description | Amount |\n")
	content.WriteString("|------|-------------|--------|\n")

	for _, t := range transactions {
		content.WriteString(fmt.Sprintf("| %s | %s | %s |\n", t.Date, t.Description, t.Amount))
	}

	if debug {
		log.Printf("DEBUG: Writing markdown file to: %s", filepath)
	}
	return os.WriteFile(filepath, []byte(content.String()), 0644)
}

func parseTransactions(text string) ([]Transaction, error) {
	debug := os.Getenv("DEBUG") == "true"
	var transactions []Transaction
	lines := strings.Split(text, "\n")

	if debug {
		log.Printf("DEBUG: Parsing text:\n%s", text)
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) >= 3 {
			date := strings.TrimSpace(parts[0])
			desc := strings.TrimSpace(parts[1])
			amount := strings.TrimSpace(parts[2])

			if date == "" || desc == "" || amount == "" {
				if debug {
					log.Printf("DEBUG: Skipping invalid transaction line: %s", line)
				}
				continue
			}

			transactions = append(transactions, Transaction{
				Date:        date,
				Description: desc,
				Amount:      amount,
			})

			if debug {
				log.Printf("DEBUG: Parsed transaction: Date=%s, Desc=%s, Amount=%s",
					date, desc, amount)
			}
		} else if debug {
			log.Printf("DEBUG: Invalid line format: %s", line)
		}
	}

	if len(transactions) == 0 {
		return nil, fmt.Errorf("no valid transactions found in the text")
	}

	return transactions, nil
}
