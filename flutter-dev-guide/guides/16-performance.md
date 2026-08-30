# 16 · 效能與量測

**管**：Flutter app 的量測紀律、profile / release 模式、啟動路徑、jank、重建與主執行緒阻塞的定位方式、優化前後如何驗證。  
**不管**：單一 widget API 設計與 lifecycle 細節（`09-widget-api`、`10-widget-state-and-lifecycle`）、安全與 log 資料治理（`17-security`）。

| 想找 | 去 |
|---|---|
| rebuild、`const`、list key、圖片尺寸 | `11-rendering-performance` |
| 非同步邊界、isolate、取消、錯誤處理 | `12-async-streams-isolates` |
| 設計系統造成的 layout／字體問題 | `14-design-system` |

---

## 一、量測紀律

Flutter 的效能討論若沒有模式、裝置與數字，就是猜測。順序固定如下：

1. 先定義問題：啟動慢、首屏卡、清單掉幀、記憶體膨脹，不能混在一起講。
2. 用 `--profile` 或 release-like build 在實機建立基準，記錄裝置、版本、操作步驟。
3. 用 DevTools timeline、CPU profiler、memory view 或 performance overlay 找出熱點。
4. 一次只改一個變因，再用同一套腳本重測。
5. 沒有改善就回退，不要把複雜度留在程式裡。

`flutter run` 的 debug 模式只能用來定位問題，**不能拿來宣稱數字變好了**。

## 二、profile mode 是預設量測模式

| 模式 | 能做什麼 | 不能做什麼 |
|---|---|---|
| `debug` | 功能開發、hot reload、粗略觀察 | 任何正式效能結論 |
| `profile` | DevTools timeline、CPU / memory 分析、啟動與 jank 量測 | 發佈驗收 |
| `release` | 最終裝置驗收、商店前 smoke test | 日常互動式分析 |

建議的基準命令：

```bash
flutter run --profile
flutter build apk --release
flutter build ios --release
```

## 三、先看啟動，再看互動

Daily News 這類 app 的第一個效能瓶頸，多半不在動畫，而在啟動與首屏資料準備：

- `main()` 與 app bootstrap 只做必要初始化。重量級 I/O、遠端設定、analytics SDK、圖片預熱都要延後。
- 首屏只渲染第一個可互動結果。首頁方塊、最近更新時間與「立即更新」按鈕先出現，次要區塊延後。
- 不要在 build 方法同步做 JSON parse、大量 sorting、Markdown 轉換或圖片解碼。

若需要把 CPU 密集工作搬離 UI isolate，優先用 `compute()` 或明確封裝的 isolate helper；但先量測，因為 isolate 建立本身也有成本。

## 四、定位 jank 的實用方式

- 打開 performance overlay，看 rasterizer 與 UI thread bar 是否超過 frame budget。
- 清單卡頓先查：item 是否缺 stable key、圖片是否沒有固定尺寸、每次 build 是否重做 expensive mapping。
- 啟動畫面卡頓先查：provider 初始化是否做太多事、是否有同步讀檔或一次建太多 controller。
- memory 持續上升先查：`ScrollController`、`AnimationController`、`StreamSubscription`、`ProviderContainer` 是否未釋放。

## 五、規則

### perf-claim-from-debug-build · MUST_NOT · new-only · manual

任何啟動、frame 或記憶體優化結論，不得來自 debug mode。

✗ 「debug 跑起來變順了」  
✓ 「Pixel 7，`flutter run --profile`，首頁首幀 820ms → 610ms」

例外：無。

### optimize-without-baseline · MUST_NOT · new-only · manual

不得先改架構、拆 provider、搬 isolate，再回頭補量測；優化前必須先留下基準。

✗ 直接把整個首頁拆成五個 provider，沒有任何前測數字  
✓ 先記錄 `profile` trace，再做單一調整

例外：明確的正確性問題，例如 build 中做同步檔案 I/O。

### blocking-work-in-build · MUST_NOT · new-only+on-touch · manual

`build()`、provider 建構、widget lifecycle hook 中不得做同步重運算或 I/O；重工作要預先計算、延後或搬到背景。

✗ `build()` 內做大型 JSON parse  
✓ 在 repository / service 預先轉好，畫面只接收結果

例外：常數時間的輕量格式化。

### startup-initializes-noncritical-sdk · SHOULD_NOT · new-only · manual

非首屏必需的 SDK 不應在 app 啟動時同步初始化；analytics、A/B 實驗、次要追蹤應延後到首屏之後。

✗ `runApp()` 前等待 analytics ready  
✓ app 先起來，再在背景初始化 adapter

例外：Crash reporting 這類必須非常早接管的最小初始化。

### list-performance-checked-without-profile-trace · SHOULD_NOT · new-only · manual

清單或捲動優化若沒有 profile trace、performance overlay 或 DevTools 證據，不應進入「已優化」狀態。

✗ 只靠肉眼感覺順  
✓ 保留 timeline / overlay 的前後比對

例外：無。

## 六、完成前檢查

```bash
flutter analyze
flutter test
flutter run --profile
python3 flutter-dev-guide/tools/check-rules.py --staged
```

PR 或任務紀錄至少要留下：裝置型號、Flutter mode、操作步驟、前後數字，以及是否仍有已知瓶頸。
