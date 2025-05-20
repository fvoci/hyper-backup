# 🛡️ hyper-backup

**hyper-backup**은 MySQL, PostgreSQL, MongoDB, Traefik 로그, 지정된 폴더 등을 자동으로 백업하고, `rclone` 또는 `rsync`를 통해 S3 호환 스토리지로 업로드할 수 있는 백업 도구입니다. Go로 작성되었으며, 도커 컨테이너 환경에 최적화되어 있습니다.

---

## 🚀 주요 기능

- ✅ MySQL, PostgreSQL, MongoDB 백업 (gzip 압축)
- ✅ Traefik JSON 로그 회전 및 USR1 시그널 전송
- ✅ 사용자 정의 폴더 백업 (`.tar.zst` 또는 `.tar.gz`)
- ✅ Rclone 또는 Rsync를 통한 외부 스토리지 업로드
- ✅ 크론 표현식 또는 간격 기반 스케줄링 지원
- ✅ 권한 감지 및 `gosu`로 사용자 전환 실행

---

## 📦 환경 변수

### 🔧 데이터베이스

| 환경변수 | 설명 |
|----------|------|
| `MYSQL_HOST`, `MYSQL_USER`, `MYSQL_PASSWORD`, `MYSQL_DATABASE` | MySQL 설정 |
| `POSTGRES_HOST`, `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB` | PostgreSQL 설정 |
| `MONGO_URI` 또는 `MONGO_HOST`, `MONGO_DB` | MongoDB 설정 |

### 📂 폴더 백업

| 환경변수 | 설명 |
|----------|------|
| `PACK_UP_HYPER_BACKUP_1`, `PACK_UP_HYPER_BACKUP_2`, ... | 백업할 폴더 경로 |
| `FILE_BACKUP_COMPRESSION` | `zstd` (기본값) 또는 `gzip` |

### 🌐 Traefik 로그 회전

| 환경변수 | 설명 |
|----------|------|
| `TRAEFIK_LOG_FILE` | Traefik 로그 파일 경로 |
| `TRAEFIK_BACKUP_DIR` _(선택)_ | 추가 백업 디렉토리 |

### ☁️ 외부 저장소 (Rclone / Rsync)

| 환경변수 | 설명 |
|----------|------|
| `RCLONE_REMOTE`, `RCLONE_PATH`, `S3_ENDPOINT`, `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` | Rclone 설정 |
| `RSYNC_SRC`, `RSYNC_DEST` | Rsync 설정 |
| `RCLONE_RETENTION_DAYS` | 삭제 보존 기간 (기본값 14일) |

---

## ⏰ 스케줄링

| 환경변수 | 설명 |
|----------|------|
| `BACKUP_SCHEDULE` | 크론 표현식 (예: `0 0 * * *`) |
| `BACKUP_INTERVAL_HOURS` | 시간 간격 (예: `6`) |

> `BACKUP_SCHEDULE` 가 우선이며, 없을 경우 `BACKUP_INTERVAL_HOURS`, 둘 다 없으면 매일 자정 실행됩니다.

---

## 🐳 Docker 사용법

```bash
docker run --rm \
  -e MYSQL_HOST=db \
  -e MYSQL_USER=root \
  -e MYSQL_PASSWORD=secret \
  -e MYSQL_DATABASE=testdb \
  -e RCLONE_REMOTE=minio \
  -e RCLONE_PATH=backup \
  -e S3_ENDPOINT=http://minio:9000 \
  -e AWS_ACCESS_KEY_ID=minioadmin \
  -e AWS_SECRET_ACCESS_KEY=minioadmin \
  -v /var/run/docker.sock:/var/run/docker.sock \
  fvoci/hyper-backup
```

---

## 🛡️ hyper-backup (English)

**hyper-backup** is a container-friendly backup tool written in Go.
It automatically backs up MySQL, PostgreSQL, MongoDB, Traefik logs, and user-specified folders, and uploads them to external S3-compatible storage using `rclone` or `rsync`.

---

## 🚀 Features

* ✅ MySQL, PostgreSQL, MongoDB backups (with gzip compression)
* ✅ Traefik log rotation and USR1 signal to container
* ✅ User-defined folder backup (`.tar.zst` or `.tar.gz`)
* ✅ Upload to external storage via Rclone or Rsync
* ✅ Supports cron expressions or interval-based scheduling
* ✅ Automatic user privilege switching via `gosu`

---

## 📦 Environment Variables

### 🔧 Database Configuration

| Variable                                                             | Description              |
| -------------------------------------------------------------------- | ------------------------ |
| `MYSQL_HOST`, `MYSQL_USER`, `MYSQL_PASSWORD`, `MYSQL_DATABASE`       | MySQL configuration      |
| `POSTGRES_HOST`, `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB` | PostgreSQL configuration |
| `MONGO_URI` or `MONGO_HOST`, `MONGO_DB`                              | MongoDB configuration    |

### 📂 Folder Backup

| Variable                                                | Description                                    |
| ------------------------------------------------------- | ---------------------------------------------- |
| `PACK_UP_HYPER_BACKUP_1`, `PACK_UP_HYPER_BACKUP_2`, ... | Absolute paths of folders to back up           |
| `FILE_BACKUP_COMPRESSION`                               | Compression method: `zstd` (default) or `gzip` |

### 🌐 Traefik Log Rotation

| Variable                          | Description                                  |
| --------------------------------- | -------------------------------------------- |
| `TRAEFIK_LOG_FILE`                | Path to Traefik's JSON log file              |
| `TRAEFIK_BACKUP_DIR` *(optional)* | Additional backup directory for rotated logs |

### ☁️ External Storage (Rclone / Rsync)

| Variable                                                                                    | Description                             |
| ------------------------------------------------------------------------------------------- | --------------------------------------- |
| `RCLONE_REMOTE`, `RCLONE_PATH`, `S3_ENDPOINT`, `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` | Rclone config for S3-compatible targets |
| `RSYNC_SRC`, `RSYNC_DEST`                                                                   | Rsync config                            |
| `RCLONE_RETENTION_DAYS`                                                                     | Retention period in days (default: 14)  |

---

## ⏰ Scheduling Options

| Variable                | Description                        |
| ----------------------- | ---------------------------------- |
| `BACKUP_SCHEDULE`       | Cron expression (e.g. `0 0 * * *`) |
| `BACKUP_INTERVAL_HOURS` | Interval in hours (e.g. `6`)       |

> If `BACKUP_SCHEDULE` is set, it takes priority.
> If not, `BACKUP_INTERVAL_HOURS` is used.
> If neither is set, defaults to daily at midnight.

---

## 🐳 Docker Usage

```bash
docker run --rm \
  -e MYSQL_HOST=db \
  -e MYSQL_USER=root \
  -e MYSQL_PASSWORD=secret \
  -e MYSQL_DATABASE=testdb \
  -e RCLONE_REMOTE=minio \
  -e RCLONE_PATH=backup \
  -e S3_ENDPOINT=http://minio:9000 \
  -e AWS_ACCESS_KEY_ID=minioadmin \
  -e AWS_SECRET_ACCESS_KEY=minioadmin \
  -v /var/run/docker.sock:/var/run/docker.sock \
  fvoci/hyper-backup
```