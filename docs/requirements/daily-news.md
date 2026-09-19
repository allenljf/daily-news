# 每日新聞 App 需求與架構規格

> **狀態：Go 遷移設計待確認**
> **原始確認日：2026-08-30；遷移更新日：2026-09-01**
> **適用範圍：** `flutter-dev-guide/`、未來的 Flutter App、Go backend、PostgreSQL、Cloud Run 與 GitHub Actions。
> **後續工作入口：** 實作前必讀本文件與 `docs/tasks/daily-news.md`；task 清單建立後，以它的未完成項目為唯一執行範圍。

## 1. 目標與範圍

建立一個僅供單一使用者使用的每日新聞 App。使用者可設定多個新聞類別與來源提示；系統每天在台北時間早上 08:00 執行擷取，依類別設定搜尋相關新聞、去重並保存。Flutter App 供使用者設定類別、閱讀與篩選新聞、將新聞設為永久或刪除。

本 repo 也要提供可獨立複製的 `flutter-dev-guide/`。它採用 Flutter 官方的分層架構原則，並將 Riverpod 與 Dio 明確標示為本專案採用的成熟社群方案，而不是 Flutter 官方指定套件。

### 非目標

- 多使用者、團隊、角色權限與多租戶資料隔離。
- 任意繞過登入、付費牆、robots、平台權限或使用條款以抓取內容。
- 把 LLM 視為可保證讀取 Facebook、Instagram、Threads、YouTube 或任何指定網站的爬蟲。
- 即時推播、全文離線閱讀、書籤、已讀狀態、類別內單獨隱藏新聞。

## 2. 固定決策

| 決策 | 結論 | 理由 |
|---|---|---|
| 使用者模型 | 單一使用者 | 不建立 `user_id`、帳號管理或多租戶模型。 |
| Client 認證 | Firebase Google Sign-In + email allowlist | 手機不保存秘密；Go API 驗證 Firebase ID token，只放行設定中的 email。 |
| 後端形態 | Go `net/http` Cloud Run service + 共用程式碼的 Cloud Run Job | API 與長時間每日擷取各自有合適的執行生命週期；不用大型 web framework、ORM 或 DI container。 |
| Go backend | `net/http`、`context`、`encoding/json`、`errors`、`testing/httptest`、`database/sql` + pgx stdlib adapter、Firebase Admin Go SDK、golang-migrate | 以明確 composition root、context propagation、錯誤對應與 transaction boundary 展示 Go 能力。 |
| 排程 | GitHub Actions 於台北 08:00 觸發 Cloud Run Job；首頁可手動觸發同一 Job | 保留使用者指定的 GitHub Actions，並讓使用者能非同步要求一次更新。 |
| GitHub→GCP 身分 | GitHub OIDC + Workload Identity Federation | 不使用長效 service-account JSON key。 |
| App 架構 | Flutter UI + data layers；View/ViewModel、Repository、Service | Flutter 官方架構建議；domain/use case 只在跨 repository 或複雜且重用時加入。 |
| State / network | Riverpod / Dio | 本案選型；Riverpod 僅處理 UI state 與 composition，Dio 只存在 Remote Service。 |
| 去重 | canonical URL 優先、標準化網頁標題次之 | 符合「網頁標題重複不儲存」並降低同 URL 不同追蹤參數的重複。 |
| 刪除 | Article 全域 soft delete | 保留去重指紋，避免下一次擷取把已刪新聞重新加入。 |
| 永久保存 | Article 全域期限設定 | 相同 Article 出現在多個 Category 時，永久／刪除一致地作用於所有類別。 |
| 分頁 | cursor pagination，固定每頁最多 20 筆 | 穩定支援依最新入庫時間排序，不受資料新增影響。 |

## 3. 共同詞彙

完整詞彙見 [CONTEXT.md](../../CONTEXT.md)。本規格使用以下名稱：

- **Category**：一組使用者新聞意圖。
- **Source Setting**：Category 內一個網站／平台／未指定網站的搜尋指示。
- **Article**：全域去重後的新聞內容。
- **Category Article**：Article 與 Category 的關聯，包含其來源 tag。
- **Ingestion Run**：一次每日排程或手動觸發的批次執行。

## 4. Flutter 開發準則庫

