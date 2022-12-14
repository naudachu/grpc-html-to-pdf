package converter

import (
	"bytes"
	"errors"
	"fmt"
	"grpc-html-to-pdf/internal/event"
	"io/ioutil"
	"log"
	"os"
	"time"

	pdf2 "github.com/SebastiaanKlippert/go-wkhtmltopdf"
	pdf "github.com/adrg/go-wkhtmltopdf"
)

func ConvertADRG(event *event.Event) error {
	object, err := pdf.NewObject(event.TempFolder + "/index.html")
	if err != nil {
		log.Print(err)
		return err
	}

	object.Header.ContentCenter = "[title]"
	object.Header.DisplaySeparator = true
	object.Footer.ContentLeft = "[date]"
	object.Footer.ContentCenter = "Sample footer information"
	object.Footer.ContentRight = "[page]"
	object.Footer.DisplaySeparator = true

	converter, err := pdf.NewConverter()
	if err != nil {
		return err
	}
	converter.Add(object)

	converter.Title = "Sample document"
	converter.PaperSize = pdf.A4
	converter.Orientation = pdf.Landscape
	converter.MarginTop = "1cm"
	converter.MarginBottom = "1cm"
	converter.MarginLeft = "10mm"
	converter.MarginRight = "10mm"

	// Convert objects and save the output PDF document.

	outfileName := fmt.Sprint(event.UUID, ".pdf")
	outFile, err := os.Create(outfileName)
	if err != nil {
		return err
	}
	defer outFile.Close()

	if err := converter.Run(outFile); err != nil {
		return err
	}
	stat, err := outFile.Stat()
	if err != nil {
		return err
	}

	if stat.Size() == 0 {
		return errors.New("out file is empty")
	}
	converter.Destroy()

	event.Dur = time.Since(event.Start)

	return nil
}

func PDFg(event *event.Event) error {
	pdfg, err := pdf2.NewPDFGenerator()
	if err != nil {
		return err
	}
	htmlfile, err := ioutil.ReadFile(event.TempFolder + "/index.html")
	if err != nil {
		return err
	}

	page := pdf2.NewPageReader(bytes.NewReader(htmlfile))
	page.EnableLocalFileAccess.Set(true)
	page.DisableExternalLinks.Set(true)
	page.DisableInternalLinks.Set(true)

	pdfg.AddPage(page)
	pdfg.Dpi.Set(600)

	// The contents of htmlsimple.html are saved as base64 string in the JSON file
	/*jb, err := pdfg.ToJSON()
	if err != nil {
		return err
	}

	// Server code
	pdfgFromJSON, err := pdf2.NewPDFGeneratorFromJSON(bytes.NewReader(jb))
	if err != nil {
		return err
	}



	err = pdfgFromJSON.Create()
	if err != nil {
		return err
	}*/

	err = pdfg.Create()
	if err != nil {
		return err
	}

	outfileName := fmt.Sprint(event.UUID, ".pdf")
	err = pdfg.WriteFile(outfileName)
	if err != nil {
		return err
	}

	log.Printf("convertation for the %s finished", event.UUID.String())
	event.Dur = time.Since(event.Start)
	return nil
}
