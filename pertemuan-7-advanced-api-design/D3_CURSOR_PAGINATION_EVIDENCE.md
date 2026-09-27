# Bukti Pengujian D.3 — Cursor Pagination Students

## Metode dan lingkungan

Pengujian dilakukan melalui endpoint HTTP Fiber `GET /api/v1/students`, bukan hanya query builder. Endpoint terhubung ke PostgreSQL database pengujian terpisah `hl_test` menggunakan data sintetis.

- Database utama `praktikum_backend` tidak digunakan untuk operasi tulis.
- Database pengujian: `hl_test`.
- Data awal: enam record students sintetis.
- Limit pagination: `2`.
- Akun pengujian: akun admin sintetis dengan permission `student:list`.
- Pengujian dilakukan dengan `Fiber app.Test` dan PostgreSQL lokal.
- Tidak ada screenshot atau log yang direkayasa.

## Skenario penyisipan record

### Halaman pertama sebelum INSERT

Request pertama menggunakan `limit=2` tanpa cursor.

```text
HTTP 200
IDs: [6, 5]
has_more: true
next_cursor: tersedia
```

Cursor tersebut disimpan dan digunakan untuk seluruh halaman lanjutan.

### Record baru

Satu student sintetis disisipkan ke database `hl_test` dengan `created_at` yang lebih baru daripada seluruh record awal.

```text
INSERTED_NEW_ID=7
```

Record ini hanya disisipkan pada database pengujian terpisah.

### Pagination menggunakan cursor lama

```text
Halaman 2: IDs [4, 3], has_more=true
Halaman 3: IDs [2, 1], has_more=false
```

Hasil aktual:

- Jumlah halaman: `3`.
- Jumlah ID unik: `6`.
- Record lama yang ditemukan: `6, 5, 4, 3, 2, 1`.
- ID `7` muncul pada halaman lanjutan dari cursor lama: `false`.
- Tidak ada ID yang berulang.
- Semua enam record lama muncul tepat satu kali.
- Halaman terakhir memiliki `has_more=false`.

Status pembuktian cursor lama: **PASS**.

### Pagination baru tanpa cursor

Pagination dimulai ulang tanpa cursor setelah record baru disisipkan.

```text
Halaman 1: IDs [7, 6]
has_more: true
next_cursor: tersedia
```

ID `7` muncul pada halaman pertama pagination baru.

Status pagination baru: **PASS**.

## Cursor tidak valid

Request menggunakan cursor invalid menghasilkan:

```text
HTTP 400
```

Status: **PASS**.

## Perbandingan perilaku

| Pemeriksaan | Hasil |
|---|---|
| Halaman pertama sebelum INSERT | `[6, 5]` |
| Halaman lanjutan dengan cursor lama | `[4, 3]`, `[2, 1]` |
| Record baru muncul pada cursor lama | Tidak |
| Record lama muncul tepat satu kali | Ya |
| ID duplikat antarhalaman | Tidak ada |
| Pagination baru setelah INSERT | `[7, 6]` |
| Cursor invalid | HTTP 400 |

Perilaku tersebut sesuai dengan keyset pagination: cursor merepresentasikan posisi terakhir yang telah dikirim, sehingga record baru yang berada sebelum posisi tersebut tidak masuk ke rangkaian halaman lanjutan dari cursor lama.

## Bukti EXPLAIN ANALYZE sebelumnya

`EXPLAIN (ANALYZE, BUFFERS)` dijalankan pada database lokal untuk query students halaman pertama dan halaman cursor. Dataset saat itu berisi enam record, sehingga PostgreSQL memilih `Seq Scan`.

### Halaman pertama

```text
Limit  (cost=17.45..17.46 rows=5 width=12) (actual time=0.035..0.036 rows=5.00 loops=1)
  Buffers: shared hit=1
  ->  Sort  (cost=17.45..18.15 rows=280 width=12) (actual time=0.034..0.035 rows=5.00 loops=1)
        Sort Key: created_at DESC, id DESC
        Sort Method: quicksort  Memory: 25kB
        Buffers: shared hit=1
        ->  Seq Scan on students  (cost=0.00..12.80 rows=280 width=12) (actual time=0.014..0.016 rows=6.00 loops=1)
              Buffers: shared hit=1
Planning Time: 0.090 ms
Execution Time: 0.054 ms
```

### Halaman dengan cursor

