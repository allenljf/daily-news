# 18 · 埋點與事件治理

**管**：analytics adapter 的邊界、事件命名、參數治理、consent gate、埋點與主流程隔離、測試方式。  
**不管**：單一 SDK 的安裝步驟與平台後台設定；這些交給各 SDK 官方文件。

| 想找 | 去 |
|---|---|
| side effect API 與 widget lifecycle | `10-widget-state-and-lifecycle` |
| PII、token、敏感資料紅線 | `17-security` |
| 測試替身與 override 策略 | `15-testing` |
| 分層與依賴方向 | `01-architecture` |

---

## 一、先抽象，再埋點

Feature code 不直接碰 analytics SDK。所有事件都先進專案自己的 adapter，例如 `AnalyticsAdapter.log(AnalyticsEvent event)`。這樣做的價值不是「好看」，而是把四件麻煩事集中在一處：

- consent gate
- event name / parameter 正規化
- 例外吞吐與降級
- 測試替身注入

## 二、事件命名

事件名稱一律小寫 snake_case，格式優先用 `<subject>_<verb-past>` 或 `<screen>_<action>`。同一語意全專案只保留一個名字。

✗ `NewsClicked`, `clickNews`, `articleTap`  
✓ `article_opened`, `manual_refresh_requested`, `category_saved`

參數 key 也集中定義，不要在各畫面手寫 map 字串。

## 三、資料最小化與 consent

- analytics 事件不得帶 PII：email、完整搜尋關鍵字、特殊需求全文、精確 URL、token、裝置唯一識別都不行。
- 需要追蹤使用者時，用後端配發的匿名 id 或安全 hash。
- 需使用者同意的事件，在同意前完全不送；不要先快取再補傳，除非產品與法遵明確要求。

## 四、埋點不阻塞主流程

埋點是 side effect，不是業務結果。送失敗不能影響首頁刷新、Category 儲存、登入或閱讀流程。

- adapter 內要捕捉例外並吞掉，同時保留最小錯誤紀錄。
- UI 事件埋點放在 callback、notifier action 或適當的 effect 內，不要直接寫在 widget build body。
- 曝光類事件若跟 widget 狀態綁定，key 要用真正識別該次曝光的值，而不是無腦 `Unit` 風格的常數。

## 五、測試方式

- 關鍵事件以 fake analytics adapter 驗證「送了什麼」與「不該送時沒送」。
- 不用人工看 console 當唯一驗證。
- 若同一流程會依 consent、錯誤或 feature flag 分岔，這些分支都要能在測試中明確覆蓋。

## 六、規則

### analytics-via-adapter-only · MUST · new-only · manual

UI、ViewModel、Repository、UseCase 不得直接呼叫 analytics SDK；只能依賴專案自訂 adapter。

✗ `FirebaseAnalytics.instance.logEvent(...)`  
✓ `analyticsAdapter.log(AnalyticsEvent.articleOpened(...))`

例外：adapter 實作模組本身。

### analytics-contains-pii · MUST_NOT · new-only+on-touch · manual

事件與參數不得包含 PII、token、完整 URL 或自由文字大欄位。

✗ `search_performed(query: "Taipei mayor election")`  
✓ `search_performed(query_length: 21, source_count: 3)`

例外：經產品與法遵明確批准的匿名化資料欄位。

### analytics-blocks-user-flow · MUST_NOT · new-only · manual

埋點送出失敗不得阻塞導航、提交或資料刷新。

✗ `await analytics.log(...); await saveCategory();`  
✓ `analytics.log(...); await saveCategory();`

例外：非 analytics 類的法遵稽核事件。

### analytics-in-build-body · MUST_NOT · new-only+on-touch · manual

不得在 widget build body 直接呼叫埋點；重建會造成重複事件。

✗ `Widget build(...) { analytics.log(...); ... }`  
✓ 在 callback、`ref.listen`、`Future.microtask` 不需要的情況下，改用明確 effect / action

例外：無。

### analytics-without-consent-gate · MUST · new-only · manual

需要同意的 analytics 事件，必須由 adapter 的單一判斷點攔住；不可分散在各呼叫端自行 if 判斷。

✗ 每個畫面各自 `if (hasConsent)`  
✓ adapter 內統一檢查 consent state

例外：不需同意的必要營運事件，且名稱與資料集已分開。

## 七、完成前檢查

```bash
flutter analyze
flutter test
python3 flutter-dev-guide/tools/check-rules.py --staged
```

人工再看兩件事：事件命名是否一致、以及新事件是否真的不含 PII。
