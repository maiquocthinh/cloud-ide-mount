# Tra Cứu Nhanh (Bookmark!)

## "Tôi cần..." → "Đọc file này"

| Cần gì | File | Mục đích |
|--------|------|---------|

## Quy Trình Xử Lý Issue

1. Đọc issue trong CHECKLIST.md
2. Xem ví dụ trong IMPLEMENTATION.md
3. Kiểm tra design decisions liên quan
4. Tạo branch: `git checkout -b fix/issue-N`
5. Viết test trước (TDD)
7. Implement code theo pattern
8. `go test -race -v ./...`
9. Push + PR

## Checklist Hàng Ngày

**Standup Sáng:**
- [ ] Xem issue trong CHECKLIST.md
- [ ] Kiểm tra blocker
- [ ] Note: multi-OS implications?

**Trước Code:**
- [ ] Đọc issue chi tiết
- [ ] Xem ví dụ code
- [ ] Đọc pattern có liên quan
- [ ] Check: Design decisions apply? (portable, multi-OS, error handling)

**Trong Code:**
- [ ] Viết test trước (TDD)
- [ ] Follow pattern
- [ ] Consider: Windows vs Linux vs macOS implications
- [ ] Run `go test -race -v ./...` thường xuyên

**Trước PR:**
- [ ] Tất cả test pass
- [ ] Không race condition
- [ ] Link issue: Fixes #N
- [ ] Cập nhật CHECKLIST.md
- [ ] Verify: Design decisions followed

## FAQ

**Q: Issue nào chọn trước?**
A: #1, #2, #3 (parallel) → #4 → #6 → #14-16 (quick wins)

Nhưng ưu tiên dựa trên design decisions:
1. Thiết lập Portable (#1 state location)
2. Sửa lỗi nghiêm trọng (#1-4)
3. Tái cấu trúc (#5)
4. Trừu tượng hóa Multi-OS (các phase sau)

**Q: Tôi bị stuck?**
A: 1) CHECKLIST.md 2) IMPLEMENTATION.md 3) Check design decisions 4) Hỏi team

**Q: Làm sao biết fix đúng?**
A: 1) `go test -race -v ./...` 2) Pattern match 3) Design decisions check 4) Code review

**Q: Tìm ví dụ code ở đâu?**
A: IMPLEMENTATION.md có code before/after cho 5 issue + platform abstraction

**Q: Cập nhật CHECKLIST.md khi nào?**
A: PR tạo → IN_PROGRESS, PR merge → DONE

**Q: Design decisions là gì?**

A: Các quyết định thiết kế hướng dẫn hành vi ứng dụng trên các thiết lập Multi-OS và Portable:

**Hỗ trợ Multi-OS (Ưu tiên 1):**
- Mục tiêu: Windows, Linux, macOS trên một codebase duy nhất
- Windows: Mount bằng drive letter (D:\cs\)
- Linux: XDG-compliant ~/.local/mnt/cs/ (không cần sudo)
- macOS: /Volumes/cs/
- Cùng lệnh ở mọi nơi, chỉ khác điểm mount
- Combine mode: mount duy nhất + cấu trúc org/repo
- Separate mode: nhiều mount riêng cho mỗi codespace

**Thiết lập Portable (Ưu tiên 2):**
- Tất cả config/data/cache trong thư mục ứng dụng
- Không gây ô nhiễm hệ thống
- Cấu trúc:
  - bin/: các tệp thực thi cloud-ide-mount, rclone, gh
  - config/: cấu hình gh, cấu hình rclone, state.json
  - data/: .ssh/cloud-ide (khóa SSH)
  - cache/: rclone-vfs-cache, logs

**Chuỗi dự phòng SSH Key:**
1. Thử ghi đè flag --key-file
2. Thử thư mục ứng dụng: {AppRoot}/data/.ssh/cloud-ide
3. Thử toàn hệ thống: ${USERPROFILE}/.ssh/codespaces.auto
4. Portable trước, tương thích ngược

**Vị trí tệp State:**
- Đường dẫn: {AppRoot}/config/state.json
- Bền vững: tồn tại sau khi khởi động lại, dọn dẹp tạm thời
- Mục đích: theo dõi trạng thái tunnel/mount

