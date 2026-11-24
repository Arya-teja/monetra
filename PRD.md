# PRD — Monetra

> Product Requirements Document · v1.0 · 24 November 2026
> Owner: Arya

---

## 1. Ringkasan

**Monetra** adalah aplikasi pencatat keuangan pribadi dan keluarga. Setiap anggota keluarga punya akun login sendiri. Aplikasi dipakai terutama di HP (iPhone) sebagai **PWA**, sehingga bisa dipasang ke Home Screen tanpa App Store dan tanpa biaya.

### Tujuan

1. Mencatat pemasukan dan pengeluaran harian dengan cepat (≤ 10 detik per transaksi).
2. Mengetahui saldo semua sumber dana (bank, e-wallet, cash) di satu tempat.
3. Mengontrol pengeluaran lewat budget per kategori, dengan peringatan sebelum over.
4. Proyek portofolio: backend **Go** dengan clean architecture, frontend **Next.js**.

### Di luar cakupan (versi ini)

- Multi-currency (hanya **Rupiah**)
- Export laporan (PDF/Excel)
- Notifikasi email atau push (hanya **in-app**)
- Sinkronisasi otomatis dengan rekening bank
- Aplikasi native iOS/Android

---

## 2. Pengguna

| Persona          | Kebutuhan                                                            |
| ---------------- | -------------------------------------------------------------------- |
| Arya (pemilik)   | Mencatat transaksi, memantau budget, melihat laporan bulanan         |
| Anggota keluarga | Punya akun sendiri; nantinya tergabung dalam satu Household (Fase 2) |

---

## 3. Lingkup Rilis

### MVP (Fase 1)

| #   | Fitur                                                                  | Prioritas |
| --- | ---------------------------------------------------------------------- | --------- |
| F1  | Autentikasi (register, login, logout, proteksi halaman)                | Wajib     |
| F2  | Akun sumber dana (bank / e-wallet / cash)                              | Wajib     |
| F3  | Kategori income & expense (default + custom, inline di form transaksi) | Wajib     |
| F4  | Transaksi income/expense, saldo akun otomatis menyesuaikan             | Wajib     |
| F5  | Budget bulanan per kategori + peringatan in-app                        | Wajib     |
| F6  | Dashboard: saldo total, % income terpakai, chart bulanan               | Wajib     |
| F7  | PWA (installable di iPhone)                                            | Wajib     |

### Fase 2

- Household / mode keluarga (menghubungkan beberapa User)
- Tag bebas pada transaksi (lintas kategori, mis. `liburan`, `darurat`)
- Sisihan bulanan otomatis (mis. Rp200.000 untuk ongkos tiap bulan)
- Analisis mingguan dan breakdown pengeluaran per akun
- Transfer antar akun

---

## 4. Kebutuhan Fungsional

### F1 — Autentikasi

- Register: nama, email (unik), password (min. 8 karakter).
- Login dengan email + password; logout.
- Saat register berhasil, sistem membuat **kategori default** otomatis (lihat Lampiran A).
- Semua halaman di luar `/login` dan `/register` wajib login.
- Setiap data (akun, kategori, transaksi, budget) **hanya bisa diakses pemiliknya**.

### F2 — Akun Sumber Dana

- Field: nama akun, tipe (`bank` | `ewallet` | `cash`), nama institusi (opsional, mis. "BSI", "GoPay"), nomor akun (opsional), saldo awal.
- CRUD akun. Saldo awal hanya diisi saat membuat akun; setelah itu saldo berubah **hanya** lewat transaksi.
- Akun yang sudah punya transaksi **tidak bisa dihapus**; pengguna bisa mengarsipkannya (disembunyikan dari pilihan form, tetap muncul di riwayat).

### F3 — Kategori

- Tipe kategori: `income` atau `expense`. Kategori income dan expense terpisah.
- Tidak ada halaman kategori terpisah. Kategori dipilih **atau dibuat langsung dari form transaksi** (combobox dengan opsi "+ Tambah kategori").
- Daftar kategori di form terfilter sesuai tipe transaksi yang dipilih.
- Kategori bisa di-rename dan dihapus dari halaman Pengaturan. Kategori yang sudah dipakai transaksi tidak bisa dihapus.
- Nama kategori unik per user per tipe.

### F4 — Transaksi

- Alur form: **tipe** (pemasukan/pengeluaran) → **akun** → **kategori** → **nominal** → tanggal (default hari ini) → catatan (opsional).
- Nominal selalu positif; arah saldo ditentukan oleh tipe.
- Saat transaksi dibuat, diubah, atau dihapus, **saldo akun disesuaikan dalam satu database transaction** (atomik).
- Saldo boleh negatif (mencerminkan kondisi nyata), tetapi UI menampilkan penanda.
- Daftar transaksi: dikelompokkan per tanggal, terbaru di atas, filter bulan / akun / kategori / tipe, pagination.

