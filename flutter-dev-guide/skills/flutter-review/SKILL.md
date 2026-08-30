---
name: flutter-review
description: Use when reviewing Flutter changes against flutter-dev-guide, checking which guide areas apply to a diff, or preparing a pre-merge self-review for Flutter code.
---

# Flutter Review

這個 skill 用來把 Flutter 變更對照回 `flutter-dev-guide/`。它不是規則全文，而是 review 的路由與報告格式。

## Trigger Hints

常見觸發詞：`flutter-review`、Flutter review、準則稽核、PR 前檢查、這份 diff 有沒有違反 guide。

## Input

輸入應包含：

- review 範圍：working tree、commit、branch diff 或指定檔案
- 主要變更類型：widget、provider、repository、service、navigation、analytics、security、tests
- 是否已有 `flutter analyze`、`flutter test`、rule checker 結果

## Routing

一律先讀 `flutter-dev-guide/guides/00-principles.md`，再依 diff 特徵補讀：

| Diff 特徵 | 加讀 |
|---|---|
| `Widget`、`ConsumerWidget`、`HookConsumerWidget` | `08-ui-state.md` `09-widget-api.md` `10-widget-state-and-lifecycle.md` |
| 清單、動畫、圖片、rebuild hot path | `11-rendering-performance.md` `16-performance.md` |
| `Notifier`、`AsyncNotifier`、provider 組合 | `08-ui-state.md` `15-testing.md` |
| Repository、service、DTO、Dio | `05-data-layer.md` `06-network.md` `15-testing.md` |
| token、外部 URL、log、PII | `17-security.md` |
| analytics / crash / consent | `18-analytics.md` |
| route / deep link | `13-navigation.md` |
| 測試檔 | `15-testing.md` |

若是場景明確的變更，再對照：

- `checklists/new-feature.md`
- `checklists/new-screen.md`
- `checklists/new-api.md`
- `checklists/refactor.md`

## Output

輸出格式固定：

```markdown
## Flutter Guide Review

**範圍**：<diff 範圍>
**已對照**：guides/00-principles.md、guides/<...>
**Checklist**：<若有，列出>

### 必修
1. <只列目前這份變更必須修的問題>

### 建議修
1. <有價值但非阻斷的問題>

### 需確認
1. <資訊不足的判斷點>

### 已檢查無問題
- <讓讀者知道已看過哪些面向>
```

## Guardrails

- 只評論變更範圍，不把未修改區域拉進來湊數。
- 先看 `flutter analyze` / `flutter test` / rule checker 結果，再看設計層面問題。
- 腳本能抓的機械問題，優先交給 rule checker；這個 skill 補的是分層、狀態、測試強度與資料治理。
- 不要在 skill 內重述 guide 規則；引用對應 guide 即可。
