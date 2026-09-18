# Pertemuan 6 - Authorization dan RBAC

Authentication menjawab pertanyaan “siapa yang login?”, sedangkan authorization menentukan tindakan apa yang boleh dilakukan user tersebut. Pada pertemuan ini authentication dari tahap sebelumnya dilanjutkan dengan RBAC (Role-Based Access Control).

## RBAC dan permission

Role yang dipakai adalah `admin`, `staff`, dan `user`. Hak akses tidak ditulis sebagai pengecekan role yang tersebar di setiap handler. Role dipetakan ke permission melalui tabel `role_permissions`, lalu dibaca oleh `RoleRepository` dan dibentuk menjadi `PermissionSet`.

`PermissionSet` bekerja secara fail closed: permission yang tidak dikenal atau permission yang tidak dimiliki akan ditolak. Middleware `RequirePermission` dipasang di route setelah `RequireAuth`. Request tanpa login menghasilkan `401`, sedangkan user yang sudah login tetapi tidak memiliki permission menghasilkan `403`.

## Authorization pada user dan student

Authorization juga dicek di service, bukan hanya di route. Ini membuat aturan tetap berlaku walaupun service dipanggil dari tempat lain.

- Akses user memakai permission seperti `user:read:any`, `user:update:any`, dan permission terkait pengelolaan user.
- Akses student memakai permission untuk membaca, membuat, mengubah, dan menghapus data.
- Student memiliki `owner_id`. User biasa hanya boleh mengubah data yang dimilikinya, sedangkan role tertentu dapat mengakses data milik user lain sesuai permission `:any`.
- `owner_id` tidak boleh diubah lewat request update biasa.
- Data lama dengan owner kosong tidak otomatis dianggap milik user yang sedang login.

JWT masih dipakai sebagai identitas awal. Role pada JWT yang sudah lama tidak dianggap sebagai sumber izin terakhir; permission aktif dibaca kembali saat authorization dilakukan. Karena itu perubahan role bisa menghentikan akses dari token lama sesuai data role yang berlaku.

## Database dan file penting

Migration authorization ada di:

- `migrations/003_rbac.sql` untuk role, permission, dan relasinya.
- `migrations/004_student_permissions.sql` untuk permission student.

Beberapa file utama:

- `helper/authz.go` berisi helper permission dan pengecekan akses.
- `middleware/authz.go` berisi `RequirePermission`.
- `app/repository/role_repository.go` membaca role dan permission.
- `app/service/authz_rules.go` mengatur akses user.
- `app/service/student_authz_rules.go` mengatur akses berdasarkan ownership dan permission.
- `app/service/user_service.go` dan `student_service.go` menerapkan aturan tersebut.
- `route/route.go` memasang authentication dan authorization pada endpoint.

## Endpoint yang dilindungi

Route tetap mencakup authentication, user, student, dan achievement. Endpoint user memakai permission untuk operasi seperti membaca, membuat, memperbarui, dan menghapus user. Endpoint student dan achievement juga dibatasi sesuai permission yang diberikan pada role.

Detail daftar route ada di `route/route.go`; aturan sebenarnya tetap diperiksa ulang di service.

## Menjalankan dan menguji

Buat `.env` berdasarkan `.env.example`, siapkan PostgreSQL, lalu jalankan migration dari folder `migrations`. Dari folder ini:

```bash
go run .
```

Test unit dijalankan dengan:

```bash
go test ./...
```

Test authorization mencakup permission yang dikenal dan tidak dikenal, role admin/staff/user, akses berdasarkan owner, larangan mengubah owner, serta middleware `401` dan `403`.
