# 🚀 Panduan Deployment ke Google Cloud (Docker Image)

Dokumen ini menjelaskan langkah **deployment ke Google Cloud Platform (GCP)**
menggunakan **Docker image**. Target utama: **Cloud Run** (paling mudah, langsung
mendapat URL publik `https://...run.app` sesuai syarat soal).

> Ganti `yourusername`, `PROJECT_ID`, dan `REGION` sesuai milik Anda.
> Contoh region: `asia-southeast2` (Jakarta).

---

## 0. Prasyarat

1. Akun GCP + sebuah **Project** (punya `PROJECT_ID`).
2. **Billing** aktif pada project tersebut.
3. **gcloud CLI** terpasang & login:
   ```bash
   gcloud auth login
   gcloud config set project PROJECT_ID
   gcloud config set run/region asia-southeast2
   ```
4. Aktifkan API yang dibutuhkan:
   ```bash
   gcloud services enable \
     run.googleapis.com \
     artifactregistry.googleapis.com \
     cloudbuild.googleapis.com \
     sqladmin.googleapis.com
   ```

---

## 1. Siapkan Artifact Registry (tempat menyimpan image)

```bash
gcloud artifacts repositories create book-repo \
  --repository-format=docker \
  --location=asia-southeast2 \
  --description="Book Management gRPC images"

# Konfigurasi Docker agar bisa push ke registry
gcloud auth configure-docker asia-southeast2-docker.pkg.dev
```

---

## 2. Build & Push Docker Image

Tetapkan nama image sebagai variabel agar konsisten:

```bash
export PROJECT_ID=$(gcloud config get-value project)
export IMAGE=asia-southeast2-docker.pkg.dev/$PROJECT_ID/book-repo/book-management:latest
```

### Opsi A — build lokal lalu push (butuh Docker di mesin Anda)
```bash
docker build -t $IMAGE .
docker push $IMAGE
```

### Opsi B — build di cloud (tanpa Docker lokal, pakai Cloud Build)
```bash
gcloud builds submit --tag $IMAGE .
```

> ⚠️ **Penting — jalankan codegen & tidy di lokal dulu.**
> Sebelum build image, hasilkan kode gRPC dari `.proto` dan buat `go.sum`:
> ```bash
> make generate     # = buf generate proto  → menghasilkan pkg/genproto/...
> go mod tidy       # menghasilkan go.sum
> ```
> `Dockerfile` sengaja **tidak** menjalankan `buf` di dalamnya (agar image ramping);
> ia menyalin `pkg/genproto` dan `go.sum` yang sudah ada. Lakukan dua perintah di
> atas sekali di mesin Anda (yang punya internet penuh), commit hasilnya, baru build.

---

## 3. Deploy ke Cloud Run

Aplikasi mengekspos **HTTP gateway (REST + Swagger)** pada satu port. Cloud Run
menyuntikkan variabel `PORT`, jadi kita arahkan `HTTP_PORT` ke `PORT`.

### 3a. Mode paling cepat — in-memory (tanpa database)
Cocok untuk demo/penilaian. Data hilang saat instance restart.

```bash
gcloud run deploy book-management \
  --image=$IMAGE \
  --platform=managed \
  --region=asia-southeast2 \
  --allow-unauthenticated \
  --port=8080 \
  --set-env-vars=DB_DRIVER=memory,JWT_SECRET=ganti-dengan-secret-kuat,HTTP_PORT=8080,GRPC_PORT=50051,SCHED_INTERVAL=1h
```

Setelah selesai, gcloud menampilkan **Service URL**, misalnya:
```
https://book-management-xxxxxxxx-et.a.run.app
```
Cek:
```bash
curl https://book-management-xxxxxxxx-et.a.run.app/healthz
# Swagger UI:
# https://book-management-xxxxxxxx-et.a.run.app/swagger/
```

### 3b. Mode produksi — dengan Cloud SQL (PostgreSQL)

1. Buat instance Cloud SQL + database:
   ```bash
   gcloud sql instances create book-db \
     --database-version=POSTGRES_15 \
     --tier=db-f1-micro \
     --region=asia-southeast2

   gcloud sql databases create bookdb --instance=book-db
   gcloud sql users set-password postgres --instance=book-db --password=RAHASIA_DB
   ```