`flutter-dev-guide/` 必須和 `android-dev-guide/` 一樣可單獨複製、可供人類與 AI agent 使用，且不綁定特定 AI 工具。預定結構：

```text
flutter-dev-guide/
├── AGENTS.md
├── README.md
├── rules.yaml
├── guides/
├── checklists/
├── skills/
├── tools/check-rules.py
└── adoption/
```

### 4.1 指南範圍

| Android 指南主題 | Flutter 對應主題 |
|---|---|
| `00-principles` | 核心原則：關注點分離、data-driven UI、SSOT、UDF、不可變與可測試性。 |
| `01-architecture` | View/ViewModel、Repository/Service、可選的 use case、依賴方向與模型轉換。 |
| `02-modularization` | feature-first 結構、何時建立 Dart package、依賴方向與避免過度拆分。 |
| `03-dependency-injection` | constructor injection、Riverpod composition root、provider override 測試。 |
| `04-domain-layer` | domain/use case 的啟用判準、純 Dart 商業規則。 |
| `05-data-layer` | Repository SSOT、錯誤模型、cache/retry/refresh policy。 |
| `06-network` | Dio、timeout、interceptor、DTO、取消與 HTTP error translation。 |
| `07-persistence` | `shared_preferences` 的限制、SQLite/Drift、migration、offline policy。 |
| `08-ui-state` | Riverpod `AsyncValue`、ViewModel state、one-off effect。 |
| `09-compose-api` | 可重用 Widget interface、callback、keys、accessibility。 |
| `10-compose-state` | Flutter widget lifecycle、state hoisting、controller/dispose、side effect。 |
| `11-compose-performance` | rebuild、list key、const、圖片與 render profile。 |
| `12-coroutines-flow` | Future、Stream、cancellation、isolate 與 async error handling。 |
| `13-navigation` | `go_router`、typed route、deep link、只傳 id／filter。 |
| `14-design-system` | Material 3、theme token、localization、dark mode、accessibility。 |
| `15-testing` | unit/widget/integration test、fake、Riverpod override。 |
| `16-performance` | app startup、DevTools、profile/release measurement。 |
| `17-security` | client secret、Firebase token、TLS、PII、URL validation。 |
| `18-analytics` | analytics adapter、consent、事件命名與 PII 禁止。 |

另須提供 `new-feature`、`new-screen`、`new-api`、`refactor` 四份 checklist，兩個 Flutter 專用 skill（實作導引與變更稽核），以及新／既有專案和 AI 工具整合的 adoption 文件。`rules.yaml` 為規則唯一真相；`tools/check-rules.py` 僅稽核可以由單檔 Dart、`pubspec.yaml` 或分析設定安全判定的規則。

### 4.2 技術基線

| 領域 | 採用 | 定位 |
|---|---|---|
| 架構 | Flutter View/ViewModel + Repository/Service | Flutter 官方建議。 |
| UI | Flutter Material 3 | Flutter 官方。 |
| State / composition | `flutter_riverpod` | 社群主流、本案選型。 |
| Network | `dio` | 社群主流、本案選型。 |
| Navigation | `go_router`、必要時 `go_router_builder` | Flutter team 維護。 |
| JSON | `json_serializable`、`json_annotation`、`build_runner` | Dart/Google 發佈。 |
| 小型設定 | `shared_preferences` 的 async API | Flutter team 發佈；不用來存新聞或秘密。 |
| 離線關聯資料 | SQLite；複雜 reactive query 時採 Drift | Flutter 官方示範 SQLite；Drift 為社群選項。 |
| Logging | `package:logging` | Dart 官方。 |
| Crash reporting | `firebase_crashlytics` | Google/Firebase 官方。 |
| 靜態檢查 | `flutter_lints`、`flutter analyze` | Flutter 官方。 |
| 測試 | `flutter_test`、`integration_test` | Flutter 官方。 |

官方與套件依據詳見[研究筆記](../research/2026-08-29-modern-flutter-architecture.md)。

## 5. 使用者體驗

### 5.1 首頁

