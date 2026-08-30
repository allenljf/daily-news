# 04 · Domain Layer

這份回答兩件事：什麼情況值得建立 use case，以及 domain 層一旦存在，裡面可以放什麼。

---

## 一、不是每個 feature 都需要 domain 層

Flutter 官方把 domain/use case 視為條件式層，不是固定樣板。以下情況才值得新增：

1. 同一條商業規則被兩個以上 ViewModel 重用
2. 單一 ViewModel 已承擔多段可獨立命名的商業流程
3. 邏輯需要脫離 UI 與資料細節，單獨以純 Dart 驗證

只有一個畫面用到、而且只是把 repository 回傳資料稍作排序或格式化，通常留在 ViewModel 就夠了。

## 二、domain 層應該保持純 Dart

domain model、use case、policy 只處理業務規則，不 import Flutter、Riverpod、Dio、SQLite、shared_preferences 或 `BuildContext`。

## 三、domain 的輸入輸出

domain 層對外使用 domain model、value object、明確的 command / result。它不回傳 DTO、entity，也不拼 UI 文案。

---

### usecase-only-for-shared-or-complex-logic · SHOULD · new-only+on-touch · manual

只有跨多個呼叫端重用，或已經複雜到需要獨立命名與測試的商業規則，才新增 use case。不要為了形式把每個 repository call 再包一層。

### domain-is-pure-dart · MUST · new-only+on-touch · regex

domain 層不 import Flutter、Riverpod、Dio、SQLite、Drift 或 shared_preferences。它應能在純 Dart 測試中直接建立與執行。

### domain-does-not-format-ui-copy · MUST_NOT · new-only+on-touch · manual

domain 回傳的是商業結果，不是 `"更新成功"` 這種 UI 文案或 `Color`、`TextStyle` 之類呈現型別。

---

## 主要來源

- Flutter architecture recommendations: <https://docs.flutter.dev/app-architecture/recommendations>
- Flutter architecture guide: <https://docs.flutter.dev/app-architecture/guide>
