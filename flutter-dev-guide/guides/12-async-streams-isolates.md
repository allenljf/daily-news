# 12 · Future、Stream、取消與 Isolate

**這份管**：Future 與 Stream 在 Flutter / Riverpod 畫面中的使用邊界、取消與清理、錯誤映射、什麼工作值得丟給 isolate。
**這份不管**：Widget 生命週期細節、導航路由定義、設計 token。

| 需要什麼 | 去哪 |
|---|---|
| `AsyncValue` 在畫面上的呈現 | `08-ui-state` |
| `dispose`、`mounted`、controller 擁有權 | `10-widget-state-and-lifecycle` |
| rebuild 熱點與 lazy list | `11-rendering-performance` |
| `go_router` 與返回值流程 | `13-navigation` |

非同步工作的責任劃分很單純：資料工作在 repository / service / provider，Widget 只消費結果與回報使用者意圖。

---

### future-created-outside-build · MUST

適用檔案：`lib/**/screens/**/*.dart`、`lib/**/widgets/**/*.dart`
安全自動檢查：regex（可掃描 `FutureBuilder(future: someCall())`）+ manual

不可在 `build()` 內即時建立新的 Future 或直接呼叫 repository / service。每次 rebuild 都可能重新觸發工作。

✗ `FutureBuilder(future: api.fetchNews(), builder: ...)`
✓ 先在 provider、`initState`、或已快取的欄位建立 Future，再交給畫面

例外：真正的靜態一次性本地 Future，且父層保證 widget 不會頻繁重建。

### stream-subscription-cancelled-on-dispose · MUST

適用檔案：`lib/**/widgets/**/*.dart`
安全自動檢查：regex（可掃描 `listen(` 與 `StreamSubscription`）+ manual

手動建立的 `StreamSubscription` 必須在 `dispose()` 取消；若來源依賴 widget 輸入，還要在 `didUpdateWidget` 換源時重建。

✗ `stream.listen(_handleEvent);` 沒有保存 subscription
✓ `late final StreamSubscription sub; ... sub.cancel();`

例外：使用 `StreamBuilder`、Riverpod stream provider 或 `ref.listen` 時，由框架接手生命週期。

### cancellation-not-swallowed · MUST

適用檔案：`lib/features/**/application/**/*.dart`、`lib/features/**/data/**/*.dart`
安全自動檢查：manual

任何廣義的 `catch` 不得把取消語意吞掉。若底層 API 有自己的取消例外或 token，必須讓它能往上終止，而不是一律轉成 generic error。

✗ `catch (error) { state = AsyncValue.error(error, stack); }`
✓ 先辨識取消，再只把真正失敗轉成畫面錯誤

例外：底層平台 API 不提供顯式取消時，可在上層忽略過期結果，但要明說這是過期資料保護，不是假裝取消。

### async-error-mapped-once · MUST

適用檔案：`lib/features/**/data/**/*.dart`、`lib/features/**/application/**/*.dart`
安全自動檢查：manual

同一個 async failure 只能在一個層級被翻譯一次。repository 把 transport / persistence error 映射成領域錯誤後，provider 不要再猜原始例外型別。

✗ widget 判斷 `DioExceptionType.connectionTimeout`
✓ repository 回 `NewsFailure.timeout`，provider / widget 只處理 `NewsFailure`

例外：最外層 crash reporting 可額外記錄原始錯誤，但不影響 UI decision。

### isolate-only-for-heavy-pure-work · SHOULD

適用檔案：`lib/**/application/**/*.dart`、`lib/**/data/**/*.dart`
安全自動檢查：manual

只有 CPU 密集、可序列化、純函式的工作才值得用 `Isolate.run` / `compute`，例如大型 JSON 轉換、摘要分群、全文索引前處理。I/O、plugin 呼叫、短小映射不要硬搬去 isolate。

✗ 只為了字串格式化就 `compute(...)`
✓ 對數千筆文章做重排序或大型 JSON parsing 時用 isolate

例外：平台 plugin 明確要求背景 isolate 並提供官方模式時，照其文件實作。

### unawaited-future-must-be-explicit · SHOULD

適用檔案：`lib/**/application/**/*.dart`、`lib/**/screens/**/*.dart`
安全自動檢查：regex（可掃描 `unawaited(`）+ manual

刻意不等待的 Future 必須用 `unawaited(...)` 標明，且只能用在 fire-and-forget 不影響當前 UI 正確性的工作。

✗ `analytics.logOpen();` 然後直接離開函式，呼叫端看不出是否故意
✓ `unawaited(analytics.logOpen());`

例外：框架 callback 本身已規定回傳 `void` 且操作同步完成。

## 尚無定論

- `FutureProvider`、`StreamProvider`、`AsyncNotifier` 都能包 async work。建議單次讀取用 `FutureProvider`，持續變化來源用 `StreamProvider`，需要重新整理、提交或複合 UI state 時再升到 `AsyncNotifier`。
- `compute` 與 `Isolate.run` 取捨以 repo 支援的 Dart SDK 為準；新專案若無舊相容壓力，優先用較直觀的 `Isolate.run`。