**Phát hiện App Root:**
1. Kiểm tra biến env CLOUD_IDE_MOUNT_ROOT
2. Dự phòng đến thư mục chứa tệp thực thi
3. Portable mặc định, linh hoạt nếu cần

**Thiết lập lần đầu:**
- Tự động tạo tất cả thư mục khi chạy lần đầu
- Hiển thị hướng dẫn các bước tiếp theo
- Thiết lập an toàn luồng

**Q: Làm sao test multi-OS?**
A: Dùng runtime.GOOS để kiểm tra hệ điều hành. Dùng build tags: //+build windows, //+build linux, //+build darwin

## Template Standup Tuần

```
Tuần [N] - Phase [X]

Hoàn Thành:
- [x] Issue #X (PR #123)

Đang Làm:
- [ ] Issue #Y (blocker: [nếu có])
  Design note: [multi-OS? portable? error handling?]

Vấn Đề Chặn:
- [danh sách]

Kế Hoạch Tuần Tới:
- [ ] Issue #A
- [ ] Issue #B

Chỉ Số:
- PR merge: [N]
- Test coverage: X% → Y%
- Critical còn lại: [N]
- Design decisions followed: [Y/N] each PR
```

## Design Decisions Checklist

Trước PR, verify mỗi decision point:

**Cân Nhắc Multi-OS:**
- [ ] Windows: Mount bằng drive letter, syscall.SysProcAttr{HideWindow} OK
  - Ví dụ: D:\cs\github\repo-1\
  - Mount mode: combine (một drive) hoặc separate (D:\, E:\, F:\)
  - Process: mã dành riêng cho Windows trong các file windows_*.go

- [ ] Linux: XDG Base Directory compliant ~/.local/mnt/cs/
  - Ví dụ: ~/.local/mnt/cs/github/repo-1/
  - Không cần sudo (người dùng sở hữu thư mục)
  - Thư mục ẩn (ứng dụng quản lý tài nguyên)
  - FUSE mount (rclone xử lý)
  - Dùng filepath.Join() cho đường dẫn đa nền tảng

- [ ] macOS: /Volumes/cs/ mount point
  - Ví dụ: /Volumes/cs/github/repo-1/
  - macOS volume mount (tương tự Linux FUSE)
  - Cùng pattern với Linux (hành vi portable)

- [ ] Các pattern trừu tượng hóa nền tảng đã áp dụng:
  - Xử lý đường dẫn dùng filepath.Join()
  - Cơ chế mount được trừu tượng hóa (Windows drive vs Linux/macOS FUSE)
  - Quản lý tiến trình dùng runtime.GOOS checks
  - Build tags phân tách mã dành riêng cho từng OS

**Thiết lập Portable:**
- [ ] Config trong thư mục ứng dụng: {AppRoot}/config/
  - gh config: {AppRoot}/config/gh/config.yml
  - rclone config: {AppRoot}/config/rclone/rclone.conf
  - State file: {AppRoot}/config/state.json

- [ ] SSH key: Dự phòng thư mục ứng dụng trước
  - Thử: {AppRoot}/data/.ssh/cloud-ide
  - Sau đó: ${USERPROFILE}/.ssh/codespaces.auto
  - Sau đó: ghi đè flag --key-file
  - Portable, tương thích ngược

- [ ] State bền vững:
  - File: {AppRoot}/config/state.json
  - Tồn tại sau khi khởi động lại
  - Theo dõi trạng thái tunnel/mount
  - Không nằm trong os.TempDir()

- [ ] Phát hiện app root:
  - Kiểm tra biến env CLOUD_IDE_MOUNT_ROOT trước
  - Dự phòng đến thư mục chứa tệp thực thi
  - Phát hiện một lần khi khởi động
  - Lưu trong biến package

**Thiết lập lần đầu:**
- [ ] Tự động tạo thư mục khi chạy lần đầu
  - Tạo: bin/, config/, config/gh/, config/rclone/, data/.ssh/, cache/logs/
  - Tạo an toàn luồng
  - Hiển thị hướng dẫn các bước tiếp theo

- [ ] Hướng dẫn người dùng:
  - Hiển thị cấu trúc thư mục
  - Giải thích vị trí cấu hình
  - Các bước tiếp theo để thiết lập (SSH key, thông tin xác thực)

