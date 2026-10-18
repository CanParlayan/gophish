# GoPhish — File Simulation V2 (Türkçe)

> **Durum:** V2 temel sürüm sunucuda test edilip çalıştırıldı. Sonraki
> dosya adı ve HTML hata akışı düzeltmeleri GitHub branch'indedir;
> bunlar için yeniden test ve kontrollü deployment gerekir.

## Amaç ve kapsam

Yetkili güvenlik farkındalığı kampanyalarında güvenli dosya indirme ve
simüle edilmiş tarayıcı önizlemesi ölçmek.

Desteklenen gerçek örnek dosyalar:

| Dosya türü | Oluşan içerik | Açılma sinyali |
|---|---|---|
| HTML (.html) | Statik HTML raporu | Uzak görsel isteği alınabilir; **kanıt değildir** |
| PDF (.pdf) | Geçerli minimal PDF | Hayır |
| Excel (.xlsx) | Makrosuz, geçerli OOXML çalışma kitabı | Hayır |
| Word (.docx) | Makrosuz, geçerli OOXML belge | Hayır |

Özel dosya yükleme, EXE, makro ve dosya açılışında yerel kod yürütme
**desteklenmez**. Bir HTML dosyasına başka uzantı verilmez.

## Yeni kampanya oluşturma

1. New Campaign → **File Download Simulation** seçin.
2. **Dosya adı** alanına `Finansal_Ozet.pdf`, `rapor.html`,
   `ornek.xlsx` veya `ornek.docx` yazın. Ayrı tür seçimi yoktur;
   sistem gerçek formatı uzantıdan belirler.
   Sadece güvenli ASCII dosya adları (en çok 80 karakter) kabul edilir.
   Eski API istemcileri `file_type` göndermeye devam edebilir; alanlar
   çelişirse istek reddedilir.
3. **Preview mode** seçin:
   - **Standard Download:** İndir kontrolü doğrudan dosyayı sunar.
   - **Preview Error:** HTML'de İndir düğmesi doğrudan HTML dosyasını
     indirir; açıldığında HTML belgesinin içinde simüle hata gösterilir.
     PDF/Word/Excel'de İndir kontrolü öncelikle tarayıcıdaki simüle
     hata sayfasına gider; kullanıcı Yenile veya Dosyayı indir seçebilir.
   - **Download + Preview:** İndir kontrolü bilgi/önizleme sayfasına
     gider; kullanıcı buradan Dosyayı indir seçer. Otomatik indirme yapılmaz.
4. Email Template, Landing Page, URL, Sending Profile ve hedef Group seçin.
5. Kapsamı doğrulayın, test alıcısıyla denedikten sonra kampanyayı başlatın.

## Landing Page uyumluluğu

GoPhish, Landing Page'in mevcut indirme butonlarını tekrar kullanır.
Yeni ve görünür bir buton eklemez. Bu iki kontrol yapısından biri gereklidir:

```html
<form id="download-intent" method="GET"></form>
<button type="submit" form="download-intent">İndir</button>
```

veya

```html
<a id="download" href="#">Download</a>
```

Sunucu, `download-intent` formuna `name="rid"` gizli alanını ekler.
GET formunun `action` query string'i `rid` taşımak için kullanılmaz.
Önizleme modlarında (HTML hata modu hariç) formun içine
`preview=1` eklenir.

Endpoint'ler:

- `/dev-file-landing?rid=...`: Seçilmiş Landing Page.
- `/dev-file-landing?rid=...&preview=1`: Simüle önizleme.
- `/download-report?rid=...`: Seçilen gerçek dosya biçimi.
- `/download-simulation?rid=...`: Eski doğrudan PDF/XLSX/DOCX
  bağlantıları için geriye dönük uyumluluk.
- `/file-open-signal?rid=...`: Yalnızca HTML uzaktan görüntü talebi.
  HTML belgesi yerel dosya olarak açıldığında tarayıcı güvenlik ayarları
  dış ağ isteğini engelleyebilir; sinyal garantili değildir.

Mevcut Apache `/dev-file-landing` ve `/download-report` yönlendirmeleri
yeni önizleme için de kullanılır; ayrı `/file-preview` proxy kuralı gerekmez.

## Event anlamları

| Olay | Sunucunun kesin olarak bildiği |
|---|---|
| Clicked Link | Landing bağlantısına HTTP isteği geldi |
| File Download Served | Dosyanın HTTP yanıtı sunucu tarafından gönderildi |
| Preview Shown | Simüle önizleme HTTP yanıtı gönderildi |
| Retry Clicked | Önizlemenin Yenile bağlantısına istek geldi |
| File Open Signal | HTML içindeki uzaktaki görsele ağ isteği geldi |

**Hiçbiri dosyanın kaydedildiğini, insan tarafından açıldığını veya
çalıştırıldığını tek başına kanıtlamaz.** PDF, XLSX ve DOCX dosyalarında
yerel dosya açma telemetrisi yoktur. Yeni indirme tıklamaları
`Submitted Data` olarak kaydedilmemelidir.

## Veri tabanı ve geriye dönük uyumluluk

Yeni migration: `20261019010000_file_simulation_v2.sql` (SQLite/MySQL).
`campaigns` tablosuna `file_name` ve `preview_mode` eklenir.
Mevcut kampanyalarda `preview_mode=standard`, boş dosya adı için
sunucu biçime uygun varsayılan isim kullanır. Önceki migration'lar korunur.
Yeni sürümden önce **DB ve binary yedeği** alınmalıdır.

## SSH ile güncelleme — önce test

```bash
cd /home/pentester/gophish-src-v0.12.1
git fetch fork feature/scenario-tracking
git switch feature/scenario-tracking
git pull --ff-only fork feature/scenario-tracking
gofmt -w models/campaign.go models/dev_html_landing_validation.go \
  controllers/dev_file_landing.go controllers/dev_html_report.go \
  controllers/file_simulation_v2.go controllers/phish.go
go test ./models ./controllers
go build -o /tmp/gophish-production-v2 .
```

`gofmt` kaynak dosyalarını değiştirir; format farklarını gerekiyorsa
ayrı commit'leyin. GitHub'a gizli anahtar, production `config.json`,
DB, hedef listesi, log veya binary göndermeyin.

## Production kabul matrisi

Yeni test kampanyasında her format için Standard Download, Preview Error,
Download + Preview doğrulanmalı (4 × 3 senaryo). HTML Preview Error akışında
ara hata sayfası yerine doğrudan indirme ve dosya içinde hata görünümü
beklenir. Her senaryoda:

1. Mevcut İndir kontrolü çalışmalı; altına üçüncü buton eklenmemeli.
2. İndirilen dosya **gerçek seçilen biçimde** açılmalı; dosya adı uygun olmalı.
3. Önizleme/Yenile olayları gerektiğinde oluşmalı; yeni Submitted Data oluşmamalı.
4. Eski HTML kampanyalarının indirmesi bozulmamalı.

Önce sadece yetkili test alıcısı kullanın. Testler geçmeden aktif
kampanyalara dokunmayın. Servis yeniden başlatma ve rollback işlemleri
için güncel production binary'si ve veritabanını ayrı ayrı yedekleyin.
