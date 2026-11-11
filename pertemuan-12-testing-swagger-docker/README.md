# Pertemuan 12 — Testing, Swagger, dan Docker

REST API SIAKAD Mini berbasis Go Fiber. Modul ini memuat unit test dan integration test, dokumentasi Swagger, serta deployment lima service dengan Docker Compose.

## Isi proyek

- `app/`, `config/`, `database/`, `helper/`, `middleware/`, `route/`, `main.go`, dan `worker.go`: aplikasi API dan worker.
- `migrations/`: bootstrap dan skema PostgreSQL untuk instalasi baru.
- `tests/integration/`: enam test C.1 dan dua test tambahan D.2. Test ini memakai build tag `integration`.
- `docs/docs.go`, `docs/swagger.json`, `docs/swagger.yaml`: artefak Swagger hasil generator.
- `Dockerfile` dan `docker-compose.yml`: API, worker, PostgreSQL, MongoDB, dan Redis.

## Konfigurasi dan Docker Compose

Gunakan Go sesuai `go.mod` untuk menjalankan aplikasi tanpa Docker. Salin `.env.example` ke `.env` dan isi nilai lokal; jangan unggah `.env`. Aplikasi memerlukan `JWT_SECRET` sedikitnya 32 karakter. Integration test membutuhkan `TEST_DATABASE_URL` ke database khusus dengan nama berakhiran `_test`. Test menolak database yang tidak memenuhi syarat itu.

Untuk Compose, set `P12_DB_PASSWORD` dan `JWT_SECRET` pada lingkungan pemanggil, lalu jalankan dari folder ini:

```powershell
docker compose config
docker compose build
docker compose up -d
docker compose ps
```

Compose memakai project `p12-testing-swagger-docker`, menyimpan PostgreSQL, MongoDB, dan Redis pada volume project, serta menerbitkan API pada `127.0.0.1:${P12_HTTP_PORT:-18080}`. Tetapkan `P12_HTTP_PORT` bila port default terpakai. `SWAGGER_ENABLED=true` mengaktifkan UI Swagger; nilai default `false` menonaktifkannya. Endpoint kesehatan: `/api/v1/health`.

## Dokumentasi dan pengujian

Spesifikasi API tersedia di [docs/swagger.json](docs/swagger.json), dengan base path `/api/v1` dan otorisasi Bearer untuk endpoint terlindungi. Untuk regenerasi dari anotasi source:

```powershell
go tool swag init -g main.go -d ".,./app/service" -o "<path-absolut-folder-proyek>/docs"
```

Unit test: `go test ./...`. Pengujian acak C.1: jalankan `go test -shuffle=on -count=5 ./app/service/` sebanyak tiga kali secara independen. Integration test: `go test -tags=integration -v ./tests/integration/` setelah database test terisolasi tersedia. Jangan gunakan database produksi atau database proyek lain.

Laporan akademik dan evidence pengujian disimpan terpisah dari source yang diperlukan untuk membangun aplikasi dari clone.