1. 初始沒有 Category 時，首頁只顯示新增 Category 的方塊與 `+`。
2. 有 Category 時，首頁以方塊顯示每個 Category；最後一格固定為新增方塊。
3. 點擊 Category 方塊開啟該 Category 的新聞列表。
4. 點擊新增方塊開啟設定 Bottom Sheet。
5. 首頁右上角顯示最近一次**成功完成** Ingestion Run 的資料更新時間；尚無成功 Run 時顯示「尚未更新」。
6. 首頁右上角提供「立即更新」按鈕。點擊後開啟確認 dialog，說明它會排入與每日排程相同的後端背景擷取工作、所需時間依類別與來源數量而定、可離開 App、完成前畫面不會立刻出現新資料。
7. dialog 確認後送出手動 Run 請求並立即關閉 dialog；首頁顯示該 Run 為排隊中或執行中，按鈕暫時停用。完成且成功後刷新更新時間與 Category／新聞資料；失敗時顯示可重試的錯誤狀態。
8. 首頁另提供「刷新頁面」按鈕；它不建立 Ingestion Run，只重新呼叫首頁所需的 API 以取得最新更新狀態與 Category 資料。即使手動 Run 正在排隊或執行中，使用者仍可使用此按鈕重新讀取資料。

### 5.2 Category 設定 Bottom Sheet

由上到下包含：

1. **新聞類別**文字輸入；必填。
2. **搜尋關鍵字**文字輸入；可留空，留空時以 Category 名稱作為查詢語意。
3. **搜尋網站**：初始顯示一個可編輯的「未指定網站」文字欄位，代表一般公開網頁搜尋；旁邊有新增按鈕。每次點擊新增按鈕，開啟 dialog 讓使用者輸入網站名稱或 URL；確認後新增一個可編輯的 Source Setting 欄位。空白欄位不保存。
4. **其他特殊需求**多行文字輸入；可留空。
5. **內容語言**選擇器；可選「繁體中文」（`zh-Hant`）或「英文」（`en`），新增與既有 Category 預設皆為繁體中文。
6. **儲存設定**按鈕；送出 `POST /v1/categories` 或既有 Category 的 `PATCH`。

輸入錯誤在欄位旁顯示；儲存中不可重複提交；成功後關閉 Bottom Sheet 並刷新首頁。Category 的編輯放在其設定入口；刪除放在 Category 新聞列表右上角，須先顯示確認 dialog。刪除會停用 Category 的 Source Setting 與後續擷取，並移除其 Category Article 關聯；不刪除仍被其他 Category 引用的 Article。

### 5.3 Category 詳情與新聞詳情

- 類別詳情依 `Category Article.inserted_at` 由新到舊顯示。
- 提供來源 tag filter；tag 對應 Source Setting，不以使用者可變更的自由文字做查詢鍵。
- 每次讀取 20 筆，以 cursor 繼續載入。
- 每筆列出標題、來源 tag、入庫時間、原始出處與期限狀態。
- 詳情以 App 內建 WebView 載入 Article 的 `canonicalUrl`，直接顯示原文網頁內容，不再只顯示一行連結文字；摘要、來源與 citation 仍由詳情 API 提供並顯示於原文之前。
- WebView 只載入通過驗證的 `https` URL 且 host 非空；非 HTTPS 或畸形 URL 不載入，改顯示明確的錯誤狀態。WebView 不啟用檔案存取或通用 bridge，頁內導覽僅允許 `http`／`https`。
- 永久為可切換狀態：第一次操作把 Article 設成永久；再次操作取消永久並恢復其原本時效。
- 使用者可全域刪除 Article；刪除成功後直接返回 Category 新聞列表，且該 Article 不再出現於列表中。

## 6. 資料與生命週期

### 6.1 關聯模型

```text
Category 1 ─── * SourceSetting
Category * ─── * Article  (透過 CategoryArticle；記錄 source_setting_id 與 inserted_at)
IngestionRun 1 ─── * IngestionAttempt
```

最少資料欄位：

| 實體 | 關鍵欄位 |
|---|---|
| `categories` | id、name、search_keywords、special_requirements、content_language（`zh-Hant` 或 `en`，預設 `zh-Hant`）、created_at、updated_at、deleted_at |
| `source_settings` | id、category_id、label、website_input、normalized_host、kind、position、created_at、deleted_at |
| `articles` | id、title、normalized_title_hash、canonical_url、canonical_url_hash、summary、published_at、first_seen_at、expires_at、deleted_at |
| `category_articles` | category_id、article_id、source_setting_id、inserted_at |
| `ingestion_runs` | id、trigger（scheduled/manual）、idempotency_key、taipei_date、started_at、finished_at、status、candidate_count、inserted_count、duplicate_count、error_count |
| `ingestion_attempts` | id、run_id、category_id、source_setting_id、status、candidate_count、inserted_count、duplicate_count、error_summary |

