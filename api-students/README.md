# Student REST API

REST API sederhana untuk mengelola data mahasiswa menggunakan **Go** dan **Fiber**.

Data masih disimpan di memory, sehingga seluruh data akan hilang ketika server dihentikan atau dijalankan ulang.

## Menjalankan Program

Masuk ke folder `api-students`, lalu jalankan:

```bash
go run .
```

Server akan berjalan di:

```text
http://localhost:3000
```

## Struktur Data Student

Contoh data mahasiswa:

```json
{
  "id": 1,
  "nim": 4343241123,
  "name": "Falah",
  "grade": 80,
  "is_active": true
}
```

Keterangan:

| Field       | Tipe    | Keterangan                                    |
| ----------- | ------- | --------------------------------------------- |
| `id`        | integer | ID mahasiswa yang dibuat otomatis oleh server |
| `nim`       | integer | NIM mahasiswa dan harus unik                  |
| `name`      | string  | Nama mahasiswa                                |
| `grade`     | number  | Nilai mahasiswa dengan rentang 0–100          |
| `is_active` | boolean | Status aktif mahasiswa                        |

## API Contract

| Method | Endpoint               | Parameter                                               | Contoh Body                                                              | Status yang Mungkin          | Contoh Response                                                                         |
| ------ | ---------------------- | ------------------------------------------------------- | ------------------------------------------------------------------------ | ---------------------------- | --------------------------------------------------------------------------------------- |
| GET    | `/api/v1/health`       | -                                                       | -                                                                        | 200                          | `{"success":true,"message":"server berjalan","data":{...}}`                             |
| GET    | `/api/v1/students`     | `page`, `limit`, `search`, `sort`, `order`, `is_active` | -                                                                        | 200                          | `{"success":true,"message":"daftar mahasiswa berhasil diambil","data":[],"meta":{...}}` |
| GET    | `/api/v1/students/:id` | `id` pada path                                          | -                                                                        | 200, 400, 404                | `{"success":true,"message":"mahasiswa ditemukan","data":{...}}`                         |
| POST   | `/api/v1/students`     | -                                                       | `{"nim":4343241123,"name":"Falah","grade":80}`                           | 201, 400, 409, 415, 422      | `{"success":true,"message":"mahasiswa berhasil ditambahkan","data":{...}}`              |
| PUT    | `/api/v1/students/:id` | `id` pada path                                          | `{"nim":4343241123,"name":"Falah Updated","grade":90,"is_active":false}` | 200, 400, 404, 409, 415, 422 | `{"success":true,"message":"data mahasiswa berhasil diganti","data":{...}}`             |
| PATCH  | `/api/v1/students/:id` | `id` pada path                                          | `{"grade":95}`                                                           | 200, 400, 404, 409, 415, 422 | `{"success":true,"message":"data mahasiswa berhasil diperbarui","data":{...}}`          |
| DELETE | `/api/v1/students/:id` | `id` pada path                                          | -                                                                        | 204, 400, 404                | Tidak ada body pada status 204                                                          |

## Query Parameter

Endpoint:

```text
GET /api/v1/students
```

mendukung beberapa query parameter.

| Parameter   | Fungsi                                       | Default              |
| ----------- | -------------------------------------------- | -------------------- |
| `page`      | Menentukan halaman data                      | `1`                  |
| `limit`     | Menentukan jumlah data per halaman           | `10`, maksimal `100` |
| `search`    | Mencari mahasiswa berdasarkan nama           | kosong               |
| `sort`      | Menentukan field pengurutan                  | `id`                 |
| `order`     | Menentukan urutan `asc` atau `desc`          | `asc`                |
| `is_active` | Menyaring mahasiswa berdasarkan status aktif | tidak difilter       |

Field yang dapat digunakan pada parameter `sort`:

```text
id
nim
name
grade
is_active
```

Jika field `sort` tidak valid, API akan menggunakan `id` sebagai nilai default.

### Contoh Pagination

```text
GET /api/v1/students?page=1&limit=2
```

Contoh metadata response:

```json
{
  "page": 1,
  "limit": 2,
  "total": 3,
  "total_pages": 2
}
```

### Contoh Search

```text
GET /api/v1/students?search=ANDI
```

Pencarian tidak membedakan huruf besar dan kecil.

### Contoh Sorting

