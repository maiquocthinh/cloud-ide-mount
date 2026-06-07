# Hướng Dẫn Triển Khai

## Quy Trình Làm Việc

1. Chọn issue từ CHECKLIST.md
2. Tạo branch: `git checkout -b fix/issue-N`
3. Viết test trước (TDD)
4. Implement code
5. `go test -race -v ./...`
6. PR + merge

## Quyết Định Thiết Kế (Áp Dụng)

### 1. Hỗ Trợ Đa Nền Tảng (Multi-OS Support)

Ứng dụng được thiết kế để hỗ trợ Windows, Linux, và macOS bằng một codebase duy nhất và workflow portable. Mặc dù các điểm mount và công cụ khác nhau giữa các hệ điều hành, hành vi cốt lõi và trải nghiệm người dùng vẫn nhất quán trên tất cả các nền tảng.

#### Các Điểm Mount Cụ Thể Theo Nền Tảng
- **Windows:** Drive letter (ví dụ `D:\cs\`)
- **Linux:** `~/.local/mnt/cs/` (XDG Base Directory compliant, do người dùng sở hữu, không cần sudo)
- **macOS:** `/Volumes/cs/`

#### Chế Độ Mount: Combine và Separate
Cả hai mode được hỗ trợ trên tất cả các hệ điều hành:

**Combine Mode (mặc định):**
- Điểm mount duy nhất với cấu trúc org/repo
- Windows: `D:\cs\github\repo-1\`
- Linux: `~/.local/mnt/cs/github/repo-1/`
- macOS: `/Volumes/cs/github/repo-1/`
- Sử dụng rclone combine remote (triển khai hiện tại)

**Separate Mode:**
- Nhiều điểm mount, một cho mỗi codespace
- Windows: `D:\`, `E:\`, `F:\`
- Linux: `~/.local/mnt/cs-1/`, `~/.local/mnt/cs-2/`
- macOS: `/Volumes/cs-1/`, `/Volumes/cs-2/`

Lựa chọn của người dùng: `cloud-ide-mount mount --mode combine|separate`

#### Nguyên Tắc Thiết Kế
- Lệnh tương tự hoạt động ở mọi nơi: workflow portable
- Cấu trúc thư mục tương tự (tổ chức org/repo)
- Sự khác biệt chỉ ở các đường dẫn điểm mount (drive vs folder)
- Trừu tượng hóa sự khác biệt về nền tảng trong code

#### Yêu Cầu Triển Khai
- [ ] Abstract path handling (điểm mount cụ thể theo nền tảng)
- [ ] Abstract mount mechanism (Windows: drive letter, Linux/macOS: FUSE)
- [ ] Abstract process management (loại bỏ syscall chỉ dành cho Windows)
- [ ] Hỗ trợ cả hai mode combine + separate trên tất cả OS
- [ ] Platform detection + conditional logic

### 2. Cấu Hình Portable

Tất cả tệp cấu hình, dữ liệu và cache được lưu trữ trong một thư mục ứng dụng duy nhất, loại bỏ ô nhiễm hệ thống và cho phép dễ dàng di động giữa các máy.

#### Cấu Trúc Thư Mục
```
cloud-ide-mount/
├── bin/
│   ├── cloud-ide-mount.exe (hoặc cloud-ide-mount trên Linux)
│   ├── rclone
│   └── gh
├── config/
│   ├── gh/
│   │   ├── config.yml
│   │   └── hosts.yml
│   ├── rclone/
│   │   └── rclone.conf
│   └── state.json
├── data/
│   └── .ssh/
│       └── cloud-ide (SSH key)
└── cache/
    ├── rclone-vfs-cache/
    └── logs/
```

#### Phát Hiện App Root
Vị trí thư mục ứng dụng được xác định bằng chuỗi ưu tiên:
1. **Biến môi trường:** `CLOUD_IDE_MOUNT_ROOT` (ghi đè người dùng)
2. **Thư mục Executable:** Thư mục tương tự như tệp nhị phân cloud-ide-mount (mặc định portable)

Triển khai: Auto-detect trong `config.Init()` và lưu trữ kết quả dưới dạng biến package.

#### Chuỗi Fallback Cho Tệp
Khi phân giải tệp cấu hình hoặc dữ liệu:
1. Kiểm tra `{AppRoot}/config/` hoặc `{AppRoot}/data/`
2. Quay lại các giá trị mặc định toàn hệ thống nếu không tìm thấy

Điều này duy trì tính portable theo mặc định trong khi vẫn linh hoạt thông qua các biến môi trường cho người dùng nâng cao.

#### Tự Động Cấu Hình Lần Đầu
Trên lần chạy đầu tiên, ứng dụng sẽ tự động:
- Tạo các thư mục bị thiếu: `config/`, `data/.ssh/`, `cache/`
- Hiển thị hướng dẫn cấu hình cho người dùng mới
- Đặt các mặc định portable thông qua các biến môi trường
- Chuẩn bị thư mục ứng dụng để sử dụng ngay lập tức

### 3. Chuỗi SSH Key Fallback

Phát hiện SSH key theo thứ tự ưu tiên ủng hộ tính portable trong khi duy trì khả năng tương thích ngược với các thiết lập hiện có.

#### Thứ Tự Ưu Tiên
1. **Command-line flag:** `--key-file` (ưu tiên cao nhất, ghi đè rõ ràng của người dùng)
2. **Vị trí portable:** `{AppRoot}/data/.ssh/cloud-ide` (phương pháp portable-first)
3. **Vị trí toàn hệ thống:** `~/.ssh/codespaces.auto` (tương thích ngược)
4. **Key hiện có đầu tiên chiến thắng** và được sử dụng để tạo tunnel

#### Lý Do
- Portable-first: Key được gói gọn với thư mục ứng dụng di chuyển cùng nó
- Linh hoạt: Hỗ trợ các cấu hình SSH key người dùng hiện có thông qua fallback
- Rõ ràng: Người dùng có thể ghi đè thông qua command-line flag
- Phát hiện một lần: Được lưu trữ trong trạng thái kết nối

### 4. Vị Trí Lưu Trữ Trạng Thái

Trạng thái ứng dụng được lưu trữ để cho phép theo dõi tunnel, mount và kết nối trên các lần khởi động lại.

#### Chi Tiết Lưu Trữ
- **Đường dẫn:** `{AppRoot}/config/state.json`
- **Vòng đời:** Tồn tại trên các lần khởi động lại ứng dụng
- **Tính Portable:** Di chuyển cùng thư mục ứng dụng
- **Di Chuyển:** Chuyển từ `os.TempDir()` đến thư mục ứng dụng (đảm bảo lưu trữ)

#### Những Gì Được Lưu Trữ
- Kết nối hoạt động và cấu hình tunnel của chúng
- Mount hoạt động và ID quá trình rclone của chúng
- Cấu hình SSH và vị trí tệp key
- Trạng thái kết nối và thông tin sức khỏe

#### Đảm Bảo
- Trạng thái tồn tại trên các hoạt động dọn dẹp tạm thời
- Trạng thái di chuyển từ các vị trí tạm thời trên lần chạy đầu tiên
- Truy cập an toàn luồng thông qua mutex locks (xem Issue #1)

### 5. Cấu Hình Lần Đầu

Ứng dụng cung cấp trải nghiệm lần đầu thân thiện với người dùng bằng cách tạo thư mục tự động và hướng dẫn rõ ràng.

#### Hành Động Tự Động Cấu Hình
- Tự động tạo các thư mục cần thiết: `config/`, `data/.ssh/`, `cache/`
- Không cần tạo thư mục thủ công
- Portable theo mặc định (sử dụng thư mục executable)
- Linh hoạt thông qua biến môi trường `CLOUD_IDE_MOUNT_ROOT`

#### Hướng Dẫn Cấu Hình
- Hiển thị các bước tiếp theo cho người dùng mới
- Giải thích cấu trúc thư mục và cấu hình
- Hiển thị ví dụ về các lệnh thường gặp
- Liên kết đến tài liệu để được trợ giúp chi tiết

#### Mặc Định Portable
- Tất cả tệp được lưu trữ trong thư mục ứng dụng
- Không ô nhiễm hệ thống
- Sẵn sàng sử dụng ngay sau khi giải nén
- Có thể đặt ở bất kỳ thư mục nào và hoạt động

## 5 Issue Quan Trọng (v0.1)

### Issue #1: Race Condition State File
- **File:** internal/state/state.go
- **Problem:** Concurrent writes làm hỏng state.json
- **Solution:** Atomic write (temp + rename) + mutex lock
- **Test:** 100 concurrent saves, verify no corruption
- **Effort:** 1-2 ngày

**Design note:** State persistence di chuyển đến app folder (DESIGN_DECISIONS), cần truy cập thread-safe.

### Issue #2: Process Kill Không Atomic
- **File:** internal/tunnel/ssh.go
- **Problem:** killPid() không đợi process exit
- **Solution:** process.Wait() thay vì time.Sleep()
- **Test:** Kill đợi exit, không sleep tùy ý
- **Effort:** 1 ngày

**Design note:** Abstract process management cho multi-OS (DESIGN_DECISIONS).

### Issue #3: Port Allocation TOCTOU
- **File:** internal/tunnel/port.go
- **Problem:** Check port rồi cấp phát (race condition)
- **Solution:** net.Listen() approach (atomic)
- **Test:** Concurrent allocation, không duplicates
- **Effort:** 1 ngày

### Issue #4: Silent Error Handling
- **Files:** rclone.go, mount.go, connection.go
- **Problem:** Errors logged nhưng không return
- **Solution:** Return all errors, không nil returns
- **Test:** All errors propagate
- **Effort:** 3-4 ngày

**Design note:** Áp dụng cho tất cả layers (CLI, business, infrastructure).

### Issue #5: mount.go Complexity
- **File:** cmd/mount.go
- **Problem:** 500+ lines, khó test
- **Solution:** Extract functions: orchestrateTunnels(), buildConfig(), mountDrives()
- **Test:** Unit test mỗi function
- **Effort:** 1-2 tuần

**Design note:** Refactor phải hỗ trợ multi-OS mount modes (combine vs separate).

## Mẫu Code

**Error Handling (Đúng):**
```go
if err := operation(); err != nil {
  return fmt.Errorf("operation failed: %w", err)
}
```

**Concurrent Safety (Đúng):**
```go
s.mu.Lock()
defer s.mu.Unlock()
// Safe operation
```

**Process Management (Đúng):**
```go
process, _ := os.FindProcess(pid)
state, _ := process.Wait()  // Wait, đừng sleep
```

**Platform Abstraction (Đúng):**
```go
var mountPoint string
switch runtime.GOOS {
case "windows":
  mountPoint = "D:\\cs\\"
case "linux":
  mountPoint = filepath.Join(os.ExpandEnv("$HOME"), ".local/mnt/cs")
case "darwin":
  mountPoint = "/Volumes/cs"
}
```

## Danh Sách Kiểm Tra PR

- [ ] `go test -race -v ./...` pass
- [ ] `go test -race ./...` không race conditions
- [ ] `go vet ./...` clean
- [ ] Tests added/updated
- [ ] Link issue: Fixes #N
- [ ] Update CHECKLIST.md
- [ ] Check: design decisions followed (multi-OS, portable, error handling)

## Lệnh Hữu Ích

```bash
go test -race -v ./...           # Tests với race detector
go vet ./...                      # Static analysis
go build -o cloud-ide-mount main.go     # Build
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out # Coverage report
```

## Thứ Tự Triển Khai (Phased)

1. **Phase 1:** Thiết lập Portable (config, đường dẫn state, phát hiện app root)
2. **Phase 2:** Sửa 5 critical issues (race, process, port, errors, refactor)
3. **Phase 3:** Hỗ trợ Multi-OS (trừu tượng hóa đường dẫn/mount/tiến trình)
