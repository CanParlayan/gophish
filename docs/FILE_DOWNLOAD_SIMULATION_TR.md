# GoPhish - Senaryo ve Dosya İndirme Simülasyonu

## Özellikler

Bu fork, standart GoPhish kampanyalarına ek olarak Scenario Type
ve File Type seçenekleri sunar.

File Download Simulation kapsamında HTML rapor indirme akışı,
dosya sunma olayı ve dosya açılma sinyali ayrı izlenir.

## Kampanya oluşturma

1. New Campaign ekranını açın.
2. Scenario Type olarak File Download Simulation seçin.
3. File Type olarak HTML Rapor Önizlemesi seçin.
4. Email Template, Landing Page, URL, Sending Profile ve Group seçin.
5. Kampanyayı başlatın.
6. Campaign Results ekranından olayları inceleyin.

## Landing Page

Mevcut uygulama, download-intent kimlikli formu kullanan
indirme butonlarını destekler.

Örnek:

<form id="download-intent" method="GET">
</form>

<button type="submit" form="download-intent">İndir</button>

Uygulama bu formun action adresini kampanyaya özel
/download-report?rid=... endpoint'ine yönlendirir.

Bu yapı kullanıldığında ayrıca görünür indirme bağlantısı
eklenmemesi hedeflenir.

## Olayların anlamları

- Email Sent: E-posta gönderildi.
- Email Opened: E-posta açılma sinyali alındı.
- Clicked Link: Kampanya bağlantısı ziyaret edildi.
- File Download Served: Sunucu dosya isteğine yanıt verdi.
- File Open Signal: Dosya açılma sinyali alındı.
- Submitted Data: Normal form gönderimi olayıdır.

File Download Simulation için Submitted Data kullanılması gerekmez.

File Download Served, dosyanın cihaza kaydedildiğini kesin olarak
kanıtlamaz. File Open Signal da dosyanın gerçekten açıldığının
kesin kanıtı değildir.

## Derleme

go test ./models ./controllers
go build -o /tmp/gophish-production-new .

## Production

Servis: gophish.service
Binary: /home/pentester/gophish/gophish
Kaynak: /home/pentester/gophish-src-v0.12.1

Production binary değiştirilmeden önce yedek alınmalıdır.

## Test durumu - 19 Ekim 2026

- Models testleri: PASS
- Controllers testleri: PASS
- Go binary derlemesi: PASS
- Production servis başlangıcı: PASS
- Binary SHA256 eşleşmesi: PASS
- Tarayıcıda indirme akışı: Henüz doğrulanmadı

## Güvenlik

Production veritabanları, config.json, SMTP parolaları,
API anahtarları, özel anahtarlar, loglar ve yedekler
GitHub'a gönderilmemelidir.

Simülasyonlar yalnızca yetkili kapsamda yürütülmelidir.
