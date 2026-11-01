# Gate B2A — C.1, 29 Oktober 2026

**C.1 sesuai spesifikasi PDF hlm.31–32 pada checkpoint 05.** Ini bukan hasil
C.2–C.8 atau D.2/D.3, yang belum dijalankan. B1 tidak diulang sebagai pembuktian;
binary direbuild hanya untuk menjalankan revisi API/worker B2A.

## Lingkungan, migration, akun

Identitas tiga container khusus, image, network dan port direkam pada
[identity.json](identity.json). PostgreSQL host 55411 menggunakan
pertemuan11_test/pbe11_baseline; MongoDB host 27111/pertemuan11_mongo dan Redis
host 6379/DB0. API Go berhasil terhubung ke PostgreSQL dan MongoDB dari host
dan menjalankan EnsureIndexes sesuai startup Modul 10. Relay/worker berkomunikasi
dengan Redis dari host; queue admin menunjukkan pemrosesan nyata.

[migrations.json](migrations.json) menyimpan output asli dan hash sepuluh file:
bootstrap 001,002,003_auth,003_rbac,004_student_permissions,005_users_cursor;
kemudian 005_courses,006_seed_loadtest,007_review_permissions,008_outbox.
Tidak ada 009. Seed memuat 13 kelas dan 60 akun loadtest yang memang tidak dapat
login. [database-final.json](database-final.json) memuat jumlah seed serta
constraint UNIQUE(outbox_id) baseline.

Tiga akun didaftarkan lewat API: mhsuji dengan alamat mhsuji@student.unair.ac.id, salahalamat dengan
alamat salah@kampus.invalid, dan adminuji yang dipromosikan menjadi admin pada
database khusus. Ketiganya berhasil login; password/token hanya berada dalam
direktori privat yang diabaikan Git. [accounts.json](accounts.json) menyimpan
identitas dan status register tanpa secret. Konfigurasi efektif pengiriman:
MAILER_FAIL_EVERY=0, MAILER_DELAY_MS=1500, concurrency5, pool10; lihat
[execution-config-safe.json](execution-config-safe.json).

## Bukti berurutan

| Checkpoint | Hasil aktual | Kesimpulan terbatas |
|---|---|---|
| 01-initial-observation | Outbox5, queue pending5, notifikasi0, email0 | Worker mendengarkan queue salah; ambil screenshot sebelum fix |
| 02-queue-fixed-drained | Task diproses; pending0, notifikasi5, email5 | Perbaikan queue membuka gejala hilir; keadaan disimpan sebelum reset |
| 03-queue-fixed-c1 | HTTP201,409,204,201,200; event/notifikasi/email5 | Outbox melalui pool tetap commit ketika enrollment duplikat rollback |
| 04-outbox-fixed-c1 | Status sama; jumlah4/4/4, tetapi Drop masih created | Transaksi benar; jenis event/subject Drop masih salah |
| 05-final-c1 | Status sama; jumlah4/4/4, event dan subject tepat | C.1 memenuhi kriteria yang diminta |

Checkpoint pertama sempat gagal mengumpulkan JSON SQL multiline. Direktori
01-baseline dibiarkan sebagai jejak kegagalan collector, lalu observasi SQL/log
disimpan di 01-initial-observation sebelum perubahan/reset. HTTP awal tidak
dipulihkan dari ingatan. Raw response C.1 lengkap tersedia mulai checkpoint03.
Parser collector diperbaiki; ini bukan salah satu puzzle dosen.

Setiap direktori checkpoint menyimpan result.json, query.sql, sql-output.txt,
salinan source, app.log.txt/mail.log.txt, dan transcript. Empat screenshot asli
Notepad ada pada checkpoint01,03,04,05. Screenshot menampilkan transcript
tersimpan, **bukan terminal live**. Output HTTP/SQL/log menjadi bukti pendamping.
Log merupakan snapshot kumulatif; processed/completed Redis bersifat kumulatif
antarcheckpoint dan bukan jumlah task satu C.1. Mail.log dikosongkan sebelum
setiap pengulangan C.1 setelah bukti sebelumnya disimpan.

## Hasil final

Status HTTP: 201,409 ALREADY_ENROLLED,204,201,200.
Jumlah outbox setelah tiap request: **1,1,2,3,4**; request duplikat tidak membuat
event tambahan. Urutan event: created,dropped,created,swapped. Notifikasi
newest-first: pindah IF102→IF302; berhasil terdaftar IF102; keluar IF102;
berhasil terdaftar IF102. Tepat empat row notifikasi dan empat baris email mhsuji.
Respons lengkap dan subject utuh ada di [05-final-c1/result.json](05-final-c1/result.json).

Hanya tiga baris source diperbaiki pada dua file: QueueNotifications pada
config/queue.go, WithTx(tx).Insert pada Enroll, dan EventEnrollmentDropped pada
Drop. [source-changes.patch](source-changes.patch) menunjukkan perubahan minimal.
[verification.json](verification.json) merangkum pemeriksaan hasil tersimpan,
keutuhan snapshot, sisa puzzle, serta pemindaian secret pada evidence.

P11-06, P11-07, P11-08 tidak diubah. Tidak menjalankan C.2–C.8, D.2/D.3 atau
migration009. API dan worker yang dibuat controller telah dihentikan setelah
bukti final diambil; penghentian ini bukan pengujian C.6. Tiga container baseline
tetap tersedia; data final C.1 dipertahankan untuk kelanjutan bertahap.

## Laporan dan kelanjutan

[D1-B2A.md](D1-B2A.md) memuat narasi dan pasangan screenshot tiga puzzle.
Tab w11 diperbarui melalui konektor: baris D.1 P11-04/05/09, hasil aktual C.1,
serta empat gambar asli berikut caption. B1 dan tab lain tidak ditimpa.
Tata letak PDF final masih menunggu finalisasi laporan seluruh modul.

Gate B2B memerlukan persetujuan terpisah. Belum ada commit/push.
