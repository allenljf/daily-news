# 01 · 架構分層與依賴方向

這份管層與層之間的邊界：Widget、ViewModel、Repository、Service 各做什麼，model 放哪裡，資料怎麼流。

| 主題 | 去哪 |
|---|---|
| feature-first 結構與 package 切分 | [02-modularization.md](02-modularization.md) |
| Riverpod wiring 與 provider override | [03-dependency-injection.md](03-dependency-injection.md) |
| use case 何時該出現 | [04-domain-layer.md](04-domain-layer.md) |

---

## 一、建議的層次

| 層 | 負責 | 不負責 |
|---|---|---|
| View | render UI、收使用者輸入、轉送 event | 資料取得、快取、HTTP、SQLite |
| ViewModel / controller | 把 domain data 轉成 UI state、處理 user intent | 直接碰 Dio、直接查 DB |
| Domain（可選） | 可重用或複雜的商業規則 | UI 文案、HTTP、持久化細節 |
| Repository | source of truth、整合 remote/local、refresh/retry/error policy | render、BuildContext、Widget state |
| Service | 封裝單一外部來源：API、SQLite、prefs、platform bridge | 組合多來源商業規則 |

對小 feature 而言，常見形狀是 `View -> ViewModel -> Repository -> Service`。只有當商業規則跨多個 ViewModel 共用，或單一 ViewModel 已經太重時，才加 use case。

## 二、model 邊界

| model | 住哪 | 誰看得到 |
|---|---|---|
| DTO | remote service / data layer | Repository 實作內部 |
| SQLite / Drift entity | local service / data layer | Repository 實作內部 |
| Domain model | repository contract / domain layer | 全 feature 或多 feature |
| UI model / UiState | presentation layer | 該畫面 |

DTO、entity、domain model、UI model 名字相近也不要共用同一型別。它們代表的是不同責任與變更節奏。

## 三、判斷表

| 問題 | 放哪 | 原因 |
|---|---|---|
| `expiresAt` 怎麼算、新聞要不要顯示永久標籤 | ViewModel 或 use case | 這是業務或呈現規則，不是 transport 細節 |
| API response 轉 app model | Repository / mapper | DTO 不外流 |
| refresh、cache、fallback to local | Repository | SSOT 與錯誤策略集中 |
| token、timeout、cancel token、interceptor | Remote service | 只跟網路客戶端有關 |
| SQLite query、prefs read/write | Local service | 只跟單一持久化技術有關 |

---

### widget-reads-repository · MUST_NOT · new-only+on-touch · regex

Widget 不直接依賴 Repository。它應依賴 ViewModel 或 presentation provider，從 state render，透過 callback / command 送事件。

### viewmodel-accesses-service · MUST_NOT · new-only+on-touch · regex

ViewModel 不直接持有 service、Dio、SQLite 或 shared_preferences。它的直接依賴應是 Repository 或必要時的 use case。

### layer-models-do-not-cross · MUST_NOT · new-only+on-touch · manual

DTO 只在 network boundary，entity 只在 persistence boundary，domain model 用在 repository contract，UI model / UiState 只留在 presentation。不要省 mapper 而讓某一層的型別向外漏。

---

## 主要來源

- Flutter architecture guide: <https://docs.flutter.dev/app-architecture/guide>
- Flutter data layer: <https://docs.flutter.dev/app-architecture/data-layer>
- Flutter dependency injection case study: <https://docs.flutter.dev/app-architecture/case-study/dependency-injection>
