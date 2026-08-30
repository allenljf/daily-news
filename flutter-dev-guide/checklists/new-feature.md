# Checklist · 新功能

依 `flutter-dev-guide/AGENTS.md` 的順序由內而外完成：契約 → 資料 → 狀態 → 畫面 → 測試。每一項都要能回答「做了」或「不適用，原因是什麼」。

## 動工前

- [ ] 已讀 `guides/00-principles.md`、`01-architecture.md`
- [ ] 已確認這個功能要放在哪個 feature / package → `guides/02-modularization.md`
- [ ] 已確認是否需要新 Repository、UseCase、route、analytics 事件
- [ ] 已列出主要成功、空資料、錯誤與權限情境

## ① 契約

- [ ] Repository / adapter 介面已定義，方法簽章反映實際行為（一次性用 `Future`、持續資料用 `Stream`）
- [ ] domain / app model 是純 Dart，不含 DTO、JSON、UI 型別
- [ ] feature 對外只暴露需要的 interface，不把 Dio、Firebase、SQLite 型別帶出來

## ② 資料

- [ ] Remote / local service 的責任已分清楚，沒有在 widget 或 notifier 直接呼叫 Dio / persistence
- [ ] DTO、entity、domain model、UI model 有明確轉換邊界
- [ ] 錯誤有轉成 app 可處理的型別，沒有把裸 exception 往上丟
- [ ] 涉及外部 URL、token、PII 時已對照 `guides/17-security.md`
- [ ] 需要埋點時只新增 adapter 與事件型別，不讓 SDK 穿透 feature code

## ③ 狀態

- [ ] Riverpod provider / notifier 的責任已界定：輸入、輸出與 side effect 清楚
- [ ] `AsyncValue` 或自訂 state 已完整表達 loading / data / error / empty
- [ ] one-off effect 不是靠隱性全域狀態或直接持有 navigator / context
- [ ] provider 可在測試中 override，沒有偷綁 production singleton

## ④ 畫面

- [ ] screen 與 content 分層清楚；可重用 widget 的 API 只收值與 callback
- [ ] 清單、圖片、輸入欄位、controller lifecycle 已對照 `guides/09-widget-api.md`、`10-widget-state-and-lifecycle.md`
- [ ] 使用者可見文案、間距、顏色、字級都走設計系統與字串資源
- [ ] build 方法內沒有同步重運算、I/O 或直接埋點

## ⑤ 測試

- [ ] Repository / mapper / notifier 有單元測試
- [ ] widget test 覆蓋 loading / empty / error / success 或至少本功能實際會出現的狀態
- [ ] 關鍵互動或平台整合有 integration test 或明確列出為何暫不適用
- [ ] fake / provider override 已準備好，沒有依賴真 API 或真 Firebase

## 最終檢查

- [ ] `flutter analyze`
- [ ] `flutter test`
- [ ] `flutter test integration_test` 或記錄本專案尚未接入 integration test
- [ ] `python3 flutter-dev-guide/tools/check-rules.py --staged` 或記錄稽核器尚未接入