`articles` 對 `canonical_url_hash` 建唯一索引；`normalized_title_hash` 建索引並作 URL 不同時的第二層去重。`category_articles` 對 `(category_id, article_id)` 建唯一索引；列表索引以 `(category_id, inserted_at DESC, article_id DESC)` 支援 cursor。

### 6.2 到期、永久與刪除

- 新 Article：`expires_at = first_seen_at + 30 days`。
- 永久：`expires_at = null`。
- 非永久 Article 到期後不出現在預設列表，但資料仍保留。
- 刪除：寫入 `deleted_at`，所有 Category 都不再顯示，日後 ingestion 因唯一指紋而略過。
- Category 刪除：soft delete Category 與其 Source Setting／Category Article 關聯；Article 僅在沒有其他有效關聯時保留為不顯示的歷史資料。

## 7. HTTP interface

所有端點前綴為 `/v1`，要求 `Authorization: Bearer <Firebase ID token>`。checked-in 的 [`docs/contracts/daily-news.openapi.json`](../contracts/daily-news.openapi.json) 是 HTTP/OpenAPI contract 的唯一真相；Go handler DTO 與資料庫 row model 不得成為彼此的 contract，Flutter 不需變更 API。

| Method | Path | 用途 |
|---|---|---|
| `GET` | `/categories` | 首頁 Category 方塊。 |
| `POST` | `/categories` | 建立 Category 與 Source Settings。 |
| `PATCH` | `/categories/{categoryId}` | 更新 Category 設定。 |
| `DELETE` | `/categories/{categoryId}` | 刪除 Category。 |
| `GET` | `/categories/{categoryId}/news?cursor=&limit=20&sourceTagId=` | 類別新聞列表。 |
| `GET` | `/categories/{categoryId}/news/{newsId}` | 新聞詳情與類別來源 tag。 |
| `PATCH` | `/news/{newsId}` | 設定永久或恢復預設期限。 |
| `DELETE` | `/news/{newsId}` | 全域 soft delete Article。 |
| `GET` | `/ingestion-runs/latest` | 首頁讀取最近成功更新時間與目前背景工作狀態。 |
| `POST` | `/ingestion-runs` | 非同步觸發手動更新；回傳 202 與新建或既有進行中 Run。 |

`limit` 預設與最大值均為 20；cursor 是不透明、具排序資訊的值。列表回應至少包含 `items` 與可為空的 `next_cursor`。錯誤回應採 `application/problem+json`：未登入為 401、email 不在 allowlist 為 403、找不到為 404、欄位驗證為 422、cursor 無效為 400、衝突為 409。

## 8. 每日擷取與來源規則

### 8.1 排程與 idempotency

GitHub Actions 預定以 UTC 00:00 觸發，目標為台北時間 08:00。GitHub schedule 可能延遲，因此 scheduled Run 以 `scheduled:<taipei_date>` 作為 idempotency key，而非假設準時啟動。workflow 透過 OIDC/WIF 取得短期 GCP 身分，觸發 Cloud Run Job。

Flutter 的「立即更新」送出 manual Run 請求。Go API 建立唯一的 `manual:<request-id>` idempotency key，再以具備最小 `run.jobs.run` 權限的 service account 啟動同一 Cloud Run Job。manual Run 可在 scheduled Run 成功後的同一天再執行，讓使用者取得較新的資料；但系統全域同時只能有一個 `queued` 或 `running` Run。若已有進行中的 Run，API 回傳該 Run 而不建立第二個工作。這保護 LLM／來源額度並避免同時寫入造成競態。

Job 依序讀取每個有效的 Category 與 Source Setting；每一組至多保留 10 個候選。每個來源的寫入互不影響：單一來源失敗只留下 Attempt 與 error summary，絕不清除既有 Article。每個 Run 記錄候選、寫入、重複與失敗數。

### 8.2 搜尋與 adapter

Category、搜尋關鍵字、來源提示、特殊需求與內容語言會組成 prompt。此版本接受繁體中文（`zh-Hant`）或英文（`en`）；LLM 的輸出只是候選與摘要輔助，寫入前必須有可驗證 URL、來源資訊與去重檢查。每個 Article 保存原文 URL、canonical URL、來源與可用 citation。

