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

func ConvertADRG(e *event.Event) error {
	object, err := pdf.NewObject(e.TempFolder + "/index.html")
	if err != nil {
		log.Print(err)
		return err
	}
	object.BlockLocalFileAccess = false
	object.UseExternalLinks = false
	object.UseLocalLinks = false
	object.EnableJavascript = false

	converter, err := pdf.NewConverter()
	if err != nil {
		return err
	}
	converter.Add(object)

	// Convert objects and save the output PDF document.

	outfileName := fmt.Sprint(e.UUID, ".pdf")
	outFile, err := os.OpenFile(outfileName, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Fatal(err)
	}
	/*outFile, err := os.Create(outfileName)
	if err != nil {
		return err
	}*/

	convertErr := converter.Run(outFile) // [ ] Fails here at the level of c-lib;
	if convertErr != nil {
		return convertErr
	}

	stat, err := outFile.Stat()
	if err != nil {
		return err
	}

	if stat.Size() == 0 {
		return errors.New("out file is empty")
	}

	converter.Destroy()
	outFile.Close()

	e.Dur = time.Since(e.Start)

	return nil
}

//PDFg
/*Converts files mentioned at the e.FilePath to PDF with SebastiaanKlippert/go-wkhtmltopdf*/
func PDFg(e *event.Event) error {
	pdfg, err := pdf2.NewPDFGenerator()
	if err != nil {
		return err
	}

	outFile, err := ioutil.ReadFile(e.TempFolder + "/index.html")
	if err != nil {
		return err
	}

	page := pdf2.NewPageReader(bytes.NewReader(outFile))
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

	err = pdfg.Create() // [ l Fails here;
	if err != nil {
		return err
	}

	outfileName := fmt.Sprint(e.UUID, ".pdf")
	err = pdfg.WriteFile(outfileName)
	if err != nil {
		return err
	}

	log.Printf("convertation for the %s finished", e.UUID.String())
	e.Dur = time.Since(e.Start)
	return nil
}

func CountTillFifty(e *event.Event) {
	log.Printf("job starter for %s", e.UUID.String())
	for i := 0; i < 49; i++ {
		e.Counter++
		time.Sleep(1 * time.Second)
	}
	log.Printf("job finished for %s, result: %s", e.UUID.String(), fmt.Sprint(e.Counter))
}