```text
Limit  (cost=15.87..15.88 rows=6 width=255) (actual time=0.042..0.044 rows=5.00 loops=1)
  Buffers: shared hit=1
  ->  Sort  (cost=15.87..16.10 rows=93 width=255) (actual time=0.041..0.042 rows=5.00 loops=1)
        Sort Key: created_at DESC, id DESC
        Sort Method: quicksort  Memory: 25kB
        Buffers: shared hit=1
        ->  Seq Scan on students  (cost=0.00..14.20 rows=93 width=255) (actual time=0.026..0.029 rows=5.00 loops=1)
              Filter: (ROW(created_at, id) < ROW('2026-09-20 18:45:54.174265+07'::timestamp with time zone, 34))
              Rows Removed by Filter: 1
              Buffers: shared hit=1
Planning Time: 0.114 ms
Execution Time: 0.066 ms
```

Index `students_created_at_id_desc_idx` tidak dipilih pada dataset enam record. `Seq Scan` dipilih oleh PostgreSQL karena biaya tabel kecil; hasil ini bukan bukti bahwa index tidak valid.

## Index pada database utama

Setelah pengujian cursor, kedua index composite diterapkan pada database utama `praktikum_backend` dalam satu transaksi terpisah. Tidak ada data yang diubah dan tidak ada migration lain yang dijalankan.

Output catalog aktual:

```text
index_name                      | table_name | indisvalid | indisready | definition
--------------------------------+------------+------------+------------+---------------------------------------------------------------
students_created_at_id_desc_idx | students   | t          | t          | CREATE INDEX students_created_at_id_desc_idx ON public.students USING btree (created_at DESC, id DESC)
users_created_at_id_desc_idx    | users      | t          | t          | CREATE INDEX users_created_at_id_desc_idx ON public.users USING btree (created_at DESC, id DESC)
```

Dengan demikian, kedua index memiliki `indisvalid=true` dan `indisready=true`.

## EXPLAIN ANALYZE setelah index diterapkan

Empat `EXPLAIN (ANALYZE, BUFFERS)` read-only dijalankan pada `praktikum_backend`:

| Query | Hasil aktual |
|---|---|
| Students halaman pertama | PASS; `Seq Scan on students`, `Sort Key: created_at DESC, id DESC`, 6 rows, execution time 0.046 ms |
| Students halaman cursor | PASS; `Seq Scan on students` dengan filter row cursor, `Sort Key: created_at DESC, id DESC`, 5 rows, execution time 0.045 ms |
| Users halaman pertama | PASS; `Seq Scan on users`, `Sort Key: created_at DESC, id DESC`, 5 rows, execution time 0.031 ms |
| Users halaman cursor | PASS; `Seq Scan on users` dengan filter row cursor, `Sort Key: created_at DESC, id DESC`, 4 rows, execution time 0.036 ms |

Planner tetap memilih `Seq Scan` karena tabel sangat kecil, bukan karena index gagal. Catalog membuktikan kedua index valid dan siap digunakan; pilihan execution plan ditentukan oleh estimasi biaya PostgreSQL.

## Regression test setelah index

Regression test HTTP read-only pada `praktikum_backend` berhasil:

```text
users:    [37, 36] -> [35, 13] -> [12]
students: [34, 33] -> [32, 6] -> [3, 1]
```

Kedua endpoint mengembalikan HTTP 200, `has_more` dan `next_cursor` sesuai, ID tidak berulang, dan halaman terakhir tidak memiliki cursor. Content negotiation JSON, CSV, XML (406), wildcard JSON, serta cursor invalid (400) juga PASS.

## Isolasi pengujian INSERT

Pengujian penyisipan record baru dilakukan secara terisolasi pada database `hl_test`, bukan `praktikum_backend`. Enam record sintetis awal menghasilkan halaman `[6, 5]`; setelah record baru ID `7` disisipkan di posisi paling atas, cursor lama menghasilkan `[4, 3]` lalu `[2, 1]` tanpa menampilkan ID `7`. Pagination baru tanpa cursor menghasilkan `[7, 6]`. Database utama tidak digunakan untuk operasi INSERT tersebut.

## Rekonsiliasi terhadap praktikum_backend

Skenario INSERT dan ID sintetis pada bagian sebelumnya adalah bukti historis `hl_test`; tidak boleh dibaca sebagai pengujian pada `praktikum_backend`. Riwayat tersebut dipertahankan apa adanya.

Audit read-only terbaru pada `praktikum_backend` menghasilkan:

```text
current_database() = praktikum_backend
students = 6
users = 5
students_created_at_id_desc_idx tersedia
users_created_at_id_desc_idx tersedia
```

`EXPLAIN (ANALYZE, BUFFERS)` untuk SELECT halaman pertama students berhasil dan memilih `Seq Scan` dengan `Sort Key: created_at DESC, id DESC`; ini konsisten dengan tabel kecil dan bukan indikasi index tidak valid. Tidak ada operasi tulis atau migration.

