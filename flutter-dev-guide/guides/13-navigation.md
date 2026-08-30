# 13 · `go_router` 導航

**這份管**：`go_router` 的集中組裝、route 參數、redirect、deep link、返回值與跨 feature 邊界。
**這份不管**：畫面狀態本身如何建模、Widget 公開 API、Material 3 樣式。

| 需要什麼 | 去哪 |
|---|---|
| 導航 trigger 怎麼存在 state 裡並清除 | `08-ui-state` |
| widget 是否可以直接持有 router | `09-widget-api` |
| `mounted`、await 後操作 context | `10-widget-state-and-lifecycle` |
| 本地化路由標題與可存取性 | `14-design-system` |

`go_router` 是 app 的導航組裝層，不是 feature 內任意可呼叫的全域捷徑。畫面之間只傳辨識資料，不傳整包模型。

---

### app-router-single-composition-root · MUST

適用檔案：`lib/core/router/**/*.dart`、`lib/app/**/*.dart`
安全自動檢查：manual

整個 app 必須有單一 `GoRouter` composition root，由 app 層組裝所有 feature routes、redirect 與 shell。feature 不得各自 new 一個 router。

✗ 每個 feature 都有自己的 `GoRouter(...)`
✓ `lib/core/router/app_router.dart` 組裝 `GoRoute` / `ShellRoute`

例外：獨立微型 demo app 或 golden test harness 可建立測試專用 router。

### route-params-are-ids-and-filters-only · MUST

適用檔案：`lib/core/router/**/*.dart`、`lib/features/**/presentation/**/*.dart`
安全自動檢查：manual

路由參數只傳 id、filter、tab、cursor 這種可序列化且穩定的值，不傳整個 article、provider state、repository object。

✗ `extra: article`
✓ `pathParameters: {'categoryId': categoryId}` 或 typed route 只帶 `articleId`

例外：真正短命、只在同一次流程內傳遞且無法序列化成穩定 id 的小結果，可用 `extra`，但要在 PR 或設計文件中說明理由。

### feature-widgets-do-not-call-go-directly · MUST_NOT

適用檔案：`lib/**/widgets/**/*.dart`、`lib/features/**/presentation/**/*.dart`
安全自動檢查：regex（可掃描 `.go(` / `.push(` / `.pop(`）+ manual

可重用 widget 與 list item widget 不得直接呼叫 `context.go` / `context.push`。它們只上拋 `onArticlePressed` 之類的意圖，由 route / screen 容器決定實際目的地。

✗ `NewsRow` 內 `context.push('/news/$id')`
✓ `NewsRow(onPressed: () => onArticlePressed(id))`

例外：最外層 route 容器 widget。

### redirect-derived-from-auth-and-app-state · MUST

適用檔案：`lib/core/router/**/*.dart`、`lib/features/**/application/**/*.dart`
安全自動檢查：manual

redirect 只能依賴可重播的 app state，例如登入狀態、allowlist 驗證結果、是否完成 onboarding。不要在 redirect 裡做網路請求或臨時副作用。

✗ `redirect` 內直接打 API 驗證 token
✓ router watch auth provider 的結果，再決定回傳哪個 path

例外：無。

### caller-owns-pop-result-handling · SHOULD

適用檔案：`lib/features/**/presentation/**/*.dart`
安全自動檢查：manual

需要等待下一頁結果時，由發起導航的 caller `await context.push<T>()` 並處理回傳值；不要讓被開啟頁面偷偷改全域 mutable state 來回傳結果。

✗ `EditCategoryPage` 存完後直接寫全域 provider，上一頁自己猜是否該刷新
✓ 上一頁 `final saved = await context.push<bool>(...)`，回來後依結果 refresh

例外：結果本來就會落入共享 repository，上一頁只需 watch 同一份資料來源即可。

### deep-link-resolves-through-route-table · MUST

適用檔案：`lib/core/router/**/*.dart`、`android/app/src/main/AndroidManifest.xml`、`ios/Runner/Info.plist`
安全自動檢查：manual

deep link 必須由 route table 與平台 URL 設定共同處理，不要在 widget 內手動解析 URI 再決定畫面。這樣 back stack、測試與平台設定才一致。

✗ 首頁 widget 啟動後自己讀 `initialUri` 再 `context.go(...)`
✓ 平台把 URI 交給 `go_router`，router 依 route table 解析

例外：從通知 payload 轉成 app 內 path 的最外層 adapter，可存在 app bootstrap，但最終仍交給 router。

## 尚無定論

- typed routing 要用 `go_router_builder` 還是手寫 route helper：兩者都可。若 route 數量多、參數複雜或常改名，建議 codegen；小型 app 手寫 helper 也夠。
- shell / nested navigation 只有在底部導覽、多分頁持久 back stack 真的存在時才引入；單一路徑 app 不要先把 router 結構做重。
