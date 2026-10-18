package controllers

// DEV HTML REPORT OPEN SIGNAL v1
// HTML is explicitly labeled as HTML. The image fetch is only a network
// signal and must never be reported as guaranteed user open or execution.
import (
	"encoding/json"
	"fmt"
	"html"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/gophish/gophish/models"
)

func devReportRecipient(r *http.Request) (models.Result, models.Campaign, bool) {
	var emptyR models.Result
	var emptyC models.Campaign
	if r.Method != http.MethodGet {
		return emptyR, emptyC, false
	}
	rid := r.URL.Query().Get(models.RecipientParameter)
	if len(rid) != 7 {
		return emptyR, emptyC, false
	}
	result, err := models.GetResult(rid)
	if err != nil {
		return emptyR, emptyC, false
	}
	campaign, err := models.GetCampaign(result.CampaignId, result.UserId)
	if err != nil {
		return emptyR, emptyC, false
	}
	if campaign.ScenarioType != "file_download" || campaign.FileType != "html" || campaign.Status == models.CampaignComplete {
		return emptyR, emptyC, false
	}
	return result, campaign, true
}

// DownloadReportHandler returns a personalized HTML document. This is an
// explicitly selected .html scenario, never a disguised .xlsx attachment.
func (ps *PhishingServer) DownloadReportHandler(w http.ResponseWriter, r *http.Request) {
	rs, c, ok := devReportRecipient(r)
	if !ok {
		http.NotFound(w, r)
		return
	}
	base := strings.TrimRight(c.URL, "/")
	parsed, err := url.Parse(base)
	if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Host == "" {
		http.Error(w, "Invalid campaign URL", http.StatusInternalServerError)
		return
	}
	signal := "https://tracker.sympanion.com/file-open-signal?rid=" + url.QueryEscape(rs.RId)
	// Personalization and telemetry RID are built on the server for each result.
	body := fmt.Sprintf(`<!doctype html>
<html lang="tr"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>Rapor Önizlemesi</title><style>
body{font:15px system-ui,Arial,sans-serif;background:#f4f7fb;color:#172536;margin:0;padding:34px}
main{max-width:950px;margin:auto;background:white;border:1px solid #dde4ee;border-radius:12px;padding:30px}
h1{font-size:24px;margin:0 0 8px}p{color:#516176}table{width:100%%;border-collapse:collapse;margin-top:25px}
th,td{border:1px solid #d9e1ea;text-align:left;padding:12px}th{background:#eff5fb}
small{color:#637589}</style></head><body>
<main><h1>Rapor Önizlemesi</h1><p>Merhaba %s. Bu belge HTML biçiminde görüntülenen bir eğitim raporu önizlemesidir.</p>
<table><thead><tr><th>Başlık</th><th>Durum</th><th>Açıklama</th></tr></thead><tbody>
<tr><td>Rapor görünümü</td><td>Hazır</td><td>Örnek tablo</td></tr>
<tr><td>Dosya biçimi</td><td>HTML</td><td>Excel çalışma kitabı değildir</td></tr>
<tr><td>Güvenlik simülasyonu</td><td>Etkin</td><td>Belge açıldığında ağ sinyali alınabilir</td></tr>
</tbody></table><p><small>Bu doküman herhangi bir program çalıştırmaz ve makro içermez.</small></p></main>
<img src="%s" width="1" height="1" alt="" style="position:absolute;left:-9999px" referrerpolicy="no-referrer">
</body></html>`, html.EscapeString(rs.FirstName), html.EscapeString(signal))

	w.Header().Set("Cache-Control", "no-store, max-age=0")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="rapor-onizlemesi.html"`)
	if _, err := w.Write([]byte(body)); err != nil {
		return
	}
	detail, _ := json.Marshal(map[string]string{"format": "html", "action": "http-response-served"})
	_ = models.AddEvent(&models.Event{Email: rs.Email, Message: "File Download Served", Details: string(detail)}, rs.CampaignId)
}

// FileOpenSignalHandler serves a transparent 1x1 image. A GET indicates that
// an image resource was requested; it is NOT proof the file was executed.
func (ps *PhishingServer) FileOpenSignalHandler(w http.ResponseWriter, r *http.Request) {
	rs, _, ok := devReportRecipient(r)
	if !ok {
		http.NotFound(w, r)
		return
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		ip = r.RemoteAddr
	}
	details := models.EventDetails{
		Payload: url.Values{"signal": []string{"remote-image-request"}},
		Browser: map[string]string{"address": ip, "user-agent": r.UserAgent()},
	}
	if err := models.AddUniqueDevFileOpenSignal(rs.CampaignId, rs.Email, details); err != nil {
		http.Error(w, "Failed to record event", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Type", "image/gif")
	// Valid single-pixel transparent GIF.
	w.Write([]byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x01, 0x00, 0x01, 0x00, 0x80, 0x00, 0x00, 0xff, 0xff, 0xff, 0x00, 0x00, 0x00, 0x21, 0xf9, 0x04, 0x01, 0x00, 0x00, 0x00, 0x00, 0x2c, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x02, 0x02, 0x44, 0x01, 0x00, 0x3b})
}
