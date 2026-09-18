# Pertemuan 3 - Database dan Repository

Pada pertemuan ini data student tidak lagi disimpan di slice seperti pertemuan sebelumnya. Data dipindahkan ke PostgreSQL dan akses database dipisahkan ke `StudentRepository`.

## Yang dikerjakan

- Membuka connection pool PostgreSQL dari konfigurasi environment.
- Menyimpan data student di tabel `students`.
- Memindahkan operasi list, detail, tambah, ubah, dan hapus ke repository.
- Menangani error data tidak ditemukan dan NIM duplikat dari hasil repository.
- Menjaga response API dan validasi request tetap berada di handler.

Migration `migrations/001_create_students.sql` membuat tabel student dengan `id`, `nim`, `name`, `grade`, `is_active`, dan `created_at`. NIM dibuat unik dan nama memiliki index untuk pencarian. File pentingnya ada di `app/model`, `app/repository`, `database`, `config`, `handler.go`, dan `main.go`.

## Endpoint

Route student masih memakai endpoint berikut:

| Method | Endpoint | Keterangan |
| --- | --- | --- |
| GET | `/api/v1/health` | Mengecek koneksi server dan database |
| GET | `/api/v1/students/` | Mengambil daftar student |
| GET | `/api/v1/students/:id` | Mengambil satu student |
| POST | `/api/v1/students/` | Menambah student |
| PUT | `/api/v1/students/:id` | Mengganti data student |
| PATCH | `/api/v1/students/:id` | Mengubah sebagian data |
| DELETE | `/api/v1/students/:id` | Menghapus student |

Query list mendukung pagination, pencarian nama, sorting, dan filter `is_active` seperti pada tahap sebelumnya.

## Menjalankan project

Siapkan PostgreSQL dan buat file `.env` berdasarkan `.env.example`. Isi nama database, user, password, host, port, dan konfigurasi lain sesuai komputer sendiri. Jangan memasukkan file `.env` ke Git.

Jalankan migration dari folder ini:

```bash
psql -U postgres -d nama_database -f migrations/001_create_students.sql
```

Lalu jalankan aplikasi:

```bash
go run .
```

Health check akan mengembalikan `503` jika aplikasi tidak bisa melakukan ping ke PostgreSQL. Pada folder ini belum ada test otomatis.