### F5 — Budget

- Budget ditetapkan per **kategori expense** per **bulan kalender** (zona waktu `Asia/Jakarta`).
- Satu kategori maksimal satu budget per bulan.
- Opsi "salin budget bulan lalu" untuk mempercepat setup.
- Status dihitung dari total pengeluaran kategori di bulan itu:
  - `< 80%` → aman
  - `80–99%` → **hampir habis** (kuning)
  - `≥ 100%` → **over budget** (merah)
- Peringatan in-app: banner di dashboard + indikator di halaman budget. Saat menyimpan transaksi expense yang membuat budget ≥ 80%, tampilkan toast peringatan.
- Peringatan **dihitung saat dibaca** (tidak ada tabel notifikasi di MVP).

### F6 — Dashboard

- Total saldo semua akun aktif.
- Pemasukan, pengeluaran, dan **% income terpakai** bulan berjalan (`expense / income × 100`; jika income 0, tampilkan "—").
- Chart batang pemasukan vs pengeluaran per bulan (6 bulan terakhir).
- Top 5 kategori pengeluaran bulan berjalan.
- Daftar budget yang hampir habis / over.
- 5 transaksi terakhir.

### F7 — PWA

- Web App Manifest (nama, ikon 192/512 + apple-touch-icon, `display: standalone`, theme color).
- Bisa di-install lewat Safari: Share → Add to Home Screen.
- Offline: app shell tampil, tetapi input transaksi tetap butuh koneksi (offline write di luar MVP).
- Safe-area iPhone (notch & home indicator) diperhatikan.

---

## 5. Kebutuhan Non-Fungsional

| Aspek           | Target                                                                                                          |
| --------------- | --------------------------------------------------------------------------------------------------------------- |
| Keamanan        | Password di-hash (bcrypt); token JWT; validasi input di backend; query parametrized; CORS hanya origin frontend |
| Integritas data | Perubahan saldo atomik; uang disimpan sebagai **integer Rupiah (BIGINT)**, bukan float                          |
| Performa        | Response API p95 < 300 ms (di luar cold start hosting gratis)                                                   |
| Aksesibilitas   | Kontras teks WCAG AA, target sentuh ≥ 44 px, label pada semua input                                             |
| Responsif       | Mobile-first (375 px), tetap rapi di desktop                                                                    |
| Kualitas kode   | SOLID, clean architecture, type-safe, tanpa dead code                                                           |

> **Catatan keputusan uang:** Rupiah tidak memakai sen, jadi `BIGINT` lebih sederhana dan aman daripada `Decimal`. Di Go tidak perlu library decimal, dan di frontend tidak ada masalah presisi.

---

## 6. Arsitektur & Tech Stack

```
[Next.js (PWA) — Vercel]  ──HTTPS/JSON──▶  [Go REST API — Render]  ──▶  [PostgreSQL — managed]
```

### Frontend

- Next.js 16 (App Router), TypeScript, Tailwind CSS, shadcn/ui, Zod (stack yang sudah terpasang).
- Prisma **dihapus dari frontend**; semua akses data lewat Go API.
- Auth: NextAuth v5 Credentials tetap dipakai sebagai session layer. `authorize()` memanggil `POST /api/v1/auth/login` di Go dan menyimpan access token di session. Dengan cara ini proteksi route yang sudah ada tidak perlu dibongkar.
- Pemanggilan API dari server (Server Components / Server Actions) melalui satu API client yang menyisipkan token.

### Backend (Go)

- Go ≥ 1.22, router `net/http` standar (sudah mendukung method + path params).
- Struktur clean architecture:

```
backend/
├── cmd/api/main.go            # entry point, wiring dependency
├── internal/
│   ├── domain/                # entity + interface repository (tanpa dependency luar)
│   ├── service/               # business logic (saldo, budget status, dll.)
│   ├── repository/postgres/   # implementasi repository
│   ├── http/
│   │   ├── handler/           # HTTP handler per fitur
│   │   ├── middleware/        # auth JWT, logging, CORS, recover
│   │   └── response/          # format response & error standar
│   └── config/                # load env
├── migrations/                # SQL migration
└── go.mod
```

- **Library (usulan, perlu persetujuan Arya sebelum di-install):**
  - `jackc/pgx` — driver PostgreSQL
  - `sqlc` — generate kode type-safe dari SQL
  - `golang-migrate` — migrasi database
  - `golang-jwt/jwt` — token
  - `golang.org/x/crypto/bcrypt` — hash password

