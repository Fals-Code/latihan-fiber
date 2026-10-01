# Spesifikasi UTS Pemrograman Backend Lanjut
## SIAKAD Mini RESTful API

Dokumen ini adalah source of truth implementasi UTS.
Implementasi tidak boleh mengurangi atau mengubah requirement berikut.

---

## 1. Studi Kasus

Membangun RESTful API backend untuk SIAKAD Mini, yaitu layanan akademik sederhana yang mengelola:

- mahasiswa
- mata kuliah
- Kartu Rencana Studi (KRS)

API wajib terdiri dari tepat 10 endpoint sesuai spesifikasi.

Role pengguna:

1. admin
   - mengelola data mahasiswa
   - mengelola data mata kuliah sesuai endpoint yang tersedia

2. mahasiswa
   - melihat profil sendiri
   - mengambil mata kuliah pada KRS
   - membatalkan mata kuliah pada KRS

---

## 2. Ketentuan Teknis

- Menggunakan Go.
- Framework mengikuti framework pada praktikum, yaitu Fiber.
- Database menggunakan PostgreSQL.
- Migration dan seeder wajib tersedia.
- Seeder minimal:
  - 1 admin
  - 20 mahasiswa
  - 10 mata kuliah
- Password harus disimpan dalam bentuk hash.
- Semua endpoint wajib menggunakan autentikasi token kecuali login.
- Source code berada di repository GitHub dengan riwayat commit bertahap.

Project UTS berada di:

uts-pbl-siakad-mini/

Project menggunakan go.mod dari root repository latihan-fiber.
Jangan membuat go.mod baru di dalam folder UTS.

---

## 3. Model Data Minimal

### users

Kolom utama:

- id
- email
- password
- role

Role hanya:

- admin
- mahasiswa

Relasi:
- 1-to-1 dengan students untuk user mahasiswa

### students

Kolom utama:

- id
- user_id
- nim
- nama
- prodi
- angkatan
- ipk_terakhir
- deleted_at

Ketentuan:

- nim unique
- deleted_at digunakan untuk soft delete
- relasi 1-to-many ke enrollments

### courses

Kolom utama:

- id
- kode_mk
- nama_mk
- sks
- semester
- kuota

Ketentuan:

- kode_mk unique
- relasi 1-to-many ke enrollments

### enrollments

Kolom utama:

- id
- student_id
- course_id
- tahun_akademik
- created_at

Constraint wajib:

UNIQUE(student_id, course_id, tahun_akademik)

---

## 4. Business Rules

### Batas SKS berdasarkan IPK terakhir

- IPK >= 3.00:
  maksimal 24 SKS

- IPK 2.50 sampai 2.99:
  maksimal 21 SKS

- IPK < 2.50:
  maksimal 18 SKS

### Aturan KRS

1. Mahasiswa tidak dapat mengambil mata kuliah yang sama dua kali pada tahun akademik yang sama.

2. Mata kuliah yang kuotanya penuh tidak dapat diambil.

3. Mahasiswa hanya dapat mengakses dan mengubah KRS miliknya sendiri.

---

# 5. Endpoint Wajib

Total akhir harus mengikuti 10 endpoint berikut.

---

## Endpoint 1

POST /api/v1/auth/login

Akses:
Publik

Request body:

- email
  - wajib
  - format email

- password
  - wajib
  - minimal 8 karakter

Response sukses:

HTTP 200

Harus mengembalikan:

- access_token
- token_type
- expires_in
- user:
  - id
  - email
  - role

Error:

- 401 jika kredensial salah
- 422 jika validasi gagal
- 429 jika gagal login lebih dari 5 kali per menit

Rate limiting wajib diimplementasikan.

---

## Endpoint 2

GET /api/v1/auth/me

Akses:
Semua role yang sudah login

Header:

Authorization: Bearer <token>

Response:

Data user yang sedang login.

Jika role mahasiswa, sertakan:

