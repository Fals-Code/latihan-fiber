# Bukti D.2 — Validasi Deklaratif Students

## Sumber dan lingkungan

Instruksi D.2 berasal dari PDF Modul 7, halaman 23 (`modul7_extracted.txt:1334-1342`). Pengujian HTTP dilakukan melalui Fiber `app.Test` terhadap PostgreSQL database pengujian terpisah `hl_test`.

- Database: `hl_test`.
- Database utama `praktikum_backend` tidak disentuh.
- Student sintetis: ID `1`, NIM `910001`.
- User pengujian memiliki permission `student:update:any`.
- Nilai student dipulihkan setelah pengujian.

## Perbandingan kode sebelum dan sesudah

### Sebelum

Kode Bagian B PDF menggunakan request PATCH dengan field pointer yang salah pada `Username` user:

```go
type PatchUserRequest struct {
    Username string  `json:"username,omitempty" validate:"omitnil,min=3,max=30,alphanum"`
    Email    *string `json:"email,omitempty" validate:"omitnil,email,max=120"`
    IsActive *bool   `json:"is_active,omitempty"`
}
```

Pada source students sebelum validasi deklaratif, aturan create/replace dilakukan melalui fungsi manual `ValidateCreate` dan `ValidateReplace`. Snapshot kode sebelum perbaikan students tidak tersedia sebagai file terpisah; contoh di atas untuk PATCH berasal langsung dari PDF Modul 7.

### Sesudah

`app/model/student.go:15-30`:

```go
type CreateStudentRequest struct {
    NIM   int     `json:"nim" validate:"required,studentnim"`
    Name  string  `json:"name" validate:"required,min=1"`
    Grade float64 `json:"grade" validate:"gte=0,lte=100"`
}

type PatchStudentRequest struct {
    NIM      *int     `json:"nim" validate:"omitnil,gt=0"`
    Name     *string  `json:"name" validate:"omitnil,min=3"`
    Grade    *float64 `json:"grade" validate:"omitnil,gte=0,lte=100"`
    IsActive *bool    `json:"is_active" validate:"omitnil"`
}
```

Per-field validation dipanggil oleh `StudentService` melalui `helper.ValidateRequest` (`app/service/student_service.go:127-129,190-192,265-267`).

## Tag validasi students

| Struct/field | Tag | Makna |
|---|---|---|
| `CreateStudentRequest.NIM` | `required,studentnim` | NIM wajib dan harus positif |
| `CreateStudentRequest.Name` | `required,min=1` | Nama wajib dan minimal satu karakter |
| `CreateStudentRequest.Grade` | `gte=0,lte=100` | Nilai 0 sampai 100 |
| `ReplaceStudentRequest.NIM` | `required,gt=0` | NIM wajib dan positif |
| `ReplaceStudentRequest.Name` | `required,min=1` | Nama wajib |
| `ReplaceStudentRequest.Grade` | `gte=0,lte=100` | Nilai 0 sampai 100 |
| `PatchStudentRequest.NIM` | `omitnil,gt=0` | Dilewati hanya jika tidak dikirim; jika dikirim harus positif |
| `PatchStudentRequest.Name` | `omitnil,min=3` | Dilewati hanya jika tidak dikirim; jika dikirim minimal 3 karakter |
| `PatchStudentRequest.Grade` | `omitnil,gte=0,lte=100` | Dilewati jika nil; jika dikirim 0 sampai 100 |
| `PatchStudentRequest.IsActive` | `omitnil` | Dilewati jika tidak dikirim |

## Custom validation `studentnim`

`helper/validator.go:41-44`:

```go
_ = v.RegisterValidation("studentnim", func(fl validator.FieldLevel) bool {
    value, ok := fl.Field().Interface().(int)
    return ok && value > 0
})
```

Terjemahan pesan berada di `helper/validator.go:98-99`:

```go
case "studentnim":
    return "NIM harus berupa angka positif"
```

## Pointer dan `omitnil` pada PATCH

Pointer membedakan dua keadaan:

- `nil`: field tidak dikirim, sehingga tidak diubah.
- pointer ke string kosong: field dikirim dengan nilai kosong, sehingga tetap divalidasi dan ditolak oleh `min=3`.

`ApplyPatch` hanya menggabungkan field yang tidak nil (`app/service/student_rules.go:11-24`).

Aturan antar-field `IsEmptyPatch` tetap manual (`student_rules.go:27-33`) karena memeriksa hubungan antarfield: setidaknya satu field harus dikirim. Aturan ini tidak dapat dinyatakan oleh tag per-field.

## Bukti HTTP aktual

### 1. PATCH `{"name":""}`

```text
HTTP 422
code: VALIDATION_ERROR
message: validasi gagal
fields: {name: minimal 3 karakter}
```

Status: **PASS**.

### 2. PATCH `{"grade":90}`

```text
HTTP 200
message: data mahasiswa berhasil diperbarui
data.grade: 90
```

Nama tidak dikirim dan tetap dipertahankan. Status: **PASS**.

### 3. PATCH `{}`

```text
HTTP 400
code: BAD_REQUEST
message: tidak ada field yang diubah
```

Status: **PASS**.

### 4. PATCH sah `{"name":"D2 Restored"}`

```text
HTTP 200
message: data mahasiswa berhasil diperbarui
data.name: D2 Restored
```

Status: **PASS**.

Nilai record sintetis dipulihkan setelah seluruh skenario.

## Test otomatis terkait

- `app/service/student_rules_test.go:10-22` membuktikan perbedaan PATCH kosong dan `name` kosong.
- `app/service/student_rules_test.go:25-31` membuktikan field nil tidak mengubah field lain.
- `helper/remediation_test.go:39-46` membuktikan pesan custom `studentnim` tidak kosong.

## Kesimpulan

D.2 untuk validasi deklaratif students, custom validator, pointer/`omitnil`, dan perilaku PATCH telah **PASS** berdasarkan source dan HTTP integration test pada `hl_test`. Bukti historis source students sebelum perbaikan manual tidak tersedia sebagai snapshot terpisah; contoh “sebelum” yang dapat dikutip langsung berasal dari PDF Modul 7.
