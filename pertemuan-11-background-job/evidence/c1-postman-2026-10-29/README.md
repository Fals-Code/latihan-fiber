# C.1 melalui Postman — 29 Oktober 2026

Rangkaian C.1 yang berhasil memenuhi kontrak yang diperiksa: HTTP
**201, 409, 204, 201, 200**, jumlah outbox **1, 1, 2, 3, 4**, empat event
yang tepat, empat notifikasi, dan empat baris email mailer tiruan. Kesimpulan
ini terbatas pada C.1. Pembaruan tab laporan w11 dan pemeriksaan halaman PDF
C.1 telah dicatat pada handoff akhir di bawah. Pemeriksaan Word masih
**PENDING**; tidak ada klaim laporan seluruh modul final.

Acuan: *Modul 11 — Background Job*, hlm.31–32, serta
[rencana C.1](../../docs/C1_POSTMAN_PLAN.md), source route/auth, course service,
notification repository, dan queue aktual. Gate B2A yang telah diterima tetap
utuh dan merupakan pengulangan berbeda. Berkas `result.json` merangkum hasil
baru; status HTTP berasal dari tampilan Postman asli, bukan transcript HTTP
yang dibuat ulang.

## Lingkup, target, dan kondisi awal

Main Agent adalah satu-satunya pengendali database, proses API/worker, dan UI.
Mutasi KRS C.1 dikirim melalui UI Postman. QA meninjau artifact secara read-only
dan hanya menulis README ini. Tidak ada C.2–C.8, D.2/D.3, migration, perbaikan
source, commit, atau push pada pekerjaan C.1 ini.

- Folder aplikasi: `pertemuan-11-background-job`.
- API: `http://127.0.0.1:3332/api/v1`; screenshot memakai variabel `{{base_url}}`.
- PostgreSQL: host `127.0.0.1:55411`, database `pertemuan11_test`, role
  `pbe11_baseline`. Nilai `server_port: 5432` pada hasil SQL adalah port internal
  PostgreSQL; `identity.json` mencatat mapping host 55411 ke container 5432.
- Redis: target baseline DB0, queue `notifications`, host port 6379.
  Identitas container khusus praktikum tercatat pada `identity.json`.
- `before.json` pada 08:00:49 WIB masih memuat hasil B2A sebelumnya dan queue
  pending/active/scheduled/retry/archived nol.
- Arsip privat sebelum reset dicatat pada `archive-metadata.json`: metadata
  ukuran/hash log, dump database, dan arsip evidence lama. Isi log dan secret
  tidak ditampilkan dalam README maupun screenshot.
- `reset.json` pada 08:01:16 WIB merekam satu transaksi: `DELETE 4`
  notifications, `DELETE 4` outbox, `DELETE 1` enrollment, `UPDATE 13` courses,
  lalu `COMMIT`. Reset menghitung ulang counter **seluruh 13 kelas**, selain
  menghapus riwayat dua akun uji. Sequence tidak direset. Setelah arsip,
  ukuran mail.log tercatat nol. Tidak ada reset ulang sesudah kendala auth.
- `runtime.json` merekam startup dua proses terpisah pada 08:01:39 WIB dan
  SHA-256 binary. Metadata startup ini tidak menyatakan status proses saat
  README dibaca.

## Kendala autentikasi yang dipertahankan

`screenshots/00-auth-header-missing.png` merekam HTTP **401 Unauthorized**
dengan pesan `header Authorization tidak ada atau salah bentuk` sebelum
enroll berhasil. `checkpoint-01.json` pada 08:05:23 WIB membuktikan outbox dan
notifikasi masih kosong; `resume-before.json` juga mencatat baseline kosong
serta queue idle. Checkpoint ini dipertahankan sebagai kegagalan autentikasi,
bukan sebagai langkah C.1 yang berhasil.

Menurut catatan pengoperasian Main Agent, konfigurasi auth root collection
dan inheritance request sudah benar. Pengisian nilai melalui `set_value`
tidak bertahan pada UI; pengetikan normal melalui `type_text` berhasil
menyimpan variabel. Postman kemudian mendeteksi token mahasiswa dan
memindahkannya ke Local Vault. GET terautentikasi menghasilkan 200, lalu
enroll dapat dilanjutkan. Penyebab operasional ini tidak diperlakukan sebagai
bug source. Bukti PNG 401 hanya membuktikan header hilang/salah bentuk;
PNG tersebut tidak membuktikan token kedaluwarsa.

