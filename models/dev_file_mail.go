package models

// DEV_SERVER_SIDE_HTML_JOURNEY_V1
// GoPhish assembles these links/pixels at SEND TIME for file_download/html.
// The SMTP template no longer needs {{.Tracker}} or special download links.
// A request to the tracking pixel is NOT reliable proof a human read the email.
import (
	"html"
	"net/url"
	"regexp"
	"strings"
)

var devMailAnchorRE = regexp.MustCompile(`(?is)<a\b[^>]*>`)
var devMailImageRE = regexp.MustCompile(`(?is)<img\b[^>]*>`)
var devMailHrefRE = regexp.MustCompile(`(?i)\bhref\s*=\s*(?:"([^"]*)"|'([^']*)')`)
var devMailSrcRE = regexp.MustCompile(`(?i)\bsrc\s*=\s*(?:"([^"]*)"|'([^']*)')`)

func devMatchURL(raw, baseURL, originalURL, rid string, paths ...string) bool {
	val := html.UnescapeString(raw)
	u, err := url.Parse(val)
	if err != nil {
		return false
	}
	b, err := url.Parse(baseURL)
	if err != nil {
		return false
	}
	if u.IsAbs() && !strings.EqualFold(u.Host, b.Host) {
		return false
	}
	if u.Query().Get(RecipientParameter) != rid {
		return false
	}
	for _, p := range paths {
		if u.Path == p {
			return true
		}
	}
	return val == originalURL
}

func devAttributeValue(tag string, rx *regexp.Regexp) (int, int, string, bool) {
	ii := rx.FindStringSubmatchIndex(tag)
	if ii == nil {
		return 0, 0, "", false
	}
	for _, ix := range []int{2, 4} {
		if ii[ix] >= 0 {
			return ii[ix], ii[ix+1], tag[ii[ix]:ii[ix+1]], true
		}
	}
	return 0, 0, "", false
}

func devReplaceLink(tag string, rx *regexp.Regexp, value string) string {
	from, to, _, ok := devAttributeValue(tag, rx)
	if !ok {
		return tag
	}
	return tag[:from] + html.EscapeString(value) + tag[to:]
}

func devMailBase(baseURL, path, rid string) string {
	return strings.TrimRight(baseURL, "/") + path + "?rid=" + url.QueryEscape(rid)
}

// The NORMAL GoPhish .URL and older direct-download URLs become landing links.
// Existing unrelated external links and CSS/images remain unchanged.
func devPrepareHTMLCampaignMail(body, originalURL, baseURL, rid string) string {
	landing := devMailBase(baseURL, "/dev-file-landing", rid)
	pixel := devMailBase(baseURL, "/dev-track", rid)
	buttons := 0
	body = devMailAnchorRE.ReplaceAllStringFunc(body, func(tag string) string {
		_, _, href, ok := devAttributeValue(tag, devMailHrefRE)
		if !ok {
			return tag
		}
		if !devMatchURL(href, baseURL, originalURL, rid, "/", "/download-report", "/download-simulation", "/dev-download-click", "/dev-file-landing") {
			return tag
		}
		buttons++
		return devReplaceLink(tag, devMailHrefRE, landing)
	})
	if buttons == 0 {
		block := `<p style="margin:24px 0"><a href="` + html.EscapeString(landing) + `" style="background:#2458a6;color:white;text-decoration:none;padding:13px 22px;display:inline-block;border-radius:6px">Raporumu Görüntüle</a></p>`
		if pos := strings.LastIndex(strings.ToLower(body), "</body>"); pos >= 0 {
			body = body[:pos] + block + body[pos:]
		} else {
			body += block
		}
	}
	// Normalize a rendered {{.Tracker}} image, so it doesn't point to PROD /track.
	pixels := 0
	body = devMailImageRE.ReplaceAllStringFunc(body, func(tag string) string {
		_, _, src, ok := devAttributeValue(tag, devMailSrcRE)
		if !ok {
			return tag
		}
		if !devMatchURL(src, baseURL, originalURL, rid, "/track", "/dev-track") {
			return tag
		}
		pixels++
		return devReplaceLink(tag, devMailSrcRE, pixel)
	})
	if pixels == 0 {
		tracker := `<img src="` + html.EscapeString(pixel) + `" alt="" width="1" height="1" style="width:1px;height:1px;border:0" />`
		if pos := strings.LastIndex(strings.ToLower(body), "</body>"); pos >= 0 {
			body = body[:pos] + tracker + body[pos:]
		} else {
			body += tracker
		}
	}
	return body
}

func devPrepareTextCampaignMail(body, originalURL, baseURL, rid string) string {
	landing := devMailBase(baseURL, "/dev-file-landing", rid)
	// Literal URLs are replaced for text/plain recipients. No tracking pixel.
	body = strings.ReplaceAll(body, originalURL, landing)
	for _, path := range []string{"/download-report", "/download-simulation", "/dev-download-click"} {
		body = strings.ReplaceAll(body, devMailBase(baseURL, path, rid), landing)
	}
	if !strings.Contains(body, landing) {
		body += "\n\nRaporu görüntüle: " + landing + "\n"
	}
	return body
}
