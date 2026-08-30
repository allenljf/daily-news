---
name: flutter-guide
description: Use when starting Flutter feature work, screen work, API integration, or refactoring and you need to decide which flutter-dev-guide documents to read before changing code.
---

# Flutter Guide

`flutter-dev-guide/` 是 Flutter 準則庫。這個 skill 的工作只有一個：先判斷任務類型，再把本次必讀文件縮到最小集合。

## Trigger Hints

常見觸發詞：`flutter-guide`、Flutter 準則、該讀哪些 guide、怎麼開始寫這個 screen、這次要看哪些文件。

## Input

輸入應包含：

- 任務類型或使用者意圖：新功能 / 新畫面 / 串 API / 重構 / 修 bug / review
- 受影響的主要檔案或層：widget、provider、repository、service、navigation、analytics、security
- 是否是新增檔案或修改既有程式碼

## Routing

先讀 `flutter-dev-guide/guides/00-principles.md`，其餘依任務選 2 到 4 份：

| 任務 | 加讀 |
|---|---|
| 新功能 | `01-architecture.md` `02-modularization.md` `08-ui-state.md` |
| 新畫面 | `08-ui-state.md` `09-widget-api.md` `10-widget-state-and-lifecycle.md` `14-design-system.md` |
| 串 API | `05-data-layer.md` `06-network.md` `15-testing.md` |
| 本地儲存 | `07-persistence.md` `15-testing.md` |
| 效能 | `11-rendering-performance.md` `16-performance.md` |
| 安全 | `17-security.md` |
| 埋點 | `18-analytics.md` |
| 導航 | `13-navigation.md` |

接著指定對應場景 checklist：

| 場景 | Checklist |
|---|---|
| 新功能 | `checklists/new-feature.md` |
| 新畫面 | `checklists/new-screen.md` |
| 新 API | `checklists/new-api.md` |
| 重構 / 修 bug | `checklists/refactor.md` |

## Output

輸出格式固定：

```markdown
## 本次 Flutter 任務

**類型**：<任務類型>
**必讀**：guides/00-principles.md、guides/<...>
**Checklist**：checklists/<...>.md

**先注意的 3-5 件事**
1. <與本次最相關的紅線或選型點>

**待確認**
- <只有真的存在的未決選型才列>
```

## Guardrails

- 不要一次把所有 guide 都讀進來。
- 不要憑印象複述規則；要實際讀檔。
- 規則細節留在 guide 本文，不要在這個 skill 重述。
- 修改既有程式碼時，提醒呼叫者只修本次碰到的區塊。
