# Checklist · 串接新 API

依順序處理：契約 → 資料 → 狀態 → 畫面 → 測試。目標不是「先打通」，而是讓 API 不把 network 細節滲進整個 app。

## 動工前

- [ ] 已讀 `guides/05-data-layer.md`、`06-network.md`、`15-testing.md`
- [ ] API contract 已確認：path、method、request、response、錯誤碼、pagination、auth 需求
- [ ] 已確認要不要快取、離線、重試、去重 → 必要時補讀 `guides/07-persistence.md`
- [ ] 涉及 token、外部 URL、個資時已讀 `guides/17-security.md`

## ① 契約

- [ ] Repository / use case 介面先定義，回傳 app/domain model，不回傳 DTO
- [ ] 方法簽章清楚表達是一次性請求、串流更新還是分頁載入
- [ ] 不把 Dio response、JSON map 或 SDK 型別帶到上層

## ② 資料

- [ ] Dio service / client 設定放在 data / remote service 層
- [ ] DTO 有明確 parser 與 mapper，nullable 欄位處理清楚
- [ ] 錯誤轉譯、timeout、retry、auth header 都在資料層統一處理
- [ ] 分頁 cursor、去重、cache policy 不在 widget 或 notifier 即席硬寫

## ③ 狀態

- [ ] provider / notifier 只處理「何時載入、如何顯示」，不自行拼 URL 或解析 JSON
- [ ] loading / error / empty / data 都能被 state 表達
- [ ] 若有手動重試、下拉刷新、分頁追加，狀態轉移已先想清楚

## ④ 畫面

- [ ] 畫面只消費 state，不直接 import remote service
- [ ] 錯誤訊息已轉成使用者看得懂的文案
- [ ] 外部文章 URL、來源 tag、到期狀態等顯示邏輯都從 UI model 取得，不在 build 內即席計算

## ⑤ 測試

- [ ] DTO / mapper 有測試，尤其是缺欄位、null、錯誤格式
- [ ] Repository 至少測成功、HTTP 錯誤、網路異常、空資料
- [ ] notifier / ViewModel 測到 loading → data / error
- [ ] 若是關鍵 API，已加 widget 或 integration 驗證第一條使用者路徑

## 最終檢查

- [ ] `flutter analyze`
- [ ] `flutter test`
- [ ] `flutter test integration_test` 或記錄未接入原因
- [ ] `python3 flutter-dev-guide/tools/check-rules.py --staged` 或記錄稽核器尚未接入
