# Pertemuan 5 - Authentication dan Security

Pada pertemuan ini project mulai mengenal user dan login. Fokusnya adalah authentication: memastikan siapa yang sedang mengakses API, bukan mengatur apakah role tersebut boleh melakukan semua tindakan.

## Yang dikerjakan

- Register user dengan validasi username, email, dan password.
- Menyimpan password dalam bentuk hash, bukan password asli.
- Login menggunakan username dan password.
- Membuat JWT access token dan refresh token.
- Refresh token disimpan dalam bentuk hash dan diputar saat proses refresh.
- Logout mencabut refresh token yang sedang digunakan.
- `RequireAuth` membaca access token dari request dan menyimpan identity ke context.
- Login rate limiter membatasi percobaan login yang terlalu sering.

JWT memakai `JWT_SECRET`, issuer, access TTL, dan refresh TTL dari environment. File `.env.example` hanya berisi nama konfigurasi dan placeholder. Buat `.env` lokal sendiri dan jangan memasukkan isinya ke Git.

## Endpoint

Route authentication ada di `/api/v1/auth`:

| Method | Endpoint | Keterangan |
| --- | --- | --- |
| POST | `/api/v1/auth/register` | Membuat user baru |
| POST | `/api/v1/auth/login` | Login dan menerima token |
| POST | `/api/v1/auth/refresh` | Meminta access token baru dari refresh token |
| POST | `/api/v1/auth/logout` | Mencabut refresh token |

Endpoint user dan student juga memakai `RequireAuth` pada bagian yang membutuhkan login. Pada tahap ini pengecekan utamanya masih authentication. Aturan role dan permission dikembangkan di Pertemuan 6.

## File penting

- `app/service/auth_service.go` menangani alur register, login, refresh, dan logout.
- `app/service/auth_rules.go` berisi validasi request authentication.
- `app/repository/user_repository.go` dan `token_repository.go` menangani penyimpanan user serta refresh token.
- `helper/security.go` menangani hashing password dan token.
- `helper/jwt.go` membuat dan membaca JWT.
- `middleware/auth.go` berisi `RequireAuth` dan rate limiter login.
- `migrations/003_auth.sql` menyiapkan tabel authentication.

## Menjalankan dan menguji

Siapkan PostgreSQL, jalankan migration yang ada di folder `migrations`, lalu buat `.env` berdasarkan `.env.example`. Dari folder ini jalankan:

```bash
go run .
```

Test unit yang tersedia dapat dijalankan dengan:

```bash
go test ./...
```

Test mencakup validasi auth, hashing password, JWT, middleware authentication, token yang tidak valid atau kedaluwarsa, serta rate limiter login.
