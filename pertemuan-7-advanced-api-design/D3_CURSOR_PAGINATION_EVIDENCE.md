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

## Kesimpulan D.3

Pembuktian penyisipan record di antara pengambilan halaman telah **PASS** pada database pengujian terpisah. Cursor lama tidak mengulang atau memasukkan record baru, sedangkan pagination baru menampilkan record baru pada halaman pertama.
