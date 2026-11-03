# Desain: Animasi UI VibeLog

## Tujuan

Menambahkan animasi menyeluruh pada VibeLog agar web terasa hidup dan nyaman
dipakai, tanpa mengorbankan gaya pixel/neo-brutalist dan tanpa library tambahan.

## Pendekatan yang dipilih

Kombinasi dua lapis, semuanya CSS-based (tanpa dependensi baru):

1. **Halus & fungsional** untuk semua interaksi: hover tombol, transisi form,
   munculnya kartu, transisi halaman.
2. **Retro playful** secukupnya pada elemen khas: efek tekan tombol ala saklar,
   blink teks pixel, kartu masuk dengan sedikit overshoot, angka statistik
   count-up.

Alternatif yang dipertimbangkan dan ditolak: memakai library (GSAP/anime.js) -
menambah berat bundle, tidak sepadan untuk animasi yang mayoritas CSS. Dan
animasi hanya-halus tanpa sentuhan retro - kurang berkarakter untuk brand pixel.

## Cakupan animasi (full)

- Transisi antar halaman (fade + slide kecil) di RouterView.
- Kartu fitur landing muncul bertahap (stagger).
- Kartu entri dashboard muncul bertahap.
- Panel form (entri & proyek) slide-in + fade saat dibuka.
- Tombol: hover lift, active press turun (translateY) + shadow berkurang.
- Tombol primary: efek "tekan" jelas.
- Teks pixel headline: blink lembut pada kursor underscore.
- Angka statistik (StatsCards): count-up saat pertama tampil.
- Chip/tag: hover dan active feedback.
- Prompt box di EntryCard: buka/tutup halus (bukan v-show instan).
- Logo mark: sedikit wiggle saat hover.
- Elemen list (fitur, stack pills): hover lift.
- Form field focus: sudah ada outline oranye; tambah transisi.

## Aksesibilitas

Semua animasi dibungkus `@media (prefers-reduced-motion: reduce)` sehingga
user dengan setting reduced-motion mendapat versi minimal (tanpa transform
dan blink, count-up langsung ke nilai akhir).

## Batasan

- Tanpa dependency baru.
- Bahasa UI tetap Bahasa Indonesia.
- Tidak mengubah logika data/API.
- Tetap memakai variabel warna yang ada (`--orange`, `--line`, dll).