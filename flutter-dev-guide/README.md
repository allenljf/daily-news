# Flutter 開發準則

一份可直接複製到任何 Flutter 專案的開發準則，基線採用 Flutter 官方的
View/ViewModel + Repository/Service 架構，並把 Riverpod、Dio 標示為本專案採用的成熟社群方案。

給 AI agent 的入口是 [AGENTS.md](AGENTS.md)。

---

## 這份管什麼、不管什麼

**管**：Flutter feature 的分層、資料邊界、依賴注入、網路與持久化責任。

**不管**：需求怎麼拆、PR 怎麼開、要不要用哪個工作流工具。這份指南只管「程式碼該長什麼樣」。

**不綁定任何 AI 工具。** `rules.yaml` 是機器可讀規則的唯一真相；之後的檢查腳本可接進任何 CI 或 git hook。

---

## 目前內容

```text
flutter-dev-guide/
├── AGENTS.md
├── README.md
├── rules.yaml
├── tools/
│   ├── check-rules.py
│   └── test_check_rules.py
└── guides/
    ├── 00-principles.md
    ├── 01-architecture.md
    ├── 02-modularization.md
    ├── 03-dependency-injection.md
    ├── 04-domain-layer.md
    ├── 05-data-layer.md
    ├── 06-network.md
    ├── 07-persistence.md
    ├── 08-ui-state.md
    ├── 09-widget-api.md
    ├── 10-widget-state-and-lifecycle.md
    ├── 11-rendering-performance.md
    ├── 12-async-streams-isolates.md
    ├── 13-navigation.md
    ├── 14-design-system.md
    ├── 15-testing.md
    ├── 16-performance.md
    ├── 17-security.md
    └── 18-analytics.md
├── checklists/
│   ├── new-feature.md
│   ├── new-screen.md
│   ├── new-api.md
│   └── refactor.md
├── adoption/
│   ├── getting-started.md
│   ├── existing-project.md
│   └── integrating-ai-tools.md
└── skills/
    ├── flutter-guide/SKILL.md
    └── flutter-review/SKILL.md
```

已交付完整的 guide、checklist、adoption、agent skill 文件，以及可接進 CI / git hook 的單檔規則稽核器。

---

## 怎麼開始

### 新專案

先讀 [guides/00-principles.md](guides/00-principles.md)，再依任務從 [AGENTS.md](AGENTS.md) 選 2 到 4 份相關 guide。

### 既有專案

先把它當成 review 標準，不要一次把全部規則硬套到老專案。`rules.yaml` 已標記每條規則的
`level`、`enforce`、`check`，後續 checker 只會自動化單檔可安全判定的規則。

## 稽核器用法

先確認規則索引與 owner guide 一致：

```bash
python3 flutter-dev-guide/tools/check-rules.py --self-check
```

掃描所有支援的檔案（`.dart`、`pubspec.yaml`、`analysis_options.yaml`）：

```bash
python3 flutter-dev-guide/tools/check-rules.py --all
```

只掃描 staged 變更：

```bash
python3 flutter-dev-guide/tools/check-rules.py --staged
```

掃描某個 commit / branch / range 相對的 diff：

```bash
python3 flutter-dev-guide/tools/check-rules.py --diff HEAD~1
```

只掃描指定檔案：

```bash
python3 flutter-dev-guide/tools/check-rules.py --files \
  lib/features/news/presentation/news_page.dart \
  pubspec.yaml \
  analysis_options.yaml
```

輸出格式固定為 `path:line: rule-id: message`；有違規時 exit code 為 `1`，無違規時為 `0`，規則設定錯誤或 git 參數錯誤時為 `2`。

若某一行需要例外，將 `// guide-ignore: rule-id` 放在同一行，或放在前一行讓它只忽略下一個非空白行。

---

## 技術基線

| 領域 | 採用 | 定位 |
|---|---|---|
| 架構 | Flutter View/ViewModel + Repository/Service | Flutter 官方建議 |
| UI | Flutter Material 3 | Flutter 官方 |
| State / composition | `flutter_riverpod` | 社群主流，本案選型 |
| Network | `dio` | 社群主流，本案選型 |
| Navigation | `go_router` | Flutter team 維護 |
| JSON | `json_serializable` | Dart / Google 發佈 |
| 小型設定 | `shared_preferences` async API | Flutter team 發佈 |
| 離線資料 | SQLite；複雜 query 時採 Drift | Flutter 官方示範 + 社群成熟方案 |

Riverpod 與 Dio 是工具選型，不改變官方建議的責任邊界：View 只看 ViewModel，ViewModel 只調 Repository，Repository 再協調 service。

---

## 修改這份指南

1. 改規則要同步兩處：對應 `guides/*.md` 章節與 [rules.yaml](rules.yaml)
2. 一條規則只由一份 guide 擁有，其他 guide 只 cross-reference，不重述
3. 改完至少重跑 `python3 flutter-dev-guide/tools/check-rules.py --self-check`

## 依情境使用

- 新功能：[checklists/new-feature.md](checklists/new-feature.md)
- 新畫面或 route：[checklists/new-screen.md](checklists/new-screen.md)
- 新 API / 資料串接：[checklists/new-api.md](checklists/new-api.md)
- 重構或修 bug：[checklists/refactor.md](checklists/refactor.md)
- 新專案導入：[adoption/getting-started.md](adoption/getting-started.md)
- 既有專案漸進導入：[adoption/existing-project.md](adoption/existing-project.md)
- AI 工具整合：[adoption/integrating-ai-tools.md](adoption/integrating-ai-tools.md)
- Agent 導讀：[skills/flutter-guide/SKILL.md](skills/flutter-guide/SKILL.md)
- 變更稽核：[skills/flutter-review/SKILL.md](skills/flutter-review/SKILL.md)

---

## 主要來源

- Flutter architecture guide: <https://docs.flutter.dev/app-architecture/guide>
- Flutter architecture recommendations: <https://docs.flutter.dev/app-architecture/recommendations>
- Flutter architecture concepts: <https://docs.flutter.dev/app-architecture/concepts>
- Daily News research note: [docs/research/2026-08-29-modern-flutter-architecture.md](../docs/research/2026-08-29-modern-flutter-architecture.md)
