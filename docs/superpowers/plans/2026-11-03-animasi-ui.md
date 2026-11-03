# Animasi UI VibeLog Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Menambahkan animasi menyeluruh (fungsional + retro playful) pada UI VibeLog memakai CSS murni tanpa dependency baru.

**Architecture:** Tiga lapis: (1) keyframes + utility animasi global di `main.css`, (2) transisi route di `App.vue`, (3) animasi lokal per komponen via `<style scoped>` dan composable count-up. Semua dibungkus guard `prefers-reduced-motion`.

**Tech Stack:** Vue 3, TypeScript, CSS (keyframes, transitions, custom properties), Web Animations via requestAnimationFrame untuk count-up.

**Spec:** `docs/superpowers/specs/2026-11-03-animasi-ui-design.md`

## Global Constraints

- Tanpa dependency npm baru.
- Bahasa UI: Bahasa Indonesia; tanpa em-dash di teks terlihat.
- Warna tetap memakai variabel: `--orange #ff5c1a`, `--line #d2d2c8`, `--ink #111`, `--muted #5c5c54`.
- Semua animasi wajib punya fallback di `@media (prefers-reduced-motion: reduce)`.
- Tidak mengubah logika data/API.

## Review Focus

- User dengan `prefers-reduced-motion: reduce`: tidak ada transform/blink, count-up langsung final, transisi tetap ada tapi minimal.
- Form dibuka/tutup berkali-kali: tidak ada animasi menumpuk atau elemen nyangkut transparan.
- Kartu muncul bertahap saat filter berubah: stagger tidak membuat kartu lama "berkedip" ulang tanpa perlu.
- Tombol disabled saat proses: tidak ikut efek press/hover yang menyesatkan.
- Prompt box dibuka/tutup: tinggi transisi tidak melompat atau memotong teks panjang.

---

### Task 1: Fondasi animasi global di `main.css`

**Files:**
- Modify: `frontend/src/assets/main.css`

**Interfaces:**
- Produces: keyframes `fade-in`, `fade-up`, `pop-in`, `blink`, `wiggle`; class `.anim-fade-up`; variabel `--ease-out`, `--dur-fast`, `--dur-mid`; rule tombol `.btn` dengan transisi + active; guard reduced-motion global.

- [ ] **Step 1: Tambah variabel easing/durasi dan keyframes ke `:root`**

Tambahkan di dalam blok `:root` (setelah `--muted`):



- [ ] **Step 2: Tambah keyframes setelah blok `:root`**



- [ ] **Step 3: Tambah utility class animasi**

Setelah keyframes:



- [ ] **Step 4: Tambah transisi + efek tekan pada `.btn`**

Ubah rule `.btn` yang ada dengan menambahkan properti transisi (jangan hapus yang lama):



Lalu ubah `.btn:hover:not(:disabled)` menjadi:



Tambah rule baru tepat setelah `.btn:hover:not(:disabled)`:



Untuk `.btn.primary:hover:not(:disabled)` tambahkan `box-shadow: 0 3px 0 #b83c0c;`.

- [ ] **Step 5: Tambah hover lift untuk chip & tag**

Ubah `.chip` dan `.tag` dengan menambah transisi, dan rule hover:



- [ ] **Step 6: Tambah transisi focus field**

Ubah `.field input, .field select, .field textarea` dengan menambah:



- [ ] **Step 7: Tambah guard reduced-motion global di akhir file**



- [ ] **Step 8: Typecheck**

Run: `npm run typecheck`
Expected: PASS tanpa error (CSS tidak dicek tsc, tapi memastikan tidak ada yang rusak).

- [ ] **Step 9: Commit**



---

### Task 2: Transisi antar halaman di `App.vue`

**Files:**
- Modify: `frontend/src/App.vue`

**Interfaces:**
- Consumes: keyframes `fade-up` dari Task 1.
- Produces: transisi route bernama `page`.

- [ ] **Step 1: Ganti template dengan RouterView bertransisi**



- [ ] **Step 2: Tambah style transisi (tambahkan blok `<style>` baru)**



- [ ] **Step 3: Typecheck**

Run: `npm run typecheck`
Expected: PASS.

- [ ] **Step 4: Commit**



---

### Task 3: Stagger kartu landing + blink headline

**Files:**
- Modify: `frontend/src/pages/LandingPage.vue`

**Interfaces:**
- Consumes: `fade-up` keyframes, `anim-fade-up` dari Task 1.
- Produces: kartu `.ref-card` masuk bertahap dengan `--i` index.

- [ ] **Step 1: Beri style stagger pada article kartu**

Ganti tag `<article>` pada `v-for="card in cards"`:



- [ ] **Step 2: Tambah animasi pada headline hero**

Beri class `hero-title` pada `<h1 class="pixel-headline">Vibe_Log</h1>`:



Tambah di scoped style:



- [ ] **Step 3: Tambah hover lift untuk kartu fitur**

Dalam scoped style tambahkan:



- [ ] **Step 4: Tambah hover lift pada pill stack**



- [ ] **Step 5: Typecheck**

Run: `npm run typecheck`
Expected: PASS.

- [ ] **Step 6: Commit**



---

### Task 4: Animasi hero-grid + plus motif landing

**Files:**
- Modify: `frontend/src/pages/LandingPage.vue`

