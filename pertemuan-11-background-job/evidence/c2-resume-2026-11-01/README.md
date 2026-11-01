# C.2 — kelanjutan checkpoint, 1 November 2026

Status akhir: **C.2 runtime PASS**, berdasarkan pengulangan lengkap pada fase 2
dan review QA bagian 8. Screenshot manual dan pemeriksaan PDF tetap PENDING.

## Fase 1 — pemulihan checkpoint terdahulu

Status pada akhir fase 1: **pemulihan backlog terbukti; penerimaan lengkap C.2
masih PENDING**. Catatan berikut mempertahankan kronologi sebelum persetujuan
pemulihan terbatas dan pengulangan lengkap; status terbaru ada pada fase 2.
Tidak ada reset, migration, seed, pengiriman HTTP, atau perubahan source aplikasi
pada checkpoint ini. C.1 tetap PASS historis. Bukti lama tidak ditimpa.

## Target dan keadaan awal

`00-preflight.json` merekam tiga container dengan ID, image, network, dan port
baseline yang cocok dengan lock lokal. PostgreSQL adalah
`127.0.0.1:55411/pertemuan11_test`, role `pbe11_baseline`; MongoDB menggunakan
container baseline pada port 27111; Redis baseline menggunakan port 6379/DB 0.
API dan worker historis sudah tidak hidup. Redis hidup dan antrean kosong.
Empat event/notifikasi C.1 ID outbox 21/23/24/25 tetap ada. Tiga event C.2
26/27/28 belum published; enrollment mahasiswa 61 masih kelas 4/5/6/7.

Log lokal sebelum proses baru disalin ke arsip privat yang diabaikan Git dan
tidak dikosongkan. `mail.log` saat preflight berukuran nol; empat email C.1
tetap dirujuk dari evidence historis, bukan diklaim ada di file log aktif.

## Eksekusi dan hasil

1. Source saat ini dikompilasi menjadi binary khusus C.2. Rekaman command,
   cwd, awal/akhir, output, exit code 0 dan hash ada di
   `pbe11-c2-current-build.json`. Ini persiapan runtime C.2, bukan C.8.
2. Hanya `pbe11-baseline-redis` dihentikan. API PID 12860 dan worker PID 9012
   kemudian dimulai dengan `MAILER_FAIL_EVERY=0`, delay 1500 ms. Karena
   proses historis telah mati, urutan ini adalah kelanjutan checkpoint,
   bukan pengulangan lengkap urutan API/worker-hidup-sebelum-Redis-mati.
3. `03-redis-off-state.json` menunjukkan pending tiga, published empat,
   notifikasi empat, dan nol email baru. Ada 29 `relay_delayed` dengan
   `published:0`. Loop berhenti pada kegagalan enqueue event 26;
   jangan menafsirkan ini sebagai 29 percobaan untuk setiap event.
4. Computer Use mendeteksi jendela Windows Security menghalangi Postman.
   Agent tidak mengendalikan dialog keamanan dan meminta pengguna menutupnya.
   Tidak ada request login, enrollment, admin, atau notifikasi dikirim.
5. Redis dipulihkan agar tidak ditinggalkan mati. Perintah aktual adalah
   `docker start pbe11-baseline-redis`, cwd nested module, melalui controller
   lokal. Helper `run` menolak nonzero exit; stdout container tercatat pada
   `05-recovery-timing.json`. stderr sukses tidak disimpan terpisah dan tidak
   direkonstruksi. Field container `state:exited` di file itu adalah identitas
   sebelum perintah start, bukan keadaan akhirnya.
6. Pengukuran konservatif dimulai sebelum pemanggilan `docker start`, pada
   09:06:29.929782 WIB. SQL pertama menunjukkan pending nol sekitar 1,531
   detik dan tujuh row claim/notifikasi sekitar 1,703 detik. Timestamp
   observasi diambil setelah query; snapshot tidak atomik lintas layanan.
   Tujuh row belum berarti mailer selesai: snapshot 06 masih active tiga.
7. Log tersimpan menunjukkan ketiga `notification_processed` ber-ID
   26/27/28, retry 0, `already_sent:false`; yang terakhir pada
   09:06:32.9681561 WIB, sekitar **3,04 detik** dari awal perintah start.
   `07-settled-state.json` memuat tiga email tiruan baru, notifikasi tujuh,
   pending nol, dan pending/active/scheduled/retry/archived queue seluruhnya
   nol. Completed tiga adalah nilai checkpoint ini, bukan jumlah C.1.
