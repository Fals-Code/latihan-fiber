# D.5 Analisis Singkat

## 1. Trade-off cursor pagination tanpa total halaman

Cursor pagination tidak menyediakan informasi seperti “halaman 3 dari 40” karena endpoint tidak menghitung seluruh jumlah baris dengan `COUNT(*)`. Dalam implementasi ini, server hanya mengambil `limit+1` baris, mengirim paling banyak `limit` baris, lalu memakai baris tambahan untuk menentukan `has_more`. Trade-off tersebut dapat diterima untuk endpoint students karena tujuan utamanya adalah mengambil data secara efisien dan stabil ketika dataset berubah, bukan menyediakan nomor halaman absolut. Client masih dapat mengetahui apakah ada halaman berikutnya melalui `has_more` dan `next_cursor`. Jika produk benar-benar memerlukan total, jumlah tersebut dapat disediakan melalui endpoint statistik terpisah atau query count khusus, tanpa menjadikan count sebagai bagian wajib setiap request pagination.

Implementasi aktual terlihat pada `app/service/student_service.go:65-74`: hasil `limit+1` dipotong, `has_more` ditentukan, dan `next_cursor` dibuat dari record terakhir yang dikirim. Pada halaman terakhir, `has_more=false` dan `next_cursor` tidak disertakan.

## 2. Informasi yang tidak boleh disimpan dalam cursor

Cursor dapat dibaca dan dipalsukan oleh siapa pun karena `helper.EncodeCursor` hanya melakukan encoding Base64 terhadap pasangan `created_at` dan `id` (`helper/errors.go:60-71`); Base64 bukan enkripsi. Karena itu cursor tidak boleh menyimpan password, token, data pribadi, kredensial, permission, role, atau informasi rahasia lain. Cursor juga tidak boleh dipercaya sebagai dasar authorization. Jika informasi rahasia dimasukkan, client dapat membacanya dan informasi tersebut berpotensi bocor melalui URL, log, history browser, atau sistem observability. Jika role atau permission dimasukkan lalu dipercaya saat decode, client dapat memalsukannya dan memperoleh akses yang tidak semestinya. Implementasi aktual hanya memakai timestamp dan ID sebagai posisi urutan, sedangkan authorization tetap dilakukan oleh middleware dan service.

`helper.DecodeCursor` memvalidasi format, timestamp, dan ID (`helper/errors.go:73-85`). Cursor rusak menghasilkan HTTP 400 pada service (`app/service/student_service.go:53-57`), bukan diam-diam dianggap sebagai halaman pertama.

## 3. Mengapa `IsEmptyPatch` tetap berada di luar tag validasi

Tag validator memeriksa aturan satu field pada satu waktu. Tag `omitnil` pada `PatchStudentRequest` membedakan field yang tidak dikirim dari field yang dikirim, sedangkan tag seperti `min`, `gt`, dan `gte` memeriksa nilai field tersebut (`app/model/student.go:26-30`). Namun aturan “PATCH harus mengandung setidaknya satu field” adalah aturan antar-field: hasilnya invalid hanya ketika `NIM`, `Name`, `Grade`, dan `IsActive` semuanya nil. Tidak ada satu field tertentu yang dapat diberi tag untuk menyatakan hubungan tersebut tanpa mengubah makna validasi per-field.

Karena itu `IsEmptyPatch` tetap melakukan pemeriksaan eksplisit (`app/service/student_rules.go:27-33`) setelah validasi tag dijalankan. Service mengembalikan HTTP 400 untuk PATCH kosong (`app/service/student_service.go:269-272`), sedangkan `{"name":""}` tetap mencapai validasi tag dan ditolak sebagai HTTP 422. Pemisahan ini menjaga perbedaan antara body tanpa perubahan sama sekali dan field yang dikirim tetapi nilainya melanggar aturan.
