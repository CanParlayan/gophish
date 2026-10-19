package controllers

// File Simulation V2: render the selected GoPhish page and bind its existing download control.
import (
	"bytes"
	"html"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
)

func (ps *PhishingServer) DevFileLandingHandler(w http.ResponseWriter, r *http.Request) {
	rs, c, ok := devReportRecipient(r)
	if !ok {
		http.NotFound(w, r)
		return
	}
	// The same externally proxied path handles a remote image request made
	// only by the downloaded HTML file. It must never count as a link click.
	if r.URL.Query().Get("document_open") == "1" {
		ps.ServeHTMLDocumentOpenSignal(w, r, rs, c)
		return
	}
	// Reuse the already proxied /dev-file-landing route; no Apache change.
	if r.URL.Query().Get("preview") == "1" {
		ps.FilePreviewHandler(w, r)
		return
	}
	// GoPhish's original handler counts Clicked Link and renders the selected
	// landing page. Don't double-record clicks on subsequent file downloads.
	recorder := httptest.NewRecorder()
	ps.PhishHandler(recorder, r)
	response := recorder.Result()
	defer response.Body.Close()
	for key, values := range response.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	if response.StatusCode != http.StatusOK || !strings.Contains(strings.ToLower(response.Header.Get("Content-Type")), "text/html") {
		w.WriteHeader(response.StatusCode)
		_, _ = io.Copy(w, response.Body)
		return
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, 2<<20))
	if err != nil {
		http.Error(w, "Unable to render landing page", http.StatusInternalServerError)
		return
	}

	// Reuse existing controls. GET forms must submit rid as a named
	// field: query strings in a GET form action may be discarded by browsers.
	target := "/download-report"
	// An HTML error simulation belongs inside the downloaded HTML file,
	// not in a separate web page reached by an "İndir" button.
	if c.PreviewMode == "web_view" || c.PreviewMode == "download_preview" ||
		(c.PreviewMode == "error" && c.FileType != "html") {
		target = "/dev-file-landing"
	}
	formRE := regexp.MustCompile(`(?is)<form\b[^>]*\bid\s*=\s*["']download-intent["'][^>]*>`)
	if position := formRE.FindIndex(body); position != nil {
		newForm := []byte(`<form id="download-intent" action="` +
			html.EscapeString(target) + `" method="GET">` +
			`<input type="hidden" name="rid" value="` +
			html.EscapeString(rs.RId) + `">` +
			previewInput(target))
		updated := make([]byte, 0, len(body)+len(newForm))
		updated = append(updated, body[:position[0]]...)
		updated = append(updated, newForm...)
		body = append(updated, body[position[1]:]...)
	} else {
		// Older pages can use an existing <a id="download"> instead.
		// Do not append a new visible button.
		anchorRE := regexp.MustCompile(`(?is)<a\b[^>]*\bid\s*=\s*["']download["'][^>]*>`)
		hrefRE := regexp.MustCompile(`(?i)\bhref\s*=\s*(?:"[^"]*"|'[^']*'|[^\s>]+)`)
		position := anchorRE.FindIndex(body)
		if position == nil {
			http.Error(w, "Landing Page: existing download control not found", http.StatusUnprocessableEntity)
			return
		}
		anchor := string(body[position[0]:position[1]])
		hrefTarget := target + "?rid=" + url.QueryEscape(rs.RId)
		if target == "/dev-file-landing" {
			hrefTarget += "&preview=1"
		}
		href := `href="` + html.EscapeString(hrefTarget) + `"`
		if hrefRE.MatchString(anchor) {
			anchor = hrefRE.ReplaceAllString(anchor, href)
		} else {
			anchor = strings.TrimSuffix(anchor, ">") + " " + href + ">"
		}
		updated := make([]byte, 0, len(body)+len(anchor))
		updated = append(updated, body[:position[0]]...)
		updated = append(updated, anchor...)
		body = append(updated, body[position[1]:]...)
	}

	// Remove the obsolete event field.
	intentRE := regexp.MustCompile(`(?is)<input\b[^>]*\bname\s*=\s*["']simulation_event["'][^>]*>`)
	body = intentRE.ReplaceAll(body, nil)

	// Remove legacy injected anchors if they are already saved in HTML.
	extraLinkRE := regexp.MustCompile(`(?is)<p\s+style="text-align:center">\s*<a\s+id="download"[^>]*>HTML Raporunu İndir</a>\s*</p>`)
	body = extraLinkRE.ReplaceAll(body, nil)

	// The supplied landing page contains legacy JS which changes the link
	// back to tracker.sympanion.com/download/<rid>. Remove ONLY scripts
	// containing this obsolete endpoint. The correct href is in HTML.
	legacyScriptRE := regexp.MustCompile(`(?is)<script\b[^>]*>.*?</script\s*>`)
	body = legacyScriptRE.ReplaceAllFunc(body, func(script []byte) []byte {
		if bytes.Contains(script, []byte("tracker.sympanion.com/download/")) {
			return nil
		}
		return script
	})
	if c.FileType == "html" {
		body = bytes.ReplaceAll(body, []byte("Excel formatındaki raporunuzu"), []byte("HTML biçimindeki raporunuzu"))
		body = bytes.ReplaceAll(body, []byte("Excel Raporunu İndir"), []byte("HTML Raporunu İndir"))
	} else if c.FileType == "pdf" && c.PreviewMode == "web_view" {
		body = bytes.ReplaceAll(body, []byte("Excel formatındaki raporunuzu"), []byte("PDF raporunuzu"))
		body = bytes.ReplaceAll(body, []byte("Excel Raporunu İndir"), []byte("PDF Raporunu İndir"))
	}
	if c.PreviewMode == "web_view" {
		// The first action opens the browser viewer; downloading is offered
		// separately inside it. Do not label that action as a direct download.
		body = bytes.ReplaceAll(body, []byte("Raporunu İndir"), []byte("Raporunu Görüntüle"))
		body = bytes.ReplaceAll(body, []byte(">Dosyayı indir<"), []byte(">Dosyayı görüntüle<"))
		body = bytes.ReplaceAll(body, []byte(">İndir<"), []byte(">Görüntüle<"))
	}
	w.Header().Del("Content-Length")
	w.Header().Del("Content-Encoding")
	w.Header().Del("ETag")
	w.Header().Set("Cache-Control", "no-store, max-age=0")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}

// previewInput is only added to preview-mode GET forms, so their click
// stays on the already-proxied /dev-file-landing endpoint.
func previewInput(target string) string {
	if target == "/dev-file-landing" {
		return `<input type="hidden" name="preview" value="1">`
	}
	return ""
}
