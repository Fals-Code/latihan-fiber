# Bukti D.4 — Content Negotiation dan Error Handling Students

## Lingkungan

Pengujian HTTP dilakukan melalui Fiber `app.Test` dengan PostgreSQL database pengujian `hl_test`. Tidak ada operasi tulis, migration, atau perubahan database yang dilakukan.

- Akun sintetis memiliki permission `student:list`.
- Endpoint diuji: `GET /api/v1/students?limit=2`.
- Data halaman pertama aktual: student IDs `[7, 6]`.

## Content negotiation

### Accept: application/json

```text
STATUS=200
Content-Type=application/json
JSON_IDS=[7 6]
```

Status: **PASS**.

### Accept: text/csv

```text
STATUS=200
Content-Type=text/csv; charset=utf-8
CSV_BODY="ID,NIM,Name,Grade,IsActive,OwnerID,CreatedAt
7,910007,Cursor Proof New Top,90,true,0x3cd27a08658,2027-01-01 07:00:00 +0700 +07
6,910006,Cursor Proof 6,85,true,0x3cd27a08680,2026-01-01 07:00:06 +0700 +07
"
CSV_RECORDS=2
CSV_IDS=[7 6]
COLUMNS=7
JSON_IDS=[7 6]
```

CSV dibaca menggunakan `encoding/csv`. Header berisi 7 kolom dan setiap baris data juga memiliki 7 kolom. ID CSV sama dengan ID JSON pada query yang sama.

Status: **PASS**.

### Accept: application/xml

```text
STATUS=406
Content-Type=application/json
```

Status: **PASS**.

### Accept: */*

```text
STATUS=200
Content-Type=application/json
JSON_IDS=[7 6]
```

Status: **PASS**.

## Pencarian helper.Fail

Perintah yang dijalankan:

```text
grep/search: helper.Fail, helper.FailValidation, Fail(, FailValidation(
path: pertemuan-7-advanced-api-design/**/*.go
```

Hasil aktual:

```text
NO_MATCHES
```

Tidak ada pemanggilan `helper.Fail` atau `helper.FailValidation` pada source Pertemuan 7. Jalur students mengembalikan `AppError` melalui service dan ErrorHandler terpusat.

Status: **PASS**.

## Jalur endpoint students

- Route dan permission: `route/route.go:45-51`; GET list memakai `RequireAuth` dan permission `student:list`.
- Authentication: `middleware/auth.go:16-32`; token hilang/salah/kedaluwarsa menghasilkan unauthorized.
- Authorization: `middleware/authz.go:11-21`; role tanpa permission menghasilkan forbidden.
- List parser dan cursor: `app/service/student_service.go:43-75`.
- Repository error translation: `app/service/student_service.go:24-40`.
- Structured ErrorHandler: `config/app.go:35-60`.
- Error constructors dan stable codes: `helper/errors.go:11-58`.

## Tabel kode error students

