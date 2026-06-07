# Kiến Trúc - 5 Layer

## Tổng Quan

```
┌─────────────────────────────────┐
│ 1. CLI Layer (cmd/)             │
├─────────────────────────────────┤
│ 2. Business Logic               │
├─────────────────────────────────┤
│ 3. Abstraction Layer            │
├─────────────────────────────────┤
│ 4. Infrastructure               │
├─────────────────────────────────┤
│ 5. External Tools               │
└─────────────────────────────────┘
```

## Cấu Trúc Thư Mục Ứng Dụng (Thiết Lập Portable)

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

## Điểm Mount Đa Nền Tảng

**Điểm Mount theo Hệ Điều Hành:**
- **Windows:** Drive letter (ví dụ D:\cs\)
- **Linux:** ~/.local/mnt/cs/
- **macOS:** /Volumes/cs/

**Lý Do:**
- **Windows:** Sử dụng drive letter cho native mount points
- **Linux:** XDG Base Directory compliant (`~/.local/mnt/cs/`), portable, không cần sudo
- **macOS:** Sử dụng `/Volumes/` theo tiêu chuẩn cho mounted filesystems

**Tùy Chọn Mode:**
- **Combine (mặc định):** Điểm mount duy nhất với cấu trúc org/repo
  - Windows: `D:\cs\github\repo-1\`
  - Linux: `~/.local/mnt/cs/github/repo-1/`
  - macOS: `/Volumes/cs/github/repo-1/`
- **Separate:** Nhiều điểm mount (một cho mỗi codespace)
  - Windows: `D:\`, `E:\`, `F:\`
  - Linux: `~/.local/mnt/cs-1/`, `~/.local/mnt/cs-2/`
  - macOS: `/Volumes/cs-1/`, `/Volumes/cs-2/`

## 8 Module Chính

1. **Provider** - Trừu tượng Cloud IDE (GitHub, Gitpod, AWS)
2. **Connection** - Quản lý vòng đời SSH tunnel
3. **Tunnel** - Điều phối SSH, cấp phát cổng
4. **Rclone** - Điều phối mount
5. **State** - Lưu trữ (JSON file, thread-safe)
6. **IDE** - Khởi chạy IDE (VS Code, Zed, IntelliJ)
7. **Config** - Tải cấu hình
8. **UI** - Các prompt CLI, hiển thị

## Luồng Dữ Liệu

**Connect:** Provider.Detect → Provider.Get → TunnelManager.Create → State.AddConnection

**Mount:** State.GetConnection → RcloneManager.ConfigureRemote → RcloneManager.Mount → State.AddMount

**Open:** State.GetConnection → IDEManager.Launch

## Các Interface Chính

```go
type Provider interface {
  Name() string
  List(ctx) ([]Workspace, error)
  Get(ctx, id) (*Workspace, error)
  GetSSHConfig(ctx, id) (*SSHConfig, error)
}

type SSHConfig struct {
  Host    string
  Port    int
  User    string
  KeyFile string
}
```

**Ưu Tiên Cấu Hình (Phát Hiện App Root):**

**Vị Trí App Root:**
1. Biến env: `CLOUD_IDE_MOUNT_ROOT` (ghi đè rõ ràng)
2. Thư mục executable (nếu env không được đặt)

**Lý Do:** Portable theo mặc định (được gửi kèm executable), linh hoạt cho người dùng nâng cao cần vị trí tùy chỉnh thông qua biến env. Logic auto-detect trong `config.Init()` xử lý cả hai trường hợp.

**Thứ Tự Tìm Kiếm SSH Key:**

1. Flag `--key-file` (ghi đè rõ ràng)
2. `{AppRoot}/data/.ssh/cloud-ide` (portable ứng dụng)
3. `~/.ssh/codespaces.auto` (toàn hệ thống)

**Lý Do:** Cố gắng thư mục ứng dụng trước để tính portable, sau đó quay lại key toàn hệ thống để tương thích ngược với các setup hiện có. Người dùng có thể ghi đè thông qua flag khi cần.

**Tìm Kiếm Cấu Hình gh:**

1. `{AppRoot}/config/gh/config.yml` (portable ứng dụng)
2. Cấu hình gh toàn cục (toàn hệ thống, mặc định)

**Lý Do:** Kiểm tra cấu hình thư mục ứng dụng trước cho các thiết lập portable, sau đó quay lại cấu hình gh toàn cục nếu cấu hình ứng dụng không tồn tại. Sử dụng biến env `GH_CONFIG_DIR` để trỏ gh CLI đến thư mục ứng dụng khi cấu hình ứng dụng có.

## Schema State

```json
{
  "connections": {
    "conn-1": {
      "id": "conn-1",
      "workspace_id": "owner/repo",
      "provider": "github",
      "tunnel_port": 2223,
      "ssh_pid": 1234,
      "created_at": "2026-06-07T10:00:00Z"
    }
  },
  "mounts": {
    "z": {
      "drive": "z",
      "rclone_pid": 5678,
      "connections": ["conn-1"]
    }
  }
}
```

**Vị Trí:** `{AppRoot}/config/state.json` (lưu trữ)

## Các Issue Quan Trọng

| # | Issue | File | Nỗ Lực | Ảnh Hưởng |
|---|-------|------|--------|---------|
| 1 | Race condition state file | internal/state/state.go | 1-2 ngày | Hỏng dữ liệu |
| 2 | Process kill không atomic | internal/tunnel/ssh.go | 1 ngày | Tunnel treo |
| 3 | Port TOCTOU | internal/tunnel/port.go | 1 ngày | Xung đột cổng |
| 4 | Silent errors | Multiple | 3-4 ngày | Khó gỡ lỗi |
| 5 | mount.go phức tạp | cmd/mount.go | 1-2 tuần | Khó bảo trì |
