# Nhật Ký: Phase 1 — Critical (v0.1)

<!-- Ghi nhật ký mỗi ngày làm việc, thêm entry mới ở đầu. -->

## 2026-06-07

### Làm
- [x] Tạo docs structure (CHECKLIST, TASKS, WORK_LOG, templates)
- [x] #1.1: Thêm mutex lock + fileMu package-level
- [x] #1.2: Atomic write (temp file + rename)
- [x] #1.3: Viết 7 test cases (concurrent, atomic, remove, v.v.)
- [x] #1.4: Đổi path từ os.TempDir() sang {AppRoot}/config/state.json

### Kết quả
- Build: ✅ pass
- Test: ✅ `go test -race -v` — 7/7 pass, không race
- Vet: ✅ `go vet ./...` — clean

### Blocker / Ghi chú
- Không có blocker. Bắt đầu với Issue #1.

## 2026-06-07 (tiếp)

### Làm
- [x] #2.1: Tạo `internal/executil/process.go` — `KillProcess(pid, timeout)` with Wait + timeout
- [x] #2.2: Xoá `killPid()` trong `cmd/unmount.go` — chuyển sang `executil.KillProcess()`
- [x] #2.3: Thay 3 chỗ `killPid()` + `time.Sleep()` trong `cmd/mount.go` bằng `executil.KillProcess()`
- [x] #2.4: Viết test `TestKillProcess` + `TestKillProcessNonExistent` trong `internal/executil/process_test.go`

### Kết quả
- Build: ✅ pass
- Test: ✅ `go test -race -v ./...` — 9/9 pass, không race
- Vet: ✅ `go vet ./...` — clean

### Ghi chú
- `killPid()` đã được xoá khỏi `cmd/unmount.go`. Không còn `time.Sleep(time.Second)` sau kill.
- Cơ chế mới: Kill → Wait(timeout) → báo lỗi nếu process không exit.
- Issue #2 done. Tiếp theo: Issue #3 (Port allocation TOCTOU).

## 2026-06-07 (tiếp)

### Làm
- [x] #3.1: Tạo `internal/tunnel/port.go` — `AllocatePort(from)` dùng `net.Listen()` atomic
  - `AllocatedPort` struct giữ listener, đảm bảo port không bị chiếm
  - `Close()` idempotent (an toàn gọi nhiều lần)
- [x] #3.2: Cập nhật `cmd/mount.go` — thay `NextFreePort` bằng `AllocatePort`
  - Listener mở trong suốt setup (sshd, detect port)
  - `ap.Close()` ngay trước `StartTunnel` → window TOCTOU giảm từ ~giây xuống ~micro-giây
- [x] #3.3: Cập nhật `internal/tunnel/tunnel.go` — `NextFreePort` deprecated, delegate sang `AllocatePort`
- [x] #3.4: Viết 6 test cases trong `internal/tunnel/port_test.go`
  - `TestAllocatePort`: allocation cơ bản, verify port đang listen
  - `TestAllocatePortRespectsBusyPort`: skip port đã được dùng
  - `TestAllocatePortReleasesOnClose`: port free sau Close
  - `TestAllocatePortCloseIdempotent`: Close 2 lần không panic
  - `TestConcurrentAllocationNoDuplicatePorts`: 20 goroutine đồng thời, không duplicate
  - `TestAllocatePortLargeOffset`: high port range

### Kết quả
- Build: ✅ pass
- Test: ✅ `go test -race -v ./...` — 17/17 pass, không race
- Vet: ✅ `go vet ./...` — clean

### Ghi chú
- `AllocatePort` thay thế hoàn toàn `NextFreePort` trong mount workflow.
- `net.Listen` là atomic operation — chỉ một process có thể bind thành công.
- `ap.Close()` set `ap.listener = nil` để idempotent, tránh double-close panic.
- Issue #3 done. Tiếp theo: Issue #4 (Silent error handling).
