# Pertemuan 1 - Dasar Go

Pertemuan ini berisi latihan dasar Go sebelum masuk ke pembuatan API. Contohnya masih kecil dan dipisahkan per tugas supaya konsepnya mudah dicoba satu per satu.

## Materi yang dikerjakan

- `tugas1-server` membuat server Fiber sederhana dengan endpoint `GET /` yang mengembalikan `Hello, World!`.
- `tugas2-variabel` mencoba variabel dengan beberapa tipe data, slice, map, menambah dan menghapus isi map, serta membaca data map.
- `tugas3-pointer` mencoba pointer melalui fungsi swap, menambah item ke slice, dan membandingkan perubahan nilai biasa dengan perubahan lewat pointer.
- `tugas4-struct` membuat struct `Student` beserta method untuk menampilkan informasi, mengubah nilai, dan mengaktifkan atau menonaktifkan status student.

## Menjalankan latihan

Masuk ke salah satu folder tugas, lalu jalankan:

```bash
go run .
```

Contoh:

```bash
cd pertemuan-1-dasar-go/tugas1-server
go run .
```

Untuk `tugas1-server`, server berjalan di port `3000` dan endpoint yang tersedia adalah `GET /`. Tiga tugas lainnya menampilkan hasil latihan langsung di terminal.

## Catatan

Folder ini masih fokus pada sintaks dan konsep dasar Go. Belum ada database, migration, atau test otomatis di dalam folder ini.
