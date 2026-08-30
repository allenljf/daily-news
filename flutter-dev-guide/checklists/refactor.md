# Checklist · 重構或修改既有程式碼

這份的重點是守住邊界。既有專案一定有舊結構，不先劃清本次改動範圍，小修改很快就會擴成整片重寫。

## 動工前

- [ ] 已列出本次真正要動的檔案與區塊
- [ ] 已讀對應領域指南：至少 `guides/00-principles.md`，再加本次涉及的資料／狀態／畫面文件
- [ ] 若是修 bug，已先想好如何重現與驗證，不是只憑症狀猜

## ① 契約

- [ ] 公開 API、provider、Repository 介面是否真的需要改；若不需要，不順手擴張
- [ ] 改動沒有把 DTO、SDK、平台型別往外層推
- [ ] 若牽涉 route / state / adapter 公開介面，已確認受影響呼叫點

## ② 資料

- [ ] 只在必要的 repository / service / mapper 區塊內修改，不擴散到無關模組
- [ ] 空 catch、裸 exception、硬編碼 URL / token / PII 若落在改動區塊內，順手修掉
- [ ] 若修一條違規會牽動大量檔案，記成後續工作，不在這次偷渡大重構

## ③ 狀態

- [ ] 既有 provider / notifier 改動後，狀態仍可推理；沒有多塞一條隱性 mutable state
- [ ] 修改的 state 轉移有測試保護，尤其是 bug fix
- [ ] navigation / dialog / snackbar 等 effect 沒有越修越分散

## ④ 畫面

- [ ] 只修改需要的 screen / widget；同檔其他無關區塊不順手重寫
- [ ] 若畫面區塊本來就違反設計系統或 lifecycle 規則，而你這次碰到它，至少把改到的部分修齊
- [ ] build 方法沒有因為臨時修補又塞進更多 I/O、mapping 或 analytics

## ⑤ 測試

- [ ] 修 bug 先有能重現問題的測試或最小驗證腳本
- [ ] 重構前後行為不變，有原測試保護或先補測試
- [ ] 沒有透過放寬測試、拿掉斷言、增加 sleep 來讓綠燈成立

## 最終檢查

- [ ] `flutter analyze`
- [ ] `flutter test`
- [ ] `flutter test integration_test` 或記錄未接入原因
- [ ] `python3 flutter-dev-guide/tools/check-rules.py --staged` 或記錄稽核器尚未接入