8. `08-final-identity.json` memverifikasi tiga container running/healthy dan
   identitas binary/proses API serta worker. Redis telah dipulihkan.

Seluruh hash 72 file Go sama antara preflight dan snapshot. Tidak ada
perbaikan puzzle baru. Koreksi P11-06 yang sudah ada kini didukung bukti
enqueue gagal tidak menandai published, kemudian backlog yang sama selesai
ketika Redis pulih. P11-07/08 tetap tidak diubah dan belum dibuktikan.

## Batas penerimaan dan langkah berikut

PASS penuh C.2 belum dinyatakan. Masih diperlukan tiga response 201 pada
source terkoreksi saat Redis mati, response Postman admin setelah lima detik
yang memuat pending tiga dan queue_error, serta response notifikasi akhir.
Response 201 historis dari source sebelum koreksi tidak diberi label sebagai
pengujian baru. Belum ada screenshot atau saved response baru.

Pemulihan terbatas data C.2 telah diajukan untuk persetujuan, belum dijalankan:
arsipkan state; batasi perubahan hanya notifikasi/event 26–28 dan enrollment
mahasiswa 61 kelas 5–7; hitung ulang counter hanya kelas 5–7; pertahankan
event/notifikasi C.1, sequence, dan Redis. Jangan menjalankan `reset11.sql`
karena akan menghapus data C.1. Setelah disetujui, verifikasi target dan
queue idle ulang; gunakan Postman untuk tiga enrollment baru yang memperoleh
ID baru, tanpa menghapus task Redis historis. Arsip lengkap dan guard SQL
harus disiapkan sebelum tindakan.

C.3 belum dimulai. Prasyaratnya ialah C.2 lengkap, arsip hasil C.2, persetujuan
tahap C.3, dan penataan state/konfigurasi failure yang sesuai PDF. Ketersediaan
container bukan bukti kesiapan penerimaan C.3.

## Provenance

Controller privat tersimpan lokal pada folder `.secrets`, tidak diunggah atau
disertakan dalam repository. Sebuah percobaan awal helper gagal karena path
root helper keliru sebelum mengakses layanan; helper diperbaiki dan preflight
baru berhasil. Docker dari sandbox awal ditolak akses named pipe, kemudian
pemeriksaan dijalankan melalui mekanisme eskalasi yang disetujui. Ini kendala
alat lokal, bukan kegagalan aplikasi. Review independen ada di `QA_REVIEW.md`.

## Fase 2 — pengulangan lengkap setelah persetujuan pengguna

Pengguna menutup dialog keamanan secara manual dan menyetujui pemulihan
terbatas C.2. Backup database disimpan privat sebelum transaksi. Hanya
notifikasi/event 26–28 dan enrollment mahasiswa 61 kelas 5–7 dipulihkan;
counter hanya kelas 5–7 dihitung ulang. Sequence dan Redis tidak direset.
Empat event/notifikasi C.1 diperiksa identik di dalam transaksi sebelum
COMMIT dan secara independen pada akhir pengujian. Lihat 09–13, termasuk
catatan kegagalan guard helper pada percobaan pertama yang tidak diterapkan.

API dan worker tetap hidup sebelum hanya Redis baseline dihentikan.
`14-rerun-before-redis-off-state.json` memuat kondisi empat event/notifikasi.
Semua HTTP di bawah dikirim melalui Postman Desktop, bukan controller/CLI.

