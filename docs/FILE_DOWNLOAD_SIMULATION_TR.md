# GoPhish — File Simulation V2 (Türkçe)

> **Durum:** Web PDF/HTML görüntüleyici değişiklikleri GitHub branch'indedir;
> sunucuda Go testleri ve gerçek tarayıcı kabul testi tamamlanmadan
> production PASS sayılmaz.

## Amaç ve kapsam

Yetkili güvenlik farkındalığı kampanyalarında güvenli dosya indirme ve
web üzerinden belge görüntüleme ve indirme etkileşimlerini ölçmek.

Desteklenen gerçek örnek dosyalar:

| Dosya türü | Oluşan içerik | Yeni kampanya akışı |
|---|---|---|
| HTML (.html) | Statik HTML raporu | Web görüntüleyici, ayrı indirme |
| PDF (.pdf) | Geçerli minimal PDF | Web görüntüleyici, ayrı indirme |
| Excel (.xlsx) | Makrosuz OOXML çalışma kitabı | Doğrudan indirme |
| Word (.docx) | Makrosuz OOXML belge | Doğrudan indirme |

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
3. Yeni PDF ve HTML kampanyalarında mevcut landing kontrolü web görüntüleyiciyi
   açar; içerik `inline` HTTP yanıtı olarak sunulursa `Web File Viewed`
   kaydedilir. Ayrı **Dosyayı indir** kontrolü `File Download Served` kaydeder.
   DOCX/XLSX yeni kampanyalarında doğrudan indirme korunur.
   Eski kampanyaların kaydedilmiş ayarları değiştirilmez.
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
Yeni PDF/HTML kampanyalarında `preview_mode=web_view` otomatik seçilir.
Mevcut landing kontrolü `/dev-file-landing?preview=1&rid=...` adresine
gider. Eski önizleme modları geriye dönük uyumluluk için korunur.

Endpoint'ler:

- `/dev-file-landing?rid=...`: Seçilmiş Landing Page.
- `/dev-file-landing?rid=...&preview=1`: Yeni web görüntüleyici
  (yalnızca `web_view` PDF/HTML için); eski modlarda eski önizleme.
- `/download-report?rid=...&inline=1`: Web görüntüleyici içine gerçek
  PDF/HTML baytları (Content-Disposition: inline). Başarılı yanıt
  `Web File Viewed` olayını kaydeder; bir indirme olayı değildir.
- `/download-report?rid=...`: Dosyayı indirme (attachment). Başarılı
  yanıt `File Download Served` olayını kaydeder.
- `/download-simulation?rid=...`: Eski doğrudan PDF/XLSX/DOCX
  bağlantıları için geriye dönük uyumluluk.

Mevcut Apache `/dev-file-landing` ve `/download-report` yönlendirmeleri
web görüntüleyici için de kullanılır; yeni Apache proxy kuralı gerekmez.

## File Download Simulation gösterimi

File Download Simulation raporunda, standart kampanya grafiklerinden ayrı,
yedi küçük yüzde göstergesi bulunur: Gönderildi, Etkileşim yok,
Açıldı (e-posta), Tıklandı, İndirildi, Web'de Açıldı, Raporlandı.

İndirildi sayısı benzersiz alıcıların `File Download Served` olayından,
Web'de Açıldı sayısı ise `Web File Viewed` olayından hesaplanır.
Aynı kişiden tekrar istek gelmesi kişi yüzdesini şişirmez.
Eski kampanyalarda ve DOCX/XLSX kampanyalarında web görüntüleyici yoksa
Web'de Açıldı **Ölçülemiyor** olarak kalır.
**Web'de Açıldı yalnızca inline belgenin HTTP üzerinden sunulduğu anlamındadır**;
gerçek kullanıcı tarafından okunduğunu veya yerel dosyanın açıldığını
kanıtlamaz. Standard Campaign raporu ve eski olay kayıtları korunur.

## Event anlamları

| Olay | Sunucunun kesin olarak bildiği |
|---|---|
| Clicked Link | Landing bağlantısına HTTP isteği geldi |
| File Download Served | İndirme yanıtı sunucu tarafından gönderildi |
| Web File Viewed | PDF/HTML içeriği tarayıcı görüntüleyicisine inline HTTP yanıtı olarak gönderildi |
| Preview Shown | Simüle önizleme HTTP yanıtı gönderildi |
| Retry Clicked | Önizlemenin Yenile bağlantısına istek geldi |
| File Open Signal (eski kayıtlar) | Önceden oluşturulmuş HTML kampanyasının tarihsel ağ isteği; yeni sinyal üretilmez |

**Hiçbiri dosyanın kaydedildiğini, insan tarafından açıldığını veya
çalıştırıldığını kanıtlamaz.** Yeni HTML belgelerinde de dış kaynak
yükleme/ayrı açılma sinyali yoktur. Eski olaylar denetim geçmişi olarak
veritabanında kalır fakat açılma grafiği veya yeni rozet olarak
gösterilmez. Sonuçlarda indirme ve web görüntüleme HTTP olayları ayrı gösterilir.
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

Yeni PDF ve HTML kampanyalarında web görüntüleyici + indirme;
DOCX/XLSX kampanyalarında doğrudan indirme test edilmelidir.
Eski kampanyalardaki önizleme seçenekleri korunur. Her senaryoda:

1. Landing'deki mevcut kontrol PDF/HTML için web görüntüleyiciyi açmalı.
2. PDF/HTML iframe yanıtı `inline` olmalı ve `Web File Viewed`
   olayı üretmeli. Bu aşamada `File Download Served` oluşmamalı.
3. **Dosyayı indir** bağlantısının `attachment` yanıtı
   `File Download Served` üretmeli; dosya adı ve format doğru olmalı.
4. DOCX/XLSX ve eski kampanyaların doğrudan indirme davranışı korunmalı.
5. Standard Campaign grafikleri değişmemeli. Yeni `Submitted Data`,
   uzaktan izleme pikseli veya `File Open Signal` oluşmamalı.

Önce sadece yetkili test alıcısı kullanın. Testler geçmeden aktif
kampanyalara dokunmayın. Servis yeniden başlatma ve rollback işlemleri
için güncel production binary'si ve veritabanını ayrı ayrı yedekleyin.
