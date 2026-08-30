# Checklist · 新畫面

這份給「新增一個新 screen / route」。若只是修改既有畫面的一小段，改走 [refactor.md](refactor.md)。

## 動工前

- [ ] 已讀 `guides/08-ui-state.md`、`09-widget-api.md`、`10-widget-state-and-lifecycle.md`、`14-design-system.md`
- [ ] 已確認進入路徑、route 參數、返回行為 → `guides/13-navigation.md`
- [ ] 已列出 loading / empty / error / success / 無權限等畫面狀態
- [ ] 已確認是否需要曝光或互動事件 → `guides/18-analytics.md`

## ① 契約

- [ ] route 只收必要參數（id、filter、tab），不傳整包 entity / DTO
- [ ] screen 對 ViewModel / notifier 的依賴已明確，沒有讓深層 widget 偷讀不該知道的 provider
- [ ] 可重用 widget 的輸入輸出已定義清楚，不依賴畫面外部隱性狀態

## ② 資料

- [ ] 畫面需要的資料來源已確認：哪個 repository / provider 提供、何時載入、是否需要 refresh
- [ ] 畫面不直接碰 Dio、SQLite、shared_preferences、Firebase SDK
- [ ] 外部圖片 / URL / 分享內容若進畫面，已有驗證與降級策略

## ③ 狀態

- [ ] 畫面 state 由 provider / notifier 持有，不把業務狀態塞進 `StatefulWidget` 私有欄位
- [ ] `TextEditingController`、`ScrollController`、`FocusNode`、動畫 controller 的生命週期責任清楚
- [ ] one-off effect（snackbar、dialog、導航）有明確消費方式
- [ ] 曝光或互動埋點不是放在 build body

## ④ 畫面

- [ ] screen 與 content 至少分兩層，content 可以在 preview / widget test 中單獨餵狀態
- [ ] 清單項目有穩定 key，圖片有合理尺寸或 placeholder
- [ ] 字型、顏色、圓角、間距不硬編碼
- [ ] 可點擊區域、語意標籤、文案與深色模式已檢查

## ⑤ 測試

- [ ] widget test 覆蓋主要畫面狀態與至少一個關鍵互動
- [ ] notifier / ViewModel 的狀態轉移有單元測試
- [ ] 若畫面是核心旅程，已有 integration test 或已明確記錄原因
- [ ] 使用 fake provider override，不依賴真後端

## 最終檢查

- [ ] `flutter analyze`
- [ ] `flutter test`
- [ ] `flutter test integration_test` 或記錄未接入原因
- [ ] `python3 flutter-dev-guide/tools/check-rules.py --staged` 或記錄稽核器尚未接入
