# Pertemuan 11 — Background Jobs

Implementasi background job pada REST API SIAKAD Mini menggunakan Go Fiber. Perubahan KRS diproses melalui PostgreSQL, sedangkan pengiriman notifikasi dilakukan secara asinkron melalui transactional outbox, Redis, dan Asynq.

## Fitur utama

- Pendaftaran, pembatalan, dan perpindahan kelas (KRS).
- Pencatatan event KRS bersama transaksi database.
- Relay outbox yang meneruskan event ke antrean Redis.
- Worker notifikasi dengan retry, idempotensi, dan graceful shutdown.
- Endpoint riwayat notifikasi mahasiswa serta monitoring antrean untuk admin.
- Varian tugas D.2 untuk notifikasi perubahan kuota (build tag `d2`).

## Teknologi

Go, Fiber, PostgreSQL (pgx), MongoDB, Redis, Asynq, dan JWT.

## Struktur proyek

| Lokasi | Isi |
| --- | --- |
| `app/model`, `app/repository`, `app/service` | Model, akses data, dan logika aplikasi |
| `route`, `middleware` | Endpoint HTTP dan kontrol akses |
| `config`, `database` | Konfigurasi serta koneksi layanan |
| `migrations` | Skema dan data awal PostgreSQL |
| `infra/baseline` | Konfigurasi Docker untuk layanan lokal |
| `tools/redeliver` | Pengiriman ulang task untuk uji idempotensi |
| `extensions/d2` | Varian perubahan kuota dan migration 009 |
| `postman` | Collection dan template environment |
| `evidence` | Bukti pengujian yang disimpan berdasarkan skenario |

Arsitektur aplikasi dijelaskan lebih lanjut pada [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md).

## Menjalankan aplikasi

1. Pastikan Go sesuai versi pada `go.mod`, serta PostgreSQL, MongoDB, dan Redis tersedia. Konfigurasi Docker lokal ada di [infra/baseline](infra/baseline/README.md).
2. Dari folder `pertemuan-11-background-job`, salin `.env.example` menjadi `.env` lalu isi parameter koneksi dan `JWT_SECRET` secara lokal. File `.env` tidak disimpan di Git.
3. Siapkan database khusus praktikum dan jalankan migration sesuai urutan: `migrations/bootstrap/001`–`005`, lalu `migrations/005_courses.sql` hingga `008_outbox.sql`. Migration D.2 (`009`) diterapkan hanya pada database terpisah untuk varian tersebut.
4. Jalankan API dan worker pada terminal berbeda:

```powershell
go run . -role=api
```

```powershell
go run . -role=worker
```

Untuk menjalankan keduanya dalam satu proses, gunakan `go run . -role=all`. Hentikan proses interaktif dengan Ctrl+C.

## Endpoint yang digunakan

Seluruh endpoint berikut memakai prefix `/api/v1`.

| Method | Endpoint | Kegunaan |
| --- | --- | --- |
| `POST` | `/auth/login` | Login pengguna |
| `POST` | `/courses/:id/enroll` | Mendaftar kelas |
| `DELETE` | `/courses/:id/enroll` | Keluar dari kelas |
| `POST` | `/enrollments/swap` | Pindah kelas |
| `GET` | `/notifications/mine` | Riwayat notifikasi |
| `GET` | `/admin/queue` | Monitoring antrean |
| `PATCH` | `/courses/:id` | Perubahan kuota oleh admin |

Endpoint selain autentikasi publik memerlukan Bearer token; monitoring dan perubahan kuota juga memerlukan permission yang sesuai.

## Ringkasan pengujian

| Bagian | Hasil |
| --- | --- |
| C.1–C.5 | PASS pada checkpoint yang terdokumentasi (alur KRS, gangguan Redis, retry, SkipRetry, idempotensi) |
| C.6 | Task notifikasi selesai sebelum worker berhenti; skenario kelas 9 dan waktu Ctrl+C belum terbukti lengkap |
| C.7 | Log dan status antrean tersedia; timeout Redis belum memiliki penyebab yang terkonfirmasi |
| C.8 | `go build ./...` dan `go vet ./...` berhasil; penerimaan keseluruhan mengikuti kelengkapan C.6–C.7 |
| D.1 | Koreksi puzzle dicatat dalam evidence |
| D.2 | Build varian dan unit test berhasil; uji runtime multi-penerima belum lengkap |
| D.3 | Uji pembanding terisolasi belum lengkap |

Bukti yang relevan tersedia di:

- [C.1 — Jalur normal dan screenshot Postman](evidence/c1-postman-2026-10-29/)
- [C.2 — Gangguan dan pemulihan Redis](evidence/c2-resume-2026-11-01/)
- [C.3–C.8 — Hasil pengujian lanjutan](evidence/completion-2026-11-01/)
- [D.1 — Perbaikan compiler](evidence/gate-b1/compiler-2026-10-27/) dan [perbaikan perilaku](evidence/gate-b2a/)

## Postman

Impor [Week-11.postman_collection.json](postman/Week-11.postman_collection.json) dan [template environment](postman/Week-11-Baseline.postman_environment.json). Isi token dan password hanya pada environment lokal, bukan pada collection yang dibagikan. Petunjuk singkat ada di [postman/README.md](postman/README.md).
