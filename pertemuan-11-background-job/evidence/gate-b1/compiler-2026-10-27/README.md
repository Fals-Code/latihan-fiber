# Gate B1 — pembuktian compiler, 27 Oktober 2026

Status: **Gate B1 selesai; build dan vet exit code 0. Perilaku aplikasi belum diuji.**

Perintah dijalankan dari nested module `pertemuan-11-background-job`, bukan
module root. `GOPROXY=off` dan `GOTOOLCHAIN=local` hanya berlaku pada child
process pemeriksaan: dependency memakai cache, tanpa mengubah konfigurasi
sistem. Hash `go.mod` dan `go.sum` sebelum/sesudah tersimpan pada metadata
masing-masing tahap dan tidak berubah. Tidak ada API, worker, migration,
request HTTP, atau operasi data PostgreSQL/MongoDB/Redis yang dijalankan.
Keputusan terbaru pengguna memisahkan kesiapan layanan dari Gate B1; catatan
readiness sebelumnya tetap dipertahankan sebagai riwayat.

## Urutan bukti asli

Setiap direktori menyimpan `output.txt` (stdout/stderr gabungan asli),
`metadata.json` (perintah, waktu Asia/Jakarta, exit code proses Go, hash),
`transcript.txt` (output asli dengan pembungkus metadata), serta salinan source
yang berlaku **sebelum perintah tahap itu**. Screenshot PNG adalah tangkapan
layar Notepad yang membuka transcript tersimpan tersebut; bukan terminal live
dan bukan gambar keluaran yang direkonstruksi. PNG tidak diedit.

| Tahap | Perintah | Exit code | Hasil dan screenshot |
|---|---|---:|---|
| 01-baseline | go build ./... | 4294967295 | Percobaan sandbox dihentikan setelah sekitar 214 detik; hanya proses build milik sesi ini dihentikan. Bukan bukti kegagalan compiler yang dipakai untuk keputusan perbaikan. |
| 02-baseline-compiler | go build ./... | 1 | [Output asli](02-baseline-compiler/output.txt), [screenshot](02-baseline-compiler/screenshot.png): P11-01 dan import tak terpakai. Build ini selesai normal. |
| 03-after-mailer | go build ./... | 1 | [Output asli](03-after-mailer/output.txt), [screenshot](03-after-mailer/screenshot.png): P11-01 hilang, import tak terpakai masih menghalangi. |
| 04-after-unused-import | go build ./... | 1 | [Output asli](04-after-unused-import/output.txt), [screenshot](04-after-unused-import/screenshot.png): P11-02 muncul. |
| 05-after-task-payload | go build ./... | 1 | [Output asli](05-after-task-payload/output.txt), [screenshot](05-after-task-payload/screenshot.png): P11-02 hilang, P11-03 muncul. |
| 06-final-build | go build ./... | 0 | [Metadata](06-final-build/metadata.json), [screenshot](06-final-build/screenshot.png): semua paket default berhasil dikompilasi, output kosong. |
| 07-final-vet | go vet ./... | 0 | [Metadata](07-final-vet/metadata.json), [screenshot](07-final-vet/screenshot.png): tidak ada diagnostik, output kosong. |

Percobaan 02 dan berikutnya memakai eksekusi dengan akses cache toolchain
yang lebih luas setelah percobaan sandbox terhenti. Tidak ada dependency
diunduh. Exit code yang relevan ialah `metadata.json.exit_code` milik proses
Go, bukan exit code harness Python yang berhasil menyimpan bukti.

## Tiga puzzle compiler untuk D.1

### P11-01 — method set Mailer

Lokasi diagnostik: `helper/mailer.go:58:9`.

```text
cannot use LogMailer{…} (value of struct type LogMailer) as Mailer value in return statement: LogMailer does not implement Mailer (method Send has pointer receiver)
```

`Send` mempunyai pointer receiver, sehingga nilai `LogMailer` tidak memenuhi
interface `Mailer`. Perbaikan terbatas pada constructor:

```diff
-return LogMailer{path: path, delay: delay, failEvery: int64(failEvery)}, nil
+return &LogMailer{path: path, delay: delay, failEvery: int64(failEvery)}, nil
```

