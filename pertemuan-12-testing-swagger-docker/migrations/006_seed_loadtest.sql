-- Migration 006: 60 akun mahasiswa untuk uji beban
-- HANYA untuk database praktikum. Jangan dijalankan di server sungguhan.
--
-- Kolom password_hash diisi '!' — bukan hash bcrypt yang sah — sehingga akun
-- ini TIDAK MUNGKIN dipakai login. Alat uji beban membuat token sendiri.
INSERT INTO users (username, email, password_hash, role)
SELECT 'mhs' || LPAD(g::TEXT, 3, '0'),
       'mhs' || LPAD(g::TEXT, 3, '0') || '@student.unair.ac.id',
       '!', 'user'
FROM generate_series(1, 60) AS g
ON CONFLICT DO NOTHING;
