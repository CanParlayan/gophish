# GoPhish DEV - HTML Landing Page Standardı v5

## Kapsam

Tüm yeni `scenario_type=file_download`, `file_type=html` kampanyaları,
kaydedilmeden önce **seçilmiş ve kaydedilmiş** Landing Page üzerinden doğrulanır.
Bunun için bir kampanya ID'si girmeniz gerekmez.

## Zorunlu indirme butonu

İstediğiniz tasarımı kullanın. Sayfada tam bir adet şunun bulunması yeterli:

```html
<a id="download" href="#">HTML Raporunu İndir</a>
```

- `id="download"` tek bir `<a>` bağlantısında olmalı.
- Sabit RID veya `tracker.sympanion.com/download/` JavaScript'i eklemeyin.
- Development GoPhish, her alıcı için indirme adresini sunucuda doldurur.
- E-posta şablonuna dosya indirme JavaScript'i yazmanız gerekmez.

## Hata / eski kampanyalar

Uyumsuz sayfada yeni bir HTML kampanyası başlatmaya çalışırsanız GoPhish
anlaşılır bir `HTML Landing Page hatası` mesajıyla işlemi reddeder.
Mevcut eski HTML kampanyalarında eksik buton yüzünden 500 yerine sade bir
indirme linki eklenir. Bu, ikinci bir kart değildir.

## Güvenlik ve kapsam sınırları

Yalnız development kaynak kodu ve binary'si değişir.
Production GoPhish, Apache, PostgreSQL/SQLite ve eski collector değiştirilmez.
XLSX, PDF ve DOCX kampanyalarının davranışı değiştirilmez.
`File Open Signal` yalnız bir uzak görsel HTTP isteğidir; kişinin dosyayı
kesin açtığını veya çalıştırdığını kanıtlamaz.
