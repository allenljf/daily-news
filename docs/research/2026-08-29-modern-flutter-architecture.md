# 現代 Flutter App Architecture 與每日新聞後端研究

> 查核日：2026-08-29。本文只引用來源擁有者的一手文件：Flutter／Dart／Google
> 文件、Flutter/Dart 官方 GitHub，以及各套件的官方 pub.dev 或官方 repository。
> 「主流」是依套件官方 pub.dev 的使用訊號與維護狀態所作的實務分類，**不是**
> Flutter 團隊背書；沒有找到官方明示推薦時一律如此標記。

## 結論摘要

對新 Flutter app，最可辯護的基線是 Flutter 官方的 **UI + data layers**：UI 採
View/ViewModel（MVVM），data 採 Repository + Service；Repository 是資料真相來源，
負責快取、錯誤、重試與刷新。Domain/use-case 是「複雜或重複邏輯」才加的選擇層，
不是每個 feature 必備的第三層。官方強烈推薦分層、Repository 與 ViewModel；這是
本案應遵循的架構決策，而非社群 Clean Architecture 版本的猜測。
[Flutter architecture guide](https://docs.flutter.dev/app-architecture/guide)／
[recommendations](https://docs.flutter.dev/app-architecture/recommendations)

Riverpod、Dio、Freezed、Drift、GetIt 都是成熟且常見的社群選項，但截至本查核日，
**Flutter 官方沒有把 Riverpod 或 Dio 列為架構推薦套件**。若產品決定採用它們，建議
採「官方架構邊界 + 社群實作」：Riverpod 只負責 UI state、生命週期、composition 與
覆寫測試；Dio 放在 stateless Service；Repository 仍是 cache／error policy 的唯一
所有者。不要讓 provider 或 HTTP client 直接成為 UI 的資料層。

```text
View (Widget) ──watch/render──> ViewModel / feature controller
                                      │ constructor / provider injection
                                      v
                               Repository (SSOT, cache, error policy)
                                  │                 │
                                  v                 v
                         Remote Service (Dio)   Local Service (DB/prefs)
```

此圖的元件職責與依賴方向來自 Flutter 的架構指南；Riverpod/Dio 是下文選擇的實作方式，
不是官方圖中指定的技術。

## 標示規則

| 標示 | 意義 |
|---|---|
| **官方明示** | Flutter/Dart/Google 文件直接推薦、或 Flutter/Dart/Google 所發佈／維護的套件。 |
| **官方範例使用** | 官方教學或 architecture case study 使用，但不等同全域推薦。 |
| **社群主流** | 套件自身的一手文件顯示為成熟、常用的選擇；無 Flutter 官方背書聲明。 |

## 1. 官方架構：採什麼、不要過度解讀什麼

### 1.1 Layered MVVM 是目前 Flutter 文件的明示建議

| 層／元件 | 應負責 | 邊界與決策 | 地位與來源 |
|---|---|---|---|
| UI：View + ViewModel | Widget 呈現和輸入；ViewModel 將 application data 轉 UI state，公開 command | View 不放資料／商業邏輯；一個 ViewModel 對一個 View（可為一組 widgets） | **官方明示**：[guide](https://docs.flutter.dev/app-architecture/guide)、[recommendations](https://docs.flutter.dev/app-architecture/recommendations) |
| Data：Repository | domain model 的 source of truth；快取、error handling、retry、refresh、polling | ViewModel 只依賴 Repository；Repository 不認識彼此 | **官方明示**：[guide—repositories](https://docs.flutter.dev/app-architecture/guide#repositories) |
| Data：Service | 封裝一種外部／平台資料來源，回傳 `Future`／`Stream`，且不持 state | API、檔案、local DB 各以 service 隔離 | **官方明示**：[guide—services](https://docs.flutter.dev/app-architecture/guide#services) |
| Domain/use-case | 跨 repository 或會在 ViewModel 重複的複雜商業邏輯 | 小型／一般 app 不必為形式而建立；複雜、可重用時加入 | **官方明示為 conditional**：[recommendations](https://docs.flutter.dev/app-architecture/recommendations) |
| 資料流 | immutable data 與單向資料流 | UI event → ViewModel command → Repository → state update → UI render | **官方明示**：[architecture concepts](https://docs.flutter.dev/app-architecture/concepts) |

實作上保持 feature-first 資料夾也可以；官方重點是責任、依賴與可測試輸入輸出，並未規定
必須使用 `core/`、每 feature 一個 package，或「永遠三層」。官方 case study 的溝通規則是
View 只知道一個 ViewModel、ViewModel constructor 接收 Repository、Repository constructor
接收 Service。[Communicating between layers](https://docs.flutter.dev/app-architecture/case-study/dependency-injection)

### 1.2 Riverpod：可採用，但不要宣稱 Flutter 官方推薦

| 判定 | 證據與建議 |
|---|---|
| **社群主流，非官方背書** | [`flutter_riverpod`](https://pub.dev/packages/flutter_riverpod) 由 `dash-overflow.net` 發佈，定位為 reactive caching/data-binding，支援 loading/error 與 UI/logic 分離；其一手 [testing guide](https://riverpod.dev/docs/how_to/testing) 支援以 `ProviderContainer` 隔離測試。Flutter 架構文件列為 conditional 的是 SDK `ChangeNotifier`／`Listenable`，並說 state-management 是選擇問題，沒有點名 Riverpod。[Flutter recommendation](https://docs.flutter.dev/app-architecture/recommendations) |
| **建議用法** | `ProviderScope` 放 app root；feature provider/controller 僅協調 ViewModel state、觀察 repository stream、處理 `AsyncValue`；repository 與 service 以 constructor 建構，provider 只在 composition root 組裝。測試時以 provider override 注入 fake repository，不 mock notifier。這保留官方架構的可測試邊界。 |
| **避免** | 不以 `ref.read` 讓 Widget 直接呼叫 Dio／DB；不把 provider 當全域 mutable singleton；不因選 Riverpod 而取消 Repository。Riverpod 可做 DI，但這是技術能力，非 Flutter 對 Service Locator 的官方轉向。 |

### 1.3 Dio：採用條件與邊界

| 判定 | 證據與建議 |
|---|---|
| **社群主流，非官方背書** | [`dio`](https://pub.dev/packages/dio) 的官方頁列出 global config、interceptors、FormData、cancel、timeout、adapter 與 transformer；維護 repo 是 [cfug/dio](https://github.com/cfug/dio)。Flutter 的入門 HTTP 路徑使用標準 [`http`](https://docs.flutter.dev/learn/pathway/tutorial/http-requests)，官方 architecture recommendation 並未指名 Dio。 |
| **何時選 Dio** | 需要共用 auth／observability interceptor、per-request cancellation、上傳下載 progress、多 API host 或自訂 adapter 時，作為 Remote Service 的 injected client 合理。建立一個配置完成的 `Dio` 實例，將 DTO decoding、HTTP exception 轉換侷限在 service/repository 邊界。 |
| **何時不要額外引入** | 單一 JSON API、無上述需求時，SDK `http` 更小；不應為「官方推薦」而選 Dio，因為沒有這項背書。快取、retry 與對上層的 domain error 仍屬 Repository 的職責。[Flutter repository responsibility](https://docs.flutter.dev/app-architecture/guide#repositories) |

## 2. 套件選型盤點

| 領域 | 建議基線 | 身分／官方證據 | 替代與限制 |
|---|---|---|---|
| Routing | `go_router`；需要 deep link、redirect、nested navigator 時優先 | **官方明示／Flutter team 維護**：Flutter recommendation 說它是 90% app 的 preferred way；[pub page](https://pub.dev/packages/go_router) 發佈者為 `flutter.dev`，且 [flutter/packages](https://github.com/flutter/packages) 說其為 core team first-party packages。 | 簡單／特殊流程可用 framework [`Navigator`](https://docs.flutter.dev/ui/navigation)；`auto_route` 等為社群方案，本文未找到 Flutter 背書。可加 [`go_router_builder`](https://pub.dev/packages/go_router_builder) 做 type-safe route codegen（同 Flutter 維護）。 |
| JSON serialization | `json_serializable` + `json_annotation` + `build_runner` | **官方**：[`json_serializable`](https://pub.dev/packages/json_serializable) 發佈者 `google.dev`，`@JsonSerializable` 產生 `fromJson`／`toJson`；[Dart build_runner](https://dart.dev/tools/build_runner) 以它為正式 builder 例。 | 手寫小 DTO 可免 codegen；資料模型不應把 API JSON 形狀漏到 Repository 外。 |
| immutable model／union | 視需要加 `freezed`，可和 `json_serializable` 併用 | **社群主流**：[`freezed`](https://pub.dev/packages/freezed) 的官方文件為 immutable data class、union、copy、JSON integration；Flutter architecture 文件也以 Freezed 建模為例（[offline-first recipe](https://docs.flutter.dev/app-architecture/design-patterns/offline-first)），但 recommendation 是「Freezed or built_value」，不是指定 Flutter 套件。 | Dart 現有 language feature 足以做簡單 immutable class；大量生成檔會增加 build time（官方 recommendation 亦提醒）。 |
| 小型持久化 | `shared_preferences` 的 `SharedPreferencesAsync`（或明確需要同步 cache 才 `WithCache`） | **官方**：[`shared_preferences`](https://pub.dev/packages/shared_preferences) 由 `flutter.dev` 發佈；官方 cookbook 用於 key-value data。新版文件將舊 `SharedPreferences` 列為 legacy，鼓勵新專案使用 Async／WithCache。 | 只存非關鍵設定；該套件明確說 async disk write 不保證 return 後已落盤，且 cache 可在 multi-isolate／multi-engine／native writer 時過期。不可當新聞資料庫或 secret store。 |
| 新聞／離線資料庫與 cache | 先在 Repository 制定 offline policy；有 query、排序、transaction、離線閱讀需求選 SQLite | **官方範例使用**：Flutter [SQLite cookbook](https://docs.flutter.dev/cookbook/persistence/sqlite) 與 [architecture SQL recipe](https://docs.flutter.dev/app-architecture/design-patterns/sql) 示範 `sqflite`；SQL recipe 也重申 Repository 隱藏 DB implementation。 | [`sqflite`](https://pub.dev/packages/sqflite) 是社群 package，非 Flutter 發佈。[`drift`](https://pub.dev/packages/drift) 是社群主流 reactive/type-safe SQLite 選擇，適合複雜 query/migration/stream；兩者皆非官方唯一答案。HTTP cache 沒有發現 Flutter 官方指定套件，應依 HTTP headers、TTL、ETag 和產品 offline policy 在 Repository 實作。 |
| structured logging | `package:logging` 後接 app-specific sink；release 不輸出 PII／token | **官方 Dart**：[`logging`](https://pub.dev/packages/logging) 由 `dart.dev` 發佈，提供 `Logger`、level 與 record handler。 | [`logger`](https://pub.dev/packages/logger) 是常見社群 developer-console formatter，非官方；不能取代 production error reporting。 |
| error/crash reporting | app init 同時處理 `FlutterError.onError` 與 `PlatformDispatcher.instance.onError`；Firebase stack 選 `firebase_crashlytics` | **官方 Flutter + Google/Firebase**：Flutter 說 framework callback errors 經 `FlutterError.onError`，其他未捕獲 error 用 `PlatformDispatcher`；[Firebase Crashlytics for Flutter](https://firebase.google.com/docs/crashlytics/flutter/get-started) 有兩者串接範例。 | Crashlytics 是 Google/Firebase 官方，非 Flutter SDK 必需項；Sentry 等可行但屬其他 vendor。務必 scrub PII、憑證與完整新聞內容。 |
| tests | `test`／`flutter_test`：Repository/mapper/controller unit tests，View widget tests，重要 end-to-end flow 用 `integration_test` | **官方明示**：Flutter 建議大量 unit + widget tests 和覆蓋主要情境的 integration tests；SDK 含 `integration_test`。[testing overview](https://docs.flutter.dev/testing/overview) | Riverpod provider 測試可依其官方 docs 用 `ProviderContainer`／override（社群實作）。Native permission／platform UI 不在 SDK integration_test 能力內；Flutter docs 將 Patrol 定位為第三方選項。[integration concepts](https://docs.flutter.dev/cookbook/testing/integration/introduction) |
| DI | constructor injection；composition root 組裝；若不用 Riverpod 則 `provider` 可作 widget-tree wiring | **官方明示**：Flutter case study 說依賴以 constructor 傳入，並在 widget tree top-level 用 `provider` 佈線；recommendation 明確推薦 DI。[case study](https://docs.flutter.dev/app-architecture/case-study/dependency-injection) | `provider` 本身為社群維護；`get_it` 是社群 Service Locator。選 Riverpod 可用 providers 組裝，但 public constructor 與 fake-injection 邊界仍應保留。不要把「Provider 範例」讀成強制 Provider 或 Service Locator。 |
| lint／static analysis | `flutter_lints` + `flutter analyze`；僅針對已同意規則加嚴 | **官方**：[`flutter_lints`](https://pub.dev/packages/flutter_lints) 由 `flutter.dev` 發佈，starter apps 已預設使用，且基於 Dart recommended lints；[Dart analysis config](https://dart.dev/tools/analysis) 要 Flutter project 用它。 | `very_good_analysis` 等是社群嚴格規則集，需團隊明確採用；避免一次導入大量 lint 而產生噪音。 |

## 3. 對本案可執行的 architecture baseline

1. 每個 feature 保持 `presentation`（widget + Riverpod controller/view model）、`data`
   （repository、remote/local service、DTO/entity mapper）；僅跨 feature 的複雜流程再加 `domain`。
2. `NewsRepository` 是文章、已讀、書籤、cache freshness 的唯一入口。Remote service 持有
   Dio，local service 持有 SQLite／preferences；兩者都不回傳給 Widget。
3. DTO、DB entity 與 UI model 分開。`json_serializable` 主要給 API DTO；若 model 需要
   immutable value equality、sealed UI state，再加 Freezed。不要因 codegen 就讓同一模型跨
   network／DB／UI 三種責任。
4. Router 維持純 route/redirect policy；以 `go_router` route builder 在 feature boundary 建立
   controller/view model。路由參數傳 id／filter，目的地回 Repository 查資料。
5. cache policy 必須是可測規格：freshness TTL、ETag/`If-None-Match`、offline fallback、
   pagination、manual refresh 與失敗時保留舊資料的處置，全部在 Repository tests 覆蓋。
6. 執行 `flutter analyze`、`flutter test`；CI 加 integration tests 覆蓋 cold start、feed
   refresh、離線閱讀／bookmarks 與登入（若有）。

## 4. 每日新聞後端：GCP、排程與來源邊界

### 4.1 Cloud Run + FastAPI 的適切性

**適合。** Google Cloud 有直接的 [FastAPI Cloud Run quickstart](https://docs.cloud.google.com/run/docs/quickstarts/build-and-deploy/deploy-python-fastapi-service)，並支援 `gcloud run deploy --source .` 由 source build/deploy。因此 FastAPI 可作為公開 app API 或內部 ingestion endpoint。若 endpoint 只供批次使用，應保留 Cloud Run IAM authentication，不要 `allow unauthenticated`。

每日工作有兩個合理模式：

| 模式 | 適用情境 | 官方依據與注意事項 |
|---|---|---|
| GitHub Actions schedule → protected FastAPI endpoint | batch CLI 與程式碼、fixture、部署 pipeline 同 repo；想從 workflow 觸發 ingestion | `schedule` 可延遲，特別是整點高負載，且 workflow 必須在 default branch；[GitHub Actions events](https://docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows)。以 GitHub OIDC + GCP Workload Identity Federation 取得短期身分，Cloud Run 對 caller principal 賦予最小範圍 `roles/run.invoker`，取得 ID token 後以 `Authorization: Bearer` 呼叫。[GCP WIF for pipelines](https://docs.cloud.google.com/iam/docs/workload-identity-federation-with-deployment-pipelines)／[Cloud Run auth](https://docs.cloud.google.com/run/docs/authenticating/service-to-service) |
| Cloud Run Job + Cloud Scheduler | 純 batch、無 HTTP API 需求，想使 scheduling/identity 留在 GCP | Cloud Run 官方文件直接支援 Cloud Scheduler schedule execution；Scheduler service account 需對 job 有 invoker 權限。[Execute jobs on a schedule](https://docs.cloud.google.com/run/docs/execute/jobs-on-schedule) |

不論哪種模式，ingestion 須有 idempotency key（來源 URL + published timestamp/content hash）、
timeout/retry/backoff、run audit log，以及「單來源失敗不清空昨日結果」的 transaction 策略。
GitHub Actions 若走 Cloud Run endpoint，避免長期 service-account JSON key；Google 的
[authentication action](https://github.com/google-github-actions/auth) 明確優先建議 Workload
Identity Federation。

### 4.2 Gemini Grounding with Google Search 與 URL Context：能力不是爬蟲保證

| 工具 | 實際能力 | 重要限制／產品規則 |
|---|---|---|
| Grounding with Google Search | API 由 model 自行判斷是否搜索、可產生一或多個 query，回傳回答及 inline `url_citation` annotations；適合一般公開網頁的發現與有引用的摘要。[official doc](https://ai.google.dev/gemini-api/docs/google-search) | 不是「指定所有網站並保證擷取」的 crawler API：model 決定 query，結果與可讀內容取決於 Google Search／頁面可近性。必須儲存並呈現 citation，不能把無 citation 的模型敘述當來源事實。Gemini 3 按模型實際執行的 query 計費。 |
| URL Context | 對 prompt 提供的公開 URL，從 index cache 或 live fetch 讀取，並回傳 URL citations；可與 Google Search 合用，讓搜索找候選 URL 後再深入讀頁面。[official doc](https://ai.google.dev/gemini-api/docs/url-context) | 一次最多 20 URLs、單 URL 34 MB；必須 public、不可 login/paywall；只讀提供的 URL，不追 nested links。支援 text/image/PDF；**明列不支援 YouTube video、video/audio、Google Workspace 檔**。URL 可能被安全檢查標為 `unsafe`。 |

因此，Grounding 的產物適合作「候選新聞與摘要輔助」，資料管線仍須保存原始 URL、擷取
時間、來源策略、模型 output 與 citation，並在 publication 前按來源規則驗證。不要以它
繞過網站登入、付費牆、robots／條款或 API 權限。

### 4.3 來源取得矩陣（Facebook／Instagram／Threads／YouTube／GitHub 不能一概當網站）

| 來源 | 可作為一般 Google Search/URL Context 候選？ | 可靠且合規的 ingestion 優先順序 | 明確不可假設的事 | 一手來源 |
|---|---|---|---|---|
| 一般新聞站／官方部落格 | **可以，但僅限公開可索引／可直讀頁。** | 站方 RSS/Atom（若實際提供）→ 公開原文 URL + URL Context → Google Search 發現；保存 citation 與 canonical URL。 | 不可保證搜索收錄、更新即時或全文擷取；paywall/login 不支援。 | [Google Search grounding](https://ai.google.dev/gemini-api/docs/google-search)、[URL Context limits](https://ai.google.dev/gemini-api/docs/url-context) |
| Facebook | **不可當任意可靠來源。** 公開、無登入的單一 HTML URL 理論上才可能被 URL Context 讀取；不應依賴。 | 需結構化／穩定資料時用 Meta **Graph/Pages API**，依 app review、token 與允許範圍取得；或由權利人提供 RSS/授權 feed。 | 不可把 Google Search 或 URL Context 當成任意 Facebook post/feed 的讀取 API，也不可繞登入／權限。 | [Meta Graph API overview](https://developers.facebook.com/docs/graph-api/overview/) |
| Instagram | **不可當任意可靠來源。** 頁面常涉及 login／動態內容，故不作 ingestion contract。 | 用 Meta **Instagram API / Instagram Graph API**，適用於授權的 professional accounts；必要時取得權利人提供的 feed／授權。 | 不能假定可搜尋或擷取任何帳號、Reel、留言或私密內容。 | [Instagram API official collection](https://www.postman.com/meta/instagram/documentation/6yqw8pt/instagram-api) |
| Threads | **不可當任意可靠來源。** 同上，僅把 public direct URL 視為 best-effort candidate。 | 用 Meta **Threads API**，由 app user 授權資料 access；或有明確授權的來源 feed。 | 不可將 Threads 視為免授權的全站搜尋／archive API。 | [Threads API official collection](https://www.postman.com/meta/threads/documentation/dht3nzz/threads-api) |
| YouTube | Google Search 可能找到 landing page，但 **URL Context 明確不支援 YouTube videos**。 | 用 **YouTube Data API `search.list`** 搜尋／列舉 video/channel/playlist metadata；若要理解影片內容，改走 Gemini 的 video-understanding 流程或取得可用 transcript／授權資料。 | 不可把 YouTube URL 丟給 URL Context 並期待影片內容被擷取。 | [URL Context unsupported types](https://ai.google.dev/gemini-api/docs/url-context)、[YouTube Data API search.list](https://developers.google.com/youtube/v3/docs/search/list) |
| GitHub | public repo page 可能是 public URL candidate，但不該是 release／event ingestion 的主要介面。 | 用 GitHub **REST API**：releases、repository events、issues/PR 等；public release 可無 authentication，並遵從 rate limit／ETag polling。RSS/Atom 僅在 GitHub 實際公開並符合需求時使用。 | 不可將搜尋結果當完整 release/security/event feed，也不可忽略 private repo auth。 | [REST releases](https://docs.github.com/en/rest/releases/releases)、[REST events + ETag polling](https://docs.github.com/en/rest/activity/events) |

這張表的保守原則是：只有「public 可直讀的 text/image/PDF URL」才符合 URL Context 的
文件前提；各平台的動態內容、登入、付費牆與影片都不構成可依賴的 generic-search contract。
若來源是產品需求，先取得該來源的官方 API、token、rate-limit 與使用條款，再把它做成
獨立 `SourceService`；若只是新聞佐證，優先引用原始官方新聞稿／文章，而非社群轉貼。

## 來源索引

* Flutter：[架構指南](https://docs.flutter.dev/app-architecture/guide)、[建議表](https://docs.flutter.dev/app-architecture/recommendations)、[分層概念](https://docs.flutter.dev/app-architecture/concepts)、[依賴注入 case study](https://docs.flutter.dev/app-architecture/case-study/dependency-injection)、[測試概覽](https://docs.flutter.dev/testing/overview)。
* Dart／Flutter packages：[json_serializable](https://pub.dev/packages/json_serializable)、[build_runner](https://dart.dev/tools/build_runner)、[shared_preferences](https://pub.dev/packages/shared_preferences)、[flutter_lints](https://pub.dev/packages/flutter_lints)、[go_router](https://pub.dev/packages/go_router)。
* Google Cloud/Gemini：[FastAPI on Cloud Run](https://docs.cloud.google.com/run/docs/quickstarts/build-and-deploy/deploy-python-fastapi-service)、[Cloud Run auth](https://docs.cloud.google.com/run/docs/authenticating/service-to-service)、[Google Search grounding](https://ai.google.dev/gemini-api/docs/google-search)、[URL Context](https://ai.google.dev/gemini-api/docs/url-context)。
