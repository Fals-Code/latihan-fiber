# Pertemuan 7 — Advanced API Design

Modul ini mengembangkan API dari Pertemuan 6 dengan fokus pada konsistensi error, validasi deklaratif, pagination yang stabil, dan dukungan beberapa format response.

## Materi dan fitur

### Centralized AppError dan ErrorHandler

`helper/errors.go` mendefinisikan `AppError` untuk membawa status HTTP, pesan aman, detail error, dan penyebab asli. Middleware, handler, service, dan repository mengembalikan error; `config/app.go` menggunakan Fiber `ErrorHandler` terpusat untuk mengubah error menjadi `WebResponse` yang konsisten.

`RequestLogger` mencatat status akhir request, termasuk status `4xx` atau `5xx` yang berasal dari error sebelum diproses oleh error handler.

### Declarative Validation dengan validator v10

DTO request menggunakan tag `validate` dan divalidasi melalui `github.com/go-playground/validator/v10`. Error validasi dipetakan kembali ke nama field JSON melalui `helper/validator.go`, sehingga response validasi tetap terstruktur dan konsisten.

### Keyset / Cursor-based Pagination

Pagination offset digantikan oleh cursor opaque untuk daftar student dan achievement. Cursor dibuat dan dibaca melalui:

- `helper.EncodeCursor`
- `helper.DecodeCursor`
- `StudentRepository.FindAfterCursor`
- `AchievementRepository.FindAfterCursor`

Query menggunakan urutan stabil `created_at DESC, id DESC`. Index pendukung tersedia pada:

- `migrations/001_create_students.sql`
- `migrations/002_create_achievements.sql`

Response daftar menyertakan metadata `next_cursor` jika masih tersedia data berikutnya. Cursor tidak valid menghasilkan error `400 Bad Request`.

### Content Negotiation

Endpoint daftar mendukung dua format berdasarkan header `Accept`:

- `application/json` untuk response JSON standar.
- `text/csv` untuk response daftar dalam format CSV.

Format selain JSON dan CSV ditolak dengan status `406 Not Acceptable`. Implementasinya tersedia di `helper/negotiation.go`.

## Struktur folder

```text
pertemuan-7-advanced-api-design/
├── app/
│   ├── model/
│   ├── repository/
│   └── service/
├── config/
├── database/
├── helper/
│   ├── errors.go
│   ├── negotiation.go
│   └── validator.go
├── middleware/
├── migrations/
├── route/
├── BUG_REPORT.md
├── .env.example
└── main.go
```

## Laporan bug

Sembilan planted bug pada Bagian B Modul 7, termasuk lokasi, gejala, akar masalah, perbaikan, dan bukti verifikasi, didokumentasikan dalam [`BUG_REPORT.md`](BUG_REPORT.md).

## Menjalankan dan menguji

Siapkan `.env` berdasarkan `.env.example`, PostgreSQL, dan jalankan migration dari folder `migrations`. Dari root repositori, gunakan:

```bash
go build ./pertemuan-7-advanced-api-design/...
go test ./pertemuan-7-advanced-api-design/...
go vet ./pertemuan-7-advanced-api-design/...
```

Untuk menjalankan server dari folder ini:

```bash
go run .
```

## Rekonsiliasi database bukti

Aplikasi tidak mengharuskan `hl_test`. Koneksi database dibaca dari `.env` melalui `DB_NAME`, dengan fallback `praktikum_backend` pada `database/postgres.go`. Bukti historis pada D.2–D.4 yang menyebut `hl_test` adalah hasil pengujian terdahulu dan tidak diubah menjadi klaim pengujian `praktikum_backend`.

Audit read-only terdahulu memverifikasi `current_database() = praktikum_backend` tanpa operasi tulis atau migration. Status **NOT VERIFIED** pada audit tersebut telah digantikan oleh pengujian HTTP aktual berikutnya pada `praktikum_backend`: D.2 PATCH dan D.3 cursor pagination **PASS** berdasarkan ringkasan self-contained pada dokumen evidence masing-masing.

D.4 content negotiation/error handling **PASS berdasarkan bukti manual terbaru dari pengguna**: JSON HTTP 200 dengan IDs `[34,33]`, CSV HTTP 200 dengan IDs `[34,33]` dan OwnerID numerik `36`/`37`, XML HTTP 406 `NOT_ACCEPTABLE`, serta wildcard `*/*` HTTP 200 JSON. Screenshot asli ada dalam laporan yang telah dikumpulkan tetapi tidak tersedia di workspace; hasil ini bukan pengujian ulang Codex. Endpoint `/auth/me` tetap **NOT VERIFIED**.
