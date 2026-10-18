package controllers

// Safe downloadable HTML awareness report. Only the HTTP file response is recorded.
import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"

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
	if campaign.ScenarioType != "file_download" || campaign.Status == models.CampaignComplete {
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
	if c.FileType != "html" {
		// All formats share this endpoint. The existing safe PDF/OOXML
		// generator is reused, and it logs File Download Served.
		ps.DownloadSimulationHandler(w, r)
		return
	}
	// Self-contained HTML: no tracking pixels or remote document resources.
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
<tr><td>Güvenlik simülasyonu</td><td>Etkin</td><td>İndirme yanıtı kaydedilir; dosyanın açılması izlenmez</td></tr>
</tbody></table><p><small>Bu doküman herhangi bir program çalıştırmaz ve makro içermez.</small></p></main>
</body></html>`, html.EscapeString(rs.FirstName))

	// In HTML error mode the simulated failure is the downloaded
	// document itself. Avoid showing one error page before download and
	// then a seemingly successful report after the user opens the file.
	if c.PreviewMode == "error" {
		body = fmt.Sprintf(`<!doctype html>
<html lang="tr"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>Rapor önizleme sorunu</title><style>
*{box-sizing:border-box}body{font:16px system-ui,Arial,sans-serif;background:#f4f6f9;color:#243449;margin:0;padding:40px 16px}
main{max-width:740px;margin:0 auto;background:#fff;border:1px solid #dce3ec;border-radius:12px;padding:36px;box-shadow:0 12px 32px #1322380b}
h1{font-size:25px;margin:0 0 16px}p{color:#52657b;line-height:1.7}
small{display:block;color:#7b8797;margin-top:42px}
</style></head><body><main>
<h1>Bu dosyayı görüntüleyemiyoruz</h1>
<p>Rapor önizlemesi oluşturulamadı. Dosyanın içeriği şu anda görüntülenemiyor.</p>
<small>Güvenlik farkındalık simülasyonu. Bu mesaj gerçek bir dosya arızasını veya bir kişinin dosyayı açtığını kanıtlamaz.</small>
</main></body></html>`)
	}

	w.Header().Set("Cache-Control", "no-store, max-age=0")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename=%q`, simulationFilename(c)))
	if _, err := w.Write([]byte(body)); err != nil {
		return
	}
	detail, _ := json.Marshal(map[string]string{"format": "html", "action": "http-response-served"})
	_ = models.AddEvent(&models.Event{Email: rs.Email, Message: "File Download Served", Details: string(detail)}, rs.CampaignId)
}