- nim
- nama
- prodi
- angkatan

Error:

- 401 jika token:
  - tidak ada
  - salah
  - kedaluwarsa

Status sukses:

HTTP 200

---

## Endpoint 3

GET /api/v1/students

Akses:
Admin

Query parameter:

- page
  - default 1

- per_page
  - default 10
  - maksimum 50

- prodi

- angkatan

- search
  - mencari nim atau nama

- sort
  - nama
  - -ipk_terakhir

Response:

- array data mahasiswa
- meta pagination:
  - current_page
  - per_page
  - total
  - last_page

Error:

- 401 unauthenticated
- 403 jika bukan admin

Status sukses:

HTTP 200

Mahasiswa yang sudah soft delete tidak boleh muncul.

---

## Endpoint 4

POST /api/v1/students

Akses:
Admin

Request body:

### nim
- wajib
- unique
- tepat 12 digit

### nama
- wajib

### email
- wajib
- unique
- format email

### prodi
- wajib

### angkatan
- wajib
- 4 digit
- tidak boleh melebihi tahun berjalan

### ipk_terakhir
- opsional
- range 0.00 sampai 4.00

Proses wajib:

Dalam SATU database transaction:

1. membuat record users
2. role = mahasiswa
3. password awal = NIM
4. password wajib di-hash
5. membuat record students

Jika salah satu proses gagal, transaction harus rollback.

Error:

- 422 jika NIM duplikat
- 422 jika email duplikat
- 422 jika validasi gagal
- 403 jika bukan admin

Status sukses:

HTTP 201

---

## Endpoint 5

GET /api/v1/students/{id}

Akses:

- admin
- mahasiswa hanya untuk data dirinya sendiri

Response:

Data mahasiswa beserta:

- daftar mata kuliah yang diambil
- total_sks
- batas_sks

Error:

- 403 jika mahasiswa mencoba mengakses mahasiswa lain
- 404 jika mahasiswa tidak ditemukan
- 404 jika mahasiswa sudah soft delete

Status sukses:

HTTP 200

---

## Endpoint 6

PUT /api/v1/students/{id}

Akses:
Admin

Field yang dapat diperbarui:

- nama
- prodi
- angkatan
- ipk_terakhir

NIM TIDAK boleh diubah.

Error:

- 404 jika data tidak ditemukan
- 422 jika validasi gagal
- 403 jika bukan admin

Status sukses:

HTTP 200

---

## Endpoint 7

DELETE /api/v1/students/{id}

Akses:
Admin

Proses:

Soft delete mahasiswa.

Isi:

deleted_at = timestamp

Jangan hard delete.

Mahasiswa yang sudah dihapus:

- tidak muncul di GET /students
- tidak dapat login

Error:

- 404 jika tidak ditemukan
- 403 jika bukan admin

Status sukses:

HTTP 204 No Content

Response 204 tidak boleh memiliki JSON body.

---

## Endpoint 8

GET /api/v1/courses

Akses:
Semua role yang sudah login

Query parameter:

- semester

- search
  - mencari kode_mk atau nama_mk

- available=true
  - hanya menampilkan mata kuliah yang kuotanya belum penuh

Setiap mata kuliah pada response harus menyertakan:

- terisi
- sisa_kuota

Nilainya dihitung berdasarkan tabel enrollments.

Error:

- 401 unauthenticated

Status sukses:

HTTP 200

---

## Endpoint 9

POST /api/v1/enrollments

Akses:
Mahasiswa

Request body:

### course_id
- wajib
- course harus tersedia

### tahun_akademik
- wajib
- format contoh:
  2026/2027-Ganjil

Proses WAJIB dilakukan dalam SATU database transaction.

Dalam transaction:

