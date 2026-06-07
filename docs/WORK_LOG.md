# Nhật Ký Làm Việc Theo Phase

Hub này liệt kê các phase đã/sẽ ghi log và link đến file log chi tiết.

---

## Phase Đang Active

- Phase 1: Critical — v0.1 → [logs/phase-1-critical.md](./logs/phase-1-critical.md)

## Phase Đã Hoàn Thành

*(chưa có)*

---

## Template File Log

Khi tạo file log mới trong `logs/`, dùng format sau:

```markdown
# Nhật Ký: Tên Phase

## YYYY-MM-DD

### Làm
- [x] Task A: mô tả
- [ ] Task B: mô tả (đang làm dở)

### Kết quả
- Build: ✅ pass / ❌ fail
- Test: ✅ pass / ❌ fail (`go test -race -v ./...`)

### Blocker / Ghi chú
- ⚠️ Vấn đề phát sinh: ...
- Quyết định: ...
```