Pada audit read-only terdahulu, pagination HTTP halaman pertama, halaman lanjutan, pemeriksaan ID duplikat, `has_more`, dan cursor invalid pada `praktikum_backend` berstatus **NOT VERIFIED** karena kredensial akun pengujian belum tersedia secara tervalidasi. Skenario INSERT belum dilakukan pada tahap tersebut; status ini digantikan oleh eksekusi HTTP aktual di bawah.

## Kesimpulan D.3

Implementasi cursor dan query dapat dibuktikan melalui source, unit test, koneksi read-only, catalog index, dan EXPLAIN. Bukti `hl_test` sebelumnya tetap historis; hasil HTTP aktual pada `praktikum_backend`, termasuk skenario INSERT, didokumentasikan pada bagian berikut.

## Eksekusi HTTP aktual pada `praktikum_backend` — 27 September 2026

Pengujian memakai aplikasi Fiber asli dan limit `2`. Ringkasan berikut dibuat self-contained agar bukti D.3 tetap dapat dibaca tanpa raw terminal log lokal yang tidak dipush ke repository.

Preflight menghasilkan `current_database=praktikum_backend`, users `[12,13,35,36,37]`, students `[1,3,6,32,33,34]`, dan `p7_test=0`. Akun sementara ID `45` dibuat melalui register, diberi role `staff` melalui UPDATE yang dibatasi ID/email lalu diverifikasi dengan SELECT, dan login melalui endpoint resmi berhasil.

### Fixture aktual

| Fixture | Student ID | NIM | `created_at` UTC |
|---|---:|---:|---|
| Batch 1 | 49 | 943096225 | 2026-09-27T09:58:19.187137Z |
| Batch 2 | 50 | 943096226 | 2026-09-27T09:58:19.362495Z |
| Batch 3 | 51 | 943096227 | 2026-09-27T09:58:19.537805Z |
| Batch 4 | 52 | 943096228 | 2026-09-27T09:58:19.740606Z |
| Batch 5 | 53 | 943096229 | 2026-09-27T09:58:20.019896Z |
| Batch 6 | 54 | 943096230 | 2026-09-27T09:58:20.184158Z |
| INSERT setelah cursor lama | 55 | 943096231 | 2026-09-27T09:58:20.534214Z |

### Snapshot sebelum INSERT

Halaman pertama sebelum INSERT adalah `[54,53]`, `has_more=true`, dan memberikan cursor lama. Snapshot lengkap melalui pagination menghasilkan `[54,53]`, `[52,51]`, `[50,49]`, `[34,33]`, `[32,6]`, `[3,1]`; hanya halaman terakhir memiliki `has_more=false` dan tidak memiliki `next_cursor`.

Snapshot ini tidak memiliki ID duplikat dan sama dengan urutan SELECT `created_at DESC, id DESC` pada database.

### Lanjutan cursor lama setelah INSERT

Setelah INSERT ID `55`, posisi aktualnya `new_index=0`, sedangkan batas cursor lama berada di `cursor_index=2`. Jadi record baru berada **sebelum** batas cursor lama.

Lanjutan dari cursor lama menghasilkan `[52,51]`, `[50,49]`, `[34,33]`, `[32,6]`, `[3,1]`. Seluruh halaman HTTP 200; halaman terakhir berakhir normal pada `has_more=false` tanpa `next_cursor`.

- Tidak ada ID duplikat.
- Urutan sama dengan proyeksi SELECT setelah batas cursor lama.
- Semua record snapshot lama yang berada setelah batas muncul tepat satu kali.
- ID `55` tidak muncul pada lanjutan cursor lama: **PASS**.

### Pagination baru setelah INSERT

Pagination baru menghasilkan `[55,54]`, `[53,52]`, `[51,50]`, `[49,34]`, `[33,32]`, `[6,3]`, `[1]`. Urutan seluruhnya sama dengan SELECT aktual `created_at DESC, id DESC`; ID `55` terlihat pada posisi pertama sesuai timestamp aktualnya. Tidak ada ID duplikat dan halaman ketujuh berakhir dengan `has_more=false` tanpa `next_cursor`.

### Cleanup

`finally` memeriksa `achievements=0` untuk setiap ID `49–55`, menghapus setiap student dengan predicate ID/owner, memeriksa dan membersihkan satu refresh token user ID `45`, kemudian menghapus user tersebut. SELECT akhir mengembalikan users `[12,13,35,36,37]`, students `[1,3,6,32,33,34]`, dan `p7_test=0`.

Status terbaru D.3 pada `praktikum_backend`: **PASS** berdasarkan ringkasan eksekusi yang sudah dicatat di dokumen ini. Bukti `hl_test` sebelumnya di dokumen ini tetap historis.
