# 17 · 安全與資料保護

**管**：Flutter client 不能持有什麼秘密、Firebase ID token 的處理方式、PII 與 log 紅線、URL 與外部輸入驗證、傳輸層與第三方 SDK 的基本防線。  
**不管**：後端 service account、Cloud Run、Secret Manager 的部署細節（那是 repo 根規格與 backend 文件的範圍）。

| 想找 | 去 |
|---|---|
| Dio / HTTP 層的責任、interceptor、錯誤轉換 | `06-network` |
| analytics 事件命名與 consent gate | `18-analytics` |
| local persistence 的一般策略 | `07-persistence` |

---

## 一、先講最重要的：Flutter app 不是保險箱

任何放進 client 的值，都要假設使用者能取出：`--dart-define`、原生字串、資源檔、混淆後常數都一樣。這代表：

- LLM key、資料庫密碼、GitHub token、Meta token、Firebase Admin key 不能進 app。
- Firebase client config 可公開，但 Firebase ID token 不可當成無害字串；它是使用者憑證。
- 高價值權限一律留在後端，由 FastAPI 驗證 token 與 allowlist 後代呼叫。

## 二、Firebase ID token 的處理

- ID token 只存在記憶體或由官方 auth SDK 受控刷新，不要自行落地到 `shared_preferences`、檔案或 debug log。
- 任何 network log、error log、crash breadcrumb 都不得包含完整 `Authorization` header 或 token body。
- token 過期、撤銷、email 不符時，client 只能把它當成 session 問題處理，不得自行推論授權結果。

## 三、PII 與 log

PII 在這份指南裡包含：email、電話、姓名、地址、精確座標、裝置唯一識別、完整新聞特殊需求文字，以及任何可直接還原使用者身份的內容。

記錄資料時只保留必要的最小資訊：

- 可以記 `categoryId`、`articleId`、匿名 user hash。
- 不可以記 email、完整 token、搜尋關鍵字原文、特殊需求全文、完整 request / response body。
- debug log 也算 log；「只在本機看」不是例外。

## 四、URL 與外部輸入驗證

Daily News 會處理外部文章 URL、deep link、分享資料與後端回傳欄位。這些值在進 UI 前都必須先驗證：

- 只接受預期的 `https` URL。
- 開外部文章前先檢查 scheme / host；未知或畸形 URL 直接拒絕。
- 不要把後端回傳字串直接拿去當 WebView、router path 或檔案路徑。

若真的需要 WebView，預設走最小權限：不要開不必要的 JavaScript、檔案存取或通用 bridge。

## 五、第三方 SDK 與最小資料原則

- 引入 SDK 前先問：它需要哪些資料？是否能延後初始化？是否能停用個資欄位？
- analytics、crash reporting、feature flag 都應經專案自己的 adapter，不直接把 SDK 型別散到 feature code。
- release build 不得留下 verbose network log、測試 token 或 debug endpoint。

## 六、規則

### secret-in-client-code · MUST_NOT · new-only+on-touch · regex

任何真正的 secret 或代表 server 身分的 token，不得出現在 Flutter、Dart、原生設定或版本控制中。

✗ `const geminiApiKey = '...'`  
✓ 由 backend 代呼叫，client 只拿自己的 Firebase session

例外：公開且設計上可暴露的 client id。

### firebase-token-logged-or-persisted · MUST_NOT · new-only+on-touch · manual

Firebase ID token 不得輸出到 log，也不得自行持久化到 `shared_preferences`、檔案或可讀取快取。

✗ `logger.info('token=$token')`  
✓ token 只交給 auth / network adapter 使用

例外：無。

### pii-in-log-or-error-surface · MUST_NOT · new-only+on-touch · manual

PII 不得出現在 log、toast、snackbar、crash breadcrumb 或 debug overlay。

✗ `SnackBar(content: Text('Login failed for $email'))`  
✓ `SnackBar(content: Text('登入失敗，請重新嘗試'))`

例外：無。

### unvalidated-external-url · MUST · new-only+on-touch · manual

來自 API、deep link、剪貼簿或分享的 URL，使用前必須驗證 scheme、host 與格式。

✗ `launchUrl(Uri.parse(article.url))`  
✓ 先驗證 `uri.isScheme('https')` 與允許的 host 條件

例外：無。

### sdk-type-bypasses-security-adapter · SHOULD_NOT · new-only · manual

feature code 不應直接依賴 auth、analytics、crash SDK 型別；要經過 app 自己的 adapter，讓敏感資料規則只在一處管控。

✗ ViewModel 直接呼叫 Firebase / analytics SDK  
✓ `AuthRepository`、`AnalyticsAdapter`、`CrashReporter`

例外：底層 adapter 實作模組本身。

## 七、完成前檢查

```bash
flutter analyze
flutter test
python3 flutter-dev-guide/tools/check-rules.py --staged
```

另外人工確認三件事：

1. diff 中沒有新增任何真實 secret 或 token。
2. 新 log / error UI 沒有帶出 PII。
3. 新增的外部 URL 都有經過驗證與安全開啟流程。
