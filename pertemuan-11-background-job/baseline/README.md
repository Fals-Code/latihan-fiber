# Snapshot baseline puzzle

puzzle-baseline.zip menyimpan seluruh source dan dokumentasi awal, termasuk
sembilan kandidat Modul 11 yang belum diperbaiki. manifest.json memuat SHA-256.
Tidak ada .env aktual, credentials, logs, screenshot, atau evidence lama.

Ekstrak ke direktori terpisah saat memerlukan salinan pembanding. Snapshot bukan
commit Git dan bukan bukti bahwa kode dapat dikompilasi atau berjalan.
File extension d2 ikut disimpan untuk kelengkapan source; varian default !d2
tetap dipilih Go selama tag d2 tidak diberikan.

Untuk D.3, setelah kode sudah diperbaiki dan dibuktikan, ambil satu perubahan
perilaku yang lolos compiler dari snapshot pada salinan kerja khusus, jalankan
uji yang membuktikan kegagalannya, lalu pulihkan. Jangan mengganti seluruh source
dengan snapshot saat hanya satu kesalahan yang hendak dibandingkan.
