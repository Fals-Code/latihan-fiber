# Pertemuan 4 - Clean Architecture

Pada pertemuan ini project student mulai dipisahkan ke beberapa layer supaya handler tidak menangani semua hal sendiri. Alur utamanya dibuat dari route ke service, lalu service memakai repository untuk mengakses PostgreSQL.

## Struktur project

- `app/model` berisi model student dan achievement.
- `app/repository` berisi akses data PostgreSQL untuk student dan achievement.
- `app/service` berisi aturan dan proses bisnis.
- `route` mendaftarkan endpoint ke service yang sesuai.
- `middleware` menangani kebutuhan umum seperti pemeriksaan JSON.
- `config` menyiapkan environment, logger, dan aplikasi Fiber.
- `database` membuat connection pool PostgreSQL.
- `migrations` berisi pembuatan tabel student dan achievement.

`main.go` merakit dependency dari repository ke service, kemudian memasangnya ke aplikasi Fiber. Aplikasi juga menunggu signal shutdown dan mencoba menutup server dengan rapi.

## Endpoint

Selain health check dan operasi student, tahap ini menambahkan resource achievement.

| Method | Endpoint | Keterangan |
| --- | --- | --- |
| GET | `/api/v1/health` | Mengecek koneksi database |
| GET | `/api/v1/students/` | Daftar student |
| GET | `/api/v1/students/:id` | Detail student |
| POST | `/api/v1/students/` | Menambah student |
| PUT | `/api/v1/students/:id` | Mengganti student |
| PATCH | `/api/v1/students/:id` | Mengubah sebagian student |
| DELETE | `/api/v1/students/:id` | Menghapus student |
| GET | `/api/v1/achievements/` | Daftar achievement |
| GET | `/api/v1/achievements/:id` | Detail achievement |
| POST | `/api/v1/achievements/` | Menambah achievement |
| PUT | `/api/v1/achievements/:id` | Mengganti achievement |
| PATCH | `/api/v1/achievements/:id` | Mengubah sebagian achievement |
| DELETE | `/api/v1/achievements/:id` | Menghapus achievement |

Detail aturan validasi student dan achievement ada di service rules. Beberapa fungsi service dan aturan tersebut sudah memiliki test di `app/service`.

## Database dan menjalankan project

Migration yang dipakai ada di folder `migrations`, termasuk tabel student dan achievement. Siapkan PostgreSQL, buat `.env` dari `.env.example`, lalu jalankan migration yang diperlukan.

Dari folder ini jalankan:

```bash
go run .
```

Untuk menjalankan test:

```bash
go test ./...
```