一般公開網站的最小 HTTP(S) Source Setting 現在可填首頁 URL。Job 以單一 composite adapter 依序嘗試（見 [`Web Source Adapter Design`](../superpowers/specs/2026-09-16-web-source-adapters-design.md)）：

1. 直接解析 RSS/Atom。
2. 若回應是 HTML，只讀取 `rel` 含 `alternate` 且 type 為 RSS/Atom 的 `<link>`，將相對 `href` 解析為絕對 URL，要求公開 HTTP(S) 且位於設定網站的 host 或其 subdomain，並取得第一個合格 feed。
3. 若沒有合格 feed，或 feed 取得／解析失敗，呼叫 SerpApi 的 Google News 搜尋（`engine=google_news`）。設定網站時以 `site:<host>` 限制在該 host，未指定網站時搜尋整個網路；以 `when:<window>` 限制近期範圍。任何 fallback 都不會停用單一來源失敗隔離、30 天到期、全域 soft delete 或 URL 優先、標題次之去重。

寫入前另有一個全域 blocklist：任何 adapter（RSS/Atom、feed discovery、SerpApi）回傳的 `youtube.com`／`youtu.be` 候選都會被丟棄；`youtube.com`／`youtu.be` 的 Source Setting 也在 router 層直接回報未支援，不會退回全 YouTube 關鍵字搜尋。

| 來源類型 | 首選方式 | 限制 |
|---|---|---|
| 一般公開新聞站／官方部落格 | 直接 RSS/Atom → 從首頁 HTML `<link rel="alternate">` 發現同 host/subdomain 的 RSS/Atom → SerpApi Google News 搜尋 | 只處理公開可索引、可直讀頁；自動發現的 feed 必須是公開 HTTP(S) 且限於設定網站的 host/subdomain；設定網站時搜尋以 `site:` 限制、未指定網站時為全網搜尋；可用 `SERPAPI_WHEN` 控制近期範圍；不得假定收錄、即時性或全文可得。 |
| YouTube | 不支援：`youtube.com`／`youtu.be` 的 Source Setting 產生明確的 failed Attempt，且所有 adapter 輸出在寫入前都會濾除 YouTube URL | 不呼叫 YouTube Data API；不保存任何 YouTube 影片連結；不因使用者填入頻道 URL 而改做全 YouTube 關鍵字搜尋。 |
| GitHub | GitHub REST API（release/event 等） | 未授權公開請求有限流；private repo 需要適當授權。 |
| Facebook／Instagram／Threads | 未來 adapter，僅限指定 Page、已授權 Professional 帳號／hashtag，或官方 Threads API | 目前不支援任意全文關鍵字搜尋；不得以 SerpApi 或其他一般搜尋取代，也不得當成可讀取任意 Meta 社群內容的 contract。 |

首頁與 feed 取得共用一個受控 HTTP client：具 request timeout、2 MiB response size limit、redirect cap，並在連線前拒絕 loopback、link-local、private 及其他非公開位址。任何 adapter 都不得在 error summary 記錄 API key、access token、prompt、頁面全文或 Firebase token。SerpApi 的 `SERPAPI_WHEN` 預設限制近期結果，以降低舊資料被收錄的機會。

## 9. 安全、秘密與設定

### 9.1 不可放入 Flutter 或版本控制的值

- LLM key、PostgreSQL 密碼、GitHub PAT、Meta access token、Firebase Admin service-account key。
- 任一用來代表 server 身分的 bearer token。

Cloud Run runtime 自 Secret Manager 取得必要秘密。GitHub Actions 不持有 service-account JSON；其必要設定使用 GitHub Variables：`GCP_PROJECT_ID`、`GCP_REGION`、`GCP_WORKLOAD_IDENTITY_PROVIDER`、`GCP_SERVICE_ACCOUNT`、`CLOUD_RUN_JOB_NAME`。

### 9.2 條件式秘密

