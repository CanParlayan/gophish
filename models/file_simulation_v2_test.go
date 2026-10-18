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
	for _, mode := range []string{"standard", "error", "download_preview"} {
		c := validSimulationCampaign()
		c.PreviewMode = mode
		if err := c.Validate(); err != nil {
			t.Errorf("mode %q: %v", mode, err)
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
