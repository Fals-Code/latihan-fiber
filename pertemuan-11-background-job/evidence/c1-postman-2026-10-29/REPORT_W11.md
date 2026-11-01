# Bahan integrasi w11 — hasil aktual pengulangan C.1 melalui Postman

Tanggal pelaksanaan: 29 Oktober 2026. Acuan penerimaan: *Modul 11 — Background Job*, hlm. 31–32. Bahan ini memuat hasil pengulangan C.1 baru pada `evidence/c1-postman-2026-10-29`; hasil dan gambar historis Gate B2A tetap dipertahankan dengan asalnya masing-masing. Tidak ada pembuktian C.2–C.8 atau D.2/D.3 dalam pengulangan ini.

## Narasi pengantar siap integrasi

Pengulangan jalur normal C.1 dilaksanakan melalui collection Postman “PBE Pertemuan 11 - C.1 Baseline” pada lingkungan praktikum khusus. Variabel `base_url` merujuk pada `http://127.0.0.1:3332/api/v1`; address bar pada gambar menampilkan nama variabel tersebut beserta path endpoint. Lima permintaan dijalankan berurutan, yaitu mendaftar kelas IF102, mengulangi pendaftaran yang sama, keluar dari kelas, mendaftar kembali, dan berpindah dari IF102 ke IF302. Respons aktual berturut-turut adalah HTTP 201, 409, 204, 201, dan 200.

Verifikasi pendamping mencatat jumlah baris outbox setelah setiap permintaan sebesar 1, 1, 2, 3, dan 4. Pada pemeriksaan akhir, empat event telah published, empat notifikasi tersimpan, dan hitungan pencatatan email akun uji berjumlah empat. Queue `notifications` tidak memiliki task pending, active, scheduled, retry, atau archived. Dengan demikian, hasil pengulangan memenuhi kriteria jalur normal C.1 pada cakupan data yang diperiksa. Hasil tersebut tidak digunakan untuk menyimpulkan perilaku saat kegagalan Redis, kegagalan pengiriman, redelivery, atau shutdown karena skenario itu belum dijalankan pada tahap ini.

## Caption dan analisis untuk placeholder w11

### C1-P01 — pendaftaran awal

**Gambar C1-P01. Respons pendaftaran awal kelas IF102 melalui Postman.**

PNG utama: `screenshots/01-enroll-201.png`.

Permintaan `POST {{base_url}}/courses/2/enroll` menghasilkan HTTP **201 Created** dengan `success: true`. Respons menunjukkan kelas ID 2, kode IF102, dan nama Basis Data Lanjut. Checkpoint pendamping setelah permintaan pertama mencatat satu event `enrollment.created` dengan ID 21. Hasil ini menunjukkan bahwa pendaftaran awal diterima dan mempunyai catatan outbox pada pengulangan yang sama.

### C1-P02 — pendaftaran duplikat

**Gambar C1-P02. Penolakan pendaftaran duplikat tanpa penambahan event outbox.**

PNG utama: `screenshots/02-duplicate-409.png`.

Pengulangan `POST {{base_url}}/courses/2/enroll` menghasilkan HTTP **409 Conflict** dengan kode `ALREADY_ENROLLED` dan pesan bahwa mahasiswa telah terdaftar di kelas tersebut. Jumlah outbox tetap satu pada checkpoint setelah permintaan kedua, dibandingkan satu pada checkpoint pertama. Kesimpulan bahwa permintaan duplikat tidak menambah event didasarkan pada pasangan checkpoint tersebut, bukan pada status 409 saja. ID 22 yang tidak muncul dalam hasil akhir tidak dihitung sebagai event; sequence yang memiliki celah tidak membuktikan keberadaan baris tambahan.

### C1-P03 — keluar kelas

**Gambar C1-P03. Respons pengunduran diri dari kelas IF102 dengan HTTP 204.**

PNG utama: `screenshots/03-drop-204.png`.

Permintaan `DELETE {{base_url}}/courses/2/enroll` menghasilkan HTTP **204 No Content** tanpa body respons. Bukti pendamping mencatat kenaikan jumlah outbox menjadi dua dan event kedua `enrollment.dropped` dengan ID 23. Notifikasi yang berkaitan memakai subjek “KRS: keluar dari IF102 Basis Data Lanjut”, sehingga tindakan keluar kelas tercatat dengan jenis event dan subjek yang sesuai.

### C1-P04 — pendaftaran kembali

**Gambar C1-P04. Pendaftaran kembali ke kelas IF102 setelah keluar.**

PNG utama: `screenshots/04-reenroll-201.png`.

