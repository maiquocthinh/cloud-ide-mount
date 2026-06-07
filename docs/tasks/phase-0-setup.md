# Phase 0: Thiết Lập

## Mục Tiêu
Thiết lập cấu trúc project, docs, CI/CD và môi trường phát triển để team có thể bắt đầu code ngay.

## Danh Sách Task

### 0.1: Khởi tạo Go module + cấu trúc project
- [x] Task 0.1.1: `go mod init cloud-ide-mount`
- [x] Task 0.1.2: Tạo cấu trúc thư mục `cmd/`, `internal/`
- [x] Task 0.1.3: Thêm dependency `cobra`

### 0.2: Tổ chức docs (7 file md)
- [x] Task 0.2.1: Tạo `docs/ARCHITECTURE.md` — kiến trúc 5 layer
- [x] Task 0.2.2: Tạo `docs/IMPLEMENTATION.md` — hướng dẫn code
- [x] Task 0.2.3: Tạo `docs/CHECKLIST.md` — theo dõi 16 issue
- [x] Task 0.2.4: Tạo `docs/QUICK_REFERENCE.md` — tra cứu hàng ngày
- [x] Task 0.2.5: Tạo `docs/CLI_REFERENCE.md` — lệnh, flags, ví dụ
- [x] Task 0.2.6: Tạo `docs/README.md` — tổng quan
- [x] Task 0.2.7: Tạo `docs/TASKS.md` + `docs/WORK_LOG.md` — theo dõi task

### 0.3: CI config (GitHub Actions)
- [x] Task 0.3.1: Tạo `.github/workflows/ci.yml` — build + test + lint
- [x] Task 0.3.2: Tạo `.github/workflows/release.yml` — release tự động
- [x] Task 0.3.3: Tạo `.golangci.yml` — linter config
- [x] Task 0.3.4: Fix cross-platform build (SysProcAttr abstraction)
- [x] Task 0.3.5: Fix lint issues (unparam, staticcheck, misspell, gofmt)
- Chạy: `git push origin setup/ci-config` → CI xanh ✅

### 0.4: Thiết lập môi trường dev (Win/Lin/Mac)
- [ ] Task 0.4.1: Script setup cho Windows (choco/scoop)
- [ ] Task 0.4.2: Script setup cho Linux (apt)
- [ ] Task 0.4.3: Script setup cho macOS (brew)
- [ ] Task 0.4.4: Tài liệu hướng dẫn cài đặt

## Ghi Chú
- Phase 0 cần hoàn thành trước khi bắt đầu Phase 1 (critical bugs)
- CI cần xanh trên cả 3 OS trước khi merge