Sebelum: tahap 02. Sesudah: tahap 03; error mailer hilang meskipun build masih
terhalang import tambahan. Source pembanding `helper__mailer.go.txt` tersedia
pada kedua direktori. Acuan kode dosen: Modul 11, halaman PDF 13.

### P11-02 — payload Asynq harus byte

Lokasi diagnostik: `app/service/outbox_relay.go:116:52` pada source sebelum fix.

```text
cannot use job (variable of struct type model.NotificationJob) as []byte value in argument to asynq.NewTask
```

`job` adalah struct; `asynq.NewTask` membutuhkan `[]byte`. Perbaikan mengimpor
`encoding/json`, memanggil `json.Marshal(job)` dengan penanganan error, lalu
memakai `payload` pada `NewTask`. Assignment enqueue menjadi `_, err = ...`
karena `err` sudah dideklarasikan. Aturan queue, retry, task ID, dan urutan
penandaan published tidak diubah. Sebelum: tahap 04. Sesudah: tahap 05;
bandingkan `app__service__outbox_relay.go.txt`. Acuan kode: halaman PDF 21.

### P11-03 — signature handler Asynq

Lokasi diagnostik pendaftaran: `worker.go:54:45`; akar masalah berada di
`app/service/notification_worker.go:44`.

```text
cannot use worker.HandleSend (value of type func(ctx "context".Context, t asynq.Task) error) as func("context".Context, *asynq.Task) error value in argument to mux.HandleFunc
```

Kontrak `HandleFunc` menerima pointer task. Perbaikan hanya parameter:

```diff
-func (w *NotificationWorker) HandleSend(ctx context.Context, t asynq.Task) error {
+func (w *NotificationWorker) HandleSend(ctx context.Context, t *asynq.Task) error {
```

Sebelum: tahap 05. Sesudah: tahap 06; bandingkan
`app__service__notification_worker.go.txt`. `worker.go` tidak diubah.
Acuan kode: halaman PDF 22.

## Penghambat integrasi tambahan — bukan puzzle dosen

Tahap 02 dan 03 membuktikan:

```text
helper\portfolio_request.go:9:2: "tugas1-go/pertemuan-11-background-job/app/repository" imported and not used
```

Import tersebut memang tidak dipakai oleh kode helper portfolio. Hanya baris
import itu yang dihapus; kontrak MongoDB dan logika portfolio tidak diubah.
Source sebelum ada pada `03-after-mailer/helper__portfolio_request.go.txt`;
sesudah pada tahap 04. Error hilang pada tahap 04. Temuan ini dicatat terpisah
agar tidak keliru dianggap salah satu dari sembilan puzzle Modul 11.

## Integritas dan batas hasil

Snapshot `baseline/puzzle-baseline.zip` dan manifest tidak diganti. SHA-256 ZIP:
`7a573d958135d5c45d4196ee7aae918eeb847564fd3652c2442732b0066d801a`.
Di awal sesi seluruh 96 entry manifest cocok dengan ZIP dan source aktif.
[Audit akhir](integrity-final.json) memverifikasi ulang ZIP, hash output,
perubahan source, serta keberadaan marker enam kandidat perilaku.
[Diff lengkap](source-changes.patch) membandingkan source/dokumen aktif dengan
snapshot; tidak ada perubahan kode selain empat file yang dijelaskan di atas.
Dokumen `docs/PUZZLE_CANDIDATES.md` diperbarui untuk menandai bukti B1.

P11-04–P11-09 belum diperbaiki: insert outbox Enroll di luar repository
transaction, event Drop salah, penandaan published sebelum hasil enqueue,
claim di luar repository transaction, pembungkusan SkipRetry yang salah,
dan nama queue worker berbeda. Pemeriksaan marker dan diff adalah bukti
statis bahwa kandidat dipertahankan, bukan bukti kegagalan runtime.

Belum dilakukan C.1–C.8, D.2, atau D.3. Gate B2 menunggu persetujuan dan
penyiapan lingkungan terpisah. Tidak ada commit/push. Bahan laporan `w11`
disiapkan pada [D1-w11.md](D1-w11.md); dokumen cloud belum diedit pada Gate B1.
