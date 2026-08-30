# 03 · 依賴注入

這份管 constructor injection、Riverpod composition root，以及測試時如何 override provider 注入 fake。

| 想找 | 去哪 |
|---|---|
| 層次與依賴方向 | [01-architecture.md](01-architecture.md) |
| repository 的責任與 error policy | [05-data-layer.md](05-data-layer.md) |
| 測試時 override 哪一層 | 後續 `15-testing.md` |

---

## 一、預設是建構子注入

Repository、service、use case、ViewModel 都應透過建構子宣告依賴。這讓依賴可見，也讓 fake 注入自然。

## 二、Riverpod 是 composition root，不是捷徑

Riverpod 很適合在 app root 與 feature presentation 做組裝：

1. 建立 Dio、database、shared preferences async instance
2. 建立 remote/local service
3. 建立 repository
4. 建立 ViewModel / controller

但 `Ref` 不應一路傳進 repository、service、use case。那些層應維持普通 Dart 物件，才能在不啟動 provider container 的情況下單測。

## 三、測試的 override seam

Widget test 與 controller test 應優先 override repository provider，注入 fake repository。這比 mock Riverpod internals 或直接 mock notifier 更穩定。

---

### riverpod-ref-stays-in-presentation · MUST_NOT · new-only+on-touch · regex

`Ref` 只留在 composition root 與 presentation。Repository、service、use case 接受的是明確依賴，不是 provider runtime。

### buildcontext-not-in-viewmodel · MUST_NOT · new-only+on-touch · regex

ViewModel 不持有 `BuildContext`。導航、SnackBar、Dialog 等效果由 View 或獨立 effect handler 執行。

### provider-overrides-enable-fakes · MUST · new-only · manual

新功能的 provider wiring 必須保留 override seam，讓 widget test / controller test 能以 fake repository 驗證邏輯，而不是起真網路或真 DB。

---

## 主要來源

- Flutter dependency injection case study: <https://docs.flutter.dev/app-architecture/case-study/dependency-injection>
- Riverpod testing guide: <https://riverpod.dev/docs/how_to/testing>
- `flutter_riverpod` package page: <https://pub.dev/packages/flutter_riverpod>