| Endpoint/operasi | HTTP | Kode | Kondisi pemicu | Lokasi source | Bukti |
|---|---:|---|---|---|---|
| Semua endpoint students | 401 | `UNAUTHORIZED` | Authorization header hilang/salah, token invalid/expired | `middleware/auth.go:19-28` | Source; belum diuji pada putaran ini |
| GET `/students` | 403 | `FORBIDDEN` | User tidak memiliki `student:list` | `middleware/authz.go:13-18`, `route/route.go:46` | Source; belum diuji pada putaran ini |
| GET `/students` | 400 | `BAD_REQUEST` | Cursor tidak valid | `student_service.go:53-57` | Bukti runtime tersedia dari pengujian pagination sebelumnya |
| GET `/students` | 500 | `INTERNAL_ERROR` | Error repository/database atau gagal encode cursor | `student_service.go:62-71`, `helper/errors.go:52-54` | Source; database failure belum diuji pada putaran ini |
| GET `/students` | 406 | `NOT_ACCEPTABLE` | `Accept` tidak didukung, misalnya `application/xml` | `helper/negotiation.go:19-24` | Bukti runtime: HTTP 406 |
| GET `/students` | 200 | — | Query berhasil, JSON atau wildcard | `student_service.go:75`, `helper/negotiation.go:16-17` | Bukti runtime JSON dan wildcard |
| GET `/students` | 200 | — | Query berhasil, CSV | `student_service.go:75`, `helper/negotiation.go:19-22,27-60` | Bukti runtime CSV dan parsing berhasil |
| GET `/students/:id` | 400 | `BAD_REQUEST` | ID path bukan angka positif | `student_service.go:84-88` | Source; belum diuji pada putaran ini |
| GET `/students/:id` | 404 | `NOT_FOUND` | Student tidak ditemukan | `student_service.go:94-103`, `translateRepositoryError:28-31` | Source; belum diuji pada putaran ini |
| GET `/students/:id` | 403 | `FORBIDDEN` | User bukan owner dan tidak memiliki `student:read:any` | `student_service.go:98-100` | Source; belum diuji pada putaran ini |
| POST `/students` | 400 | `BAD_REQUEST` | JSON body malformed | `student_service.go:122-125` | Source; belum diuji pada putaran ini |
| POST `/students` | 422 | `VALIDATION_ERROR` | Tag validation create dilanggar | `student_service.go:127-129`, `helper/errors.go:56-58` | Source; D.2 validation evidence mendukung pola |
| POST `/students` | 409 | `CONFLICT` | NIM duplicate | `student_service.go:142-145`, `translateRepositoryError:33-36` | Source; belum diuji pada putaran ini |
| POST `/students` | 401 | `UNAUTHORIZED` | Tidak ada current user | `student_service.go:116-119` | Source; route auth juga melindungi |
| PUT `/students/:id` | 400 | `BAD_REQUEST` | ID invalid atau body malformed | `student_service.go:161-165,185-188` | Source; belum diuji pada putaran ini |
| PUT `/students/:id` | 422 | `VALIDATION_ERROR` | Field PUT/tag atau field wajib tidak lengkap | `student_service.go:190-209` | Source; belum diuji pada putaran ini |
| PUT `/students/:id` | 403 | `FORBIDDEN` | Tidak berwenang mengubah owner lain | `student_service.go:175-176` | Source; belum diuji pada putaran ini |
| PUT `/students/:id` | 404 | `NOT_FOUND` | Student/owner tidak ditemukan | `student_service.go:171-180` | Source; belum diuji pada putaran ini |
| PATCH `/students/:id` | 400 | `BAD_REQUEST` | ID invalid, body malformed, atau `{}` | `student_service.go:236-240,260-272` | D.2 HTTP evidence membuktikan empty patch 400 |
| PATCH `/students/:id` | 422 | `VALIDATION_ERROR` | Pointer field/tag dilanggar | `student_service.go:265-267`, `helper/errors.go:56-58` | D.2 HTTP evidence membuktikan name kosong 422 |
| PATCH `/students/:id` | 403 | `FORBIDDEN` | Tidak berwenang mengubah owner lain | `student_service.go:250-251` | Source; belum diuji pada putaran ini |
| PATCH `/students/:id` | 404 | `NOT_FOUND` | Student/owner tidak ditemukan | `student_service.go:246-255` | Source; belum diuji pada putaran ini |
| DELETE `/students/:id` | 400 | `BAD_REQUEST` | ID invalid | `student_service.go:299-303` | Source; belum diuji pada putaran ini |
| DELETE `/students/:id` | 403 | `FORBIDDEN` | Permission/ownership tidak cukup | `student_service.go:309-310` | Source; belum diuji pada putaran ini |
| DELETE `/students/:id` | 404 | `NOT_FOUND` | Student tidak ditemukan | `student_service.go:312-315` | Source; belum diuji pada putaran ini |
| DELETE `/students/:id` | 204 | — | Delete berhasil | `student_service.go:317-320` | Source; tidak diuji karena read-only |

## Status D.4

- JSON students berisi data: **PASS**.
- CSV students berisi header/data, parsing valid, kolom konsisten, ID sama dengan JSON: **PASS**.
- `Accept: application/xml` → 406: **PASS**.
- `Accept: */*` → JSON 200: **PASS**.
- Tidak ada `helper.Fail` pada source Pertemuan 7: **PASS**.
- Tabel kode error students: **PASS berdasarkan pembacaan source**; hanya sebagian kondisi diuji runtime.

Tidak ada bug source yang diperbaiki pada audit ini.

## Pengujian Regresi CSV OwnerID

### Masalah yang ditemukan

Output CSV students sebelumnya menserialisasi field `OwnerID *int` dengan `fmt.Sprint` langsung pada nilai reflect. Akibatnya, pointer berisi nilai muncul sebagai alamat memori, misalnya `0x3cd27a08658`, dan pointer nil muncul sebagai `<nil>`.

### Perubahan serialisasi

`helper/negotiation.go` kini menggunakan helper `csvValue` untuk setiap field:

- pointer non-nil didereference sehingga `*int` menghasilkan angka ID;
- pointer nil menghasilkan string kosong;
- field biasa tetap diserialisasi seperti sebelumnya;
- `writer.Flush()` tetap dipanggil dan `writer.Error()` tetap diperiksa.

### Unit test

`helper/negotiation_test.go` menguji nilai OwnerID, OwnerID nil, header tujuh kolom, jumlah kolom konsisten, parsing dengan `encoding/csv`, tidak adanya alamat pointer, dan kompatibilitas JSON.

Hasil:

