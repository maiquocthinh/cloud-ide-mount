# Phase 0: Thiết Lập

## Mục Tiêu
Khởi tạo project structure, docs và CI để làm nền tảng cho các phase sau.

## Danh Sách Task

### #0.1: Khởi tạo Go module + cấu trúc project
- [x] Task 0.1.1: `go mod init cloud-ide-mount`
- [x] Task 0.1.2: Tạo cấu trúc thư mục `cmd/`, `internal/`, `.github/`
- [x] Task 0.1.3: Cài dependencies: cobra, ...
- [x] Task 0.1.4: File `main.go` — entry point gọi `cmd.Execute()`

### #0.2: Tổ chức docs (7 file md)
- [x] Task 0.2.1: Tạo `README.md` — tổng quan dự án
- [x] Task 0.2.2: Tạo `docs/CHECKLIST.md` — bảng theo dõi tiến độ
- [x] Task 0.2.3: Tạo `docs/tasks/` — task chi tiết cho từng phase
- [x] Task 0.2.4: Tạo `docs/logs/` — nhật ký làm việc

### #0.3: CI config (GitHub Actions)
- [x] Task 0.3.1: Tạo `.github/workflows/ci.yml` — build + test + vet
- [x] Task 0.3.2: Verify CI chạy pass trên push/PR

### #0.4: Thiết lập môi trường dev (Win/Lin/Mac)
- [x] Task 0.4.1: Script setup cho Windows (PowerShell) — `scripts/setup.ps1`
  - Check/install: Go, Git, gh, rclone qua winget
  - Generate SSH key, hướng dẫn gh auth
  - Build + test tự động
- [x] Task 0.4.2: Script setup cho Linux/macOS (bash) — `scripts/setup.sh`
  - Check/install: Go, Git, gh, rclone qua brew/apt/dnf/pacman
  - Generate SSH key, hướng dẫn gh auth
  - Build + test tự động
- [x] Task 0.4.3: Tài liệu hướng dẫn cài đặt dependencies (gh, rclone)
