# Lệnh & Flags

## Lệnh Chính

### list
Liệt kê các workspace khả dụng

```bash
cloud-ide-mount list                          # Tất cả workspace
cloud-ide-mount list --provider github        # Lọc theo provider
```


### connect
Tạo SSH tunnel đến workspace

```bash
cloud-ide-mount connect https://github.com/owner/repo
cloud-ide-mount connect owner/repo --provider github
```

**Trả về:** Connection ID, cấu hình SSH (host, port, user)


### mount
Mount kết nối dưới dạng ổ đĩa

```bash
cloud-ide-mount mount conn-1 --drive z
cloud-ide-mount mount conn-1 conn-2 --mode combine
cloud-ide-mount mount conn-1 --mode separate
```

**Chế Độ Mount:**

**Combine mode (mặc định):** Điểm mount duy nhất với cấu trúc org/repo
- Windows: `D:\cs\github\repo-1\`
- Linux: `~/.local/mnt/cs/github/repo-1/`
- macOS: `/Volumes/cs/github/repo-1/`
- Sử dụng rclone combine remote (triển khai hiện tại)
- Tất cả workspace có thể truy cập từ một điểm mount với cấu trúc thư mục có tổ chức

**Separate mode:** Nhiều điểm mount (một cho mỗi codespace)
- Windows: `D:\`, `E:\`, `F:\` (drive letter riêng biệt)
- Linux: `~/.local/mnt/cs-1/`, `~/.local/mnt/cs-2/`, `~/.local/mnt/cs-3/`
- macOS: `/Volumes/cs-1/`, `/Volumes/cs-2/`, `/Volumes/cs-3/`
- Mỗi codespace có điểm mount riêng
- Hữu ích khi bạn cần workspace riêng biệt hoặc thích gán drive trực tiếp

**Người dùng chọn mode:** `cloud-ide-mount mount --mode combine|separate`


### unmount
Ngắt mount ổ đĩa

```bash
cloud-ide-mount unmount z:
cloud-ide-mount unmount --all
```


### open
Mở IDE với remote connection

```bash
cloud-ide-mount open conn-1 --ide vscode
cloud-ide-mount open conn-1 --ide zed
```

**Hỗ trợ IDEs:** vscode, zed, intellij, neovim


### status
Hiển thị trạng thái connection và mount

```bash
cloud-ide-mount status                   # Tất cả
cloud-ide-mount status --connections    # Chỉ connections
cloud-ide-mount status --mounts         # Chỉ mounts
```


## Flag Toàn Cục

- `--start-port N` - Cổng SSH tunnel (mặc định: 2223)
- `--key-file PATH` - Khóa SSH private (mặc định: xem thứ tự tìm kiếm bên dưới)
- `--combine-remote NAME` - Tên remote rclone
- `-f, --force` - Bỏ qua xác nhận
- `--verbose` - Đầu ra chi tiết
- `--config PATH` - Đường dẫn tệp cấu hình


## Thứ Tự Tìm Kiếm SSH Key

Khi flag `--key-file` không được chỉ định, ứng dụng cố gắng tìm keys theo thứ tự sau:

1. `{AppRoot}/data/.ssh/cloud-ide` (tương đối với ứng dụng, portable)
2. `~/.ssh/codespaces.auto` (toàn hệ thống)
3. Ghi đè của người dùng thông qua flag `--key-file` (ghi đè rõ ràng)

**Key đầu tiên tìm thấy được sử dụng.** Phương pháp này ưu tiên tính portable: key có thể được gói gọn với ứng dụng và sử dụng trên bất kỳ hệ thống nào, nhưng quay lại key toàn hệ thống nếu key tương đối với ứng dụng không tồn tại, duy trì tương thích ngược với các thiết lập hiện có.

**Lý Do:** Thiết kế portable-first cho phép ứng dụng và SSH key di chuyển cùng nhau, trong khi dự phòng toàn hệ thống hỗ trợ quy trình làm việc hiện có. Người dùng cũng có thể ghi đè rõ ràng với flag.


## Phát Hiện App Root

Ứng dụng xác định thư mục gốc của nó bằng ưu tiên sau:

1. Biến môi trường: `CLOUD_IDE_MOUNT_ROOT` (nếu được đặt)
2. Vị trí thư mục executable (mặc định nếu env var không được đặt)

Tất cả tệp cấu hình, dữ liệu và cache được lưu trữ tương đối với app root này.

**Ví Dụ:**
```bash
# Portable (mặc định) - sử dụng thư mục tương tự executable cloud-ide-mount
./cloud-ide-mount list        # Sử dụng ./data, ./config, ./cache folders

# Vị trí tùy chỉnh
CLOUD_IDE_MOUNT_ROOT=/opt/cloud-ide-mount cloud-ide-mount list    # Sử dụng /opt/cloud-ide-mount/data, etc.

