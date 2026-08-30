# 15 · 測試

**管**：Flutter 要測到哪一層、Riverpod override 與 fake 的使用方式、Widget 與 integration test 的邊界、非同步測試的穩定寫法、假綠燈防治。  
**不管**：分層與依賴方向本身（`01-architecture`、`05-data-layer`、`08-ui-state`）、單次畫面的 render 最佳化（`16-performance`）。

| 想找 | 去 |
|---|---|
| Repository／Service 邊界、錯誤模型、mapper 放哪 | `01-architecture` `05-data-layer` `06-network` |
| `AsyncValue`、ViewModel state、one-off effect | `08-ui-state` |
| Widget public API、controller lifecycle | `09-widget-api` `10-widget-state-and-lifecycle` |
| profile mode 與 frame／startup 量測 | `16-performance` |

---

## 一、測試金字塔

| 受測對象 | 一定要測 | 不要測 | 建議工具 |
|---|---|---|---|
| 純 Dart model／mapper／validator／use case | 邊界值、空集合、錯誤分支、轉換規則 | Flutter framework 本身 | `package:test` 或 `flutter_test` |
| Repository | 成功、HTTP 錯誤、逾時、空資料、快取／刷新決策 | Dio、SQLite、shared_preferences 的框架行為 | fake service / fake local store |
| Riverpod Notifier／ViewModel | `loading → data / error` 狀態轉移、使用者動作、effect 旗標 | Widget 排版細節 | `ProviderContainer` + override |
| Widget | loading / empty / error / success 畫面、互動、無障礙語意 | 真 API、真資料庫、整條 app flow | `testWidgets` |
| Integration | 登入、建立 Category、讀第一頁新聞、手動更新等關鍵旅程 | 每個商業分支都靠 E2E 驗 | `integration_test` |

原則只有一個：**選能正確驗證行為的最低層**。能在純 Dart 驗的不要拖進 widget test；能在 widget test 驗的不要拖進 integration。

## 二、fake 優先，override 明確

Flutter + Riverpod 的測試穩定性，多半取決於替身策略。自家介面優先用 fake，原因和 production code 一樣務實：介面一改，fake 會跟著編譯失敗；寬鬆 mock 常常繼續全綠，卻已經不再代表真實行為。

```dart
test('load emits data after repository succeeds', () async {
  final container = ProviderContainer(
    overrides: [
      newsRepositoryProvider.overrideWithValue(
        FakeNewsRepository(items: [testArticle()]),
      ),
    ],
  );
  addTearDown(container.dispose);

  final notifier = container.read(newsListControllerProvider.notifier);
  await notifier.load();

  expect(
    container.read(newsListControllerProvider).items,
    [testArticle()],
  );
});
```

- provider override 一律在測試入口建立，不要在 production provider 裡偷塞 test 分支。
- fake 要能表達狀態，而不是只回固定值；例如 `setResponse()`、`seedArticles()`。
- mock 只保留給「驗證副作用是否有送出」這種沒有可觀察結果的依賴，例如 analytics adapter。

## 三、Widget 與 integration 的分工

- Widget test 的目標是畫面邏輯，不是整條 app stack。用 override 提供 fake repository / fake auth state，讓測試可以直接進入 loading、empty、error、success。
- Integration test 只守關鍵旅程：Google Sign-In、儲存 Category、第一頁新聞、手動更新流程。不要把所有錯誤分支都搬進 integration。
- `pumpAndSettle()` 不是萬能解。若畫面有持續動畫、輪播或永不結束的 loading spinner，它會讓測試卡住。優先用有限次 `pump()` 或等明確條件出現。

## 四、非同步測試寫法

- Future / notifier 測試優先 `await` 具體方法，不要靠任意延遲等待狀態自己變。
- Stream 驗多個 emission 時，收集明確序列；只看最後值常會漏掉中間的 loading 或 error。
- Widget test 禁用固定睡眠。等待畫面更新用 `pump()`、`pump(const Duration(...))`、或輪詢直到找到特定節點。
- 任何會建立 `ProviderContainer`、`StreamController`、`TextEditingController`、`AnimationController` 的測試，都要在測試結束釋放。

## 五、規則

### repository-test-uses-real-network · MUST_NOT · new-only · manual

Repository test 不得打真 API、真 Firebase、真 Cloud Run；一律以 fake service、fake adapter 或受控測試伺服器驗證。

✗ `await Dio().get("https://api.example.com/news")`  
✓ `FakeNewsApiService(result: [...])`

例外：獨立標示的 smoke test，可跑在 CI 以外的驗收流程。

### provider-override-missing-in-test · MUST · new-only · manual

測試需要控制 provider 輸入時，必須用 `ProviderContainer(overrides: ...)` 或 `ProviderScope(overrides: ...)`，不得讓 widget test 默默吃到 production provider。

✗ `await tester.pumpWidget(const MyApp())` 後直接期待 API 狀態  
✓ `ProviderScope(overrides: [...], child: MyApp())`

例外：驗證 production wiring 的少數 integration test。

### fake-state-not-observable · SHOULD · new-only · manual

fake 應提供可觀察的狀態變化，而不是每次只回固定常數；否則 ViewModel / notifier 的狀態轉移根本測不到。

✗ `class FakeRepo implements Repo { Future<List<Item>> fetch() async => []; }`  
✓ `FakeRepo(seed: [], nextError: TimeoutException(...))`

例外：只驗單一路徑的極小純函式測試。

### widget-test-asserts-only-final-state · MUST_NOT · new-only+on-touch · manual

畫面有明確 loading 或 error 過程時，不得只斷言最後成功畫面；至少要驗證一個中間狀態，避免假綠燈。

✗ `await tester.pumpAndSettle(); expect(find.text('Latest'), findsOneWidget);`  
✓ 先驗 loading，再觸發完成，最後驗 success

例外：同步純展示 widget。

### fixed-delay-in-test · MUST_NOT · new-only+on-touch · regex

測試不得用 `Future.delayed`、`sleep` 或任意秒數等待 UI 更新；等待必須綁定 widget tree 或受測方法完成。

✗ `await Future<void>.delayed(const Duration(seconds: 2));`  
✓ `await tester.pump();` / `await notifier.load();`

例外：無。

### integration-test-covers-branch-logic · SHOULD_NOT · new-only · manual

欄位驗證、mapper 邊界、Repository 錯誤轉換等分支不應只靠 integration test 覆蓋；這些必須在單元或 widget 層有較快的回饋。

✗ 只有一條「新增 Category 成功」E2E，沒有 ViewModel / Repository test  
✓ 單元與 widget 先覆蓋規則，再留一條整體串接驗證

例外：平台插件的真機特性，只能在 integration 驗。

## 六、完成前檢查

至少確認這些命令能跑，並把結果記錄到任務完成紀錄：

```bash
flutter analyze
flutter test
flutter test integration_test
python3 flutter-dev-guide/tools/check-rules.py --staged
```

若專案尚未有 `integration_test` 或 `check-rules.py`，要在變更說明寫清楚是「未接入」，不是默默省略。
