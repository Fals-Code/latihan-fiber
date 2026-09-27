# BUG_REPORT — Draf Audit D.1

Sumber baseline: kode Bagian B pada PDF Modul 7, dibandingkan dengan source lokal setelah perbaikan. Nomor baris pada kolom **Berkas/baris** adalah lokasi source lokal atau lokasi hasil rekonstruksi temp; nomor halaman PDF dicantumkan agar kode Bagian B dapat ditelusuri.

| # | Berkas/baris | Jenis | Gejala | Akar masalah | Perbaikan | Bukti |
|---:|---|---|---|---|---|---|
| C-1 | PDF hlm. 9, `config/app.go` ErrorHandler; rekonstruksi `d1-repro/compile/cause/config/main.go:5` | Compile | `e.cause undefined (cannot refer to unexported field cause)` | `cause` adalah field tidak diekspor dari package `helper`, sehingga package `config` tidak dapat mengaksesnya langsung. | Source lokal memakai cause accessor/internal handling tanpa akses field lintas package (`config/app.go:52-56`). | **TERBUKTI.** Output aktual tersimpan di `d1-repro/cause-output.txt`. |
| C-2 | PDF hlm. 14, `PatchUserRequest.Username`; rekonstruksi `d1-repro/compile/patch/main.go:3-5` | Compile | `invalid operation: req.Username != nil (mismatched types string and untyped nil)` | Field `Username` bertipe `string`, tetapi kontrak PATCH memerlukan pointer untuk membedakan field tidak dikirim dan string kosong. | `Username` menjadi `*string` di `app/model/user.go:27`. | **TERBUKTI.** Diagnostic aktual tersimpan di `d1-repro/patch-output.txt`. |
| C-3 | PDF hlm. 14, `ApplyPatch`; rekonstruksi `d1-repro/compile/patch/main.go:5` | Compile | `invalid operation: cannot indirect req.Username (variable of type string)` | Kode melakukan dereference pada nilai `string` non-pointer. C-2 dan C-3 adalah dua diagnostic dari satu akar masalah tipe PATCH. | Source lokal menggunakan pointer dan dereference hanya setelah nil check. | **TERBUKTI.** Diagnostic aktual tersimpan di `d1-repro/patch-output.txt`. |
| B-1 | PDF hlm. 8, `helper/errors.go:410-415` pada blok Bagian B | Behavioral | Validasi dikembalikan sebagai HTTP 400, padahal Bagian C mewajibkan 422. | Constructor `Validation` menetapkan `fiber.StatusBadRequest`. | `NewValidationError` lokal menetapkan status 422 (`helper/errors.go:43-51`). | **STATIC TERBUKTI;** HTTP baseline sebelum perbaikan belum direproduksi. |
| B-2 | PDF hlm. 9, `config/app.go:487-500` pada blok Bagian B | Behavioral | Level log 4xx/5xx tertukar: 4xx dicatat ERROR dan 5xx WARN. | Kondisi status memilih logger yang berlawanan dengan aturan C.7. | Lokal memakai WARN untuk 4xx dan ERROR untuk 5xx (`middleware/middleware.go:60-65`, `config/app.go:49-57`). | **STATIC TERBUKTI;** log baseline historis belum direproduksi. |
| B-3 | PDF hlm. 10, `middleware/middleware.go:536-550` pada blok Bagian B | Behavioral | Access log dapat mencatat status bawaan 200 walaupun ErrorHandler kemudian mengirim 4xx/5xx. | Logger memakai `c.Response().StatusCode()` sebelum ErrorHandler berjalan dan memakai nilai itu saat log. | Lokal memakai status dari `AppError`/error (`middleware/middleware.go:39-45,52`). | **STATIC TERBUKTI;** log HTTP baseline historis belum direproduksi. |
| B-4 | PDF hlm. 11, `app/service/user_service.go:599-607` | Behavioral | Unknown repository error diterjemahkan menjadi `nil`, bukan error 500 fail-closed. | `default` pada `translateError` mengembalikan `nil`. | Lokal mengembalikan `helper.Internal(err)` pada default translation. | **TERBUKTI.** Harness temp menghasilkan `UNKNOWN_ERROR_RESULT=nil`; output berasal dari `d1-repro/translate.go`. HTTP baseline belum direproduksi. |
| B-5 | PDF hlm. 12, `helper/validator.go:664-666` dan PDF C.3 | Behavioral | Password umum `password123` diterima walaupun harus ditolak. | Validator hanya memeriksa huruf/angka atau membalik hasil `passwordStrength`, tanpa daftar weak password yang diwarisi Modul 5. | Lokal menambahkan daftar weak password dan pesan `password terlalu umum` (`helper/validator.go:11-18,24-31,88-93`). | **TERBUKTI.** Sebelum perbaikan: `password123 => map[string]string(nil)`. Setelah perbaikan: test regresi PASS dan password ditolak. |
| B-6 | PDF hlm. 16, `app/repository/user_repository.go:920-921` | Behavioral | Hasil pagination users berurutan ASC, berlawanan dengan spesifikasi DESC. | Query baseline memakai `ORDER BY created_at ASC, id ASC`. | Lokal memakai `ORDER BY created_at DESC, id DESC` (`user_repository.go:81-85`). | **TERBUKTI STATIC/RUNTIME SQL.** Query ASC menghasilkan `[12,13]`; DESC menghasilkan `[37,36]` pada `praktikum_backend`. HTTP baseline sebelum perbaikan belum direproduksi. |