**Xử lý lỗi & Các trường hợp đặc biệt:**
- [ ] Đường dẫn dùng forward slash ở những nơi cần OS-agnostic
- [ ] Symlink được xử lý theo từng OS (Windows vs Unix)
- [ ] Biến môi trường được kiểm tra theo thứ tự ưu tiên

## Ghi Chú Theo Hệ Điều Hành

**Windows:**
- Mount dưới dạng drive letter
- Dùng `syscall.SysProcAttr{HideWindow}` (giữ nguyên OK)
- Đường dẫn: `D:\cs\`
- Combine mode: cùng drive với thư mục con org/repo
- Separate mode: D:\, E:\, F:\ cho nhiều codespace

**Linux:**
- Mount qua FUSE (rclone xử lý)
- Tuân thủ XDG Base Directory
- Đường dẫn: `~/.local/mnt/cs/`
- Phải dùng `filepath.Join()` cho đường dẫn
- Không cần sudo (người dùng sở hữu thư mục)
- Quản lý thư mục portable và ẩn

**macOS:**
- Mount qua macOS volume mount
- Đường dẫn: `/Volumes/cs/`
- Tương tự Linux nhưng điểm mount khác
- Dùng `filepath.Join()` cho đường dẫn

**Code Pattern:**
```go
switch runtime.GOOS {
case "windows":
  // Windows-specific code
  // Drive letter detection
  // syscall handling OK
  // Example: D:\cs\...
case "linux":
  // Linux-specific code
  // FUSE mount via rclone
  // XDG compliance
  // Example: ~/.local/mnt/cs/...
case "darwin":
  // macOS-specific code
  // Volume mount
  // Example: /Volumes/cs/...
}
```

**Cross-Platform Path Handling:**
```go
// ✅ GOOD: works everywhere
path := filepath.Join(appRoot, "config", "state.json")

// ❌ BAD: platform-specific
path := appRoot + "\config\state.json"

// ✅ Mount points platform-specific
switch runtime.GOOS {
case "windows":
  mountPoint = "D:\\cs"
case "linux":
  mountPoint = filepath.Join(os.ExpandEnv("$HOME"), ".local", "mnt", "cs")
case "darwin":
  mountPoint = "/Volumes/cs"
}
```

## Các Giai Đoạn Triển Khai

**Phase 1: Thiết lập Portable** (trước Phase 2)
- Phát hiện App Root
  - Kiểm tra biến env CLOUD_IDE_MOUNT_ROOT
  - Dự phòng đến thư mục chứa tệp thực thi
  - Lưu cache khi khởi động
- State file trong thư mục ứng dụng
  - Di chuyển từ os.TempDir() đến {AppRoot}/config/state.json
  - Đảm bảo bền vững
- Chuỗi dự phòng SSH key
  - Thử thư mục ứng dụng trước: {AppRoot}/data/.ssh/cloud-ide
  - Sau đó toàn hệ thống: ${USERPROFILE}/.ssh/codespaces.auto
  - Sau đó flag --key-file
- Vị trí cấu hình gh
  - Kiểm tra {AppRoot}/config/gh/config.yml trước
  - Dùng cấu hình gh toàn cục làm dự phòng
  - Đặt biến env GH_CONFIG_DIR nếu app config tồn tại
- Thiết lập tự động lần đầu
  - Tự động tạo cấu trúc thư mục
  - Hiển thị hướng dẫn thiết lập
  - Khởi tạo an toàn luồng

**Phase 2: Sửa 5 Critical Issues**
- Giả sử thiết lập portable đã xong
- Sửa: race conditions, quản lý tiến trình, cấp phát cổng, xử lý lỗi, refactor

**Phase 3: Hỗ trợ Multi-OS**
- Trừu tượng hóa xử lý đường dẫn (điểm mount theo nền tảng)
- Trừu tượng hóa cơ chế mount (Windows: drive letter, Linux/macOS: FUSE)
- Trừu tượng hóa quản lý tiến trình (loại bỏ syscall chỉ dành cho Windows)
- Hỗ trợ cả combine + separate mode trên tất cả OS
- Phát hiện nền tảng + logic điều kiện
- Kiểm thử trên Linux/macOS
