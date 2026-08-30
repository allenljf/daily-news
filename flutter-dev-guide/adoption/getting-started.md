# 在新專案導入

新專案最適合直接把 Flutter guide 全套接上，因為還沒有歷史包袱。

## 1. 複製指南

```bash
cp -r flutter-dev-guide/ /path/to/your-project/
```

整個目錄的目標是自足：人類可直接閱讀，AI agent 可從 `AGENTS.md` 與 `skills/` 進入，稽核器則由後續 `tools/check-rules.py` 提供。

## 2. 先定骨架，再寫功能

開始第一個 feature 前，先把這幾件事定下來：

- app 的 feature / package 切法 → `guides/02-modularization.md`
- Repository / Service / ViewModel 邊界 → `guides/01-architecture.md`
- Riverpod provider 放哪、怎麼 override → `guides/03-dependency-injection.md`、`08-ui-state.md`
- 錯誤模型與資料轉換 → `guides/05-data-layer.md`
- 第一條 API 與第一個 screen 的寫法 → `checklists/new-api.md`、`checklists/new-screen.md`

第一個功能盡量刻意做完整，讓它成為之後複製的範本。

## 3. 把檢查接進日常流程

最小集合：

```bash
flutter analyze
flutter test
python3 flutter-dev-guide/tools/check-rules.py --staged
```

之後再接：

- pre-commit：檢查 staged 檔案
- CI：檢查 branch diff
- integration test：守首頁、登入、第一頁新聞、手動更新

## 4. 先讀哪些文件

如果是新功能，先讀：

1. `guides/00-principles.md`
2. `guides/01-architecture.md`
3. `guides/08-ui-state.md`
4. 對應場景 checklist

如果是資料串接，再加：

1. `guides/05-data-layer.md`
2. `guides/06-network.md`
3. `guides/15-testing.md`

## 5. 常見第一天就該避免的坑

- 直接在 widget 或 notifier 內 new Dio / Firebase / persistence client
- provider 一開始就和 production singleton 綁死，導致測試 override 困難
- 把 DTO 當 app model 到處傳
- analytics / crash SDK 直接散進 feature code
- 把 secret 或代表 server 身分的 token 放到 client

這些都不是「以後再整理」的好債；越晚修，越會變成整條資料流一起改。