| 名稱 | 需要時機 | 建議位置 |
|---|---|---|
| `SERPAPI_API_KEY` | 啟用一般網站／未指定網站的 Google News 搜尋 fallback 時 | Secret Manager。 |
| `SERPAPI_WHEN` | 選用 | Job 環境設定，非秘密；控制搜尋近期範圍，有預設值（`7d`）。 |
| `DB_PASSWORD` | PostgreSQL password authentication | Secret Manager。 |
| `GITHUB_NEWS_TOKEN` | 需要較高 GitHub API rate limit 或 private 資源時 | Secret Manager。 |
| Meta platform access token | 啟用 Facebook／Instagram／Threads 且已授權的未來 Meta adapter 時 | Secret Manager。 |

Facebook／Instagram／Threads adapter 不進行任意全文關鍵字搜尋，且不得以 SerpApi 或 LLM 代理讀取；未配置憑證或未取得授權時只讓該 Source Setting 產生 failed Attempt。

Firebase client configuration 是可公開的 client 設定，不是 server secret；Firebase Admin 在 Cloud Run 以 Application Default Credentials 運作。後端以環境設定 `ALLOWED_USER_EMAIL` 進行 allowlist 比對，log 不得輸出完整 ID token、prompt、秘密或個人資料。

## 10. 品質與驗收

### 10.1 Flutter

- `flutter analyze` 零錯誤。
- Repository、mapper 與 Riverpod controller 有 unit tests，且透過 provider override 注入 fake。
- Bottom Sheet、空／有 Category 首頁、新聞列表的 loading/empty/error/success 有 widget tests。
- Google Sign-In、儲存 Category 與讀取第一頁新聞有 integration tests。

### 10.2 Go 與 PostgreSQL

- checked-in OpenAPI artifact 必須鎖定 `/v1` paths、JSON schema、status codes 與 `application/problem+json`；Flutter API client 不需改動。
- migration 可由空資料庫升級，並驗證必要 index／unique constraint。
- 測試 Firebase token 缺失、無效與 email 不符時分別拒絕。
- 測試 Category CRUD、cursor、來源 tag filter、20 筆 page size、Article 全域永久與 soft delete。
- 測試 canonical URL／標題去重、30 天到期、scheduled 台北日期 idempotency、manual Run 合併進行中工作、單一來源失敗不影響其他來源。
- Go 驗證至少執行 `gofmt`、`go vet ./...` 與 `go test ./...`；HTTP handler test 使用 `httptest`，PostgreSQL integration test 沿用 schema assertions。

### 10.3 部署與操作

- Cloud Run service 只接受通過 Firebase 驗證與 allowlist 的 client request。
- Job 具有最小權限的 service account；workflow 透過 WIF 啟動 Job。
- 每次 Run 可從結構化 log 與資料庫讀回 run id、trigger、來源、候選／重複／寫入／失敗統計。
- 排程 smoke test 驗證 workflow 可取得 OIDC 身分、觸發 Job 並留下 Run 記錄。

## 11. 後續執行協議

待 `docs/tasks/daily-news.md` 建立後，每次新對話必須：

1. 讀取 `AGENTS.md`、本文件與 task 清單。
2. 選擇未勾選且所有依賴已完成的 task；不自行擴大範圍。
3. 同一階段只放相互相關的 task；只有標示為可平行者可同時進行。
4. 每個 task 完成後執行指定驗證、勾選 checkbox，並在 task 下記錄結果與實際新增的依賴。
5. 規格變更先修改本文件與 task 清單，再開始實作。

## 12. 參考來源

- [Flutter app architecture guide](https://docs.flutter.dev/app-architecture/guide)
- [Flutter architecture recommendations](https://docs.flutter.dev/app-architecture/recommendations)
- [go_router](https://pub.dev/packages/go_router)
- [json_serializable](https://pub.dev/packages/json_serializable)
- [Flutter SQLite recipe](https://docs.flutter.dev/cookbook/persistence/sqlite)
- [Firebase Crashlytics for Flutter](https://firebase.google.com/docs/crashlytics/flutter/get-started)
- [Cloud Run Go service quickstart](https://cloud.google.com/run/docs/quickstarts/build-and-deploy/deploy-go-service)
- [GitHub OIDC/WIF for deployment pipelines](https://cloud.google.com/iam/docs/workload-identity-federation-with-deployment-pipelines)
- [Google Custom Search JSON API](https://developers.google.com/custom-search/v1/overview)
- [SerpApi Google News](https://serpapi.com/google-news-api)
- [完整官方來源研究](../research/2026-08-29-modern-flutter-architecture.md)