Permintaan keempat, `POST {{base_url}}/courses/2/enroll`, menghasilkan HTTP **201 Created** dan kembali menunjukkan kelas IF102 Basis Data Lanjut. Capture ini berasal dari langkah pendaftaran kembali, berbeda dari capture pendaftaran awal. Jumlah outbox meningkat dari dua menjadi tiga; event baru adalah `enrollment.created` dengan ID 24. Riwayat notifikasi menyimpan subjek pendaftaran untuk event ini secara terpisah dari event awal ID 21.

### C1-P05 — perpindahan kelas

**Gambar C1-P05. Perpindahan dari IF102 ke IF302 melalui endpoint swap.**

PNG utama: `screenshots/05-swap-200.png`.

Permintaan `POST {{base_url}}/enrollments/swap` memuat `from_course_id: 2` dan `to_course_id: 4`. Postman memperlihatkan HTTP **200 OK**, pesan “berhasil pindah kelas”, serta tujuan IF302 Keamanan Informasi. Jumlah outbox bertambah menjadi empat melalui satu event `enrollment.swapped` dengan ID 25. Pada pengulangan ini perpindahan direpresentasikan oleh satu event swap, sesuai penerimaan C.1.

### C1-P06a/b — empat notifikasi dalam satu respons

**Gambar C1-P06a. Bagian atas respons riwayat notifikasi setelah rangkaian C.1.**

PNG utama: `screenshots/06a-notifications-top.png`.

**Gambar C1-P06b. Lanjutan respons yang sama hingga akhir daftar empat notifikasi.**

PNG utama: `screenshots/06b-notifications-bottom.png`.

Permintaan `GET {{base_url}}/notifications/mine` menghasilkan HTTP **200 OK**. Kedua gambar memperlihatkan bagian atas dan bawah dari satu respons yang sama; elemen yang bertumpang tindih menunjukkan kontinuitas daftar, bukan dua pengambilan dengan hasil berbeda. Respons memuat empat notifikasi dengan urutan `outbox_id` 25, 24, 23, dan 21. Subjeknya berturut-turut adalah:

1. “KRS: pindah dari IF102 Basis Data Lanjut ke IF302 Keamanan Informasi”.
2. “KRS: berhasil terdaftar di IF102 Basis Data Lanjut”.
3. “KRS: keluar dari IF102 Basis Data Lanjut”.
4. “KRS: berhasil terdaftar di IF102 Basis Data Lanjut”.

Keempat subjek sesuai dengan urutan terbaru terlebih dahulu pada penerimaan modul. Kesamaan subjek pendaftaran pada dua elemen berasal dari dua tindakan berbeda, yang dapat dibedakan melalui `outbox_id` 24 dan 21. Urutan tampilan ini menunjukkan susunan riwayat aplikasi; bukti tersebut tidak menetapkan urutan kedatangan email di kotak masuk.

### C1-P07 — pemantauan admin

**Gambar C1-P07. Keadaan outbox dan queue setelah pemrosesan jalur normal C.1.**

PNG utama: `screenshots/07-admin-queue-200.png`; respons tersimpan: `admin-response.json`.

Permintaan admin `GET {{base_url}}/admin/queue` menghasilkan HTTP **200 OK** dengan `outbox_pending: 0`, `outbox_published: 4`, dan `notifications_sent: 4`. Pada queue `notifications`, pending, active, scheduled, retry, dan archived masing-masing bernilai nol. Counter `completed` bernilai 22 dan `processed_today` bernilai 4 pada saat capture. Counter queue mempunyai cakupan historis atau periode pemantauan sehingga `completed: 22` tidak ditafsirkan sebagai 22 task yang dihasilkan oleh C.1 baru. Jumlah empat event dan empat notifikasi pengulangan dibuktikan melalui hasil PostgreSQL yang terkait.

### C1-S01 — rekonsiliasi SQL

**Gambar C1-S01a. Query rekonsiliasi event dan notifikasi C.1 pada editor VS Code.**

PNG utama: `screenshots/08a-vscode-query.png`.

**Gambar C1-S01b. Keluaran asli pemeriksaan SQL C.1 yang disimpan dan ditampilkan pada editor VS Code.**

PNG utama: `screenshots/08b-vscode-sql-output.png`.

Gambar C1-S01a memperlihatkan salinan teks query `08-verification-query.txt`, dengan pengaturan pemenggalan baris untuk keterbacaan; query kanonisnya tersedia pada `08-verification.sql`. Gambar C1-S01b memperlihatkan keluaran asli pemeriksaan PostgreSQL melalui CLI yang telah disimpan dalam `08-sql-output.txt`. Kedua gambar merupakan capture editor VS Code, bukan terminal SQL live. Hasil mengidentifikasi database `pertemuan11_test` dan role `pbe11_baseline`, kemudian menampilkan empat event published, empat subjek notifikasi, serta hitungan outbox dan notifikasi masing-masing empat.

