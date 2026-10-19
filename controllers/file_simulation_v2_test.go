package controllers

import (
	"archive/zip"
	"bytes"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gophish/gophish/models"
)

func TestFileSimulationV2GeneratedFormats(t *testing.T) {
	pdf := buildAwarenessPDF()
	if !bytes.HasPrefix(pdf, []byte("%PDF-1.4")) ||
		!bytes.Contains(pdf, []byte("startxref")) ||
		!bytes.Contains(pdf, []byte("%%EOF")) {
		t.Fatal("generated PDF is missing its expected structure")
	}
	for _, format := range []string{"xlsx", "docx"} {
		payload, err := buildAwarenessOfficeFile(format)
		if err != nil {
			t.Fatalf("%s generation: %v", format, err)
		}
		zr, err := zip.NewReader(bytes.NewReader(payload), int64(len(payload)))
		if err != nil {
			t.Fatalf("%s is not a ZIP container: %v", format, err)
		}
		hasContentTypes := false
		for _, f := range zr.File {
			if f.Name == "[Content_Types].xml" {
				hasContentTypes = true
			}
			if strings.Contains(strings.ToLower(f.Name), "vbaproject") {
				t.Fatal("unexpected executable macro content")
			}
		}
		if !hasContentTypes {
			t.Errorf("%s is missing OOXML content types", format)
		}
	}
}

func TestFileSimulationV2Filenames(t *testing.T) {
	c := models.Campaign{FileType: "pdf", FileName: "Finansal_Ozet.pdf"}
	if got := simulationFilename(c); got != "Finansal_Ozet.pdf" {
		t.Errorf("filename = %q", got)
	}
	for _, name := range []string{"../unsafe.pdf", "a.pdf\r\nHeader", "report.docx"} {
		c.FileName = name
		if got := simulationFilename(c); got != "awareness-simulation.pdf" {
			t.Errorf("unsafe name %q yielded %q", name, got)
		}
	}
	c.FileType, c.FileName = "html", ""
	if got := simulationFilename(c); got != "rapor-onizlemesi.html" {
		t.Errorf("HTML legacy name = %q", got)
	}
}

func TestFileSimulationV2PreviewInput(t *testing.T) {
	if previewInput("/download-report") != "" {
		t.Fatal("standard download must not have preview input")
	}
	if !strings.Contains(previewInput("/dev-file-landing"), `name="preview" value="1"`) {
		t.Fatal("preview form must have a named preview input")
	}
}

func TestBlankPDFMatchesReferencePage(t *testing.T) {
	pdf := buildBlankSimulationPDF()
	if !bytes.HasPrefix(pdf, []byte("%PDF-1.4")) ||
		!bytes.Contains(pdf, []byte("startxref")) ||
		!bytes.Contains(pdf, []byte("%%EOF")) ||
		!bytes.Contains(pdf, []byte("/MediaBox [0 0 612 792]")) ||
		!bytes.Contains(pdf, []byte("/Length 0")) {
		t.Fatal("blank PDF is not a valid single Letter page")
	}
	if bytes.Contains(pdf, []byte("Authorized security awareness simulation")) {
		t.Fatal("blank PDF must not include the legacy visible training text")
	}
}

func TestHTMLDocumentSignalURL(t *testing.T) {
	c := models.Campaign{URL: "https://portal.sympanion.com/old/path?bad=1#fragment"}
	got, err := htmlOpenSignalURL(c, "AbC1234")
	if err != nil {
		t.Fatal(err)
	}
	want := "https://portal.sympanion.com/dev-file-landing?document_open=1&rid=AbC1234"
	if got != want {
		t.Errorf("signal URL = %q, want %q", got, want)
	}
	c.URL = "http://portal.sympanion.com"
	if _, err := htmlOpenSignalURL(c, "AbC1234"); err == nil {
		t.Fatal("HTML open tracking must require HTTPS")
	}
}

func TestHTMLSignalRejectsOtherFormats(t *testing.T) {
	c := models.Campaign{ScenarioType: "file_download", FileType: "pdf", PreviewMode: "pdf_blank"}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/dev-file-landing?rid=AbC1234&document_open=1", nil)
	(&PhishingServer{}).ServeHTMLDocumentOpenSignal(w, r, models.Result{}, c)
	if w.Code != 404 {
		t.Fatalf("PDF must not trigger an HTML open event: %d", w.Code)
	}
}
