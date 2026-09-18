# Pertemuan 2 - REST API HTTP

Folder `api-students` berisi latihan membuat REST API student dengan Go dan Fiber. Data disimpan di memory menggunakan slice, jadi data tidak bertahan setelah server dihentikan.

## Endpoint

| Method | Endpoint | Keterangan |
| --- | --- | --- |
| GET | `/api/v1/health` | Mengecek server |
| GET | `/api/v1/students` | Mengambil daftar student |
| GET | `/api/v1/students/:id` | Mengambil satu student |
| POST | `/api/v1/students` | Menambah student |
| PUT | `/api/v1/students/:id` | Mengganti data student |
| PATCH | `/api/v1/students/:id` | Mengubah sebagian data |
| DELETE | `/api/v1/students/:id` | Menghapus student |

Daftar student bisa memakai `page`, `limit`, `search`, `sort`, `order`, dan `is_active`. Request yang memiliki body wajib memakai `application/json`.

## Struktur singkat

- `main.go` membuat aplikasi Fiber, middleware, dan route.
- `handler.go` menangani proses list, detail, tambah, ubah, dan hapus student.
- `model.go` berisi struct student, request, response, dan metadata pagination.
- `helper.go` berisi pembuat response, validasi query, sorting, dan pagination.

Validasi yang ada antara lain NIM harus positif dan unik, nama tidak boleh kosong, nilai harus 0–100, serta PUT dan PATCH memiliki aturan field masing-masing. Response memakai status seperti `200`, `201`, `204`, `400`, `404`, `409`, `415`, dan `422` sesuai kondisi request.

## Menjalankan project

```bash
cd pertemuan-2-rest-api-http/api-students
go run .
```

Server berjalan di `http://localhost:3000`. Contoh:

```http
GET http://localhost:3000/api/v1/students?page=1&limit=10
```
