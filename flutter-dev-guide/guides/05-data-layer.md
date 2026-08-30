# 05 · Data Layer

這份管 Repository 與 mapper 的責任：source of truth、remote/local 協調、錯誤翻譯、refresh 與 cache policy。

| 想找 | 去哪 |
|---|---|
| Dio 與 HTTP 細節 | [06-network.md](06-network.md) |
| shared_preferences、SQLite、Drift | [07-persistence.md](07-persistence.md) |
| use case 是否值得新增 | [04-domain-layer.md](04-domain-layer.md) |

---

## 一、Repository 是資料真相來源

Repository 對上層提供穩定 contract，隱藏資料實際來自 API、SQLite、prefs 或記憶體 cache。ViewModel 不應知道資料怎麼刷新或 fallback。

## 二、remote 與 local 的協調責任

Repository 決定：

1. 先讀 local 還是先打 remote
2. remote 成功後如何更新 local
3. 錯誤時是否保留舊資料
4. refresh / retry / TTL / pagination 的政策

Service 只回傳單一來源的結果，不自己替整個 feature 決定快取策略。

## 三、錯誤模型

HTTP、serialization、database、permission 這些 infrastructure error 應在 Repository 邊界被翻譯成 app 能理解的錯誤類型。不要把 `DioException` 或 SQL 例外一路丟到 Widget。

## 四、mapper 只做轉換

mapper 做欄位對應、缺值補預設、簡單型別轉換。商業規則放在 Repository 或 use case；UI 格式化放在 ViewModel。

---

### repository-owns-source-of-truth · MUST · new-only+on-touch · manual

每一份業務資料的寫入與刷新策略由 Repository 集中決定。若 Widget、ViewModel 或 service 也各自維護一份可寫入副本，就不再是 SSOT。

### repository-returns-domain-models · MUST · new-only+on-touch · manual

Repository 對外回傳 domain model、value object 或明確 result，不直接回 DTO、entity、Dio response 或 shared_preferences 原始值。

### repository-translates-infra-errors · MUST · new-only+on-touch · manual

Repository 在邊界處理 transport、serialization、storage 失敗，轉成上層可處理的錯誤。UI 不該知道底層是 Dio timeout 還是 SQLite lock。

### mapper-stays-dumb · SHOULD · new-only · manual

mapper 不藏商業判斷與 UI 規則。它應該能靠固定輸入輸出單測，不需要 provider、clock 或 network。

---

## 主要來源

- Flutter architecture guide repositories section: <https://docs.flutter.dev/app-architecture/guide#repositories>
- Flutter architecture recommendations: <https://docs.flutter.dev/app-architecture/recommendations>
