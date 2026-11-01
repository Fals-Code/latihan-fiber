# Gate B1 — pemeriksaan awal lingkungan

Tanggal: 27 Oktober 2026 (Asia/Jakarta).

Status: BLOCKED pada kesiapan lingkungan. Compiler belum dijalankan.
Dokumen ini bukan bukti kegagalan compiler maupun kelulusan build/vet.

## Snapshot dan module Go

Pemeriksaan baca-saja menemukan 96 berkas dalam snapshot. Hash arsip cocok
dengan manifest; tidak ada perbedaan antara manifest, isi ZIP, dan source kerja
yang dicakup manifest sebelum dokumen ini dibuat.

Output asli pemeriksaan integritas:

```text
SNAPSHOT {"files": 96, "archive_hash_matches": true, "archive_manifest_differences": [], "working_manifest_differences": []}
```

Output asli versi Go:

```text
go version go1.26.5 windows/amd64
```

Perintah berikut dijalankan dari `pertemuan-11-background-job`:

```powershell
go env GOMOD GOWORK GOVERSION
```

Output asli:

```text
C:\Users\falah\latihan-fiber\pertemuan-11-background-job\go.mod

go1.26.5
```

`GOWORK` kosong. Nested module dikenali. Asynq v0.26.0 tercatat dalam go.mod.
Tidak dilakukan pengunduhan dependency atau kompilasi pada tahap ini.

## Konfigurasi

Tidak ditemukan .env pada root repository, folder Pertemuan 10, maupun folder
Pertemuan 11. Pemeriksaan environment proses menunjukkan DB_HOST, DB_PORT,
DB_NAME, DB_USER, DB_PASSWORD, MONGO_URI, MONGO_DATABASE, REDIS_ADDR, REDIS_DB,
REDIS_PASSWORD, JWT_SECRET, WORKER_CONCURRENCY, dan DB_MAX_CONNS belum disetel.
Tidak ada nilai password atau token yang dicetak.

Nama tujuan pada template adalah pertemuan11_test, pertemuan11_mongo, dan
Redis logical DB 11. Keberadaan serta isolasinya belum terbukti.

## Layanan dan port

Inventaris layanan Windows menunjukkan MongoDB dan postgresql-x64-18 berjalan.
Tidak ditemukan layanan bernama Redis/Memurai pada inventaris yang diperiksa.

Output asli inventaris listener pada port 5432, 27017, 6379, dan 3332:

```json
[{"LocalAddress":"::","LocalPort":5432,"OwningProcess":6936},{"LocalAddress":"127.0.0.1","LocalPort":27017,"OwningProcess":5740},{"LocalAddress":"0.0.0.0","LocalPort":5432,"OwningProcess":6936}]
```

Tidak ditemukan listener 6379 atau 3332. Hal ini tidak membuktikan Redis tidak
terpasang atau tidak berjalan pada alamat/port lain.
MongoDB berjalan dari instalasi server 8.3; mongosh dan Redis CLI tidak ditemukan
pada PATH. Konektivitas MongoDB, database khusus, akun uji, dan penggunaan layanan
oleh proyek lain belum diverifikasi.

## Percobaan PostgreSQL baca-saja

Perintah memakai `--no-password` dan `-X`, tanpa prompt password maupun psqlrc.
SQL hanya meminta daftar nama database dalam transaction READ ONLY.

```powershell
psql -X --no-password --dbname='host=127.0.0.1 port=5432 dbname=postgres user=postgres connect_timeout=3' -A -t -c 'BEGIN READ ONLY; SELECT datname FROM pg_database WHERE datistemplate = false ORDER BY datname; ROLLBACK;'
```

Output asli (exit code 1):

```text
psql: error: connection to server at "127.0.0.1", port 5432 failed: fe_sendauth: no password supplied
```

Autentikasi gagal sebelum SQL dijalankan. Keberadaan database khusus, schema,
akun mhsuji/salahalamat/admin, serta pemakaian oleh proyek lain belum dapat
dipastikan. Tidak ada perubahan data database.

## Titik berhenti dan pekerjaan tertunda

Instruksi Gate B1 mensyaratkan lingkungan aman sebelum pemeriksaan compiler.
Syarat tersebut belum terpenuhi, sehingga tidak menjalankan go build atau go vet.
Tiga kandidat compiler belum direproduksi; tidak ada perbaikan source, output
compiler, ataupun screenshot compiler yang dibuat pada sesi ini.

Tidak menjalankan migration, API, worker, Redis, skenario C.1-C.8, D.2, atau D.3.
Tidak menghentikan layanan, mengubah konfigurasi/.env, commit, atau push.
Snapshot baseline dan enam kandidat puzzle perilaku tetap utuh.

Untuk meneruskan Gate B1, diperlukan konfigurasi akses lokal PostgreSQL,
verifikasi MongoDB dan akun uji, serta alamat Redis khusus praktikum. Kredensial
perlu disediakan secara lokal tanpa dimasukkan ke laporan atau chat.
Keputusan D.2 satu task per peserta tetap ditunda sampai baseline selesai diuji.
