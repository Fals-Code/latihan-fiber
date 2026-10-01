# SIAKAD Mini

REST API akademik untuk autentikasi, pengelolaan mahasiswa, daftar mata kuliah, dan KRS.

## Persiapan lokal

1. Dari root repository, salin `uts-pbl-siakad-mini/.env.example` menjadi `uts-pbl-siakad-mini/.env`, lalu sesuaikan credential PostgreSQL lokal. Konfigurasi process environment memiliki prioritas; nilai yang tidak disetel dari process environment dimuat dari file UTS tersebut. Pastikan `DB_NAME=siakad_mini`; aplikasi tidak menjalankan migration atau seeder saat start.
2. Dari root repository, buat database `siakad_mini` bila belum ada, lalu jalankan migration dan seeder secara eksplisit:

```powershell
createdb -U postgres siakad_mini
go run ./uts-pbl-siakad-mini/cmd/dbsetup migrate
go run ./uts-pbl-siakad-mini/cmd/dbsetup seed
```

Jangan jalankan perintah pada database praktikum. Jika tool `createdb` tidak tersedia, buat database secara manual melalui psql/pgAdmin.

## Akun seed

- Admin: `admin@siakad.local` / `Admin123!`
- Mahasiswa: email berbentuk `<NIM>@student.siakad.local`; password awal adalah NIM masing-masing.

Password disimpan menggunakan bcrypt. Ganti password seed lokal setelah penggunaan awal bila diperlukan.

## Jalankan API

Persyaratan: Go dan PostgreSQL. Dari root repository setelah database tersedia dan `.env` dikonfigurasi:

```powershell
go run ./uts-pbl-siakad-mini
```

Base URL lokal: `http://localhost:3000` (atau port dari `APP_PORT`). Migrasi dan seeder tidak dijalankan otomatis saat aplikasi start.

## Endpoint

Semua endpoint selain login memerlukan bearer token.

| Method | Path | Akses |
| --- | --- | --- |
| POST | `/api/v1/auth/login` | Publik |
| GET | `/api/v1/auth/me` | Pengguna terautentikasi |
| GET | `/api/v1/students` | Admin |
| POST | `/api/v1/students` | Admin |
| GET | `/api/v1/students/:id` | Admin / mahasiswa pemilik |
| PUT | `/api/v1/students/:id` | Admin |
| DELETE | `/api/v1/students/:id` | Admin |
| GET | `/api/v1/courses` | Pengguna terautentikasi |
| POST | `/api/v1/enrollments` | Mahasiswa |
| DELETE | `/api/v1/enrollments/:id` | Mahasiswa pemilik |
