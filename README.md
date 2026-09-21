# Latihan Fiber

Repositori pembelajaran backend menggunakan Go, Fiber, PostgreSQL, dan pendekatan berlapis dari dasar API hingga desain API tingkat lanjut.

## Struktur Pertemuan

```text
.
├── pertemuan-1-dasar-go/
├── pertemuan-2-rest-api-http/
├── pertemuan-3-database-repository/
├── pertemuan-4-clean-architecture/
├── pertemuan-5-authentication-security/
├── pertemuan-6-authorization-rbac/
├── pertemuan-7-advanced-api-design/
├── go.mod
└── go.sum
```

Setiap folder pertemuan memiliki aplikasi dan dokumentasi sesuai materi tahap tersebut. Modul dapat dijalankan dari folder masing-masing dengan `go run .` setelah konfigurasi environment dan PostgreSQL disiapkan.

## Pertemuan 7 — Advanced API Design

Folder `pertemuan-7-advanced-api-design` merupakan pengembangan dari Modul 6 dengan fokus pada desain API yang lebih konsisten, aman, dan siap digunakan pada aplikasi produksi.

Materi yang diterapkan:

- **Centralized Error Handling & AppError**: error dari middleware, handler, service, dan repository dikembalikan sebagai error terstruktur. Fiber `ErrorHandler` terpusat mengubahnya menjadi response API yang konsisten tanpa menulis response error langsung dari business logic.
- **Declarative Validation**: DTO/request menggunakan tag validasi deklaratif dari `github.com/go-playground/validator/v10`, termasuk aturan nilai wajib, batas angka, panjang string, serta pemetaan error ke nama field JSON.
- **Keyset / Cursor-based Pagination**: pagination offset digantikan cursor opaque melalui `EncodeCursor` dan `DecodeCursor`. Query menggunakan urutan deterministik `created_at DESC, id DESC`, repository menyediakan `FindAfterCursor`, dan migration menambahkan index komposit yang sesuai.
- **Content Negotiation**: endpoint daftar mendukung `Accept: application/json` dan `Accept: text/csv`. Format lain ditolak dengan status `406 Not Acceptable`.
- **Request Logging**: access log menggunakan status akhir dari error yang dikembalikan, sehingga error `4xx`/`5xx` tidak tercatat keliru sebagai `200`.
- **Dokumentasi laporan bug**: sembilan planted bug Modul 7, gejala, akar masalah, perbaikan, dan bukti verifikasinya dicatat di [`pertemuan-7-advanced-api-design/BUG_REPORT.md`](pertemuan-7-advanced-api-design/BUG_REPORT.md).

### File penting Modul 7

- `config/app.go`: konfigurasi aplikasi dan centralized Fiber error handler.
- `helper/errors.go`: `AppError`, validation error, serta encode/decode cursor.
- `helper/validator.go`: integrasi `validator/v10` untuk DTO.
- `helper/negotiation.go`: response JSON/CSV berdasarkan header `Accept`.
- `middleware/middleware.go`: request logger dan validasi `Content-Type`.
- `app/repository/student_repository.go` dan `achievement_repository.go`: query keyset pagination.
- `migrations/001_create_students.sql` dan `002_create_achievements.sql`: index cursor `(created_at DESC, id DESC)`.
- `BUG_REPORT.md`: berita acara ringkas perbaikan sembilan bug Modul 7.

### Menjalankan dan menguji Modul 7

Dari root repositori:

```bash
go build ./pertemuan-7-advanced-api-design/...
go test ./pertemuan-7-advanced-api-design/...
go vet ./pertemuan-7-advanced-api-design/...
```

Untuk menjalankan server:

```bash
cd pertemuan-7-advanced-api-design
go run .
```

Siapkan `.env` berdasarkan `.env.example`, PostgreSQL, dan jalankan migration sebelum menggunakan endpoint yang membutuhkan database.