Untuk monitoring admin, override `{{admin_access_token}}` diverifikasi oleh
Main Agent dan GET admin menghasilkan 200. `login-safe.json` serta
`admin-login-resumed-safe.json` hanya menyimpan status dan identitas akun,
tanpa nilai credential/token. Tidak ada screenshot login body, Authorization,
Vault, atau Console yang membocorkan secret.

## Hasil berurutan dan rekonsiliasi

| Langkah | Hasil UI aktual | Outbox akun mhsuji | Bukti pendamping |
|---|---|---:|---|
| Enroll IF102 | POST `/courses/2/enroll`, 201; id 2, code IF102 | 1 | `checkpoint-01-success.json` |
| Enroll duplikat | POST route sama, 409; `ALREADY_ENROLLED` | 1 | `checkpoint-02.json` |
| Keluar IF102 | DELETE route sama, 204; body kosong | 2 | `checkpoint-03.json` |
| Enroll kembali | POST route sama, 201; id 2, code IF102 | 3 | `checkpoint-04.json` |
| Pindah 2 ke 4 | POST `/enrollments/swap`, 200; body from 2/to 4, respons IF302 | 4 | `checkpoint-05.json` |

Checkpoint berhasil berlangsung pada 08:12:08, 08:12:55, 08:13:29, 08:13:55,
dan 08:14:39 WIB. Semua event baru memiliki `published_at` terisi:

| Outbox ID | Event | Notifikasi terkait |
|---:|---|---:|
| 21 | `enrollment.created` | 19 |
| 23 | `enrollment.dropped` | 20 |
| 24 | `enrollment.created` | 21 |
| 25 | `enrollment.swapped` | 22 |

ID 22 yang terlewat konsisten dengan sequence BIGSERIAL yang tidak ikut
rollback saat permintaan duplikat. Bukti tidak adanya event tambahan adalah
jumlah outbox tetap satu pada dua checkpoint pertama, bukan sekadar gap ID.

GET `/notifications/mine` menghasilkan 200 dengan tepat empat elemen,
urut `outbox_id` menurun: 25, 24, 23, 21. Subjeknya berurutan:

1. `KRS: pindah dari IF102 Basis Data Lanjut ke IF302 Keamanan Informasi`.
2. `KRS: berhasil terdaftar di IF102 Basis Data Lanjut`.
3. `KRS: keluar dari IF102 Basis Data Lanjut`.
4. `KRS: berhasil terdaftar di IF102 Basis Data Lanjut`.

Capture notifikasi pukul 08:15:50 WIB berada lebih dari lima detik sesudah
checkpoint swap. Dua frame menunjukkan satu respons yang sama, dengan bagian
tumpang tindih untuk menjaga kontinuitas. Urutan daftar ini tidak digunakan
untuk menyimpulkan urutan kedatangan email.

`08-verification.sql` dan `08-sql-output.txt` memuat query baca dengan filter
`mhsuji`, empat event published, empat subjek, serta count outbox/notifikasi
**4/4**. `09-mail-count.txt` merekam hitungan **4** baris `mhsuji@` sesudah
arsip dan pengosongan awal. Ini bukti pencatatan mailer tiruan, bukan bukti
penerimaan di kotak masuk eksternal. `09-runtime-events.jsonl` berisi empat
event `notification_processed` untuk outbox 21,23,24,25, seluruhnya retry 0
dan `already_sent: false`, pada 08:11:36–08:14:30 WIB.

`admin-response.json` dan screenshot 07 memperlihatkan HTTP 200,
`outbox_pending: 0`, `outbox_published: 4`, `notifications_sent: 4`, queue
`notifications`, pending/active/scheduled/retry/archived nol, completed 22,
processed_today 4, failed_today 0, tanpa `queue_error`. Completed merupakan
counter kumulatif Redis; nilainya 22 tidak berarti C.1 menghasilkan 22 email.

## Screenshot asli dan batas pemeriksaan

