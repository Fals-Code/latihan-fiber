# Postman — Pertemuan 11

Folder ini berisi collection API serta template environment untuk pengujian background job.

## File yang dipakai

- [Week-11.postman_collection.json](Week-11.postman_collection.json) — request C.1–C.8 dan tugas lanjutan.
- [Week-11-Baseline.postman_environment.json](Week-11-Baseline.postman_environment.json) — template variabel tanpa credential.

## Cara menggunakan

1. Impor kedua file melalui **Postman → Import**.
2. Pilih environment yang sesuai, lalu periksa `base_url` (contoh lokal: `http://127.0.0.1:3332/api/v1`).
3. Jalankan request login untuk akun pengujian yang tersedia.
4. Gunakan Bearer token sesuai peran mahasiswa atau admin.
5. Kirim request sesuai skenario yang akan diuji dan bandingkan hasilnya dengan log worker serta data outbox.

Password dan access token tidak disimpan dalam file yang dibagikan. Variabel sensitif harus diisi di lingkungan lokal.

## Bukti pengujian

Bukti C.1 tersedia di [evidence/c1-postman-2026-10-29](../evidence/c1-postman-2026-10-29/). Untuk C.2–C.8, gunakan catatan pada [evidence/c2-resume-2026-11-01](../evidence/c2-resume-2026-11-01/) dan [evidence/completion-2026-11-01](../evidence/completion-2026-11-01/).

Saved Example menyimpan respons pada saat pengujian sebelumnya, bukan hasil request baru.