Hasil aktual pada `result.json` mencatat pasangan event berikut: ID 21 `enrollment.created`, ID 23 `enrollment.dropped`, ID 24 `enrollment.created`, dan ID 25 `enrollment.swapped`. Seluruhnya memiliki `published_at`. Notifikasi berpasangan dengan empat `outbox_id` yang sama, sehingga tindakan berhasil pada C.1 dapat direkonsiliasi dengan riwayat notifikasi. Jumlah outbox per permintaan adalah 1, 1, 2, 3, dan 4; khusus permintaan duplikat, tidak terjadi kenaikan jumlah baris.

### C1-S02 — pencatatan email

**Gambar C1-S02a. Hitungan empat pencatatan email akun uji C.1 yang ditampilkan pada editor VS Code.**

PNG utama: `screenshots/09a-vscode-mail-count.png`.

**Gambar C1-S02b. Empat baris log asli pemrosesan notifikasi C.1 yang dipilih dan ditampilkan pada editor VS Code.**

PNG utama: `screenshots/09b-vscode-runtime-log.png`.

Gambar C1-S02a menampilkan hasil hitungan tersimpan pada `09-mail-count.txt`: filter akun uji `mhsuji@` menghasilkan empat baris pencatatan email setelah persiapan awal yang disetujui. Isi log email mentah tidak ditampilkan. Gambar C1-S02b menampilkan empat baris log asli terpilih dalam `09-runtime-events.jsonl`, masing-masing berpesan `notification_processed` untuk `outbox_id` 21, 23, 24, dan 25, dengan `retry: 0` serta `already_sent: false`. Kedua gambar adalah capture editor atas hasil yang disimpan, bukan terminal live. Bukti ini mendampingi hasil jalur normal C.1; empat baris terpilih tersebut tidak membuktikan kelengkapan log atau ketiadaan error pada seluruh skenario C.7. Empat pencatatan email menunjukkan hasil mailer praktikum dan tidak membuktikan penerimaan oleh kotak masuk email eksternal.

## Keterangan troubleshooting yang dipertahankan

`screenshots/00-auth-header-missing.png` merupakan bukti troubleshooting HTTP 401 sebelum rangkaian C.1 berhasil. Gambar itu tidak dipilih sebagai gambar utama penerimaan final. Evidence tersebut tetap disimpan dan tidak disajikan sebagai respons pendaftaran awal yang berhasil. Uraian penyebab dan penanganan mengikuti catatan aktual Main Agent; bahan ini tidak menambahkan dugaan akar masalah atau mengklaim adanya perubahan source.

## Penempatan gambar dan status laporan

Dua belas PNG utama telah diperiksa secara visual: 01–05, 06a, 06b, 07, 08a, 08b, 09a, dan 09b. Menurut handoff capture Main Agent, seluruhnya merupakan PNG asli native **1920 × 1080**, rasio **16:9**, yang diambil dari aplikasi Windows melalui alur Computer Use menggunakan Windows GDI, tanpa resize. Tampilan Postman memuat method, path endpoint, HTTP status, serta bagian respons yang dibahas; nilai token tidak terlihat pada gambar utama. Teks `TARGET` pada nama request adalah label collection, sedangkan klaim hasil di atas mengikuti status respons aktual pada panel hasil. Capture VS Code dipilih dari empat berkas final yang disebut di atas; dua capture yang ditolak tidak dipilih sebagai gambar laporan.

PNG asli ditempatkan dengan rasio 16:9. Pada lebar penempatan 440 pt, tinggi proporsionalnya 247,5 pt; tinggi tersebut adalah ukuran layout, bukan perubahan piksel gambar sumber. Main Agent melaporkan bahwa dua belas PNG telah terintegrasi pada tab w11 dan gambar embedded PDF mempertahankan resolusi 1920 × 1080. Lima gambar lama tetap ada sehingga jumlah inline image w11 menjadi tujuh belas. Dua gambar notifikasi dipertahankan untuk menunjukkan respons lengkap.

