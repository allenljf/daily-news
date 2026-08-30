# 在既有專案導入

既有專案不要一次全開。導入失敗最常見的原因，不是規則太嚴，而是一下子把所有歷史違規都變成當前阻塞。

## 第 0 步：先盤點，不要先整改

當 `tools/check-rules.py` 可用後，先跑：

```bash
python3 flutter-dev-guide/tools/check-rules.py --all
```

把結果當基準線，不是當前待辦。數量多很正常，代表專案走過很多迭代。

## 第 1 步：只要求新程式碼

先讓團隊習慣這份 guide 的語言與檢查方式：

- 新檔案要符合指南
- 既有檔案先不擴散
- code review 用對應 checklist 輔助，而不是看到舊結構就整片翻修

## 第 2 步：改到哪裡，就把哪裡修到位

等團隊熟悉後，再提高到 `on-touch` 心智模型：

- 你新增的程式碼要完全符合
- 你改到的函式／widget／provider 區塊要順手修齊
- 同檔其他沒碰到的區塊先不動

若修一條問題會牽動十幾個檔案，它就不是這次改動的順手修，而是另開重構任務。

## 第 3 步：用場景清單取代抽象說教

最實用的導入方式通常不是叫大家背全部 guide，而是：

- 新 feature 看 `checklists/new-feature.md`
- 新 screen 看 `checklists/new-screen.md`
- 新 API 看 `checklists/new-api.md`
- 重構／修 bug 看 `checklists/refactor.md`

這樣大家會先知道「這次工作要看哪幾份」，而不是一次吞完整套文件。

## 第 4 步：先修低風險高報酬

適合第一批逐步收斂的項目：

- provider override 補齊，讓測試能穩定注入 fake
- widget build 內搬出 I/O、mapping、analytics
- 明顯的 PII log 與 token 處理修正
- DTO / UI model / domain model 邊界拉開
- 新增 `flutter analyze` 與 `flutter test` 到 CI

不適合混在日常 PR 裡偷改的：

- 整個 state 管理框架更換
- 全庫導航重做
- 所有 feature 的 package 拆分
- 一次改寫全部 Repository 錯誤模型

## 第 5 步：AI 工具也要跟著導入

只讓人類讀 guide 不夠。若團隊常用 AI agent，請同步把 `flutter-dev-guide/AGENTS.md` 與對應 skill 接到工具設定，否則最容易出現「人知道規則，寫碼的 agent 不知道」的落差。