2. Ambil **connection name**:
   ```bash
   export CONN=$(gcloud sql instances describe book-db --format='value(connectionName)')
   echo $CONN   # contoh: PROJECT_ID:asia-southeast2:book-db
   ```

3. Deploy dengan koneksi Cloud SQL (via unix socket):
   ```bash
   gcloud run deploy book-management \
     --image=$IMAGE \
     --region=asia-southeast2 \
     --allow-unauthenticated \
     --port=8080 \
     --add-cloudsql-instances=$CONN \
     --set-env-vars=DB_DRIVER=postgres,JWT_SECRET=ganti-dengan-secret-kuat,HTTP_PORT=8080 \
     --set-env-vars=DATABASE_URL="host=/cloudsql/$CONN user=postgres password=RAHASIA_DB dbname=bookdb sslmode=disable"
   ```

> Aplikasi otomatis menjalankan `AutoMigrate` (membuat tabel `users`, `books`,
> `borrowed_books`) saat start dengan `DB_DRIVER=postgres`.

---

## 4. Variabel Lingkungan (Environment Variables)

| Variabel          | Default   | Keterangan |
|-------------------|-----------|------------|
| `HTTP_PORT`       | `8080`    | Port REST gateway + Swagger (Cloud Run pakai ini). |
| `GRPC_PORT`       | `50051`   | Port server gRPC internal. |
| `JWT_SECRET`      | *(wajib)* | Kunci penandatangan JWT. **Wajib diisi kuat di produksi.** |
| `JWT_TTL`         | `24h`     | Masa berlaku token. |
| `DB_DRIVER`       | `memory`  | `memory` atau `postgres`. |
| `DATABASE_URL`    | -         | DSN PostgreSQL (jika driver `postgres`). |
| `SCHED_INTERVAL`  | `1h`      | Interval job scheduler cek buku kadaluarsa. |
| `BORROW_DURATION` | `168h`    | Durasi pinjam default (7 hari). |

Untuk menyimpan `JWT_SECRET`/password lebih aman, gunakan **Secret Manager**:
```bash
echo -n "secret-kuat" | gcloud secrets create jwt-secret --data-file=-
gcloud run deploy book-management --image=$IMAGE \
  --update-secrets=JWT_SECRET=jwt-secret:latest
```

---

## 5. Verifikasi Setelah Deploy

```bash
export URL=$(gcloud run services describe book-management \
  --region=asia-southeast2 --format='value(status.url)')

# 1) Health check
curl $URL/healthz

# 2) Register user
curl -X POST $URL/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"budi","password":"rahasia123"}'

# 3) Login → dapat token
TOKEN=$(curl -s -X POST $URL/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"budi","password":"rahasia123"}' | jq -r .token)

# 4) Buat buku (pakai token)
curl -X POST $URL/v1/books \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Clean Architecture","author":"Robert C. Martin","published_date":"2017-09-10T00:00:00Z"}'

# 5) Swagger UI di browser
echo "$URL/swagger/"
```

---

## 6. Update / Rollback

```bash
# Deploy versi baru (build ulang image lalu deploy ulang)
gcloud builds submit --tag $IMAGE .
gcloud run deploy book-management --image=$IMAGE --region=asia-southeast2

# Lihat daftar revisi
gcloud run revisions list --service=book-management --region=asia-southeast2

# Rollback ke revisi lama
gcloud run services update-traffic book-management \
  --to-revisions=REVISION_NAME=100 --region=asia-southeast2
```

---

## 7. Membersihkan Resource (opsional, hindari biaya)

```bash
gcloud run services delete book-management --region=asia-southeast2
gcloud sql instances delete book-db
gcloud artifacts repositories delete book-repo --location=asia-southeast2
```

---

### Ringkasan Alur Deployment
```
kode → docker build → push ke Artifact Registry → gcloud run deploy → URL publik
                                                   └─(opsional) Cloud SQL Postgres
```
Selesai — endpoint gRPC/REST Anda kini dapat diakses publik melalui URL Cloud Run,
lengkap dengan Swagger di `/swagger/`.
