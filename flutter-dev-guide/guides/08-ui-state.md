# 08 · UI 狀態與 Riverpod AsyncValue

**這份管**：畫面狀態的形狀、`AsyncValue` 該怎麼落到 widget、Riverpod `Notifier` / `AsyncNotifier` 的責任邊界、一次性效果（snackbar、dialog、navigation trigger）怎麼建模。
**這份不管**：Widget 建構子的公開 API、`StatefulWidget` 生命週期、效能最佳化、`go_router` route 結構。

| 需要什麼 | 去哪 |
|---|---|
| Widget 建構子、`super.key`、callback 型別、語意標籤 | `09-widget-api` |
| `initState` / `dispose` / `mounted` / controller 擁有權 | `10-widget-state-and-lifecycle` |
| rebuild 範圍、list key、`const`、長清單 | `11-rendering-performance` |
| Future / Stream / isolate / async error | `12-async-streams-isolates` |
| 導航狀態怎麼觸發與消費 | `13-navigation` |
| 文案、本地化、Material 3 與無障礙 | `14-design-system` |

畫面狀態的單一真相通常在 Riverpod provider。Widget 只負責 render 與回傳使用者意圖，不直接組裝業務狀態。

---

### asyncvalue-exhaustive-rendering · MUST

適用檔案：`lib/features/**/presentation/**/*.dart`、`lib/**/screens/**/*.dart`
安全自動檢查：manual（需確認 `AsyncValue` 的 loading/error/data 是否都被處理）

任何直接 render `AsyncValue<T>` 的畫面都必須明確處理 loading、error、data 三種分支；`data` 分支若可能為空集合，還要再明確處理 empty。

✗ `ref.watch(newsProvider).valueOrNull ?? []`
✓ `ref.watch(newsProvider).when(loading: ..., error: ..., data: ...)`

例外：最外層已把 `AsyncValue<T>` 轉成 sealed screen state，且下層 widget 不再直接接觸 `AsyncValue`。

### provider-state-owned-by-notifier · MUST

適用檔案：`lib/features/**/application/**/*.dart`、`lib/features/**/presentation/**/*.dart`
安全自動檢查：regex（可掃描 widget 內對 provider state 的直接賦值）+ manual

Riverpod provider 的可變狀態只能由對應的 `Notifier` / `AsyncNotifier` / `StateNotifier` 更新。Widget 不得持有第二份可變畫面資料，也不得直接改 provider 的內部欄位。

✗ 在 widget 內 `items.add(...)` 後期待 provider 自動同步
✓ `ref.read(newsControllerProvider.notifier).refresh()`

例外：純 UI 元素狀態（tab index、sheet 開關、捲動位置）可留在 widget 端，見 `10-widget-state-and-lifecycle`。

### one-off-effect-as-state · MUST

適用檔案：`lib/features/**/application/**/*.dart`、`lib/features/**/presentation/**/*.dart`
安全自動檢查：manual（需看事件是否能在 rebuild / 重進頁面後重放或遺失）

snackbar、dialog、導向下一頁、顯示 toast 等一次性效果，必須以可清除的 state 表達，並由 widget 在消費後呼叫 notifier 清除；不要把它做成 UI 自行猜測的副作用。

✗ `ref.listen` 收到布林後直接導航，但沒有清除來源 state
✓ `state.pendingMessage` / `state.pendingRoute` + `messageShown()` / `navigationHandled()`

例外：來源本來就在 UI 的即發即忘動作，例如按下按鈕後立刻開外部瀏覽器。

### asyncvalue-not-mixed-with-view-flags · SHOULD

適用檔案：`lib/features/**/application/**/*.dart`
安全自動檢查：manual

`AsyncValue<T>` 負責遠端資料生命週期；畫面上的暫態旗標（`isRefreshing`、`selectedFilterId`、`isSubmitting`）應放在外層 immutable screen state，而不是把所有東西都塞進 `AsyncValue.data(...)` 的 payload 裡。

✗ `AsyncValue.data(NewsState(items, isRefreshing: true))` 又在別處另外追 `selectedTag`
✓ `NewsScreenState(feed: AsyncValue<List<Article>>, selectedTagId: ..., isRefreshing: ...)`

例外：若 provider 只服務單一簡單 widget，且額外旗標只有一個，維持單一 `AsyncValue<T>` 可接受。

### family-provider-for-route-input · MUST

適用檔案：`lib/features/**/application/**/*.dart`、`lib/core/router/**/*.dart`
安全自動檢查：regex（可掃描 `.family` / codegen family 使用）+ manual

來自 route 的 id、filter、cursor 等輸入必須成為 provider 參數，而不是先讀全域 mutable state 再在 provider 內分支。這讓畫面可重建、可測試，也避免不同頁面互相污染。

✗ `selectedCategoryIdProvider` 被多個詳情頁共用
✓ `categoryNewsProvider(categoryId: ..., sourceTagId: ...)`

例外：真正的 app-wide session state，例如目前登入身份。

### provider-error-mapped-before-widget · SHOULD

適用檔案：`lib/features/**/application/**/*.dart`、`lib/features/**/data/**/*.dart`
安全自動檢查：manual

Widget 不得直接理解 `DioException`、repository exception 或原始 stack trace。provider 層要先把錯誤轉成畫面能決策的型別或文案 key，UI 只處理顯示。

✗ `error.toString()` 直接丟給 `Text`
✓ `NewsLoadFailure.offline`、`NewsMessageKey.retry`

例外：debug-only 工具頁可暫時顯示原始錯誤，但不得流入正式流程。

## 尚無定論

- `AsyncNotifier` 與 `Notifier<AsyncValue<T>>` 兩種寫法都可行。建議預設用 `AsyncNotifier<T>` 處理單一遠端主資源；同時還要管理篩選、編輯草稿或多塊 UI state 時，再包一層自訂 immutable screen state。
- `ref.listen` 可用來銜接 effect，但前提仍是 effect 來源本身可回到 state 並能被清除；不要把 `listen` 當成 side-effect event bus。