# Trên Windows
set CLOUD_IDE_MOUNT_ROOT=D:\cloud-ide
cloud-ide-mount list          # Sử dụng D:\cloud-ide\data, D:\cloud-ide\config, etc.
```


## Biến Môi Trường

- `CS_MOUNT_START_PORT=2223` - Cổng tunnel bắt đầu
- `CS_MOUNT_KEY_FILE=~/.ssh/cloud-ide` - Ghi đè khóa SSH
- `CS_MOUNT_COMBINE_REMOTE=combined` - Tên remote rclone
- `CS_MOUNT_VERBOSE=1` - Đầu ra chi tiết
- `CLOUD_IDE_MOUNT_ROOT=/path/to/app` - Vị trí thư mục ứng dụng


**Vị Trí Cấu Hình GitHub:**

Ứng dụng kiểm tra cấu hình gh CLI theo thứ tự sau:

1. Thư mục ứng dụng trước: Kiểm tra `{AppRoot}/config/gh/config.yml`
2. Dự phòng toàn cục: Sử dụng cấu hình gh toàn hệ thống (mặc định)
3. Biến môi trường: `GH_CONFIG_DIR` được đặt nếu cấu hình ứng dụng tồn tại

**Lý Do:** Phương pháp portable-first cho phép cấu hình gh CLI được gói gọn với ứng dụng, đảm bảo nó hoạt động trên các hệ thống khác nhau. Quay lại cấu hình toàn cục của người dùng nếu cấu hình ứng dụng không tồn tại.


## Mã Thoát

- `0` - Thành công
- `1` - Lỗi chung
- `2` - Đối số không hợp lệ
- `3` - Kết nối thất bại
- `4` - Mount thất bại


## Điểm Mount theo OS

| OS | Đường Dẫn Mặc Định | Mode | Chi Tiết |
|----|------|------|---------|
| **Windows** | Drive letter (combine) / Multiple drives (separate) | **Combine:** `D:\cs\github\repo-1\` | Mount duy nhất với drive letter, sử dụng cơ chế gán drive letter của Windows. Nhiều repo được tổ chức theo org/repo. |
| | | **Separate:** `D:\`, `E:\`, `F:\` | Mỗi workspace có drive letter riêng. |
| **Linux** | `~/.local/mnt/cs/` | **Combine:** `~/.local/mnt/cs/github/repo-1/` | XDG Base Directory compliant (tiêu chuẩn cho Linux apps). Người dùng sở hữu thư mục, không cần sudo. Thư mục ẩn được quản lý bởi ứng dụng. Portable trên bất kỳ hệ thống Linux nào. |
| | | **Separate:** `~/.local/mnt/cs-1/`, `~/.local/mnt/cs-2/` | Mỗi workspace có mount folder riêng. |
| **macOS** | `/Volumes/cs/` | **Combine:** `/Volumes/cs/github/repo-1/` | macOS volume mount. Điểm mount duy nhất với cấu trúc org/repo. |
| | | **Separate:** `/Volumes/cs-1/`, `/Volumes/cs-2/` | Mỗi workspace mount như separate macOS volume. |


## Khắc Phục Sự Cố

**Cổng đã sử dụng:**
```bash
cloud-ide-mount mount --start-port 2224
```

**Kết nối SSH thất bại:**
- Kiểm tra: `test -f ~/.ssh/cloud-ide` hoặc vị trí tương đối với ứng dụng
- Permissions: `chmod 600 ~/.ssh/cloud-ide`
- gh CLI: `gh --version`

**Mount thất bại:**
- Kiểm tra rclone: `rclone version`
- Drive đang sử dụng: `cloud-ide-mount status`
- Connection hoạt động: Kiểm tra PID
- Kiểm tra permissions trong app folder

**Thiết lập lần đầu:**

Trên lần chạy đầu, ứng dụng tự động tạo cấu trúc thư mục sau:

- `{AppRoot}/config/` - Thư mục cấu hình (lưu trữ config.yaml, gh config, rclone config)
- `{AppRoot}/data/.ssh/` - Thư mục khóa SSH (lưu trữ cloud-ide, codespaces keys)
- `{AppRoot}/cache/` - Thư mục cache (lưu trữ rclone-vfs-cache, logs)

**Quy trình thiết lập:**
1. Ứng dụng phát hiện lần chạy đầu
2. Tự động tạo tất cả thư mục cần thiết
3. Hiển thị hướng dẫn thiết lập với các bước tiếp theo
4. Người dùng làm theo hướng dẫn để cấu hình ban đầu

**Ví Dụ cấu trúc thư mục tạo:**
```
cloud-ide-mount/
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

Thông báo hướng dẫn được hiển thị để giúp người dùng mới hiểu thiết lập và các bước tiếp theo để kết nối workspace.
