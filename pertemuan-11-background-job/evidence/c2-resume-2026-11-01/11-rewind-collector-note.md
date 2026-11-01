# Kendala helper pemulihan C.2

Percobaan pertama transaksi pada `10-approved-rewind.sql` gagal sebelum DELETE
pertama. Tool eksekusi mengembalikan exception psql dengan pesan
`ERROR: column reference "n" is ambiguous`; alias tabel `n` bertabrakan dengan
variabel PL/pgSQL. Ini kendala helper pengujian, bukan puzzle aplikasi.

Catatan ini ditulis dari hasil tool yang diamati, bukan file stderr mentah.
stderr mentah tidak disimpan terpisah. Sesi psql menggunakan ON_ERROR_STOP,
berakhir sebelum COMMIT, dan transaksi tidak diterapkan.

Sebelum percobaan kedua, helper membandingkan seluruh hasil query database
dengan `09-before-approved-rewind.json` dan menolak jika berbeda. Guard itu
lolos. SQL kedua hanya mengganti nama variabel penghitungan menjadi
`affected`; ruang lingkup perubahan data sama. SQL pertama tetap disimpan.
Commit yang berhasil beserta pemeriksaan C.1 ada di
`12-approved-rewind-corrected.sql` dan `13-approved-rewind-result.json`.
