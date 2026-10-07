# Laporan Ujian Tengah Semester

## Implementasi RESTful API Back End SIAKAD Mini

**Nama:** Ahmad Mathlaul Falah  
**NIM:** 434241123  
**Kelas:** C-8  
**Program Studi:** Teknik Informatika  
**Dosen Pengampu:** MOH. NASRUL AZIZ

> Versi laporan final dengan seluruh bukti gambar dan tangkapan layar: https://docs.google.com/document/d/1aGWMPVHizbQAKcqc0YeaJRCsTNAZ4bHc3-y_zOjO598/edit

## BAB I Pendahuluan

### 1.1 Latar Belakang

Pada UTS ini, SIAKAD Mini dikembangkan sebagai RESTful API back end untuk menangani data mahasiswa, mata kuliah, dan Kartu Rencana Studi (KRS). Akses sistem dibedakan menjadi dua role. Admin mengelola data mahasiswa, sedangkan mahasiswa dapat melihat data akademiknya sendiri serta mengambil atau membatalkan mata kuliah pada KRS.

Back end dibangun menggunakan Go dan Fiber, dengan PostgreSQL sebagai database. Agar alur program lebih mudah dibaca dan diuji, kode dipisahkan ke dalam handler, service, dan repository. Sistem juga menerapkan autentikasi token, pembatasan akses berdasarkan role, validasi data, soft delete, migration, seeder, transaksi database, dan pengujian.

### 1.2 Tujuan

- Membangun 10 endpoint RESTful API SIAKAD Mini sesuai spesifikasi UTS.
- Menerapkan autentikasi token dan pembatasan akses untuk role admin dan mahasiswa.
- Mengelola data mahasiswa, mata kuliah, dan KRS sesuai aturan yang ditentukan pada soal.
- Menyiapkan migration dan seeder PostgreSQL agar struktur serta data awal database dapat dibuat ulang.
- Memisahkan kode ke beberapa layer agar alur program lebih mudah dibaca, diuji, dan dikembangkan.

### 1.3 Ruang Lingkup

Ruang lingkup pengerjaan mengikuti sepuluh endpoint API versi v1 yang ditentukan pada soal, mulai dari login dan profil pengguna hingga pengelolaan mahasiswa, mata kuliah, dan KRS. Front end tidak dibuat karena tugas berfokus pada implementasi layanan back end.

### 1.4 Repositori Proyek

- Repository: https://github.com/Fals-Code/latihan-fiber/tree/main/uts-pbl-siakad-mini
- Folder project: `uts-pbl-siakad-mini`
- Folder laporan: `UTS`
- Branch: `main`

Pengerjaan proyek dilakukan secara bertahap. Setiap bagian utama, mulai dari setup database, autentikasi, pengelolaan mahasiswa, mata kuliah, KRS, hingga dokumentasi, dicatat melalui commit agar riwayat pengerjaan dapat ditelusuri.

## BAB II Analisis Kebutuhan dan Perancangan

### 2.1 Ketentuan Teknis

- Bahasa pemrograman: Go 1.26.5
- Framework: Fiber v2.52.14
- Database: PostgreSQL
- Driver database: pgx/v5
- Autentikasi: JWT HS256
- Password hashing: bcrypt
- Konfigurasi environment: godotenv + process environment

### 2.2 Model Data

Model data terdiri dari empat tabel utama:

1. `users`: id, email, password, role. Email unik dan role dibatasi admin/mahasiswa.
2. `students`: user_id, nim, nama, prodi, angkatan, ipk_terakhir, deleted_at. user_id dan NIM unik serta mendukung soft delete.
3. `courses`: kode_mk, nama_mk, sks, semester, kuota. kode_mk unik dan kuota harus lebih dari 0.
4. `enrollments`: student_id, course_id, tahun_akademik, created_at. Kombinasi student-course-tahun akademik harus unik.

### 2.3 Aturan Bisnis KRS

1. IPK >= 3,00 maksimal 24 SKS; IPK 2,50-2,99 maksimal 21 SKS; IPK < 2,50 maksimal 18 SKS.
2. Mahasiswa tidak dapat mengambil mata kuliah yang sama dua kali pada tahun akademik yang sama.
3. Mata kuliah dengan kuota penuh tidak dapat diambil.
4. Mahasiswa hanya dapat mengakses detail dirinya sendiri dan hanya dapat menghapus enrollment miliknya sendiri.

