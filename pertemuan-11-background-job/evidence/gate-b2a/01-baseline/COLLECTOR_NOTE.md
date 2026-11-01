# Catatan collector awal

Percobaan C.1 awal selesai mengirim request, tetapi collector gagal saat
mem-parse JSON PostgreSQL multiline: JSONDecodeError karena hanya baris pertama
diambil. Tidak ada hasil HTTP checkpoint ini yang dipulihkan atau direkonstruksi.
Keadaan database/log disimpan melalui checkpoint 01-initial-observation sebelum
perubahan source dan reset berikutnya. Pengulangan checkpoint03 dan seterusnya
menyimpan seluruh response C.1. Catatan ini menjelaskan kegagalan alat evidence,
bukan salah satu puzzle dosen dan bukan hasil C.1 lengkap.
