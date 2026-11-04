# Setup Verifikasi Email dengan OTP

VibeLog sekarang sudah dilengkapi dengan verifikasi email menggunakan OTP (One-Time Password) 6 digit.

## Fitur

- ✅ User harus verifikasi email setelah register
- ✅ OTP 6 digit dikirim ke email user
- ✅ OTP berlaku 15 menit
- ✅ Bisa resend OTP dengan cooldown 60 detik
- ✅ User tidak bisa login sebelum email terverifikasi
- ✅ Support berbagai SMTP provider (Gmail, Mailgun, SendGrid, dll)

## Flow User

1. **Register** → User input nama, email, password
2. **Email dikirim** → Backend generate OTP 6 digit & kirim ke email
3. **Halaman verifikasi** → User diarahkan ke `/verify-otp`
4. **Input OTP** → User masukkan kode dari email
5. **Verified** → Email terverifikasi, dapat token, redirect ke dashboard
6. **Login berikutnya** → Langsung bisa login (email sudah verified)

## Setup Development (Mode Console)

Tanpa konfigurasi SMTP, OTP akan di-log ke console:

```bash
cd backend
go run .
```

Saat user register, lihat console untuk kode OTP:

```
=== OTP EMAIL ===
To: User Name <user@example.com>
Code: 123456
=================
```

## Setup Production (Email SMTP)

### 1. Menggunakan Gmail

Buat App Password di Google Account:
1. Buka https://myaccount.google.com/security
2. Aktifkan 2-Step Verification
3. Buka "App passwords"
4. Generate password untuk "Mail"
5. Gunakan password tersebut di `.env`

```bash
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-16-char-app-password
SMTP_FROM_EMAIL=noreply@yourdomain.com
SMTP_FROM_NAME=VibeLog
```

### 2. Menggunakan Mailgun

```bash
SMTP_HOST=smtp.mailgun.org
SMTP_PORT=587
SMTP_USERNAME=postmaster@your-domain.mailgun.org
SMTP_PASSWORD=your-mailgun-smtp-password
SMTP_FROM_EMAIL=noreply@yourdomain.com
SMTP_FROM_NAME=VibeLog
```

### 3. Menggunakan SendGrid

```bash
SMTP_HOST=smtp.sendgrid.net
SMTP_PORT=587
SMTP_USERNAME=apikey
SMTP_PASSWORD=your-sendgrid-api-key
SMTP_FROM_EMAIL=noreply@yourdomain.com
SMTP_FROM_NAME=VibeLog
```

## Database Migration

Database akan otomatis ter-update saat backend pertama kali jalan:

- Tabel `users` ditambah kolom `email_verified` dan `email_verified_at`
- Tabel baru `otp_codes` untuk menyimpan kode OTP
- Index untuk query OTP yang cepat

User yang sudah ada sebelumnya akan punya `email_verified = 0` secara default. Kamu bisa update manual via SQL:

```sql
-- Verifikasi semua user existing
UPDATE users SET email_verified = 1, email_verified_at = datetime('now') WHERE email_verified = 0;
```

Atau paksa satu user tertentu untuk re-verify:

```sql
-- Paksa user untuk verify ulang
UPDATE users SET email_verified = 0, email_verified_at = NULL WHERE email = 'user@example.com';
```

## API Endpoints

### POST `/api/auth/register`

**Request:**
```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "password123"
}
```

**Response (Success):**
```json
{
  "user": {
    "id": 1,
    "name": "John Doe",
    "email": "john@example.com",
    "email_verified": false,
    "created_at": "2024-01-01T00:00:00Z"
  },
  "message": "Akun berhasil dibuat. Silakan cek email untuk kode verifikasi."
}
```

### POST `/api/auth/verify-otp`

**Request:**
```json
{
  "email": "john@example.com",
  "code": "123456"
}
```

**Response (Success):**
```json
{
  "token": "eyJhbGc...",
  "user": {
    "id": 1,
    "name": "John Doe",
    "email": "john@example.com",
    "email_verified": true,
    "created_at": "2024-01-01T00:00:00Z"
  },
  "message": "Email berhasil diverifikasi!"
}
```

### POST `/api/auth/resend-otp`

**Request:**
```json
{
  "email": "john@example.com"
}
```

**Response (Success):**
```json
{
  "message": "Kode verifikasi baru telah dikirim ke email kamu."
}
```

### POST `/api/auth/login`

Jika email belum verified, akan return error:

**Response (Email Not Verified):**
```json
{
  "error": {
    "code": "email_not_verified",
    "message": "email belum diverifikasi. Silakan cek email untuk kode verifikasi."
  }
}
```

## Troubleshooting

### OTP tidak diterima

1. **Cek spam folder** di email
2. **Cek console backend** - ada log error saat kirim email?
3. **Test SMTP config** dengan tool seperti `telnet smtp.gmail.com 587`
4. **Gunakan resend OTP** di halaman verifikasi

### Gmail: "Username and Password not accepted"

- Pastikan 2-Step Verification sudah aktif
- Gunakan App Password, bukan password akun utama
- Cek apakah "Less secure app access" perlu diaktifkan (jarang)

### OTP expired

- OTP berlaku 15 menit
- Gunakan tombol "Kirim ulang" untuk mendapat kode baru
- OTP lama akan otomatis tidak valid setelah digunakan

### Database locked error

SQLite menggunakan WAL mode dan busy_timeout 5000ms. Seharusnya tidak ada masalah, tapi kalau terjadi:

```bash
# Backup database
cp data/vibelog.db data/vibelog.db.backup

# Restart backend
```

## Customization

### Ubah waktu expired OTP

Edit `backend/handlers.go`:

```go
// Default: 15 menit
if err := CreateOTP(s.db, user.ID, code, "email_verification", 15); err != nil {
```

Ganti `15` dengan jumlah menit yang diinginkan.

### Ubah cooldown resend

Edit `frontend/src/pages/VerifyOTPPage.vue`:

```typescript
// Default: 60 detik
resendCooldown.value = 60
```

### Customize email template

Edit `backend/otp.go` di fungsi `SendOTPEmail`:

```go
body := fmt.Sprintf(`Halo %s,

Terima kasih telah mendaftar di VibeLog!

Kode verifikasi kamu adalah: %s

...`, toName, code)
```

## Security Notes

- OTP hanya valid sekali (marked as `used` setelah dipakai)
- OTP expired setelah 15 menit
- Rate limiting tetap berlaku (20 req/menit untuk auth endpoints)
- OTP di-cleanup otomatis setelah 24 jam (implementasi optional)
- Password di-hash dengan bcrypt sebelum disimpan

## Future Improvements

Beberapa enhancement yang bisa ditambahkan:

- [ ] Background job untuk cleanup OTP expired (cron)
- [ ] Email template HTML yang lebih menarik
- [ ] Support 2FA optional untuk user yang mau extra security
- [ ] Brute force protection untuk OTP attempts
- [ ] Log audit untuk semua aktivitas verifikasi

---

Happy coding! 🚀