```text
GET /api/v1/students?sort=grade&order=desc
```

Data akan diurutkan berdasarkan nilai dari yang terbesar ke terkecil.

### Contoh Filter

```text
GET /api/v1/students?is_active=false
```

Hanya mahasiswa dengan status tidak aktif yang akan ditampilkan.

## POST Student

Digunakan untuk menambahkan mahasiswa baru.

```text
POST /api/v1/students
```

Header:

```text
Content-Type: application/json
```

Body:

```json
{
  "nim": 4343241123,
  "name": "Falah",
  "grade": 80
}
```

Jika berhasil, server mengembalikan:

```text
201 Created
```

dan header:

```text
Location: /api/v1/students/1
```

Field `id` dibuat otomatis oleh server dan `is_active` secara default bernilai `true`.

## PUT Student

PUT digunakan untuk mengganti seluruh data mahasiswa.

```text
PUT /api/v1/students/1
```

Seluruh field berikut wajib dikirim:

```text
nim
name
grade
is_active
```

Contoh:

```json
{
  "nim": 4343241123,
  "name": "Falah Updated",
  "grade": 90,
  "is_active": false
}
```

ID tidak ikut berubah karena ID merupakan identitas resource yang ditentukan melalui endpoint.

## PATCH Student

PATCH digunakan untuk mengubah sebagian data mahasiswa.

```text
PATCH /api/v1/students/1
```

Contoh jika hanya ingin mengubah nilai:

```json
{
  "grade": 95
}
```

Pada request tersebut hanya `grade` yang berubah. Field lain seperti `nim`, `name`, dan `is_active` tetap menggunakan nilai sebelumnya.

## DELETE Student

Digunakan untuk menghapus mahasiswa berdasarkan ID.

```text
DELETE /api/v1/students/1
```

Jika berhasil, server mengembalikan:

```text
204 No Content
```

Response `204` tidak memiliki body.

## Status HTTP

| Status                       | Penggunaan                                                         |
| ---------------------------- | ------------------------------------------------------------------ |
| `200 OK`                     | Pengambilan atau perubahan data berhasil                           |
| `201 Created`                | Mahasiswa berhasil ditambahkan                                     |
| `204 No Content`             | Mahasiswa berhasil dihapus                                         |
| `400 Bad Request`            | Request tidak valid, misalnya ID bukan angka atau JSON tidak valid |
| `404 Not Found`              | Mahasiswa tidak ditemukan                                          |
| `409 Conflict`               | NIM sudah digunakan mahasiswa lain                                 |
| `415 Unsupported Media Type` | `Content-Type` bukan `application/json`                            |
| `422 Unprocessable Entity`   | Isi request tidak lolos validasi                                   |

## Validasi

Beberapa aturan validasi yang diterapkan:

- NIM harus berupa angka positif.
- NIM tidak boleh sama dengan mahasiswa lain.
- Nama tidak boleh kosong.
- Grade harus berada pada rentang 0 sampai 100.
- POST mewajibkan `nim`, `name`, dan `grade`.
- PUT mewajibkan `nim`, `name`, `grade`, dan `is_active`.
- PATCH hanya mengubah field yang benar-benar dikirim.

## Header

POST, PUT, dan PATCH menggunakan:

```text
Content-Type: application/json
```

Setiap request juga memiliki header:

```text
X-Request-Id
```

yang digunakan sebagai identitas unik setiap request.

POST yang berhasil juga menghasilkan header:

```text
Location
```

yang menunjukkan alamat resource yang baru dibuat.

## Struktur Project

```text
api-students/
├── main.go
├── model.go
├── helper.go
├── handler.go
└── README.md
```

- `main.go` berisi konfigurasi aplikasi, middleware, dan route.
- `model.go` berisi struct Student, request, response, dan metadata.
- `helper.go` berisi helper response dan pengolahan query parameter.
- `handler.go` berisi proses GET, POST, PUT, PATCH, dan DELETE.

## Sumber Bantuan

Pengerjaan mengacu pada **Modul Praktikum Pertemuan 2 — REST API & HTTP Deep Dive**.

AI digunakan sebagai alat bantu untuk memahami materi, mengecek implementasi kode, membantu proses debugging, dan membantu penyusunan dokumentasi API. Implementasi serta pengujian endpoint dilakukan langsung pada project.
