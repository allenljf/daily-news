# 10 · Widget 狀態、擁有權與生命週期

**這份管**：`StatefulWidget` / `ConsumerStatefulWidget` 何時需要、controller 與 subscription 誰擁有、`initState` / `didUpdateWidget` / `dispose` / `mounted` 的正確用法。
**這份不管**：provider screen state 的形狀、導航模型、效能量測。

| 需要什麼 | 去哪 |
|---|---|
| Riverpod `AsyncValue`、一次性 effect | `08-ui-state` |
| 公開 widget API 與 callback 邊界 | `09-widget-api` |
| build 內避免做昂貴工作與 rebuild 最小化 | `11-rendering-performance` |
| Future / Stream / isolate 的取消與錯誤 | `12-async-streams-isolates` |
| router / deep link | `13-navigation` |

誰建立資源，誰就釋放資源。生命週期程式碼要集中、成對，不能把初始化和清理拆散在不同層。

---

### dispose-owned-controller-and-focusnode · MUST

適用檔案：`lib/**/widgets/**/*.dart`、`lib/**/screens/**/*.dart`
安全自動檢查：regex（可掃描建立 `TextEditingController` / `AnimationController` / `ScrollController` / `FocusNode`）+ manual

widget state 自己建立的 `TextEditingController`、`ScrollController`、`AnimationController`、`TabController`、`FocusNode`、`StreamSubscription` 都必須在 `dispose()` 中釋放。

✗ `final controller = TextEditingController();` 但沒有 `dispose`
✓ `@override void dispose() { controller.dispose(); super.dispose(); }`

例外：controller 明確由父層注入時，由父層負責釋放。

### subscribe-once-resubscribe-on-input-change · MUST

適用檔案：`lib/**/widgets/**/*.dart`
安全自動檢查：manual

依賴 widget 輸入參數的 listener / subscription 必須在 `initState` 建立、在 `didUpdateWidget` 比較舊新值後重掛、在 `dispose` 清理。不要把它藏在 `build()`。

✗ 在 `build()` 內 `stream.listen(...)`
✓ `initState` 訂閱，`didUpdateWidget` 換來源時取消重建，`dispose` 清理

例外：一次性使用 `StreamBuilder` / `FutureBuilder` 時，交給 builder 元件管理。

### no-side-effect-in-build · MUST_NOT

適用檔案：`lib/**/widgets/**/*.dart`、`lib/**/screens/**/*.dart`
安全自動檢查：manual

`build()` 不得呼叫 `showDialog`、`Navigator`、寫 provider、啟動計時器、建立 subscription 或發 HTTP request。`build()` 可能被頻繁重跑，副作用必須移到 callback、provider 或生命週期鉤子。

✗ `if (state.shouldShowError) ScaffoldMessenger.of(context).showSnackBar(...)`
✓ 由 listener / post-frame callback 在 effect state 變更時執行，並清除來源 state

例外：無。

### mounted-check-after-await · MUST

適用檔案：`lib/**/widgets/**/*.dart`
安全自動檢查：regex（可掃描 `await` 後使用 `context` / `setState`）+ manual

在 widget method 內 `await` 之後，若還要用 `context`、`setState`、`Navigator`、`ScaffoldMessenger`，必須先確認 `mounted` 仍為 true。

✗ `await repo.save(); Navigator.of(context).pop();`
✓ `await repo.save(); if (!mounted) return; Navigator.of(context).pop();`

例外：`Notifier` / provider 端不持有 `BuildContext`，因此不適用。

### ephemeral-ui-state-stays-local · SHOULD

適用檔案：`lib/**/widgets/**/*.dart`、`lib/**/screens/**/*.dart`
安全自動檢查：manual

只影響呈現、不需要跨頁或跨重建保存的狀態，應留在 widget 本地，例如展開收合、目前 tab、輸入 focus、捲動位置。不要把每個 UI 細節都抬進 Riverpod。

✗ `selectedTabProvider` 只給單一頁籤 widget 使用
✓ `DefaultTabController` 或 widget 自身 `State`

例外：該狀態同時是資料查詢輸入，或使用者離開畫面再回來仍必須保留。

### didupdatewidget-syncs-derived-local-state · MUST

適用檔案：`lib/**/widgets/**/*.dart`
安全自動檢查：manual

若本地 state 是由父層輸入衍生而來，而且輸入改變時應同步更新，必須在 `didUpdateWidget` 中明確處理；不要假設 `initState` 跑過一次就夠。

✗ `initState` 用 `widget.initialQuery` 建 controller，之後父層更新 query 但 UI 不跟著變
✓ `didUpdateWidget` 比較 `oldWidget.initialQuery` 後更新 controller

例外：本地 state 只在首次建立時消費輸入，後續刻意與父層脫鉤。

## 尚無定論

- 表單草稿要存在 widget 還是 provider：單步驟、只在一個 page 存活的表單，放 widget 較簡單；跨頁或送出前要和遠端資料合併時，再提升到 provider。
- `WidgetsBinding.instance.addPostFrameCallback` 只應用在真的需要等第一幀完成後的 UI 動作，例如 request focus 或打開 dialog；若只是想避開 `build()` 內副作用，先回頭檢查狀態建模是否錯了。