### Hosting (gratis)

| Komponen   | Layanan                              | Catatan                                                     |
| ---------- | ------------------------------------ | ----------------------------------------------------------- |
| Frontend   | Vercel                               | Free tier                                                   |
| Backend Go | Render                               | Free web service, tidur setelah ±15 menit idle (cold start) |
| Database   | PostgreSQL managed (Neon / Supabase) | Cek limit free tier saat setup                              |

### Struktur Repo (usulan)

```
monetra/
├── frontend/ # Next.js (project yang sudah ada dipindah ke sini)
├── backend/  # Go
├── PRD.md
└── CLAUDE.md
```

---

## 7. Model Data (konseptual)

> Diselaraskan dengan `schema.prisma` yang sudah ada saat migrasi ke SQL. Semua tabel memiliki `id` (UUID), `created_at`, `updated_at`.

| Tabel          | Kolom utama                                                                                                    |
| -------------- | -------------------------------------------------------------------------------------------------------------- |
| `users`        | name, email (unique), password_hash                                                                            |
| `accounts`     | user_id, name, type (`bank`/`ewallet`/`cash`), institution_name?, account_number?, balance BIGINT, is_archived |
| `categories`   | user_id, name, type (`income`/`expense`), is_default · unique(user_id, type, name)                             |
| `transactions` | user_id, account_id, category_id, type, amount BIGINT (> 0), occurred_at DATE, note?                           |
| `budgets`      | user_id, category_id, month DATE (hari pertama bulan), amount BIGINT · unique(user_id, category_id, month)     |

Index: `transactions(user_id, occurred_at DESC)`, `transactions(account_id)`, `transactions(category_id)`.

---

## 8. Kontrak API (v1)

Base URL: `/api/v1` · Auth: `Authorization: Bearer <token>` (kecuali register/login)

**Format response**

```json
{ "data": {}, "meta": { "page": 1, "pageSize": 20, "total": 120 } }
{ "error": { "code": "VALIDATION_ERROR", "message": "…", "details": { "amount": "must be greater than 0" } } }
```

| Method | Endpoint             | Keterangan                                                                                                   |
| ------ | -------------------- | ------------------------------------------------------------------------------------------------------------ |
| POST   | `/auth/register`     | Buat user + kategori default                                                                                 |
| POST   | `/auth/login`        | Return access token                                                                                          |
| GET    | `/auth/me`           | Profil user login                                                                                            |
| GET    | `/accounts`          | `?includeArchived=true`                                                                                      |
| POST   | `/accounts`          | Buat akun (dengan saldo awal)                                                                                |
| PATCH  | `/accounts/{id}`     | Ubah nama/tipe/institusi/arsip                                                                               |
| DELETE | `/accounts/{id}`     | 409 jika sudah ada transaksi                                                                                 |
| GET    | `/categories`        | `?type=income\|expense`                                                                                      |
| POST   | `/categories`        | Dipakai dari form transaksi                                                                                  |
| PATCH  | `/categories/{id}`   | Rename                                                                                                       |
| DELETE | `/categories/{id}`   | 409 jika sudah dipakai                                                                                       |
| GET    | `/transactions`      | `?month=2026-11&accountId&categoryId&type&page&pageSize`                                                     |
| POST   | `/transactions`      | Saldo akun ikut berubah (atomik)                                                                             |
| GET    | `/transactions/{id}` | Detail                                                                                                       |
| PATCH  | `/transactions/{id}` | Saldo dikoreksi (atomik)                                                                                     |
| DELETE | `/transactions/{id}` | Saldo dikembalikan (atomik)                                                                                  |
| GET    | `/budgets`           | `?month=2026-11` → termasuk `spent`, `percentage`, `status`                                                  |
| POST   | `/budgets`           | Buat budget                                                                                                  |
| POST   | `/budgets/copy`      | Salin dari bulan sebelumnya                                                                                  |
| PATCH  | `/budgets/{id}`      | Ubah nominal                                                                                                 |
| DELETE | `/budgets/{id}`      | Hapus                                                                                                        |
| GET    | `/dashboard/summary` | `?month=2026-11` → saldo total, income, expense, % terpakai, top kategori, budget alerts, transaksi terakhir |
| GET    | `/dashboard/trend`   | `?months=6` → income vs expense per bulan                                                                    |

Kode error standar: `VALIDATION_ERROR` (400), `UNAUTHORIZED` (401), `FORBIDDEN` (403), `NOT_FOUND` (404), `CONFLICT` (409), `INTERNAL_ERROR` (500).

---

## 9. Desain UI

### Arah visual

