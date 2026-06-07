# Phase 1: Critical — v0.1

## Mục Tiêu
Sửa 5 bug critical + thêm logging để đảm bảo ổn định cơ bản cho ứng dụng.

## Danh Sách Task

### #1: Race Condition State File
- File: `internal/state/state.go`

- [x] Task 1.1: Thêm mutex lock vào State struct
  - Chi tiết: `sync.RWMutex` cho data + `sync.Mutex` cho file I/O
- [x] Task 1.2: Triển khai atomic write (temp file + rename)
  - Chi tiết: `os.CreateTemp()` → ghi → `os.Rename()`
- [x] Task 1.3: Viết test 100 concurrent saves
  - Chi tiết: `go test -race` verify không corruption
- [x] Task 1.4: Di chuyển state file từ `os.TempDir()` đến `{AppRoot}/config/state.json`
  - Ghi chú: Dùng `CLOUD_IDE_MOUNT_ROOT` env var hoặc thư mục executable

### #2: Process Kill Không Atomic
- File: `internal/tunnel/ssh.go`, `cmd/unmount.go`, `cmd/mount.go`

- [x] Task 2.1: Thay `time.Sleep()` bằng `process.Wait()`
  - Chi tiết: Đợi process exit thực sự, không sleep cứng
- [x] Task 2.2: Xử lý timeout nếu process không chịu exit
  - Chi tiết: `Wait()` với timeout + `SIGKILL` nếu quá lâu
- [x] Task 2.3: Viết test kill + verify process đã exit

### #3: Port Allocation TOCTOU
- Files: `internal/tunnel/port.go`, `internal/tunnel/port_test.go`, `cmd/mount.go`

- [x] Task 3.1: Thay thế check-then-allocate bằng `net.Listen()` atomic
  - Chi tiết: `AllocatePort()` dùng `net.Listen()` → Listen thành công → port được giữ; fail → port tiếp theo
- [x] Task 3.2: Giữ listener mở trong suốt setup (sshd, detect port), release ngay trước `StartTunnel`
  - Chi tiết: Window TOCTOU giảm từ ~giây (network calls) xuống ~micro-giây (function call)
- [x] Task 3.3: Viết test concurrent allocation, verify không duplicate port
  - Chi tiết: 6 tests (basic, busy, release, idempotent, concurrent, large offset)

### #4: Silent Error Handling
- Files: các file trong `internal/` (rclone.go, mount.go, connection.go)

- [x] Task 4.1: Soát tất cả `if err != nil { log... }` không return
  - Chi tiết: Phát hiện 16 vị trí silent error trong internal/ và cmd/
- [x] Task 4.2: Sửa thành `return fmt.Errorf("context: %w", err)` — propagate error lên trên
  - Chi tiết: Fix 5 files (rclone.go, ide.go, mount.go, unmount.go, open.go)
- [x] Task 4.3: Đảm bảo CLI layer hiển thị lỗi rõ ràng cho người dùng
  - Chi tiết: `RunE` return errors được cobra display ra stderr + os.Exit(1)
- [x] Task 4.4: Viết test verify tất cả errors propagate đúng
  - Chi tiết: 10 tests mới (3 ide + 7 cmd)

### #5: mount.go Complexity
- File: `cmd/mount.go`

- [ ] Task 5.1: Tách `orchestrateTunnels()` — logic tạo tunnel riêng
- [ ] Task 5.2: Tách `buildConfig()` — cấu hình mount riêng
- [ ] Task 5.3: Tách `mountDrives()` — gắn kết ổ đĩa riêng
- [ ] Task 5.4: Viết unit test cho từng function mới
- [ ] Task 5.5: Verify vẫn hỗ trợ combine + separate mode

### #6: Logging
- File: thêm `internal/logging/`

- [ ] Task 6.1: Thiết lập logging package (struct logger, level, output)
  - Chi tiết: support log level (debug/info/warn/error), output file + stdout
- [ ] Task 6.2: Thay `fmt.Println` / `log.Println` rải rác bằng logger
- [ ] Task 6.3: Thêm context fields (workspace ID, connection ID) cho dễ trace
- [ ] Task 6.4: Viết test logging output

## Ghi Chú
- Quy tắc: mỗi function chỉ làm 1 việc, test kèm theo
- Kiểm tra: `go test -race -v ./...` và `go vet ./...` trước PR
- Design decisions: portable-first, multi-OS support, error không silent