**Interfaces:**
- Consumes: `fade-up` keyframes.
- Produces: hero-grid muncul bertahap.

- [ ] **Step 1: Animasi hero-grid**

Tambah di scoped style:



- [ ] **Step 2: Animasi plus motif berdenyut halus**



Karena `.plus-motif` sudah ada di scoped style, cukup tambahkan properti `animation`.

- [ ] **Step 3: Typecheck**

Run: `npm run typecheck`
Expected: PASS.

- [ ] **Step 4: Commit**



---

### Task 5: Composable count-up + StatsCards

**Files:**
- Create: `frontend/src/composables/useCountUp.ts`
- Modify: `frontend/src/components/StatsCards.vue`

**Interfaces:**
- Produces: `useCountUp(target: number, duration?: number): Ref<number>` dari `useCountUp.ts`; StatsCards menampilkan angka total/success/partial/failed dengan count-up saat pertama tampil.

- [ ] **Step 1: Tulis composable**



- [ ] **Step 2: Pakai di StatsCards**

Ganti `<script setup>`:



Ganti template angka `{{ summary.total }}` → `{{ total }}`, dan seterusnya untuk success/partial/failed.

- [ ] **Step 3: Tambah animasi strip masuk**

Tambah di scoped style:



- [ ] **Step 4: Typecheck**

Run: `npm run typecheck`
Expected: PASS.

- [ ] **Step 5: Commit**



---

### Task 6: Animasi form panel dashboard

**Files:**
- Modify: `frontend/src/pages/DashboardPage.vue`

**Interfaces:**
- Consumes: `pop-in` keyframes.
- Produces: panel form proyek & entri animasi masuk.

- [ ] **Step 1: Tambah class animasi ke form panel entri**

Ubah `div v-if="showEntryForm" class="panel form-panel"` menjadi:



- [ ] **Step 2: Tambah animasi ke ProjectForm**

Ubah `<ProjectForm v-if="showProjectForm"` dengan menambahkan wrapper animasi:



- [ ] **Step 3: Stagger kartu entri**

Ubah `EntryCard` v-for dengan tambahan class dan delay. Karena EntryCard adalah komponen, bungkus dengan wrapper index. Ubah:



- [ ] **Step 4: Animasi empty state dan panel**

Tambah di scoped style:



- [ ] **Step 5: Animasi headline dashboard**

Tambah class ke h1:



Tambah style:



- [ ] **Step 6: Typecheck**

Run: `npm run typecheck`
Expected: PASS.

- [ ] **Step 7: Commit**



---

### Task 7: Prompt box transition di EntryCard

**Files:**
- Modify: `frontend/src/components/EntryCard.vue`

**Interfaces:**
- Consumes: tidak ada (transition lokal).

- [ ] **Step 1: Ganti v-show dengan Transition**

Ubah:



menjadi:



Perhatikan: tutup `</div>` yang sesuai, hapus duplikasi isi.

- [ ] **Step 2: Tambah style transisi prompt**



- [ ] **Step 3: Tambah hover pada tombol arrow-cell dan prompt-toggle**

Arrow-cell sudah punya hover warna di main.css. Tambah transisi di scoped:



- [ ] **Step 4: Typecheck**

Run: `npm run typecheck`
Expected: PASS.

- [ ] **Step 5: Commit**



---

### Task 8: Animasi Logo wiggle

**Files:**
- Modify: `frontend/src/components/Logo.vue`

**Interfaces:**
- Consumes: `wiggle` keyframes.

- [ ] **Step 1: Tambah wiggle saat hover**

Tambah di scoped style:



- [ ] **Step 2: Typecheck**

Run: `npm run typecheck`
Expected: PASS.

- [ ] **Step 3: Commit**



---

### Task 9: Animasi auth shell dan NotFound

**Files:**
- Modify: `frontend/src/components/AuthShell.vue`
- Modify: `frontend/src/pages/NotFoundPage.vue`

**Interfaces:**
- Consumes: `fade-up`, `pop-in`.

- [ ] **Step 1: Animasi auth main**

Tambah di AuthShell scoped style:



- [ ] **Step 2: Animasi 404**

Tambah di NotFoundPage scoped style:



- [ ] **Step 3: Typecheck**

Run: `npm run typecheck`
Expected: PASS.

- [ ] **Step 4: Commit**



---

### Task 10: Verifikasi build penuh

**Files:**
- Tidak ada file baru.

- [ ] **Step 1: Typecheck**

Run: `npm run typecheck`
Expected: PASS tanpa error.

- [ ] **Step 2: Build produksi**

Run: `npm run build`
Expected: selesai tanpa error, output di `frontend/dist`.

- [ ] **Step 3: Cek visual manual**

Jalankan `npm run dev` di `frontend`, buka:
- `/` → kartu fitur muncul bertahap, headline masuk, hover kartu/pill lift, tombol press.
- `/login` → auth shell fade-up.
- `/app` (setelah login) → statistik count-up, panel slide/pop, kartu entri stagger, prompt box transisi halus.
- `/halaman-acak` → 404 pop-in.
- Simulasikan reduced-motion (DevTools > Rendering > Emulate `prefers-reduced-motion: reduce`) → animasi minimal, count-up langsung final.

- [ ] **Step 4: Commit akhir (jika ada penyesuaian)**


```