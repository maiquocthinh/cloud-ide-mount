# Bảng Theo Dõi Dự Án

**Tổng:** 24 hạng mục · 🔴 5 Critical · 🟡 3 High · 🔵 16 Normal
**Tiến độ:** ▰▰▰▰▰▰▰▰▰▰▰▰ 25% (6/24)

---

## 🟢 Phase 0: Thiết Lập

| # | Hạng Mục | Loại | Status | Task Chi Tiết |
|---|----------|------|--------|--------------|
| 0.1 | Khởi tạo Go module + cấu trúc project | Setup | ✅ | [tasks](./tasks/phase-0.md#01-khởi-tạo-go-module--cấu-trúc-project) |
| 0.2 | Tổ chức docs (7 file md) | Docs | ✅ | [tasks](./tasks/phase-0.md#02-tổ-chức-docs-7-file-md) |
| 0.3 | CI config (GitHub Actions) | Setup | ✅ | [tasks](./tasks/phase-0.md#03-ci-config-github-actions) |
| 0.4 | Thiết lập môi trường dev (Win/Lin/Mac) | Setup | ✅ | [tasks](./tasks/phase-0.md#04-thiết-lập-môi-trường-dev-winlinmac) |

## ⏳ Phase 1: Critical — v0.1

| # | Hạng Mục | Loại | Status | Task Chi Tiết |
|---|----------|------|--------|--------------|
| 1 | Race condition state file | Bug | ✅ | [tasks](./tasks/phase-1-critical.md#1-race-condition-state-file) |
| 2 | Process kill không atomic | Bug | ✅ | [tasks](./tasks/phase-1-critical.md#2-process-kill-không-atomic) |
| 3 | Port allocation TOCTOU | Bug | ✅ | [tasks](./tasks/phase-1-critical.md#3-port-allocation-toctou) |
| 4 | Silent error handling | Bug | ✅ | [tasks](./tasks/phase-1-critical.md#4-silent-error-handling) |
| 5 | mount.go complexity | Refactor | ⬜ | — |
| 6 | Logging | Feature | ⬜ | — |

## ⬜ Phase 2: High — v0.2

| # | Hạng Mục | Loại | Status | Task Chi Tiết |
|---|----------|------|--------|--------------|
| 7 | Health checks | Feature | ⬜ | [tasks](./tasks/phase-2-high.md#7-health-checks) |
| 8 | SSH port detection không ổn định | Bug | ⬜ | — |
| 9 | Per-user state profiles | Feature | ⬜ | — |

## ⬜ Phase 3: Production — v1.0

| # | Hạng Mục | Loại | Status | Task Chi Tiết |
|---|----------|------|--------|--------------|
| 10 | Input validation | Quality | ⬜ | [tasks](./tasks/phase-3-production.md#10-input-validation) |
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