Pemeriksaan visual render ekspor PDF C.1 telah dilakukan: penulis memeriksa halaman 177–182 dan 184–185; Main Agent memeriksa halaman 176 dan 183. Pada halaman yang diperiksa penulis, gambar, caption, serta analisis tidak terpotong atau saling menimpa. Pasangan gambar notifikasi pada halaman 181 dan pasangan gambar hitungan/log pada halaman 184 berada bersama caption dan uraian terkait. Rasio gambar tampak proporsional, dengan caption tebal dan posisi gambar terpusat. Teks screenshot lebih kecil daripada teks isi saat seluruh halaman ditampilkan, sehingga pembacaan detail screenshot dapat memerlukan zoom PDF; pemeriksaan ini tidak menyatakan seluruh detail nyaman dibaca pada cetakan A4 tanpa pembesaran. Ekspor Word tidak diperiksa dan tidak diperlukan pada tahap ini. Pemeriksaan terbatas ini bukan finalisasi seluruh laporan Pertemuan 11.

Readback final Main Agent mengonfirmasi bahwa paragraf kualitas pada cloud telah diperbarui dengan status pemeriksaan PDF aktual. Tab w11 memuat dua belas gambar C.1 baru, tanpa placeholder C.1 tersisa, dan empat caption historis B2A tetap dipertahankan. Jumlah inline image adalah tujuh belas, terdiri atas lima gambar lama dan dua belas gambar baru. Judul bagian C.1 berada pada paragraf tersendiri; caption 11 pt tebal, gambar terpusat berukuran 440 × 247,5 pt, dan pengaturan keep-with-caption telah diterapkan. Penulis tidak melakukan cloud edit.

| Kebutuhan | Status pada bahan ini |
|---|---|
| Hasil lima permintaan C.1 | TERSEDIA — screenshot asli 01–05 dan result.json |
| Empat notifikasi serta subjek modul | TERSEDIA — screenshot 06a/b dan hasil PostgreSQL |
| Keadaan akhir admin dan queue | TERSEDIA — screenshot 07, admin-response.json, result.json |
| Rekonsiliasi empat event dan empat pencatatan email | TERSEDIA — hasil aktual serta capture final VS Code 08a/b dan 09a/b telah diperiksa visual |
| Source dan snapshot baseline tetap | TERCATAT — checks source_unchanged dan baseline_snapshot_unchanged bernilai true pada result.json |
| Bukti B2A historis | DIPERTAHANKAN — tidak diganti atau dilabeli ulang sebagai Postman |
| Integrasi tab w11 oleh Main Agent | SELESAI pada cakupan C.1 — readback final mengonfirmasi 12 PNG baru, nol placeholder C.1, empat caption historis B2A, dan 17 inline image total |
| Pemeriksaan render ekspor PDF C.1 | DIPERIKSA VISUAL — halaman 177–182,184–185 oleh penulis; halaman 176,183 oleh Main Agent; tidak ditemukan crop/overlap pada halaman yang diperiksa penulis |
| Status paragraf kualitas cloud | DISINKRONKAN — Main Agent memperbarui status PDF menjadi hasil inspeksi aktual |
| Pembacaan detail screenshot tanpa zoom pada cetakan A4 | BELUM DIBUKTIKAN — teks screenshot lebih kecil daripada teks isi pada tampilan satu halaman |
| Ekspor Word | TIDAK DIPERIKSA — tidak diperlukan pada tahap ini |
| C.2–C.8 dan D.2/D.3 | PENDING — di luar pengulangan C.1 ini |

## Catatan persiapan integrasi — telah ditangani Main Agent

Catatan persiapan berikut dipertahankan sebagai jejak handoff, bukan pekerjaan cloud yang masih tertunda: penggantian judul rencana menjadi judul hasil aktual, pengisian placeholder C1-P/S, serta penyelarasan kalimat C.1 pada lampiran. Judul lama semula berbagi paragraf dengan inline image B2A-4 sehingga pemisahan judul harus menjaga gambar tersebut. Readback final Main Agent telah mengonfirmasi integrasi C.1 selesai dan gambar historis tetap utuh. C.2–C.8, D.2/D.3, serta tab pertemuan lain tidak menjadi bagian penyuntingan C.1 ini.

## Handoff

Bahan ini disusun dari pembacaan `result.json`, `admin-response.json`, checkpoint kedua, dokumen persiapan C.1, struktur tab w11 yang diaudit sebelumnya, pemeriksaan visual dua belas screenshot utama, serta render PDF halaman 177–182 dan 184–185. Render berasal dari ekspor privat `report-layout-final.pdf` yang disediakan Main Agent; berkas privat tersebut tidak dipublikasikan oleh penulis. `git status --short` dibaca sebelum penulisan awal. Tidak ada pengujian, HTTP, operasi database/layanan, perubahan source, capture UI baru, atau penyuntingan cloud yang dilakukan oleh penulis bahan ini. File yang ditulis hanya `REPORT_W11.md`. Laporan Pertemuan 11 secara keseluruhan belum dinyatakan final atau siap dikumpulkan.
