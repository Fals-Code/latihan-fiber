# Pertemuan 10 — API dengan MongoDB

Implementasi tugas mandiri Bagian C dibuat sebagai folder terpisah dan mengikuti pola layer Pertemuan 7. Data mahasiswa tetap berada di PostgreSQL; dokumen portfolio disimpan di collection MongoDB `student_portfolios`.

## Keputusan data dan API

- Setiap dokumen menyimpan satu `student_id`, `headline`, `bio`, `projects`, dan `certificates`.
- `student_id` menggunakan integer dari `students.id` dan dijaga unique index MongoDB.
- Portfolio dan project ID ditampilkan sebagai string pada model Go. Konversi ke BSON ObjectID menjadi tanggung jawab repository.
- Dokumen utama disimpan dengan `_id` ObjectID; project embedded memakai `_id` ObjectID dan dikembalikan sebagai ID hexadecimal.
- Projects dan certificates merupakan embedded documents karena dikirim dan dibaca sebagai bagian dari portfolio. Batas 20 projects berlaku untuk PUT dan POST.
- PUT mengganti keseluruhan representasi portfolio: headline, bio, dan kedua array wajib disertakan. Array kosong berarti menghapus semua item pada array itu.
- Project ID dibuat oleh server. Karena PUT mengganti seluruh array, ID project yang dikirim pada response PUT adalah ID baru. Client harus membaca kembali portfolio setelah PUT sebelum menggunakan project ID untuk DELETE.
- Repository memeriksa batas 20 pada payload PUT sebelum operasi database. Untuk membuat atau mengganti dokumen, repository memakai `$set` dan `$setOnInsert` agar `_id` portfolio lama tetap stabil.
- POST menggunakan filter atomik `projects.19: {$exists: false}` bersamaan dengan `$push`. Batas tidak diperiksa dengan membaca panjang array pada aplikasi.
- Unique index `student_id_unique` menjamin satu dokumen portfolio per student dan mendukung query berdasarkan `student_id`.
- Statistik teknologi dihitung melalui `$unwind` pada `projects`, lalu `$unwind` pada `projects.technologies`, kemudian `$group`, `$sort`, dan `$limit`.

## Validasi dan keamanan request

Body portfolio dan project didekode ke struct bertipe tetap. Decoder menolak field JSON yang tidak dikenal, termasuk operator MongoDB, dan menolak lebih dari satu nilai JSON. Validasi memastikan array projects pada PUT berisi paling banyak 20 item. Pengujian HTTP integrasi memverifikasi request injection dan JSON malformed ditolak sebelum data MongoDB berubah.

## Environment dan status pengerjaan

### Menjalankan dari clean clone

Gunakan Go 1.26.5 atau versi yang kompatibel, PostgreSQL, dan MongoDB. Proyek memakai `go.mod` di root repository, bukan module Go terpisah. Pertemuan 10 mengimpor repository, autentikasi, middleware, dan konfigurasi Pertemuan 7. Parameter variadik `routeExtensions` pada `pertemuan-7-advanced-api-design/config/app.go` diperlukan agar route portfolio didaftarkan sebelum handler 404; pemanggilan Pertemuan 7 tanpa parameter tambahan tetap didukung.

Siapkan database PostgreSQL **khusus test** `pertemuan10_test` beserta schema Pertemuan 7. Untuk database baru yang kosong, urutan migration ialah `001_create_students.sql`, `002_create_achievements.sql`, `003_auth.sql`, `003_rbac.sql`, `004_student_permissions.sql`, lalu `005_users_cursor_pagination.sql`, semuanya dari `pertemuan-7-advanced-api-design/migrations/`. Migration tidak dijalankan otomatis saat aplikasi hidup. Jangan menjalankan migration pada database proyek lain atau mengulang provisioning database test yang sudah tersedia. Startup MongoDB menyiapkan unique index pada collection `portfolio_test.student_portfolios`.

Dari root repository, salin template hanya jika `.env` root belum ada:

