# GoPhish — File Simulation V2 (Türkçe)

> **Durum:** Temel V2 sunucuda çalışıyor. Dosya indirme gösterimini
> sadeleştiren son arayüz düzeltmeleri GitHub branch'indedir;
> sunucuda test/deploy edilmeden canlı kabul edilmemelidir.

## Amaç ve kapsam

Yetkili güvenlik farkındalığı kampanyalarında güvenli dosya indirme ve
simüle edilmiş tarayıcı önizlemesi ölçmek.

Desteklenen gerçek örnek dosyalar:

| Dosya türü | Oluşan içerik | Tek dosya metriği |
|---|---|---|
| HTML (.html) | Yerel açılabilen statik HTML raporu | HTTP dosya yanıtı |
| PDF (.pdf) | Geçerli minimal PDF | HTTP dosya yanıtı |
| Excel (.xlsx) | Makrosuz, geçerli OOXML çalışma kitabı | HTTP dosya yanıtı |
| Word (.docx) | Makrosuz, geçerli OOXML belge | HTTP dosya yanıtı |

**Tercih edilen tek yöntem GoPhish HTTP takibidir.** Ayrı ajan, endpoint
entegrasyonu, belge içine doğrulama bağlantısı veya uzaktan görsel
yükleyerek açılma takibi yoktur. Özel dosya yükleme, EXE, makro ve dosya
açılışında yerel kod yürütme **desteklenmez**. Bir HTML dosyasına başka
uzantı verilmez.

## Yeni kampanya oluşturma

1. New Campaign → **File Download Simulation** seçin.
2. **Dosya adı** alanına `Finansal_Ozet.pdf`, `rapor.html`,
   `ornek.xlsx` veya `ornek.docx` yazın. Ayrı tür seçimi yoktur;
   sistem gerçek formatı uzantıdan belirler.
   Sadece güvenli ASCII dosya adları (en çok 80 karakter) kabul edilir.
   Eski API istemcileri `file_type` göndermeye devam edebilir; alanlar
   çelişirse istek reddedilir.
3. Yeni kampanyalarda önizleme modu yoktur: İndir doğrudan güvenli örnek
   dosyayı sunar. Önceden oluşturulmuş kampanyaların ayarları değiştirilmez.
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
Eski önizleme modlarını kullanan mevcut kampanyalarda `preview=1`
yönlendirmesi geriye dönük uyumluluk için korunur.

Endpoint'ler:

- `/dev-file-landing?rid=...`: Seçilmiş Landing Page.
- `/dev-file-landing?rid=...&preview=1`: Simüle önizleme.
- `/download-report?rid=...`: Seçilen gerçek dosya biçimi.
- `/download-simulation?rid=...`: Eski doğrudan PDF/XLSX/DOCX
  bağlantıları için geriye dönük uyumluluk.

Mevcut Apache `/dev-file-landing` ve `/download-report` yönlendirmeleri
yeni önizleme için de kullanılır; ayrı `/file-preview` proxy kuralı gerekmez.

## File Download Simulation gösterimi

File Download Simulation raporunda, standart kampanya grafiklerinden ayrı,
yedi küçük yüzde göstergesi bulunur: Gönderildi, Etkileşim yok,
Açıldı (e-posta), Tıklandı, İndirildi, Dosya açıldı, Raporlandı.

İndirildi sayısı benzersiz alıcıların `File Download Served` HTTP olayından
hesaplanır; tekrar eden istekler ayrı kişi sayılmaz. Dosya açıldı göstergesi
bilinmeyen değerdir (**Ölçülemiyor**), yüzde sıfır değildir: indirilen
yerel PDF/Office/HTML dosyalarının açılışı GoPhish tarafından ölçülemez.
Standard Campaign raporu ve eski olay kayıtları korunur.

## Event anlamları

| Olay | Sunucunun kesin olarak bildiği |
|---|---|
| Clicked Link | Landing bağlantısına HTTP isteği geldi |
| File Download Served | Dosyanın HTTP yanıtı sunucu tarafından gönderildi |
| Preview Shown | Simüle önizleme HTTP yanıtı gönderildi |
| Retry Clicked | Önizlemenin Yenile bağlantısına istek geldi |
| File Open Signal (eski kayıtlar) | Önceden oluşturulmuş HTML kampanyasının tarihsel ağ isteği; yeni sinyal üretilmez |

**Hiçbiri dosyanın kaydedildiğini, insan tarafından açıldığını veya
çalıştırıldığını kanıtlamaz.** Yeni HTML belgelerinde de dış kaynak
yükleme/ayrı açılma sinyali yoktur. Eski olaylar denetim geçmişi olarak
veritabanında kalır fakat açılma grafiği veya yeni rozet olarak
gösterilmez. Sonuçlarda yalnızca HTTP dosya yanıtı metriği bulunur.
Yeni indirme tıklamaları `Submitted Data` olarak kaydedilmemelidir.

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

Yeni kampanyalarda güvenli **doğrudan indirme** akışı doğrulanmalıdır.
Eski kampanyalardaki önizleme seçenekleri veri uyumluluğu için korunur.
Her senaryoda:

1. Mevcut İndir kontrolü çalışmalı; altına üçüncü buton eklenmemeli.
2. İndirilen dosya **gerçek seçilen biçimde** açılmalı; dosya adı uygun olmalı.
3. Dosya indirme olayı benzersiz alıcıya yazılmalı; yeni Submitted Data oluşmamalı.
4. Eski HTML kampanyalarının indirmesi bozulmamalı.
5. Yeni HTML dosyasında uzak görsel/izleme pikseli bulunmamalı; artık
   `File Open Signal` üretilmemeli. Eski olay geçmişi korunmalı.

Önce sadece yetkili test alıcısı kullanın. Testler geçmeden aktif
kampanyalara dokunmayın. Servis yeniden başlatma ve rollback işlemleri
için güncel production binary'si ve veritabanını ayrı ayrı yedekleyin.
