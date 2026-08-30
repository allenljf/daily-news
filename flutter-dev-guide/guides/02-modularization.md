# 02 · 模組化

這份管 feature-first 結構、何時該拆 Dart / Flutter package，以及跨 feature 的依賴邊界。

---

## 一、先按功能切，不先按層切

預設從 feature-first 開始：

```text
lib/
  features/
    news_feed/
      presentation/
      data/
      domain/
```

先把同一個功能放在一起，比把全 app 切成 `widgets/`, `repositories/`, `services/` 更能維持可讀性與可改性。

## 二、何時值得拆 package

只有在以下情況才值得把一塊程式碼拆成獨立 package：

1. 被多個 feature 或多個 app 重用
2. 需要獨立釋出或獨立測試矩陣
3. 有清楚的 public contract，且可以不讀 internals 就使用
4. 獨立後能減少耦合，而不是只增加 import 路徑

單一畫面、單一 repository、兩三個 helper 通常不值得為了「看起來整齊」拆 package。

## 三、跨 feature 共享的做法

共用 domain model、repository contract 或 design token 時，抽到共享 package 或 `core/`。不要讓 `features/news_feed` import `features/category_settings` 的 internals。

---

### feature-first-structure · SHOULD · new-only · manual

新的功能預設採 feature-first 結構，讓 presentation、data、domain 一起演化。只有證明會被多 feature 重用時才往共享層抽。

### avoid-cross-feature-internals · MUST_NOT · new-only+on-touch · manual

一個 feature 不應直接依賴另一個 feature 的私有檔案、provider 或 UI model。跨 feature 共享要經由公開 contract 或共享 package。

### package-only-when-boundary-is-real · SHOULD · new-only · manual

拆 package 前先說清楚 public API、擁有者、測試邊界與依賴方向。若答不出來，先留在原 feature。

---

## 主要來源

- Flutter architecture case study: <https://docs.flutter.dev/app-architecture/case-study>
- Flutter architecture guide: <https://docs.flutter.dev/app-architecture/guide>