### 2.4 Daftar 10 Endpoint

1. `POST /api/v1/auth/login` — publik — login dan menghasilkan access token — 200.
2. `GET /api/v1/auth/me` — semua role — profil pengguna aktif — 200.
3. `GET /api/v1/students` — admin — daftar mahasiswa + pagination/filter/search — 200.
4. `POST /api/v1/students` — admin — tambah mahasiswa dan akun user — 201.
5. `GET /api/v1/students/:id` — admin / mahasiswa pemilik — detail mahasiswa + total/batas SKS — 200.
6. `PUT /api/v1/students/:id` — admin — perbarui data mahasiswa — 200.
7. `DELETE /api/v1/students/:id` — admin — soft delete mahasiswa — 204.
8. `GET /api/v1/courses` — semua role terautentikasi — daftar mata kuliah + sisa kuota — 200.
9. `POST /api/v1/enrollments` — mahasiswa — tambah mata kuliah ke KRS — 201.
10. `DELETE /api/v1/enrollments/:id` — mahasiswa pemilik — batalkan mata kuliah dari KRS — 204.

## BAB III Implementasi Sistem

### 3.1 Struktur Proyek

Kode dipisahkan menjadi handler, service, dan repository agar proses menerima request, menjalankan aturan bisnis, dan mengakses database tidak berada dalam satu bagian. File `main.go` menjadi titik awal aplikasi untuk memuat konfigurasi, membuka koneksi database, menyusun dependency antar-layer, membuat instance Fiber, dan mendaftarkan route.

### 3.2 Konfigurasi Database, Migration, dan Seeder

Migration dan seeder tidak dijalankan otomatis ketika server dinyalakan. Keduanya dijalankan melalui command `dbsetup` sehingga pembuatan tabel dan pengisian data awal dapat dilakukan secara terpisah saat menyiapkan database.

```bash
createdb -U postgres siakad_mini
go run ./uts-pbl-siakad-mini/cmd/dbsetup migrate
go run ./uts-pbl-siakad-mini/cmd/dbsetup seed
```

Implementasi migration database terdiri dari empat file yang membentuk tabel users, students, courses, dan enrollments. Seeder dibuat idempotent dan mengisi data awal berupa 1 akun admin, 20 akun mahasiswa beserta data students, serta 10 mata kuliah. Seluruh password disimpan dalam bentuk hash bcrypt, sedangkan password awal mahasiswa menggunakan NIM masing-masing.

### 3.3 Autentikasi dan Otorisasi

Saat login, server memvalidasi email dan password lalu mencocokkan password dengan hash bcrypt. Jika valid, server menerbitkan JWT menggunakan HS256. Semua endpoint selain login memerlukan Bearer token. Middleware kemudian memeriksa token, status akun, dan role pengguna. Percobaan login gagal dibatasi maksimal lima kali per menit untuk setiap IP.

### 3.4 Pengelolaan Mahasiswa

Fitur mahasiswa mencakup pagination, filter prodi dan angkatan, pencarian berdasarkan NIM atau nama, serta pengurutan berdasarkan nama atau IPK. Admin dapat menambah mahasiswa, melihat detail, memperbarui data, dan melakukan soft delete.

### 3.5 Daftar Mata Kuliah

Endpoint mata kuliah mendukung filter semester dan pencarian berdasarkan kode atau nama mata kuliah. Parameter `available=true` digunakan untuk menampilkan mata kuliah yang kuotanya masih tersedia. Response juga menyertakan jumlah `terisi` dan `sisa_kuota` agar kondisi kuota dapat langsung diketahui.

### 3.6 Pengelolaan KRS

Pengambilan KRS dijalankan di dalam transaction PostgreSQL. Sebelum enrollment ditambahkan, data mahasiswa dan mata kuliah dikunci dengan `SELECT ... FOR UPDATE`, kemudian sistem memeriksa duplikasi, kuota, dan batas SKS. Pembatalan KRS hanya dapat dilakukan pada enrollment milik mahasiswa yang sedang login.

