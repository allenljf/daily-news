# 00 · 核心原則

這份檔案定義 Flutter 實作的共同原則。它回答的是「為什麼要這樣切」，不是每個 widget 或 service 的細節寫法。

| 想找 | 去哪 |
|---|---|
| 層與層之間怎麼切、model 放哪 | [01-architecture.md](01-architecture.md) |
| Riverpod provider 與建構子注入怎麼分工 | [03-dependency-injection.md](03-dependency-injection.md) |
| Repository、cache、error、refresh 誰負責 | [05-data-layer.md](05-data-layer.md) |

---

## 1 · 關注點分離

Widget 負責 render 與回報事件。資料從哪來、何時刷新、失敗怎麼翻譯，不是 Widget 的責任。

## 2 · Data-driven UI

畫面應該是狀態的函式。相同 state 要得到相同畫面，不靠 widget 內偷偷維護第二份真相。

## 3 · 單一資料來源

一份資料只能有一個可寫入擁有者。對 Flutter app 而言，這通常是 Repository。

## 4 · 單向資料流

事件往上送到 ViewModel，狀態往下流到 View。中間任何一段都不逆向寫入。

## 5 · 不可變狀態

公開給 UI 的 state 與 model 應保持不可變，靠新值取代舊值，而不是到處改同一個物件。

## 6 · 依賴反轉

上層依賴抽象，下層提供實作。ViewModel 依賴 `NewsRepository`，不是 `NewsApiService`。

## 7 · 可測試性是設計約束

介面定下來時就要知道如何用 fake 驗證它。需要 provider override、時間注入或 repository fake，就在設計階段留出 seam。

---

### single-source-of-truth · MUST · new-only+on-touch · manual

每一種業務資料都要有單一寫入擁有者，通常是 Repository。Widget、ViewModel、service 不得各自保存可寫入副本。

### immutable-public-state · MUST · new-only+on-touch · manual

公開的 state、domain model、UI model 應預設不可變。要改變時產生新值，讓重建與測試都能靠資料比對判斷差異。

### unidirectional-data-flow · MUST · new-only+on-touch · manual

UI 只能透過 command/event 呼叫 ViewModel，不能直接修改 Repository、service 或 provider 內部可變欄位。

---

## 主要來源

- Flutter architecture concepts: <https://docs.flutter.dev/app-architecture/concepts>
- Flutter architecture recommendations: <https://docs.flutter.dev/app-architecture/recommendations>