Main Agent mengambil PNG melalui Computer Use langsung dari aplikasi aktual.
QA membuka 13 PNG terpilih berikut menggunakan `view_image` dan membaca dimensi
dari header PNG: semuanya **1920 × 1080**. Tidak ada pengeditan raster,
resize, atau rekonstruksi gambar pada audit QA ini. Waktu tabel adalah
LastWriteTime lokal berkas, bukan timestamp HTTP presisi.

| PNG di `screenshots/` | Aplikasi / checkpoint | Dimensi | Waktu WIB |
|---|---|---|---|
| `00-auth-header-missing.png` | Postman, kegagalan awal 401 | 1920 × 1080 | 08:05:42 |
| `01-enroll-201.png` | Postman, enroll awal 201 | 1920 × 1080 | 08:12:08 |
| `02-duplicate-409.png` | Postman, duplikat 409 | 1920 × 1080 | 08:12:56 |
| `03-drop-204.png` | Postman, drop 204 tanpa body | 1920 × 1080 | 08:13:30 |
| `04-reenroll-201.png` | Postman, enroll kembali 201 | 1920 × 1080 | 08:13:56 |
| `05-swap-200.png` | Postman, request swap dan respons 200 | 1920 × 1080 | 08:14:40 |
| `06a-notifications-top.png` | Postman, bagian atas respons notifikasi 200 | 1920 × 1080 | 08:15:50 |
| `06b-notifications-bottom.png` | Postman, lanjutan respons yang sama | 1920 × 1080 | 08:16:05 |
| `07-admin-queue-200.png` | Postman, monitoring admin 200 | 1920 × 1080 | 08:18:17 |
| `08a-vscode-query.png` | VS Code, salinan query baca untuk tampilan | 1920 × 1080 | 13:20:43 |
| `08b-vscode-sql-output.png` | VS Code, output asli PostgreSQL CLI tersimpan | 1920 × 1080 | 13:21:23 |
| `09a-vscode-mail-count.png` | VS Code, ringkasan hitungan mailer tersimpan | 1920 × 1080 | 13:21:45 |
| `09b-vscode-runtime-log.png` | VS Code, empat baris runtime asli terpilih | 1920 × 1080 | 13:23:58 |

Method, suffix route, status, serta data penting terbaca pada file asli.
Postman menampilkan `{{base_url}}`; target literal dijelaskan oleh metadata
konfigurasi aman. Frame 06b dimulai pada bagian tengah respons; frame 06a
menyediakan pembuka dan konteksnya. Petunjuk bantuan Postman pada respons
error tidak menutupi status atau pesan error. Tidak ada secret yang tampak
pada 13 PNG terpilih yang diperiksa.

Empat gambar VS Code menampilkan berkas evidence tersimpan di editor,
**bukan terminal atau SQL client live**. Capture pada 13:20–13:23 WIB tidak menandakan
pengiriman ulang request ataupun eksekusi ulang SQL. `08a` menampilkan
`08-verification-query.txt`, salinan `08-verification.sql` yang telah
dibandingkan secara offline dan identik setelah whitespace diabaikan.
Salinan teks menghindari validator dialek MSSQL pada editor. `08b`
memperlihatkan output asli PostgreSQL CLI, termasuk identitas database/role,
empat event published, empat subjek, dan count 4/4. `09a` memperlihatkan
ringkasan hitungan aman; `09b` memperlihatkan empat baris asli
`notification_processed` yang dipilih dari app.log berdasarkan empat ID
outbox dan offset awal sebelum pengulangan. Makna dan timestamp pada baris
terpilih tetap dipertahankan; wrapping editor hanya mengubah tampilan.

`rejected-08a-overlay.png` dan `rejected-08a-mssql-editor.png` dipertahankan
sebagai capture yang tidak dipilih. Kendalanya adalah overlay NVIDIA dan
diagnostik dialek editor MSSQL, **bukan kegagalan runtime PostgreSQL**.
Indikator diagnostik editor yang masih tampak pada tab/status bar tidak
digunakan sebagai hasil pengujian SQL. Tooltip kecil di status bar gambar
09a tidak menghalangi bukti hitungan.

Pemeriksaan tampilan PDF terbaru dijelaskan pada handoff akhir di bawah.
**PENDING:** pemeriksaan Word. Teks halus screenshot dapat memerlukan zoom
meskipun PNG asli berukuran Full HD.

