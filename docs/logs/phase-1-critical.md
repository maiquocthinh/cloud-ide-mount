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
