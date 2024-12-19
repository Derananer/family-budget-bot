package main

import (
	"fmt"
	"image/jpeg"
	"log"
	"os"
	"path/filepath"

	"github.com/gen2brain/go-fitz"
)

func convertPDFToImages(pdfPath, outputDir string) ([]string, error) {
	debug := os.Getenv("DEBUG") == "true"
	log.Printf("INFO: Opening PDF file: %s", pdfPath)

	doc, err := fitz.New(pdfPath)
	if err != nil {
		return nil, fmt.Errorf("error opening PDF: %v", err)
	}
	defer doc.Close()

	if debug {
		log.Printf("DEBUG: PDF has %d pages", doc.NumPage())
	}

	var imagePaths []string

	for n := 0; n < doc.NumPage(); n++ {
		if debug {
			log.Printf("DEBUG: Processing page %d", n)
		}

		img, err := doc.Image(n)
		if err != nil {
			return nil, fmt.Errorf("error extracting page %d: %v", n, err)
		}

		imagePath := filepath.Join(outputDir, fmt.Sprintf("page_%03d.jpg", n))
		f, err := os.Create(imagePath)
		if err != nil {
			return nil, fmt.Errorf("error creating image file: %v", err)
		}

		err = jpeg.Encode(f, img, &jpeg.Options{Quality: jpeg.DefaultQuality})
		if err != nil {
			f.Close()
			return nil, fmt.Errorf("error encoding image: %v", err)
		}
		f.Close()

		imagePaths = append(imagePaths, imagePath)
		log.Printf("INFO: Created image: %s", imagePath)
	}

	return imagePaths, nil
}
