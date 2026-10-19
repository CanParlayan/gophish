package models

import (
	"strings"
	"testing"
)

func validSimulationCampaign() Campaign {
	return Campaign{
		ScenarioType: "file_download",
		FileType: "pdf",
		Name: "Authorized exercise",
		Template: Template{Name: "Test template"},
		Page: Page{Name: "Test landing"},
		SMTP: SMTP{Name: "Test mailer"},
		Groups: []Group{{Name: "Test group"}},
	}
}

func TestFileSimulationV2Validation(t *testing.T) {
	c := validSimulationCampaign()
	if err := c.Validate(); err != nil {
		t.Fatalf("default campaign: %v", err)
	}
	if c.PreviewMode != "standard" || c.FileName != "awareness-simulation.pdf" {
		t.Fatalf("unexpected defaults: mode=%q filename=%q", c.PreviewMode, c.FileName)
	}
	for _, kind := range []string{"pdf", "docx", "xlsx", "html"} {
		c := validSimulationCampaign()
		c.FileType = kind
		if err := c.Validate(); err != nil {
			t.Errorf("%s is not valid: %v", kind, err)
		}
	}
	for _, mode := range []string{"standard", "error", "download_preview", "web_view"} {
		c := validSimulationCampaign()
		c.PreviewMode = mode
		if err := c.Validate(); err != nil {
			t.Errorf("mode %q: %v", mode, err)
		}
	}
	for _, kind := range []string{"xlsx", "docx"} {
		c := validSimulationCampaign()
		c.FileType = kind
		c.PreviewMode = "web_view"
		if err := c.Validate(); err == nil {
			t.Errorf("web viewer wrongly allowed for %s", kind)
		}
	}
	for _, name := range []string{"../report.pdf", "report.xlsx", "report.pdf\r\nHeader: x", ".pdf", strings.Repeat("a", 90)+".pdf"} {
		c := validSimulationCampaign()
		c.FileName = name
		if err := c.Validate(); err == nil {
			t.Errorf("unsafe filename accepted: %q", name)
		}
	}
	c = validSimulationCampaign()
	c.FileType = "exe"
	if err := c.Validate(); err == nil {
		t.Fatal("executable file type accepted")
	}
}

func TestFileSimulationFilenameDeterminesType(t *testing.T) {
	for _, name := range []string{"rapor.html", "rapor.pdf", "rapor.xlsx", "rapor.docx"} {
		c := validSimulationCampaign()
		c.FileType = ""
		c.FileName = name
		if err := c.Validate(); err != nil {
			t.Errorf("filename %q rejected: %v", name, err)
		}
		if c.FileType != strings.TrimPrefix(name, "rapor.") {
			t.Errorf("filename %q yielded type %q", name, c.FileType)
		}
	}
	for _, name := range []string{"rapor.exe", "rapor", "../rapor.pdf", "rapor.pdf\r\n"} {
		c := validSimulationCampaign()
		c.FileType = ""
		c.FileName = name
		if err := c.Validate(); err == nil {
			t.Errorf("invalid filename accepted: %q", name)
		}
	}
	c := validSimulationCampaign()
	c.FileType, c.FileName = "html", "rapor.pdf"
	if err := c.Validate(); err == nil {
		t.Fatal("conflicting legacy file_type and filename accepted")
	}
}

func TestFileSimulationV2NonFileScenarioClearsOptions(t *testing.T) {
	c := validSimulationCampaign()
	c.ScenarioType = "standard"
	c.FileName, c.PreviewMode = "report.pdf", "error"
	if err := c.Validate(); err != nil {
		t.Fatal(err)
	}
	if c.FileType != "" || c.FileName != "" || c.PreviewMode != "" {
		t.Fatal("stale file selection on non-file campaign")
	}
}

func TestFileSimulationV2LandingControls(t *testing.T) {
	validPages := []string{
		`<html><body><form id="download-intent"></form><button form="download-intent" type="submit">İndir</button></body></html>`,
		`<html><body><a id="download" href="#">Download</a></body></html>`,
	}
	for _, markup := range validPages {
		if err := validateDevHTMLLanding(markup); err != nil {
			t.Errorf("supported control rejected: %v", err)
		}
	}
	bad := `<html><body><form id="other"></form><button form="other" type="submit">İndir</button></body></html>`
	if err := validateDevHTMLLanding(bad); err == nil {
		t.Fatal("unsupported button would lead to 422 on the landing handler")
	}
}