## Audit QA dan handoff

QA menjalankan pembacaan file aman dengan PowerShell, `git status --short`,
perhitungan SHA-256, pembacaan header PNG, serta `view_image`. Audit offline
memastikan **87** file yang tercantum di `source-before.json` tidak berubah,
count checkpoint 1,1,2,3,4, dan empat runtime event sesuai ID/event di atas.
SHA-256 `baseline/puzzle-baseline.zip` tetap
`7a573d958135d5c45d4196ee7aae918eeb847564fd3652c2442732b0066d801a`.
Ini verifikasi artifact lokal; QA tidak mengulang HTTP, SQL, server, worker,
build/vet, reset, maupun pembacaan log/secret aktual.

`verify_actual.py` dibaca untuk mengetahui asal rekonsiliasi, tetapi tidak
dijalankan oleh QA karena script melakukan query database dan menulis
artifact. README ini merupakan satu-satunya file yang ditulis QA.

Pemindaian offline pada tahap audit screenshot terhadap **25 berkas teks publik** pada direktori evidence
ini menghasilkan nol kecocokan untuk pola JWT, private key, service API key,
URI yang mengandung username/password, dan nilai secret berformat JSON yang
tidak kosong. Nilai privat `.env`/token tidak dibaca untuk audit ini.
Pemeriksaan pola ini terbatas pada bentuk yang dicari dan tidak menggantikan
pemeriksaan visual 13 gambar terpilih; hasilnya tidak diklaim sebagai jaminan
universal ketiadaan secret.

Hasil perilaku C.1 baru terbukti oleh rangkaian evidence di atas. Audit visual
PNG Postman dan VS Code terpilih telah selesai. Evidence B2A, snapshot
baseline, source, dan bukti 401 awal dipertahankan. Tidak ada klaim PASS
untuk laporan final seluruh modul atau tahap lain.

## Handoff akhir — laporan dan batas gate

`completion.json`, `report-verification.json`, dan `screenshot-manifest.json`
telah dibaca QA setelah Main Agent menyerahkan hasil akhir. Manifest memuat
13 PNG 1920 × 1080: 12 dipilih untuk w11, sedangkan gambar 401 tetap disimpan
sebagai evidence autentikasi dan tidak dimasukkan ke laporan.

Main Agent memperbarui tab w11 dan melakukan readback: **12 gambar baru**
terpasang, placeholder C.1 tersisa **0**, empat caption historis B2A tetap
ada, dan total gambar w11 **17** (5 lama + 12 baru). Menurut readback yang
sama, 21 gambar lama tab w10 tidak berubah. Pembaruan hanya melengkapi C.1;
hasil gate sebelumnya tetap dipertahankan.

Ekspor PDF berhasil diperoleh melalui fallback fetch, kemudian halaman
**176–185** ditinjau: Main Agent memeriksa 176 dan 183; penulis laporan
memeriksa 177–182, 184, dan 185. `report-verification.json` mencatat tidak
ditemukan crop konteks, tumpang tindih, atau distorsi; penempatan gambar
440 × 247,5 pt mempertahankan rasio 16:9 dan piksel gambar baru yang
tertanam tetap 1920 × 1080. Teks screenshot yang halus dapat memerlukan
zoom. **Ekspor Word belum diperiksa.** Hasil review PDF ini berasal dari
Main Agent/penulis laporan dan artifact verifikasinya; QA tidak mengulang
operasi cloud atau render PDF pada handoff ini.

Pada pemeriksaan akhir read-only Main Agent, proses API PID 45256 dan worker
PID 42748 cocok dengan binary milik proyek, dan port API 3332 terbuka.
Keduanya dipertahankan berjalan seperti saat resume; tidak ada aksi stop
atau pengujian C.6. Metadata proses adalah observasi saat pemeriksaan, bukan
jaminan proses tetap berjalan pada waktu berikutnya.

Gate pekerjaan dihentikan pada C.1 dan pembaruan bukti laporannya. Tidak ada
eksekusi C.2–C.8, D.2/D.3, reset ulang setelah resume, atau perubahan source.
Pemeriksaan Word dan pekerjaan modul berikutnya tetap mengikuti persetujuan
tahap yang bersangkutan.
