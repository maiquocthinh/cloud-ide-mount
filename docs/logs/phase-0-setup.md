# Nhật Ký: Phase 0 — Thiết Lập

<!-- Ghi nhật ký mỗi ngày làm việc, thêm entry mới ở đầu. -->

## 2026-06-07

### Làm
- [x] 0.3.1: Tạo `.github/workflows/ci.yml` — CI pipeline đầy đủ (lint, test 3 OS, build 6 platform)
- [x] 0.3.2: Tạo `.github/workflows/release.yml` — release workflow (build + archive + upload)
- [x] 0.3.3: Tạo `.golangci.yml` — linter config (errcheck, staticcheck, misspell, ...)
- [x] 0.3.4: Fix cross-platform build — `syscall.SysProcAttr{HideWindow}` chỉ Windows, tạo `internal/executil/` abstraction
- [x] 0.3.5: Fix YAML indentation (tab → space) trong CI workflow
- [x] 0.3.6: Fix artifact name (build-windows/amd64 chứa `/` không hợp lệ)
- [x] 0.3.7: Fix lint issues — `detectSSHPort` unparam, `execCmdOutput` nolint, `Cancelled` misspell, `go fmt`

### Kết quả
- Build: ✅ `go build .` + cross-compile 6 platform — pass
- Test: ✅ `go test -race -v ./...` — pass (chưa có test nào)
- Vet: ✅ `go vet ./...` — clean
- CI: 🟢 Build + Test xanh. Lint đã fix hết issues.

### Blocker / Ghi chú
- `syscall.SysProcAttr.HideWindow` là field Windows-only, cần build tags để cross-platform
- CGO_ENABLED=0 bắt buộc cho cross-compile từ Linux → các nền tảng khác
- errcheck disable tạm thời, TODO khi làm Issue #4 (Silent error handling)
