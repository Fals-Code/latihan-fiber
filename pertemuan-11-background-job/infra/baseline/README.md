# Infrastruktur lokal Pertemuan 11

Konfigurasi pada `compose.yaml` menyediakan PostgreSQL, MongoDB, dan Redis untuk praktikum. Seluruh port aplikasi database dipublikasikan pada alamat loopback `127.0.0.1`.

| Layanan | Container | Port host |
| --- | --- | --- |
| PostgreSQL | `pbe11-baseline-postgres` | `55411` |
| MongoDB | `pbe11-baseline-mongo` | `27111` |
| Redis | `pbe11-baseline-redis` | `6379` |

Konfigurasi dan versi image dirujuk dari [compose.yaml](compose.yaml) serta [images.lock.json](images.lock.json).

## Menyiapkan lingkungan

- Gunakan file `.env` lokal untuk koneksi dan credential. Jangan commit file tersebut.
- Script `prepare_local.py` dan `provision.py` ditujukan untuk penyiapan awal. Periksa nama container, port, network, dan volume sebelum menjalankannya agar tidak berbenturan dengan lingkungan lain.
- Aplikasi menggunakan database praktikum tersendiri; migration tidak dijalankan otomatis.

Periksa container yang berjalan dengan:

```powershell
docker compose -f infra/baseline/compose.yaml ps
```

Pada pengujian historis, baseline memakai PostgreSQL 18.6, MongoDB 8.0.34, dan Redis 7.4.12. Versi aktual bergantung pada image dan konfigurasi yang dijalankan.

Untuk skenario gangguan Redis, pastikan targetnya hanya container baseline ini. Jangan menghentikan layanan Redis milik proyek lain.
