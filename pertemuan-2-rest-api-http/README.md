# Pertemuan 2 - REST API HTTP

Pada pertemuan ini saya mulai membuat REST API sederhana untuk mengelola data student menggunakan Go dan Fiber. Data masih disimpan di memory, jadi akan kembali kosong setiap kali server dijalankan ulang.

## Yang tersedia

API memakai response JSON dengan bentuk `success`, `message`, `data`, dan `meta` untuk daftar data. Data student memiliki field `id`, `nim`, `name`, `grade`, dan `is_active`.

Endpoint utama:

| Method | Endpoint | Keterangan |
| --- | --- | --- |
| GET | `/api/v1/health` | Mengecek server |
| GET | `/api/v1/students` | Mengambil daftar student |
| GET | `/api/v1/students/:id` | Mengambil satu student |
| POST | `/api/v1/students` | Menambah student |
| PUT | `/api/v1/students/:id` | Mengganti seluruh data |
| PATCH | `/api/v1/students/:id` | Mengubah sebagian data |
| DELETE | `/api/v1/students/:id` | Menghapus student |

Daftar student mendukung `page`, `limit`, `search`, `sort`, `order`, dan `is_active`. Field sorting yang diterima adalah `id`, `nim`, `name`, `grade`, dan `is_active`. Request POST, PUT, dan PATCH harus menggunakan `Content-Type: application/json`.

## Validasi dan response

NIM harus positif dan tidak boleh sama dengan data lain. Nama wajib diisi, nilai berada di antara 0 sampai 100, dan PUT harus mengirim semua field yang diperlukan. PATCH hanya mengubah field yang dikirim. API juga membedakan beberapa kondisi melalui status HTTP seperti `400`, `404`, `409`, `415`, dan `422`.

Helper di `helper.go` dipakai untuk membuat response sukses, response list dengan pagination, response `201 Created`, response `204 No Content`, dan response error. Handler student ada di `handler.go`, sedangkan model request dan response ada di `model.go`.

## Menjalankan project

Masuk ke folder API:

```bash
cd pertemuan-2-rest-api-http/api-students
go run .
```

Server berjalan di `http://localhost:3000`. Contoh request:

```http
GET http://localhost:3000/api/v1/students?page=1&limit=10
```

Belum ada database atau test otomatis pada folder ini. Semua data dikelola oleh slice `students` selama proses server masih berjalan.
