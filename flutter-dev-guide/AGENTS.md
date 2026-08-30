# Flutter 開發準則 — AI Agent 入口

這份檔案是 AI coding agent 的入口。人類請看 [README.md](README.md)。

本指南只定義 Flutter 實作的技術邊界、分層、資料責任與品質原則。它不定義需求分析、任務拆分、PR 流程或特定 AI 工具做法。

---

## 每次都讀

[guides/00-principles.md](guides/00-principles.md)。

它是所有 Flutter 實作的共同紅線：SSOT、不可變狀態、單向資料流、分層與可測試性。

## 依任務讀

| 任務 | 讀 |
|---|---|
| 新功能 | `01-architecture` `02-modularization` `03-dependency-injection` |
| 串接 API | `06-network` `05-data-layer` `04-domain-layer` |
| 加本地儲存或 cache | `07-persistence` `05-data-layer` |
| 判斷要不要開 use case | `04-domain-layer` `01-architecture` |
| 整理 provider wiring / fake 注入 | `03-dependency-injection` |
| 重構 feature 資料邊界 | `01-architecture` `05-data-layer` `02-modularization` |

---

## 實作順序

由內而外，讓依賴方向一路往下：

1. 定義 domain model、repository contract、必要時的 use case contract
2. 寫 remote/local service、DTO / entity 與 mapper
3. 在 repository 決定 SSOT、cache、refresh、error translation
4. 用 Riverpod 在 composition root 組裝依賴
5. 寫 ViewModel / controller，把 domain model 收斂成 UI state
6. 寫 Widget / screen，只 render state、回報 event
7. 補齊 repository、mapper、ViewModel 測試；用 provider override 注入 fake

如果你從 Widget 開始，DTO、entity、HTTP error 很容易一路漏進 UI；那通常代表順序反了。

---

## 通用紅線

1. Widget 不直接讀 Repository、service、Dio、SQLite、shared_preferences
2. ViewModel 不直接持有 service、Dio、BuildContext、SQLite、shared_preferences
3. Riverpod 的 `Ref` 只留在 presentation 與 composition root，不傳進 repository、service、use case
4. DTO、SQLite entity、domain model、UI model 各守其層，不共用同一型別跨邊界
5. Dio 只存在 remote service；Repository 只看 service 回傳結果與 DTO mapper
6. `shared_preferences` 只放小型設定，不存新聞資料、cache 清單或 secrets

完整規則索引在 [rules.yaml](rules.yaml)。
