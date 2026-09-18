# Student REST API — Database & Repository Pattern

Project ini merupakan pengembangan dari Student REST API pada Pertemuan 2.
Pada Pertemuan 3, penyimpanan data mahasiswa dipindahkan dari in-memory slice ke PostgreSQL dan akses database dipisahkan menggunakan Repository Pattern.

Dengan perubahan ini, data mahasiswa tetap tersimpan meskipun server Go dihentikan dan dijalankan kembali.

## Teknologi

- Go
- Fiber v2
- PostgreSQL 18
- pgx v5
- pgxpool
- godotenv

## Struktur Project

```text
pertemuan-3-database-repository/
│
├── app/
│   ├── model/
│   │   └── student.go
│   │
│   └── repository/
│       └── student_repository.go
│
├── config/
│   └── env.go
│
├── database/
│   └── postgres.go
│
├── migrations/
│   └── 001_create_students.sql
│
├── .env.example
├── handler.go
├── helper.go
├── main.go
└── README.md
```

## Database Schema

Data mahasiswa disimpan pada tabel `students`.

```sql
CREATE TABLE IF NOT EXISTS students (
    id SERIAL PRIMARY KEY,
    nim BIGINT NOT NULL UNIQUE,
    name VARCHAR(100) NOT NULL,
    grade NUMERIC(5,2) NOT NULL CHECK (grade >= 0 AND grade <= 100),
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS students_name_lower_idx
    ON students (LOWER(name));
```

Struktur utama tabel:

| Field        | Tipe         | Keterangan                   |
| ------------ | ------------ | ---------------------------- |
| `id`         | SERIAL       | Primary key                  |
| `nim`        | BIGINT       | NIM mahasiswa dan harus unik |
| `name`       | VARCHAR(100) | Nama mahasiswa               |
| `grade`      | NUMERIC(5,2) | Nilai 0–100                  |
| `is_active`  | BOOLEAN      | Status mahasiswa             |
| `created_at` | TIMESTAMPTZ  | Waktu data dibuat            |

Selain primary key dan unique index pada `nim`, terdapat index tambahan pada `LOWER(name)` untuk mendukung pencarian nama.

## Environment Variables

Konfigurasi database tidak ditulis langsung di source code.

Buat file `.env` berdasarkan `.env.example`.

Contoh:

```env
APP_PORT=3000
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password_database
DB_NAME=praktikum_backend
DB_SSLMODE=disable
DB_MAX_CONNS=10
```

File `.env` sudah dimasukkan ke `.gitignore` sehingga credential lokal tidak ikut masuk ke repository.

## Setup dari Nol

### 1. Clone repository

```bash
git clone https://github.com/Fals-Code/latihan-fiber.git
cd latihan-fiber
```

### 2. Download dependency

```bash
go mod download
```

### 3. Buat database PostgreSQL

Masuk ke PostgreSQL:

```bash
psql -U postgres
```

Buat database:

```sql
CREATE DATABASE praktikum_backend;
```

Keluar dari PostgreSQL:

```text
\q
```

### 4. Masuk ke project Pertemuan 3

```bash
cd pertemuan-3-database-repository
```

### 5. Buat file `.env`

Salin `.env.example` menjadi `.env`, kemudian isi konfigurasi PostgreSQL sesuai environment lokal.

### 6. Jalankan migration

```bash
psql -U postgres -d praktikum_backend -f migrations/001_create_students.sql
```

Untuk memastikan tabel sudah terbentuk:

```bash
psql -U postgres -d praktikum_backend
```

Kemudian:

```text
\d students
```

### 7. Jalankan aplikasi

```bash
go run .
```

Server secara default berjalan pada:

```text
http://localhost:3000
```

## Health Check

```http
GET /api/v1/health
```

Ketika aplikasi dan PostgreSQL tersedia:

```text
200 OK
```

Contoh response:

```json
{
  "success": true,
  "message": "server berjalan",
  "data": {
    "timestamp": "2026-08-24T11:00:00+07:00"
  }
}
```

Endpoint health melakukan `Ping()` ke PostgreSQL.

Jika PostgreSQL tidak tersedia, API mengembalikan:

```text
503 Service Unavailable
```

```json
{
  "success": false,
  "message": "database tidak tersedia"
}
```

## Endpoint Student

| Method | Endpoint               | Fungsi                             |
| ------ | ---------------------- | ---------------------------------- |
| GET    | `/api/v1/students/`    | Mengambil daftar mahasiswa         |
| GET    | `/api/v1/students/:id` | Mengambil mahasiswa berdasarkan ID |
| POST   | `/api/v1/students/`    | Menambahkan mahasiswa              |
| PUT    | `/api/v1/students/:id` | Mengganti seluruh data mahasiswa   |
| PATCH  | `/api/v1/students/:id` | Mengubah sebagian data mahasiswa   |
| DELETE | `/api/v1/students/:id` | Menghapus mahasiswa                |

## Contoh Request

### Menambahkan mahasiswa

```http
POST /api/v1/students/
Content-Type: application/json
```

```json
{
  "nim": 20260001,
  "name": "Budi",
  "grade": 85
}
```

Response:

```text
201 Created
```

```json
{
  "success": true,
  "message": "mahasiswa berhasil ditambahkan",
  "data": {
    "id": 1,
    "nim": 20260001,
    "name": "Budi",
    "grade": 85,
    "is_active": true
  }
}
```

### Mengambil mahasiswa

```http
GET /api/v1/students/1
```

### Mengganti seluruh data

```http
PUT /api/v1/students/1
Content-Type: application/json
```

```json
{
  "nim": 20260001,
  "name": "Budi Santoso",
  "grade": 90,
  "is_active": true
}
```

