package controllers

import (
	"net/http"

	"github.com/gophish/gophish/models"
)

// ServeHTMLDocumentOpenSignal responds to the HTTP image request embedded in
// a downloaded HTML test document. Browsers may suppress remote requests
// from local files; a recorded request does not prove a human viewed content.
func (ps *PhishingServer) ServeHTMLDocumentOpenSignal(w http.ResponseWriter, r *http.Request, rs models.Result, c models.Campaign) {
	if r.Method != http.MethodGet || c.ScenarioType != "file_download" ||
		c.FileType != "html" || c.PreviewMode != "html_signal" {
		http.NotFound(w, r)
		return
	}
	if err := models.AddUniqueHTMLDocumentOpenSignal(rs.CampaignId, rs.Email); err != nil {
		http.Error(w, "Unable to log document signal", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Cache-Control", "no-store, no-cache, max-age=0")
	w.Header().Set("Content-Type", "image/gif")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Length", "43")
	_, _ = w.Write([]byte{
		0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x01, 0x00, 0x01, 0x00,
		0x80, 0x00, 0x00, 0xff, 0xff, 0xff, 0x00, 0x00, 0x00, 0x21,
		0xf9, 0x04, 0x01, 0x00, 0x00, 0x00, 0x00, 0x2c, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x02, 0x02,
		0x44, 0x01, 0x00, 0x3b,
	})
}