## Bukti CSV dan rekonsiliasi jumlah

Reproduksi terisolasi fungsi `WriteUsersCSV` dari PDF menunjukkan bahwa tanpa `writer.Flush()` output kosong:

```text
WITHOUT_FLUSH_START
WITHOUT_FLUSH_END
```

Setelah `Flush()` pada salinan temp, output berisi header dan dua baris data. Source lokal sudah memanggil `w.Flush()` di `helper/negotiation.go:59`; tidak ada perubahan pada tahap ini.

Modul menyatakan tepat enam behavioral bug, tetapi audit kandidat menghasilkan tujuh perilaku yang berbeda jika CSV tanpa `Flush()` juga dihitung: B-1 sampai B-6 di atas ditambah CSV. Karena CSV terbukti sebagai defect pada reproduksi potongan PDF, daftar resmi enam bug belum dapat dipetakan secara unik tanpa kunci baseline atau instruksi dosen yang menentukan kandidat mana yang tidak termasuk enam planted bug. CSV dicatat sebagai **ketidaksesuaian tambahan**, bukan dipaksakan menjadi baris ketujuh D.1.

## Ringkasan bukti

- Diagnostic compiler aktual: 3.
- Lokasi compile independen: 2.
- Akar masalah compile independen: 2.
- Behavioral bug yang dibuktikan static atau runtime: 6 pada matriks utama.
- Kandidat behavioral tambahan yang juga terbukti pada reproduksi terisolasi: CSV tanpa Flush.
- Bukti HTTP/curl baseline sebelum perbaikan: belum tersedia untuk B-1, B-2, B-3, B-4, dan B-6.
- Tidak ada database, source aplikasi, migration, commit, atau push yang diubah selama reproduksi ini.

## Reproduksi baseline tambahan

Reproduksi berikut dilakukan pada harness Fiber terisolasi di luar repository:

`C:\Users\falah\AppData\Local\Temp\opencode\d1-repro\http_baseline.go`

Output tersimpan di:

`C:\Users\falah\AppData\Local\Temp\opencode\d1-repro\http-baseline-output.txt`

Output ini adalah **REPRODUKSI BASELINE**, bukan log historis asli dari pengerjaan awal modul.

### B-1 — Validation 400 versus 422

Perintah:

```text
go run C:\Users\falah\AppData\Local\Temp\opencode\d1-repro\http_baseline.go
```

Hasil baseline:

```text
BASELINE /validation HTTP=400
```

Hasil setelah perbaikan yang dimodelkan dalam harness:

```text
FIXED /validation HTTP=422
```

Kode PDF yang menyebabkan gejala berada di `helper/errors.go`, PDF halaman 8, constructor `Validation`, yang menetapkan `Status: fiber.StatusBadRequest`. Source lokal memakai `NewValidationError` dengan status 422 (`helper/errors.go:56-58`).

### B-2 — Tingkat log 4xx/5xx

Hasil baseline harness:

```text
BASELINE LOGS:
time=... level=INFO msg=request level=ERROR status=200
...
```

Dalam harness, field `level=ERROR` menunjukkan klasifikasi aplikasi baseline untuk request 4xx maupun 5xx; ketiganya menggunakan level aplikasi ERROR karena kondisi baseline belum membedakan status. Ini menunjukkan 4xx tidak dipetakan ke WARN.

Hasil fixed:

```text
FIXED LOGS:
time=... level=INFO msg=request level=WARN status=422
time=... level=INFO msg=request level=WARN status=404
time=... level=INFO msg=request level=ERROR status=500
```

`level=INFO` adalah level outer handler harness, sedangkan field `level=WARN/ERROR` adalah klasifikasi aplikasi yang diuji. Source lokal memetakan 4xx ke WARN dan 5xx ke ERROR (`middleware/middleware.go:60-65`, `config/app.go:49-57`).

### B-3 — Status access log

Hasil baseline:

```text
BASELINE /bad HTTP=404
BASELINE LOGS:
time=... level=INFO msg=request level=ERROR status=200
```

HTTP aktual adalah 404, tetapi status yang dicatat baseline adalah 200. Hasil fixed:

```text
FIXED /bad HTTP=404
FIXED LOGS:
time=... level=INFO msg=request level=WARN status=404
```

Akar masalahnya adalah pembacaan `c.Response().StatusCode()` sebelum ErrorHandler final mengisi response. Source lokal menghitung status dari `AppError`/error (`middleware/middleware.go:39-45`) dan mencatat nilai tersebut (`:52`).

### Status bukti baseline

| Kandidat | Baseline HTTP/log terisolasi | Historis asli | Status |
|---|---|---|---|
| B-1 Validation 400 | HTTP 400, fixed HTTP 422 | Tidak tersedia | **REPRODUKSI BASELINE** |
| B-2 Level log | Baseline mengklasifikasikan request sebagai ERROR; fixed 4xx WARN dan 5xx ERROR | Tidak tersedia | **REPRODUKSI BASELINE** |
| B-3 Access-log status | Baseline HTTP 404 tetapi log status 200; fixed log status 404 | Tidak tersedia | **REPRODUKSI BASELINE** |

Reproduksi ini melengkapi bukti perilaku untuk B-1 sampai B-3 tanpa mengubah source aplikasi atau database. B-4 dan B-6 tetap memiliki bukti static/runtime SQL yang tercatat pada matriks utama, tetapi belum memiliki HTTP baseline penuh sebelum perbaikan.