### Mengubah sebagian data

```http
PATCH /api/v1/students/1
Content-Type: application/json
```

```json
{
  "grade": 95
}
```

### Menghapus mahasiswa

```http
DELETE /api/v1/students/1
```

Response:

```text
204 No Content
```

## Query Parameter

Endpoint:

```http
GET /api/v1/students/
```

mendukung beberapa query parameter.

| Parameter   | Contoh           | Fungsi                     |
| ----------- | ---------------- | -------------------------- |
| `page`      | `page=1`         | Halaman data               |
| `limit`     | `limit=10`       | Jumlah data per halaman    |
| `search`    | `search=budi`    | Pencarian berdasarkan nama |
| `sort`      | `sort=grade`     | Field untuk sorting        |
| `order`     | `order=desc`     | Urutan `asc` atau `desc`   |
| `is_active` | `is_active=true` | Filter status mahasiswa    |

Field sorting yang diperbolehkan:

```text
id
nim
name
grade
is_active
```

Contoh:

```http
GET /api/v1/students/?search=bud
```

```http
GET /api/v1/students/?sort=grade&order=desc
```

```http
GET /api/v1/students/?page=1&limit=2
```

```http
GET /api/v1/students/?is_active=true
```

Pencarian, filter, sorting, pagination, dan perhitungan total data dilakukan pada PostgreSQL menggunakan:

- `ILIKE`
- `WHERE`
- `ORDER BY`
- `LIMIT`
- `OFFSET`
- `COUNT(*)`

## Repository Pattern

Akses PostgreSQL tidak dilakukan langsung dari HTTP handler.

Alur aplikasi:

```text
HTTP Request
     ↓
Fiber Handler
     ↓
StudentRepository
     ↓
pgxpool
     ↓
PostgreSQL
```

`StudentRepository` menyediakan operasi:

```go
FindAll(...)
FindByID(...)
Create(...)
Update(...)
Delete(...)
```

Dengan pemisahan ini, handler bertanggung jawab terhadap HTTP request, validasi, dan HTTP response, sedangkan repository bertanggung jawab terhadap akses data dan SQL.

Repository tidak bergantung pada Fiber.

## SQL Safety

Nilai yang berasal dari client dikirim ke PostgreSQL menggunakan parameter SQL seperti:

```sql
WHERE id = $1
```

dan:

```sql
VALUES ($1, $2, $3, $4)
```

Nama kolom untuk `ORDER BY` tidak dimasukkan langsung dari input client. Repository menggunakan whitelist field sorting agar query tetap aman.

## Error Mapping

Beberapa error database diterjemahkan menjadi error domain sebelum dikirim menjadi HTTP response.

### Data tidak ditemukan

Ketika PostgreSQL menghasilkan `pgx.ErrNoRows`:

```text
pgx.ErrNoRows
→ ErrNotFound
→ 404 Not Found
```

### NIM duplikat

Kolom `nim` memiliki constraint `UNIQUE`.

Ketika PostgreSQL menghasilkan SQLSTATE `23505`:

```text
PostgreSQL 23505
→ ErrDuplicate
→ 409 Conflict
```

Contoh response:

```json
{
  "success": false,
  "message": "NIM sudah digunakan"
}
```

### Database tidak tersedia

Health endpoint:

```text
PostgreSQL tidak tersedia
→ 503 Service Unavailable
```

Sedangkan kegagalan database pada operasi repository yang tidak memiliki mapping khusus dikembalikan sebagai:

```text
500 Internal Server Error
```

`503 Service Unavailable` digunakan pada health check karena server aplikasi masih berjalan, tetapi dependency utama berupa PostgreSQL sedang tidak tersedia.

## Persistence

Pada versi Pertemuan 2, data mahasiswa disimpan pada slice di memory sehingga data akan hilang ketika aplikasi dihentikan.

Pada Pertemuan 3, data disimpan di PostgreSQL sehingga data tetap tersedia setelah aplikasi Go dihentikan dan dijalankan kembali.

Pengujian persistence dilakukan dengan:

1. Menambahkan data mahasiswa melalui API.
2. Memastikan data tersimpan.
3. Menghentikan server Go.
4. Menjalankan server kembali.
5. Mengakses endpoint `GET /api/v1/students/`.
6. Memastikan data sebelumnya masih tersedia.

## HTTP Status

Beberapa HTTP status yang digunakan:

| Status                       | Penggunaan                                  |
| ---------------------------- | ------------------------------------------- |
| `200 OK`                     | Request berhasil                            |
| `201 Created`                | Data berhasil dibuat                        |
| `204 No Content`             | Data berhasil dihapus                       |
| `400 Bad Request`            | Request atau ID tidak valid                 |
| `404 Not Found`              | Mahasiswa tidak ditemukan                   |
| `409 Conflict`               | NIM sudah digunakan                         |
| `415 Unsupported Media Type` | Request body bukan JSON                     |
| `422 Unprocessable Entity`   | Validasi data gagal                         |
| `500 Internal Server Error`  | Operasi server/database gagal               |
| `503 Service Unavailable`    | PostgreSQL tidak tersedia pada health check |

## Validasi Project

Project dapat diperiksa dengan:

```bash
go test ./...
```

dan:

```bash
go vet ./...
```

## Penggunaan AI

Dalam pengerjaan praktikum ini, AI digunakan sebagai alat bantu untuk memahami materi, mendiskusikan struktur project, membantu proses debugging, serta memberikan masukan terhadap implementasi.

Kode tetap dijalankan, diuji, dan diverifikasi secara langsung pada environment lokal sebelum digunakan pada hasil akhir praktikum.
