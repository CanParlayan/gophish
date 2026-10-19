# GoPhish - Dosya Simülasyonu (indirme öncelikli)

> Durum: Kod GitHub `feature/scenario-tracking` branch'indedir. Sunucuda
> `go test ./models ./controllers` ve manuel kabul testi geçmeden canlı
> PASS olarak kabul edilmez.

## Akış

Yeni **File Download Simulation** kampanyalarında dosya adının uzantısı
formatı belirler. Kullanıcıya mevcut Landing Page üzerindeki indirme
düğmesi gösterilir; **web görüntüleyici, ara önizleme veya ikinci onay
ekranı kullanılmaz**.

| Yeni dosya | Sunulan içerik | Olay |
|---|---|---|
| `.html` | Referans `dm3(2).pdf` gibi boş beyaz sayfa | HTTP indirmesi + şartlara bağlı HTML açılma sinyali |
| `.pdf` | Referans gibi boş, geçerli Letter PDF | HTTP indirmesi; yerel açılma ölçülemez |
| `.docx`, `.xlsx` | Önceki güvenli makrosuz örnekler | HTTP indirmesi |

Eski kampanyaların `preview_mode` değerleri/verileri değiştirilmez.
Yeni HTML kampanyaları `html_signal`, yeni PDF kampanyaları `pdf_blank`
iç modunu kullanır. Eski standart PDF/HTML kampanyalarının içerik
davranışı korunur. **Standard Campaign hiç değiştirilmez.**

## Ölçümler

- **Clicked Link**: E-posta / landing bağlantısına istek alındı.
- **File Download Served**: Sunucu indirme yanıtını gönderdi; diske
  kaydedildiğine dair kesin kanıt değildir.
- **HTML Open Signal**: İndirilen yerel HTML tarayıcıda açılırken
  `https://<kampanya-hostu>/dev-file-landing?rid=...&document_open=1`
  adresine yüklenen 1x1 görselin HTTP isteği alındı. Bu, insanın dosyayı
  okuduğuna dair kesin kanıt değildir. Tarayıcı güvenlik politikası,
  ağ erişimi, engelleyici yazılımlar veya kurumsal proxy sinyali engelleyebilir.
- **PDF açıldı**: PDF'nin Adobe Reader / Preview / Chrome / Edge gibi
  okuyucularda yerel açılışını güvenilir şekilde bildiren pasif ve
  platformlar-arası bir mekanizma yoktur. Bu format için **Ölçülemiyor**
  gösterilir. PDF içine otomatik URL açtıran JavaScript veya tehlikeli
  çalıştırılabilir öğe konulmaz.

HTML sinyali aynı alıcı ve kampanya için bir kez kaydedilir.
Gösterge benzersiz alıcıları sayar; sonuç ekranındaki HTML açılma
yüzdesi ağ sinyallerine dayanır. `File Open Signal` (eski kayıt)
geçmişte kalır; yeni sinyalin adı `HTML Open Signal`dır.

## Web yönlendirme

Var olan `/dev-file-landing` Apache reverse-proxy yolu tekrar kullanılır.
Yeni proxy kuralı gerekmez:

- `/dev-file-landing?rid=...`: Landing page.
- `/dev-file-landing?rid=...&document_open=1`: Sadece yeni HTML
  kampanyalarında görsel yanıtı ve benzersiz sinyal.
- `/download-report?rid=...`: İçerik indir (`Content-Disposition: attachment`).

Veritabanında yeni sütun yoktur; önceden kurulmuş
`20261019010000_file_simulation_v2` migration'ı yeterlidir.

## Sunucuda güvenli dağıtım

```bash
cd /home/pentester/gophish-src-v0.12.1
bash deploy.sh
```

Script fetch, Go testleri, build, binary ve frontend yedeği,
SQLite `.backup`, service restart ve hata halinde binary/frontend
geri alma adımlarını içerir. Hata durumunda üretimi değiştirmeden dur.

## Manuel kabul testi (yalnızca yetkili test alıcısı)

1. Yeni `deneme.html` kampanyası: landing'de İndir, doğrudan
   `deneme.html` indirilsin ve `File Download Served` kaydı oluşsun.
2. Dosyayı yerel Safari/Chrome'da aç. Görünüm boş sayfa olsun.
   Network erişimi izin veriyorsa `HTML Open Signal` oluşsun; engellenirse
   açılışı olmuş varsayma.
3. Yeni `deneme.pdf` kampanyası: gerçek tek sayfalık boş PDF indirilsin
   ve açılabilsin. `File Download Served` yazılsın, PDF açılması
   **Ölçülemiyor** kalsın.
4. Standard Campaign ekranı, önceki kampanya kayıtları ve
   sonuç grafikleri değişmemiş olsun.

Bu simülasyonlar makro, EXE veya kullanıcı cihazında kod çalıştırmaz.
