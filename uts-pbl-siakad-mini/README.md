# SIAKAD Mini — Phase 1

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

Dari root repository setelah database tersedia:

```powershell
go run ./uts-pbl-siakad-mini
```

Endpoint belum diimplementasikan pada Phase 1.
