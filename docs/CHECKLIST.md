# Bảng Theo Dõi Dự Án

**Tổng:** 24 hạng mục · 🔴 5 Critical · 🟡 3 High · 🔵 16 Normal
**Tiến độ:** ▰▰▰▰▰▰▰▰▰▰ 4% (1/24)

---

## 🟢 Phase 0: Thiết Lập

| # | Hạng Mục | Loại | Status | Task Chi Tiết |
|---|----------|------|--------|--------------|
| 0.1 | Khởi tạo Go module + cấu trúc project | Setup | ✅ | — |
| 0.2 | Tổ chức docs (7 file md) | Docs | ✅ | — |
| 0.3 | CI config (GitHub Actions) | Setup | ✅ | — |
| 0.4 | Thiết lập môi trường dev (Win/Lin/Mac) | Setup | ⬜ | — |

## ⏳ Phase 1: Critical — v0.1

| # | Hạng Mục | Loại | Status | Task Chi Tiết |
|---|----------|------|--------|--------------|
| 1 | Race condition state file | Bug | ⬜ | [tasks](./tasks/phase-1-critical.md) |
| 2 | Process kill không atomic | Bug | ⬜ | — |
| 3 | Port allocation TOCTOU | Bug | ⬜ | — |
| 4 | Silent error handling | Bug | ⬜ | — |
| 5 | mount.go complexity | Refactor | ⬜ | — |
| 6 | Logging | Feature | ⬜ | — |

## ⬜ Phase 2: High — v0.2

| # | Hạng Mục | Loại | Status | Task Chi Tiết |
|---|----------|------|--------|--------------|
| 7 | Health checks | Feature | ⬜ | [tasks](./tasks/phase-2-high.md) |
| 8 | SSH port detection không ổn định | Bug | ⬜ | — |
| 9 | Per-user state profiles | Feature | ⬜ | — |

## ⬜ Phase 3: Production — v1.0

| # | Hạng Mục | Loại | Status | Task Chi Tiết |
|---|----------|------|--------|--------------|
| 10 | Input validation | Quality | ⬜ | [tasks](./tasks/phase-3-production.md) |
| 11 | Cleanup on crash | Reliability | ⬜ | — |
| 12 | Config file support | Feature | ⬜ | — |
| 13 | Comprehensive testing | Test | ⬜ | — |
| 14 | Regex recompile | Optimize | ⬜ | — |
| 15 | Magic numbers | Optimize | ⬜ | — |
| 16 | Error messages | Quality | ⬜ | — |
| 17 | Multi-OS abstraction | Arch | ⬜ | — |

---

## ✅ Tiêu Chí Hoàn Thành

**v0.1:**
- [ ] 5 bug critical đã sửa
- [ ] Tất cả test pass
- [ ] Logging hoạt động

**v1.0:**
- [ ] 17 issues resolved
- [ ] >50% test coverage
- [ ] Multi-OS support

---

### Chú Thích Trạng Thái

| Ký hiệu | Ý nghĩa |
|---------|---------|
| ⬜ | Chờ xử lý |
| 📋 | Đã lên kế hoạch |
| ⏳ | Đang làm |
| ⚠️ | Bị chặn |
| ✅ | Hoàn thành |
