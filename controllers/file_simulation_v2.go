package controllers

// Safe, generated file-simulation preview pages. These are HTTP interaction
// signals, not evidence that an Office/PDF file was opened on a workstation.
import (
	"encoding/json"
	"html"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/gophish/gophish/models"
)

var fileSimulationNameRE = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,79}$`)

func simulationFilename(c models.Campaign) string {
	name := c.FileName
	if fileSimulationNameRE.MatchString(name) &&
		strings.HasSuffix(strings.ToLower(name), "."+c.FileType) {
		return name
	}
	if c.FileType == "html" {
		return "rapor-onizlemesi.html"
	}
	return "awareness-simulation." + c.FileType
}

func logFilePreviewEvent(rs models.Result, message string, mode string) {
	detail, _ := json.Marshal(map[string]string{
		"preview_mode": mode,
		"meaning": "http-request-signal",
	})
	_ = models.AddEvent(&models.Event{
		Email: rs.Email, Message: message, Details: string(detail),
	}, rs.CampaignId)
}

// FilePreviewHandler renders a simulated browser preview. No uploaded,
// arbitrary or executable document is involved.
func (ps *PhishingServer) FilePreviewHandler(w http.ResponseWriter, r *http.Request) {
	rs, c, ok := devReportRecipient(r)
	if !ok {
		http.NotFound(w, r)
		return
	}
	if c.PreviewMode == "" || c.PreviewMode == "standard" {
		http.Redirect(w, r, "/download-report?rid="+url.QueryEscape(rs.RId), http.StatusSeeOther)
		return
	}
	if c.PreviewMode != "error" && c.PreviewMode != "download_preview" {
		http.NotFound(w, r)
		return
	}

	filename := html.EscapeString(simulationFilename(c))
	rid := html.EscapeString(rs.RId)
	var page strings.Builder
	page.WriteString(`<!doctype html><html lang="tr"><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>Dosya önizlemesi</title>
<style>
*{box-sizing:border-box}body{background:#f4f6f9;color:#243449;font:15px system-ui,-apple-system,Arial,sans-serif;margin:0;padding:36px 16px}
main{max-width:740px;margin:0 auto;background:white;border:1px solid #e1e5ee;border-radius:12px;padding:30px;box-shadow:0 12px 32px #1322380b}
header{font-size:15px;color:#516176;border-bottom:1px solid #e1e5ee;padding-bottom:18px;overflow-wrap:anywhere}
h1{font-size:23px;margin:30px 0 8px}p{color:#56677b;line-height:1.6}
.actions{display:flex;gap:12px;flex-wrap:wrap;margin-top:28px;align-items:center}
button,.button{display:inline-block;padding:11px 20px;border:0;border-radius:6px;text-decoration:none;font:inherit;cursor:pointer}
button{background:#0868b5;color:white}.button{background:#eef2f7;color:#23354a}
.note{font-size:12px;color:#7c8797;margin-top:32px}
</style></head><body><main><header>📄 `)
	page.WriteString(filename)
	page.WriteString(`</header>`)
	if c.PreviewMode == "error" {
		page.WriteString(`<h1>Bu dosyayı açamıyoruz</h1><p>Önizleme sırasında bir sorun oluştu. Yeniden deneyebilir veya dosyayı indirebilirsiniz.</p>`)
	} else {
		page.WriteString(`<h1>Dosya indirilmeye hazır</h1><p>Bu dosya için güvenli bir örnek oluşturuldu. İndirmek için aşağıdaki düğmeyi kullanın.</p>`)
	}
	page.WriteString(`<div class="actions"><form action="/download-report" method="GET"><input type="hidden" name="rid" value="`)
	page.WriteString(rid)
	page.WriteString(`"><button type="submit">Dosyayı indir</button></form>`)
	if c.PreviewMode == "error" {
		page.WriteString(`<a class="button" href="/dev-file-landing?preview=1&amp;rid=`)
		page.WriteString(url.QueryEscape(rs.RId))
		page.WriteString(`&amp;retry=1">Yenile</a>`)
	}
	page.WriteString(`</div><p class="note">Güvenlik farkındalık simülasyonu: Bu ekran bir dosya açılma kanıtı değildir.</p></main></body></html>`)

	w.Header().Set("Cache-Control", "no-store, max-age=0")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; form-action 'self'; base-uri 'none'")
	if _, err := w.Write([]byte(page.String())); err != nil {
		return
	}
	logFilePreviewEvent(rs, "Preview Shown", c.PreviewMode)
	if r.URL.Query().Get("retry") == "1" {
		logFilePreviewEvent(rs, "Retry Clicked", c.PreviewMode)
	}
}
