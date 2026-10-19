package controllers

// Browser document viewer for opt-in PDF/HTML awareness simulations.
// Only successfully served inline document bytes produce a Web File Viewed
// event. This does not observe local file opens or document reading.
import (
	"html"
	"net/http"
	"net/url"
	"strings"

	"github.com/gophish/gophish/models"
)

const webDocumentViewerHTML = `<!doctype html>
<html lang="tr">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>Belge Görüntüleyici</title>
<style>
*{box-sizing:border-box}
body{background:#f3f5f8;color:#27364a;font:16px system-ui,-apple-system,Segoe UI,Arial,sans-serif;margin:0;padding:24px}
main{max-width:1120px;margin:0 auto;border:1px solid #dce3ec;border-radius:14px;background:#fff;box-shadow:0 12px 34px #1b2d430d;overflow:hidden}
header{display:flex;justify-content:space-between;align-items:center;flex-wrap:wrap;gap:16px;padding:20px 24px;border-bottom:1px solid #e1e7ef}
.document{display:flex;align-items:center;gap:11px;min-width:0}
.document-icon{font-size:28px}
.document-name{font-weight:650;font-size:18px;overflow-wrap:anywhere}
.subtitle{font-size:12px;color:#748194;margin-top:5px}
.download{background:#1169b4;color:#fff;text-decoration:none;border-radius:7px;padding:12px 20px;font-weight:600;display:inline-block}
.download:hover{background:#0a5597}
.frame{background:#eceff3;padding:18px}
iframe{display:block;background:#fff;border:1px solid #dce3ec;border-radius:7px;width:100%;height:calc(100vh - 230px);min-height:470px}
footer{padding:12px 24px;color:#728095;font-size:12px}
@media(max-width:620px){body{padding:0}main{border-radius:0}header{padding:18px}.frame{padding:8px}iframe{height:70vh;min-height:400px}}
</style>
</head>
<body>
<main>
<header>
<div class="document"><span class="document-icon" aria-hidden="true">📄</span>
<div><div class="document-name">{{FILENAME}}</div>
<div class="subtitle">{{FORMAT}} • Web belge görüntüleyici</div></div></div>
<a class="download" href="{{DOWNLOAD_URL}}">Dosyayı indir</a>
</header>
<div class="frame">
<iframe src="{{INLINE_URL}}" title="Belge içeriği" loading="eager" referrerpolicy="no-referrer">
Bu tarayıcı belgeyi gömülü görüntüleyemiyor. Yukarıdaki indirme bağlantısını kullanabilirsiniz.
</iframe>
</div>
<footer>Güvenlik farkındalık simülasyonu. Burada görüntüleme isteği ölçülür; indirilmiş yerel bir dosyanın açılması ölçülmez.</footer>
</main>
</body>
</html>`

func (ps *PhishingServer) webDocumentViewer(w http.ResponseWriter, rs models.Result, c models.Campaign) {
	if c.ScenarioType != "file_download" || c.PreviewMode != "web_view" ||
		(c.FileType != "pdf" && c.FileType != "html") {
		http.Error(w, "Document viewer unavailable", http.StatusNotFound)
		return
	}
	rid := url.QueryEscape(rs.RId)
	filename := html.EscapeString(simulationFilename(c))
	format := "PDF"
	if c.FileType == "html" {
		format = "HTML"
	}
	page := strings.NewReplacer(
		"{{FILENAME}}", filename,
		"{{FORMAT}}", format,
		"{{INLINE_URL}}", html.EscapeString("/download-report?rid="+rid+"&inline=1"),
		"{{DOWNLOAD_URL}}", html.EscapeString("/download-report?rid="+rid),
	).Replace(webDocumentViewerHTML)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store, max-age=0")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; frame-src 'self'; style-src 'unsafe-inline'; base-uri 'none'; form-action 'none'")
	_, _ = w.Write([]byte(page))
}