### 3.7 Registrasi Route

Seluruh endpoint menggunakan prefix `/api/v1`. Middleware autentikasi dan role dipasang sesuai kebutuhan masing-masing route.

## BAB IV Pengujian

### 4.1 Pengujian Otomatis

Pengujian otomatis dibuat untuk middleware, route, handler, service, dan repository. Perintah berikut dijalankan dari root repository:

```bash
go test ./uts-pbl-siakad-mini/...
```

Hasil terminal menunjukkan seluruh package yang memiliki file test berstatus `ok` dan tidak ada pengujian yang gagal.

### 4.2 Skenario Pengujian Endpoint

Pengujian endpoint dilakukan untuk memastikan setiap fungsi memberikan respons dan status HTTP sesuai spesifikasi UTS. Pengujian mencakup kondisi berhasil maupun beberapa kondisi gagal yang berkaitan dengan autentikasi, otorisasi, validasi input, data yang tidak ditemukan, duplikasi KRS, dan pembatasan percobaan login.

1. `POST /api/v1/auth/login`: berhasil 200, password salah 401, validasi request gagal 422, dan lebih dari lima kegagalan login dalam satu menit 429.
2. `GET /api/v1/auth/me`: token valid 200; tanpa token, token tidak valid, atau token kedaluwarsa 401.
3. `GET /api/v1/students`: admin 200, mahasiswa 403, query tidak valid 422.
4. `POST /api/v1/students`: data valid 201, NIM/email duplikat atau input tidak valid 422, mahasiswa 403.
5. `GET /api/v1/students/:id`: admin/pemilik 200, bukan pemilik 403, data tidak ditemukan atau soft deleted 404.
6. `PUT /api/v1/students/:id`: admin 200, perubahan NIM/field tidak valid 422, non-admin 403.
7. `DELETE /api/v1/students/:id`: admin 204, non-admin 403, data tidak ditemukan 404.
8. `GET /api/v1/courses`: token valid 200, tanpa token 401.
9. `POST /api/v1/enrollments`: mahasiswa 201, duplikat 409, kuota/SKS 422, admin 403.
10. `DELETE /api/v1/enrollments/:id`: pemilik 204, bukan pemilik 403, data tidak ditemukan 404.

### 4.3 Hasil Pengujian API

Pengujian manual melalui Postman dilakukan untuk seluruh endpoint utama. Bukti tangkapan layar lengkap tersedia pada versi Google Docs final yang ditautkan di bagian awal laporan.

### 4.4 Verifikasi Database

Verifikasi database dilakukan setelah seluruh pengujian endpoint selesai. Query akhir menunjukkan 22 data pada tabel users, 20 mahasiswa aktif, 10 mata kuliah, dan 0 enrollment. Jumlah users menjadi 22 karena satu akun mahasiswa dibuat saat pengujian `POST /students`. Data student akun tersebut kemudian di-soft delete, sedangkan akun user tetap tersimpan. Enrollment kembali 0 setelah KRS uji dibatalkan.

## BAB V Penutup

### 5.1 Kesimpulan

Pengujian melalui `go test` dan Postman menunjukkan bahwa 10 endpoint yang ditentukan pada soal UTS dapat dijalankan sesuai skenario utama yang diuji. API dibangun menggunakan Go, Fiber, dan PostgreSQL, dengan JWT untuk autentikasi, bcrypt untuk password, pembatasan akses berdasarkan role, migration dan seeder, soft delete mahasiswa, informasi kuota mata kuliah, serta pengelolaan KRS.

Pada alur KRS, sistem mencegah pengambilan mata kuliah yang sama pada tahun akademik yang sama, mengecek kuota dan batas SKS berdasarkan IPK, serta memastikan mahasiswa hanya dapat membatalkan enrollment miliknya sendiri. Proses pengambilan mata kuliah menggunakan transaction dan row locking untuk menjaga konsistensi data.

## Lampiran Akun Seed untuk Pengujian

- Admin: `admin@siakad.local` / `Admin123!`
- Mahasiswa: `<NIM>@student.siakad.local` / password awal = NIM masing-masing

Catatan: akun seed digunakan untuk pengujian lokal, bukan produksi.
