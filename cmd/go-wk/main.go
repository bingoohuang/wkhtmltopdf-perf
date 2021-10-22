package main

import (
	"fmt"
	"log"
	"os"

	pdf "github.com/adrg/go-wkhtmltopdf"
)

func main() {
	if len(os.Args[1:]) < 2 {
		fmt.Printf("%s <web page url> <output file>\n", os.Args[0])
		os.Exit(0)
	}

	// Initialize library.
	if err := pdf.Init(); err != nil {
		log.Fatal(err)
	}
	defer pdf.Destroy()

	// Create object from URL.
	pdfObj, err := pdf.NewObject(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	// Create converter.
	converter, err := pdf.NewConverter()
	if err != nil {
		log.Fatal(err)
	}
	defer converter.Destroy()

	// Add created objects to the converter.
	converter.Add(pdfObj)

	// Convert objects and save the output PDF document.
	outFile, err := os.Create(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}
	defer outFile.Close()

	if err := converter.Run(outFile); err != nil {
		log.Fatal(err)
	}
}
