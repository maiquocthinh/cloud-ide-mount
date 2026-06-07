# Phase 2: High — v0.2

## Mục Tiêu
Giải quyết các issue High priority sau khi hoàn thành Phase 1 (Critical).

## Danh Sách Task

### 7. Health checks
- [x] #7.1: Tạo `internal/health/checker.go` — Checker + Status models
  - Status type (Alive/Dead/Error)
  - CheckTunnelPort(port) — TCP port check qua net.DialTimeout
  - CheckProcess(pid) — process existence qua os.FindProcess
  - CheckMountDrive(drive) — mount path accessibility qua os.Stat
  - CheckAll(state, codespaces) — full health report
  - Function vars (findProcess, osStat, dialTimeout) để mock trong test
- [x] #7.2: Tạo `internal/health/checker_test.go` — 14 tests
  - CheckTunnelPort: alive, dead, dialTimeout override
  - CheckProcess: alive (self), dead (0/negative), non-existent, findProcess mock
  - CheckMountDrive: alive (TempDir), empty, non-existent, osStat mock
  - CheckAll: empty, with data, missing codespace
  - Status.String() coverage
- [x] #7.3: Tạo `cmd/status.go` — `cs-mount status` command
  - Cobra command với RunE pattern
  - Load state, fetch codespace list, run health checks
  - Display formatted table với emoji indicators (✅/❌/⚠️)
  - Handle edge cases: no state, no tunnels/mounts, gh fetch failure
- [x] #7.4: Update docs
  - CHECKLIST.md — đánh dấu item 7 là ✅
  - phase-2-high.md — task tracking (file này)

### 8. SSH port detection không ổn định
- [ ] Chưa bắt đầu
  - File: `internal/tunnel/tunnel.go` (`detectSSHPort` shadowing)
  - Ghi chú: ...

### 9. Per-user state profiles
- [ ] Chưa bắt đầu
  - File: `internal/state/state.go`
  - Ghi chú: ...

## Ghi Chú
- Health checks sử dụng function var pattern (giống `var execLook` trong mount.go) để dễ test.
- `CheckProcess` trên Windows có thể báo alive giả (os.FindProcess luôn thành công).
- Các check đáng tin cậy nhất trên Windows: tunnel port (TCP) và mount path (os.Stat).
