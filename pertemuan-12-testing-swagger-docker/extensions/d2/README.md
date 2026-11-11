# D.2 — Notifikasi Perubahan Kuota

Varian D.2 menambahkan event `course.quota_changed` ketika admin memperbarui kuota kelas. Implementasinya dipisahkan dari baseline menggunakan build tag `d2`.

## Komponen

- `app/service/course_quota_d2.go`: perubahan kuota dan pencatatan event dalam satu transaksi.
- `app/model/quota_event_d2.go`: payload event perubahan kuota.
- `app/service/quota_relay_d2.go` dan `app/service/quota_worker_d2.go`: pemrosesan event dan notifikasi.
- `migrations/009_notification_recipients.sql`: idempotensi berdasarkan pasangan event, pengguna, dan channel.

Setiap penerima perlu memiliki pencatatan pengiriman tersendiri agar retry tidak mengulangi penerima yang sebelumnya berhasil.

## Menjalankan varian

```powershell
go build -tags d2 ./...
```

Gunakan database terpisah yang telah menerapkan migration 009. Database baseline dengan constraint migration 008 tidak boleh dicampur dengan perubahan skema D.2.

## Hasil sementara

Build dengan tag `d2` dan unit test terkait sudah berhasil. Pengujian runtime dengan banyak peserta, kegagalan sementara, dan pengiriman ulang masih perlu diselesaikan sebelum penerimaan D.2 dinyatakan lengkap.