1. identifikasi mahasiswa yang sedang login
2. cek course
3. lakukan row locking terhadap data yang relevan untuk mencegah race condition
4. cek apakah course yang sama sudah pernah diambil mahasiswa pada tahun akademik tersebut
5. cek kapasitas/kuota
6. hitung total SKS mahasiswa pada tahun akademik tersebut
7. tentukan batas SKS berdasarkan ipk_terakhir
8. pastikan SKS course baru tidak membuat total melebihi batas
9. insert enrollment jika seluruh rule valid

Concurrency harus ditangani dengan benar.

Error:

### 409
Jika mata kuliah sudah pernah diambil pada tahun akademik yang sama.

### 422
Jika:

- kuota penuh
- total SKS akan melebihi batas

Untuk error batas SKS, message harus menyebutkan sisa SKS yang masih dapat diambil.

### 403
Jika bukan mahasiswa.

Status sukses:

HTTP 201

---

## Endpoint 10

DELETE /api/v1/enrollments/{id}

Akses:
Mahasiswa, hanya enrollment miliknya sendiri.

Proses:

- hapus record enrollment
- setelah enrollment dihapus, slot kuota otomatis tersedia kembali berdasarkan perhitungan jumlah enrollment

Error:

- 403 jika enrollment milik mahasiswa lain
- 404 jika enrollment tidak ditemukan

Status sukses:

HTTP 204 No Content

Response 204 tidak boleh memiliki JSON body.

---

# 6. Format Response

Semua response JSON harus memiliki struktur yang konsisten.

Contoh sukses:

{
  "success": true,
  "message": "Data mahasiswa berhasil diambil",
  "data": [
    {
      "id": 1,
      "nim": "187221000001",
      "nama": "Rina Putri",
      "prodi": "Sistem Informasi",
      "angkatan": 2022,
      "ipk_terakhir": 3.45
    }
  ],
  "meta": {
    "current_page": 1,
    "per_page": 10,
    "total": 20,
    "last_page": 2
  }
}

Contoh validation error:

{
  "success": false,
  "message": "Validasi gagal",
  "errors": {
    "nim": [
      "NIM sudah terdaftar"
    ],
    "email": [
      "Format email tidak valid"
    ]
  }
}

---

# 7. HTTP Status Code Wajib

Implementasi harus dapat menggunakan status berikut sesuai kondisi:

- 200 OK
- 201 Created
- 204 No Content
- 401 Unauthorized
- 403 Forbidden
- 404 Not Found
- 409 Conflict
- 422 Unprocessable Entity
- 429 Too Many Requests
- 500 Internal Server Error

Untuk HTTP 500 pada production:

JANGAN membocorkan stack trace atau detail internal aplikasi.

---

# 8. Ketentuan Implementasi Project Ini

Folder yang boleh dimodifikasi:

uts-pbl-siakad-mini/

Folder berikut adalah materi/referensi praktikum dan TIDAK boleh dimodifikasi:

- pertemuan-1-dasar-go
- pertemuan-2-rest-api-http
- pertemuan-3-database-repository
- pertemuan-4-clean-architecture
- pertemuan-5-authentication-security
- pertemuan-6-authorization-rbac
- pertemuan-7-advanced-api-design

Boleh membaca implementasinya sebagai referensi.

Gunakan terutama materi:

- Clean Architecture
- Authentication & Security
- Authorization / RBAC
- Advanced API Design

Jangan menambahkan fitur di luar kebutuhan UTS jika tidak diperlukan.

Hindari overengineering.

Implementasi harus cukup sederhana untuk project mahasiswa tetapi tetap benar dan dapat dipertanggungjawabkan.

---

# 9. Database Safety

Database UTS:

siakad_mini

Jangan pernah menggunakan atau memodifikasi:

praktikum_backend

atau database praktikum lain.

Migration dan seeder tidak boleh otomatis dijalankan saat aplikasi start.

Migration dan seeder harus dijalankan secara eksplisit.

main.go hanya bertanggung jawab untuk:

- load environment
- setup logger bila diperlukan
- membuka koneksi PostgreSQL
- bootstrap dependency
- membuat Fiber app
- register route
- menjalankan server

---
