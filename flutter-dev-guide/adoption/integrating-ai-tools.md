# 接到你的 AI 工具

這份 Flutter guide 不綁特定 AI 工具。共同原則只有兩個：

1. 讓工具知道 `flutter-dev-guide/AGENTS.md` 是入口。
2. 讓工具在動手前先讀對應的 `guides/` 與 `checklists/`，不是靠印象回答。

## Codex / 多數 agent

在專案根目錄的 `AGENTS.md` 加一段，指向：

- `flutter-dev-guide/AGENTS.md`
- `flutter-dev-guide/skills/flutter-guide/SKILL.md`
- `flutter-dev-guide/skills/flutter-review/SKILL.md`

工作流建議：

- 要開始實作前先跑 `flutter-guide`
- 要做自我 review 或 PR 前檢查時跑 `flutter-review`

## Claude Code

在 `CLAUDE.md` 裡加入最小指示：

```markdown
Flutter 程式碼遵守 flutter-dev-guide/AGENTS.md。
動手前依任務類型讀對應的 guides/ 與 checklists/。
新增檔案全面遵守；修改既有檔案只修本次碰到的區塊。
```

若工具支援 skills，將 `skills/flutter-guide` 與 `skills/flutter-review` 複製到其可搜尋的 skills 目錄。

## Cursor

建立 `.cursor/rules/flutter-guide.mdc`：

```markdown
---
description: Flutter 開發準則
alwaysApply: true
---

Flutter 程式碼遵守 @flutter-dev-guide/AGENTS.md。
動手前依任務類型讀對應的 flutter-dev-guide/guides/ 與 checklists/。
```

## GitHub Copilot

在 `.github/copilot-instructions.md` 放最短版本，避免超長提示：

```markdown
Flutter 程式碼遵守 flutter-dev-guide/AGENTS.md。
重點：UI 不直接碰 Dio / Firebase / persistence；Riverpod state 可 override 測試；
DTO、app model、UI model 分層；token / PII 不得進 log；analytics 只經 adapter。
```

## 委派或 subagent 的注意事項

若工具會開 subagent / background agent，記得把 guide 路徑一起帶進去。多數 agent 不會自動繼承主 agent 讀過的上下文。

最小指示模板：

```text
實作前先讀：
- flutter-dev-guide/guides/00-principles.md
- flutter-dev-guide/guides/<本次領域檔>
- flutter-dev-guide/checklists/<本次場景>.md

新增檔案全面遵守；修改既有檔案只修本次碰到的區塊。
有選型不確定時先停下來回報，不要自己補規則。
```

## AI 工具整合時要驗的事

- 工具真的會讀 guide，而不是只讀到一小段摘要
- review 任務會對照 checklist，不只是跑靜態分析
- skills 的描述只負責「何時使用」，不偷塞流程摘要
- subagent 指示有把 guide 路徑明確帶上