```powershell
if (Test-Path .env) {
    throw "File .env sudah ada; periksa dan sesuaikan secara lokal tanpa menimpanya."
}
Copy-Item pertemuan-10-api-mongodb/.env.example .env
```

Isi `DB_USER`, `DB_PASSWORD`, dan `JWT_SECRET` melalui editor lokal. `JWT_SECRET` wajib berupa nilai acak kriptografis minimal 32 karakter, bukan teks contoh. Pertahankan `DB_NAME=pertemuan10_test` dan `MONGO_DATABASE=portfolio_test`. Template sengaja tidak berisi password atau secret yang dapat digunakan. `.env` diabaikan Git. `godotenv.Load()` membaca `.env` dari direktori kerja dan tidak menimpa environment proses yang sudah disetel; periksa override lokal tanpa mencetak nilainya. Jangan memakai `.env` Pertemuan 7 tanpa menyesuaikan nama database secara eksplisit.

Setelah konfigurasi lokal selesai, jalankan dari root:

```powershell
$env:DB_NAME = "pertemuan10_test"
$env:MONGO_DATABASE = "portfolio_test"
$env:APP_PORT = "3331"
go run ./pertemuan-10-api-mongodb
```

Server menggunakan `http://127.0.0.1:3331` untuk konfigurasi ini. Tanpa `APP_PORT`, default aplikasi adalah 3000. Startup melakukan ping PostgreSQL, memuat permission, lalu ping MongoDB dan memastikan index; kegagalan menghentikan startup. Schema dan fixture PostgreSQL harus sudah tersedia sebelum pengujian database. Gunakan endpoint register/login Pertemuan 7 untuk JWT lokal; simpan token hanya di memori proses dan jangan menaruhnya dalam source atau output dokumentasi.

Konfigurasi MongoDB menggunakan `MONGO_URI` dan `MONGO_DATABASE` di environment lokal. Default dan `.env.example` mengarah ke `portfolio_test`; integration test juga menolak database selain `pertemuan10_test` dan `portfolio_test` sebelum operasi tulis. PostgreSQL test memakai `DB_NAME`, `DB_HOST`, dan `DB_PORT` yang ditetapkan secara eksplisit. File `.env` tidak boleh ditambahkan ke repository.

Integration test lokal dijalankan dengan build tag agar operasi database tidak berjalan pada unit test biasa. Sebelum menjalankannya, pastikan `DB_USER`, `DB_PASSWORD`, dan `JWT_SECRET` tersedia hanya pada environment lokal. Test repository memeriksa student sintetis ID 1 melalui repository Pertemuan 7, memastikan collection dan index, membaca portfolio, lalu menguji penolakan duplicate key. Test HTTP memakai JWTManager, middleware, repository, dan kedua database aplikasi; test menambahkan user dan mahasiswa sintetis, serta menyimpan portfolio sintetis tanpa menghapus data uji yang sudah ada.

```powershell
$env:DB_HOST = "127.0.0.1"
$env:DB_PORT = "5432"
$env:DB_NAME = "pertemuan10_test"
$env:MONGO_URI = "mongodb://127.0.0.1:27017"
$env:MONGO_DATABASE = "portfolio_test"
$env:P10_TEST_STUDENT_ID = "1"
go test -tags=integration -run TestPortfolioRepositoryAgainstIsolatedDatabases -v ./pertemuan-10-api-mongodb/app/repository
```

Phase 1 menyiapkan model, DTO, validasi request, dan keputusan kontrak. Phase 2 menyiapkan konfigurasi environment, startup ping, repository, unique index, PUT project guard, atomic POST filter, `$pull`, dan pipeline statistik. Phase 3 menambahkan HTTP API menggunakan response envelope, JWT, dan pemeriksaan ownership dari Pertemuan 7. Startup melakukan ping MongoDB dan berhenti jika server tidak dapat dijangkau atau unique index gagal disiapkan. Saat MongoDB terputus setelah startup, endpoint portfolio memetakan gangguan koneksi atau timeout menjadi `503 MongoDB tidak tersedia`; error database lain tetap menjadi error internal. Endpoint yang hanya memakai PostgreSQL tidak bergantung pada MongoDB dan tetap dapat dilayani selama proses aplikasi berjalan.