```text
go test -count=1 ./pertemuan-7-advanced-api-design/helper -run 'TestToCSVSerializesOwnerPointers|TestStudentJSONPreservesOwnerID'
PASS
```

### HTTP regression test aktual

Pengujian dilakukan dengan Fiber `app.Test` dan PostgreSQL `hl_test`. Query JSON dan CSV memakai path, filter, cursor, dan limit yang sama: `/api/v1/students?limit=2`. Database terverifikasi dengan `current_database() = hl_test`.

JSON:

```text
status=200
Content-Type=application/json
IDs=[7 6]
owner_id=1 untuk kedua record
```

Body JSON aktual:

```json
{"success":true,"message":"daftar mahasiswa berhasil diambil","data":[{"id":7,"nim":910007,"name":"Cursor Proof New Top","grade":90,"is_active":true,"owner_id":1},{"id":6,"nim":910006,"name":"Cursor Proof 6","grade":85,"is_active":true,"owner_id":1}],"meta":{"limit":2,"next_cursor":"...","has_more":true}}
```

CSV:

```text
status=200
Content-Type=text/csv; charset=utf-8
ID,NIM,Name,Grade,IsActive,OwnerID,CreatedAt
7,910007,Cursor Proof New Top,90,true,1,2027-01-01 07:00:00 +0700 +07
6,910006,Cursor Proof 6,85,true,1,2026-01-01 07:00:06 +0700 +07
```

Hasil parsing aktual:

```text
CSV records=2
columns=7
CSV IDs=[7 6]
JSON IDs=[7 6]
pointer_repr=false
```

OwnerID terisi sebagai angka `1`, tidak ada `0x` atau `<nil>`, dan ID CSV sama dengan JSON.

Content negotiation regression:

```text
Accept: application/xml -> HTTP 406, application/json error response
Accept: */*            -> HTTP 200, application/json
```

Status HTTP OwnerID regression: **PASS**.

### Validasi setelah perbaikan

- `gofmt` — PASS.
- `go test -count=1 -timeout 90s ./pertemuan-7-advanced-api-design/...` — PASS.
- `go build ./pertemuan-7-advanced-api-design/...` — PASS.
- `go vet ./pertemuan-7-advanced-api-design/...` — PASS.

## Regresi Serialisasi CSV OwnerID

Bagian ini adalah pembaruan dokumentasi regresi serialisasi. Bukti HTTP historis pada bagian sebelumnya tetap dipertahankan sebagai bukti sebelum perbaikan; bagian ini tidak menggantikan atau mengklasifikasikan ulang bukti historis tersebut.

### Output unit test sebelum dan sesudah perbaikan

Sebelum perbaikan, unit test regresi CSV berada pada kondisi **RED** karena `OwnerID *int` diserialisasi sebagai nilai pointer reflect. OwnerID bernilai `42` muncul sebagai alamat memori, bukan `42`; OwnerID nil juga tidak menjadi kolom kosong yang diharapkan.

Sesudah perbaikan, unit test berikut **PASS**:

```text
go test -count=1 ./pertemuan-7-advanced-api-design/helper -run 'TestToCSVSerializesOwnerPointers|TestStudentJSONPreservesOwnerID'
PASS
```

Cakupan unit test mencakup OwnerID pointer bernilai `42` menjadi `42`, OwnerID nil menjadi string kosong, header dan seluruh baris konsisten tujuh kolom, tidak ada representasi alamat pointer, serta JSON tetap memuat `"owner_id":42`.

### Penyebab dan perubahan serialisasi

Penyebab masalah sebelumnya adalah `fmt.Sprint` dipanggil langsung pada nilai `reflect.Value` yang masih berupa pointer, sehingga hasilnya berupa alamat memori (`0x...`) alih-alih nilai yang ditunjuk.

`helper/negotiation.go` sekarang menggunakan `csvValue`: pointer non-nil didereference sebelum diformat, pointer nil menghasilkan string kosong, dan field non-pointer tetap diformat seperti sebelumnya.

### Validasi build dan test setelah perbaikan

```text
go test -count=1 ./pertemuan-7-advanced-api-design/...
PASS

go build ./pertemuan-7-advanced-api-design/...
PASS

go vet ./pertemuan-7-advanced-api-design/...
PASS
```

### Status HTTP setelah perbaikan

**NOT VERIFIED.** Pengujian HTTP setelah perbaikan tidak berhasil dijalankan. Percobaan terakhir berhenti pada kegagalan parsing skrip PowerShell sebelum koneksi database dibuat; tidak ada status HTTP, Content-Type, atau output CSV baru yang dapat dijadikan bukti pascaperbaikan. Skrip tersebut tidak diulang.

Bukti HTTP yang tersedia pada bagian sebelumnya tetap berstatus historis/sebelum perbaikan dan tidak boleh dianggap sebagai verifikasi HTTP setelah perbaikan.
