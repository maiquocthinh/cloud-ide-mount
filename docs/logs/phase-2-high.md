# Nhật Ký: Phase 2 — High (v0.2)

## 2026-06-07

### Làm
- [x] #7.1: Tạo `internal/health/checker.go` — Checker + Status models
  - Status type (Alive/Dead/Error) with String()
  - CheckTunnelPort(port): TCP port check via dialTimeout
  - CheckProcess(pid): process existence via findProcess
  - CheckMountDrive(drive): mount accessibility via osStat
  - CheckAll(state, codespaces): full health report
  - Function vars (findProcess, osStat, dialTimeout) để mock trong test
- [x] #7.2: Tạo `internal/health/checker_test.go` — 15 tests
  - All check functions tested with real and mocked dependencies
  - CheckAll with empty, full, and partial data
- [x] #7.3: Tạo `cmd/status.go` — `cs-mount status` command
  - Cobra command, table display with emoji indicators
  - Handle: no state, empty tunnels/mounts, gh fetch failure
- [x] #7.4: Update docs
  - CHECKLIST.md — item 7 ✅ (10/24 = 42%)
  - tasks/phase-2-high.md — task tracking
  - logs/phase-2-high.md — log file này

### Kết quả
- Build: ✅ `go build ./...` pass
- Test: ✅ `go test -race ./...` — 63+ tests pass (15 health tests mới)
- Vet: ✅ `go vet ./...` — clean

### Ghi chú
- Issue #7 (Health checks) done! 🎉
- Phát hiện: đã có sẵn `cmd/status.go` cũ (simple inline checks) — đã thay thế bằng implementation mới dùng `health` package.
- Health checks package sử dụng function var pattern để mock OS calls trong test.
- `CheckProcess` trên Windows có hạn chế: `os.FindProcess` luôn thành công. Cần cải thiện sau nếu cần.
- Tiếp theo: Issue #8 (SSH port detection) hoặc Issue #9 (Per-user state profiles).

## 2026-06-07 (tiếp)

### Làm
- [x] #8.1: Refactor `detectSSHPort` → `tunnel.DetectSSHPort` với retry + parsing tốt hơn
  - Move từ `cmd/mount.go` sang `internal/tunnel/tunnel.go`
  - Bỏ sudo dependency (dùng `cat` thay vì `sudo grep`)
  - Thêm 3 retry attempts với 1s backoff
  - Parse sshd_config đúng (skip comments, case-insensitive, validate range)
  - Dùng function var (`execSSHCommand`) để mock trong test
- [x] #8.2: Xoá dead code `GetCsSshPort` khỏi `internal/tunnel/tunnel.go`
  - Dead function với logic sai (probe port 22 sau khi tunnel đã mở)
  - Shadowing issue (err variable shadowed trong inner scope)
- [x] #8.3: Tạo `internal/tunnel/sshport_test.go` — 13 tests (10 parse + 3 Detect)
- [x] #8.4: Update docs
- [x] #8.5: Clean up `cmd/mount.go` — xoá `detectSSHPort`, `strconv` import, test cũ

### Kết quả
- Build: ✅ `go build ./...` pass
- Test: ✅ `go test -race ./...` — 76+ tests pass (13 SSH port tests mới)
- Vet: ✅ `go vet ./...` — clean

### Ghi chú
- Issue #8 (SSH port detection) done! 🎉
- `parseSSHPort` bỏ qua dòng comment, case-insensitive, validate range port 1-65535
- `DetectSSHPort` thử `cat` trước (ko cần sudo), fallback `sudo cat` nếu lỗi, default 22
- `execSSHCommand` pattern giúp test dễ dàng (mock exec output)
- Tiếp theo: Issue #9 (Per-user state profiles).