## API portfolio

Semua route berada di bawah `/api/v1` dan route portfolio mahasiswa memakai JWT Bearer. Mahasiswa harus ada di PostgreSQL. Pemilik mahasiswa boleh mengakses portfolio miliknya; pengguna lain memerlukan permission Pertemuan 7 yang sesuai (`student:read:any` untuk GET dan `student:update:any` untuk perubahan). Statistik teknologi memerlukan autentikasi, mengikuti route resource Pertemuan 7.

| Method dan path | Perilaku | Status sukses |
| --- | --- | --- |
| `GET /api/v1/students/:id/portfolio` | Membaca portfolio setelah lookup student dan ownership | `200` |
| `PUT /api/v1/students/:id/portfolio` | Upsert dan mengganti seluruh field portfolio serta array projects/certificates | `201` dibuat, `200` diganti |
| `POST /api/v1/students/:id/portfolio/projects` | Menambahkan embedded project secara atomik, maksimal 20 | `201` |
| `DELETE /api/v1/students/:id/portfolio/projects/:projectId` | Menghapus satu project berdasarkan ObjectID string | `204` |
| `GET /api/v1/portfolios/stats/technologies` | Menghitung 10 teknologi paling sering dipakai | `200` |

Request body didekode ke DTO bertipe dan field yang tidak dikenal ditolak sebelum repository mahasiswa maupun portfolio dipanggil. PUT wajib mengirim `headline`, `bio`, dan array `projects` serta `certificates`; array kosong menghapus seluruh itemnya. Project ID dibuat ulang oleh server pada PUT, sehingga klien perlu membaca ulang portfolio sebelum memakai ID untuk DELETE. POST menegakkan batas melalui filter MongoDB `projects.19: {$exists: false}` bersamaan dengan `$push`, bukan melalui pemeriksaan jumlah di application layer.

Untuk unit test dan regresi yang tidak membuka koneksi database, gunakan environment `RUN_DB_INTEGRATION_TESTS=0`, lalu `go test ./pertemuan-10-api-mongodb/... ./pertemuan-7-advanced-api-design/...`. Test HTTP biasa memakai repository test double. Test database Pertemuan 10 hanya aktif dengan build tag `integration`; test PostgreSQL Pertemuan 7 juga memerlukan opt-in `RUN_DB_INTEGRATION_TESTS=1`.

Integration test database memerlukan `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD`, `MONGO_URI`, `MONGO_DATABASE`, `P10_TEST_STUDENT_ID`, dan `JWT_SECRET` yang disetel lokal. Perintah `go test -tags=integration -run 'TestMongoPortfolioDatabasePreflightReadOnly|TestPortfolioHTTPAgainstIsolatedDatabases|TestPortfolioRepositoryAgainstIsolatedDatabases' -v ./pertemuan-10-api-mongodb/database ./pertemuan-10-api-mongodb/route ./pertemuan-10-api-mongodb/app/repository` menjalankan pemeriksaan dan dapat menulis fixture sintetis; jalankan hanya setelah database test serta fixture baseline diperiksa. Pada validasi Phase 3, ketiga integration test tersebut PASS. Evidence Phase 4 mencatat concurrency 18+10, curl nyata, aggregation, explain, dan Compass. Bukti mentah tidak disertakan dalam publikasi source; laporan praktikum memuat bukti yang relevan. Outage/recovery dengan menghentikan database nyata tidak diuji.

Alat inspeksi read-only berada di `evidence/phase-4/tools/mongo-evidence`. Alat ini hanya menerima URI lokal `mongodb://127.0.0.1:27017` dan database `portfolio_test`; mode yang tersedia ialah `snapshot`, `portfolio`, `stats`, `explain`, dan `indexes`. Source client concurrency lama tidak disertakan karena endpoint-nya mengarah tetap ke fixture evidence student 8 dan dapat menambah project jika dijalankan ulang.