Hangat, kalem, dan "earthy": terasa seperti buku catatan keuangan yang rapi, bukan dashboard korporat. Banyak ruang kosong, sudut membulat lembut, dan angka sebagai elemen visual utama.

### Palet warna

| Token              | Hex       | Peran                                                    |
| ------------------ | --------- | -------------------------------------------------------- |
| `primary`          | `#8B9A6E` | Tombol utama, nav aktif, chart income                    |
| `background`       | `#F7F2EB` | Background aplikasi                                      |
| `surface`          | `#EAE2D6` | Card, input, border                                      |
| `muted`            | `#EEEEEE` | Skeleton, divider, elemen netral                         |
| `foreground`       | `#2F3526` | Teks utama (tambahan, kontras AA di atas cream)          |
| `foreground-muted` | `#6B6F60` | Teks sekunder (tambahan)                                 |
| `income`           | `#5F7A4E` | Nominal pemasukan (tambahan)                             |
| `expense`          | `#B5654A` | Nominal pengeluaran / over budget (tambahan, terracotta) |
| `warning`          | `#C9A24A` | Budget hampir habis (tambahan)                           |

> Pastikan teks putih di atas `#8B9A6E` memenuhi kontras; jika tidak, pakai `foreground` di atas tombol primary atau gelapkan primary untuk state hover/teks.

### Tipografi

- **Plus Jakarta Sans** via `next/font` (tanpa dependency tambahan).
- Nominal memakai `tabular-nums` agar angka rata.
- Format Rupiah: `Rp1.250.000` (`Intl.NumberFormat('id-ID')`).

### Pola layout (mobile-first)

- **Bottom navigation** (mobile): Beranda · Transaksi · **(+) Tambah** · Budget · Akun. Di desktop berubah menjadi sidebar.
- Tombol **(+)** di tengah membuka form transaksi sebagai bottom sheet (Drawer) di mobile dan Dialog di desktop.
- Form transaksi: segmented control Pemasukan/Pengeluaran di atas, input nominal besar, lalu pilihan akun dan kategori.
- Kartu budget: progress bar dengan warna sesuai status.

### Halaman (route group `(dashboard)`)

| Route                 | Isi                             |
| --------------------- | ------------------------------- |
| `/login`, `/register` | Form auth                       |
| `/dashboard`          | Ringkasan (F6)                  |
| `/transactions`       | Daftar + filter (F4)            |
| `/budgets`            | Budget bulan terpilih (F5)      |
| `/accounts`           | Daftar akun & saldo (F2)        |
| `/settings`           | Profil, kelola kategori, logout |

---

## 10. Rencana Pengerjaan

Pendekatan: **fondasi backend dulu, lalu vertical slice per fitur** (endpoint Go → UI Next.js untuk fitur yang sama).

| Tahap                   | Pekerjaan                                                                                                          |
| ----------------------- | ------------------------------------------------------------------------------------------------------------------ |
| 0. Setup                | Restruktur repo (`frontend/`, `backend/`), setup Go project, PostgreSQL, migrasi awal, middleware, format response |
| 1. Auth                 | Register/login/me di Go → hubungkan NextAuth ke Go API, hapus Prisma dari web                                      |
| 2. Akun                 | Endpoint accounts → halaman `/accounts`                                                                            |
| 3. Kategori + Transaksi | Endpoint categories & transactions (saldo atomik) → form transaksi + `/transactions`                               |
| 4. Budget               | Endpoint budgets + status → `/budgets` + toast peringatan                                                          |
| 5. Dashboard            | Endpoint summary & trend → `/dashboard` + chart                                                                    |
| 6. PWA & Deploy         | Manifest, ikon, safe-area → deploy Render + Vercel + DB managed                                                    |

### Definition of Done per fitur

- Endpoint berjalan dengan validasi dan error format standar.
- Unit test untuk business logic di `service/` (saldo, status budget).
- UI responsif, accessible, dan terhubung ke API sungguhan.
- Tidak ada dead code, unused import, atau TODO.

---

## 11. Metrik Keberhasilan

- Seluruh fitur MVP dipakai sehari-hari oleh Arya selama 1 bulan penuh.
- Input satu transaksi ≤ 10 detik di HP.
- Tidak ada selisih saldo antara akun dan total transaksinya.

---

## Lampiran A — Kategori Default

| Income    | Expense            |
| --------- | ------------------ |
| Gaji      | Makanan & Minuman  |
| Bonus     | Transportasi       |
| Freelance | Belanja            |
| Hadiah    | Tagihan & Utilitas |
| Lainnya   | Kesehatan          |
|           | Hiburan            |
|           | Pendidikan         |
|           | Liburan            |
|           | Lainnya            |