| Checkpoint | Hasil aktual | Bukti |
|---|---|---|
| POST kelas 5 | 201, 28 ms | 15-enroll-5-ui.json dan body.json |
| POST kelas 6 | 201, 17 ms | 16b-enroll-6-cached-ui.json dan 16-enroll-6-body.json |
| POST kelas 7 | 201, 13 ms | 17b-enroll-7-cached-ui.json dan 17-enroll-7-body.json |
| Admin Redis mati | 200, 76 ms; pending 3, published 4, notifications_sent 4, queue_error | 18-admin-off-ui.json dan body.json; 19-rerun-off-state.json |
| Redis pulih | SQL pertama pending nol sekitar 2,375 detik; tujuh row terlihat sekitar 3,437 detik sejak perintah start | 20-rerun-recovery-timing.json dan 21-rerun-recovered-state.json |
| GET notifikasi | Dikirim sekitar 3,956 detik setelah awal pemulihan; 200, 6 ms; tujuh elemen | 23-notifications-recovered-ui.json dan body.json |
| Admin setelah pulih | 200, 18 ms; pending nol, published tujuh, notifications_sent tujuh, queue idle | 22-admin-recovered-ui.json dan body.json |
| Penyelesaian worker | Tiga mailer baru dan tiga notification_processed ID29–31 selesai sekitar 4,853 detik; retry0/already_sentfalse | 24-rerun-settled-state.json |

Admin-off dikirim 35,895 detik setelah POST terakhir: jeda minimal lima detik
terpenuhi, tetapi bukan pengukuran tepat pada detik kelima. Request admin
akhir dikirim setelah GET notifikasi; urutan tampilan collection tetap
admin-off, admin-pulih, notifikasi. Batas sepuluh detik dibuktikan oleh polling
SQL, waktu pengiriman GET notifikasi beserta body aktual, dan log penyelesaian.
Nilai completed enam mencakup dua rangkaian pemulihan hari ini, bukan enam
email baru pada pengulangan terakhir. Tidak ada klaim email eksternal.

Snapshot aksesibilitas 16/17 tertangkap sebelum tree response diperbarui,
meskipun screenshot observasi UI sudah menunjukkan response. File tersebut
dipertahankan. File 16b/17b membaca cache response yang sama tanpa Send dan
menjadi bukti status/durasi yang dapat diaudit. Body keenam response disalin
melalui tombol Copy response Postman dan disimpan tanpa header autentikasi.
Pembacaan clipboard memerlukan akses host; percobaan sandbox pertama tidak
menyimpan file karena validasi menolak data kosong.

Enam example ACTUAL dibuat dari body aktual tersebut. `26-postman-saved-examples.json`
mencatat readback kesamaan body, status, latency dan penghilangan autentikasi.
Metadata URL example kemudian disesuaikan menjadi string endpoint aktual
karena bentuk object pada adapter sempat tampil sebagai `[object Object]`.
Perubahan hanya metadata tampilan example; body/status hasil tidak diubah.
File 28 merekam pemeriksaan tampilan sebelum koreksi tersebut. UI sesudah
koreksi memperlihatkan endpoint admin dengan benar; tidak ada HTTP baru
dikirim untuk verifikasi example.

`27-final-identity-integrity.json` menghitung ulang 72 hash source aktual dan
memastikan semuanya identik dengan preflight. Field source_sha256 pada
snapshot sebelumnya hanya mereferensikan manifest preflight; pembuktian hash
independen berada pada file 27. File ini juga memverifikasi C.1 tetap identik,
tiga container healthy, dan identitas API/worker. Redis sudah running dan
antrean idle. `29-postman-secret-safety.json` memastikan shared value password
dan access token tetap kosong; nilai autentikasi hanya lokal.

Tidak ada puzzle baru diperbaiki. Bukti awal P11-06 tetap berasal dari
percobaan gagal 29 Oktober; kode koreksi yang sudah ada kini mempunyai
pembanding runtime sesudah koreksi. P11-07/08 belum diubah atau dibuktikan.
C.3 tidak dijalankan. Persetujuan pemulihan C.2 tidak mencakup reset11.sql
untuk C.3 yang juga mengubah akun uji lain dan menghitung ulang semua counter.

## Screenshot manual

Buka workspace Week 11, folder C2, perluas masing-masing request, lalu buka
example ACTUAL 2026-11-01. Jangan tekan Send/Try/Run. Ambil enam frame
Postman sesuai Gambar23a–23f dan frame tambahan VS Code untuk file20/24
(timing, outbox, mailer, log). Labeli sebagai hasil tersimpan. Latency hanya
boleh ditampilkan jika memang terlihat pada aplikasi; metadata tersimpan
bukan izin menambahkan teks ke gambar. Placeholder/caption berada di laporan
dan docs/W11_SCREENSHOT_MANUAL.md. PNG manual, pemasangan, dan pemeriksaan
keterbacaan PDF belum dilakukan pada tahap ini.
