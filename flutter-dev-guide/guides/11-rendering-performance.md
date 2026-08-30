# 11 · Rendering 與 Rebuild 效能

**這份管**：build 成本、rebuild 範圍、`const`、list key、長清單與圖片、什麼情況值得做局部優化。
**這份不管**：資料層快取、隔離執行緒、app 啟動與 DevTools 量測流程。

| 需要什麼 | 去哪 |
|---|---|
| 本地 state / lifecycle 錯誤導致重建 | `10-widget-state-and-lifecycle` |
| Future / Stream / isolate 與非同步錯誤 | `12-async-streams-isolates` |
| token、Material 3 與圖片語意 | `14-design-system` |
| 全域效能量測與 profile/release 檢查 | `16-performance` |

Flutter 效能優化的原則是先縮小變動範圍，再考慮更進一步的技巧。大多數畫面不需要花俏技巧，只需要正確的 widget 邊界與 list key。

---

### const-where-widget-is-static · SHOULD

適用檔案：`lib/**/widgets/**/*.dart`、`lib/**/screens/**/*.dart`
安全自動檢查：analyzer（`prefer_const_constructors` / `prefer_const_literals_to_create_immutables`）

沒有依賴 runtime 值的 widget、edge insets、文字樣式常數，優先標成 `const`，讓 framework 可跳過重建與重新配置。

✗ `return Padding(padding: EdgeInsets.all(16), child: Icon(Icons.add));`
✓ `return const Padding(padding: EdgeInsets.all(16), child: Icon(Icons.add));`

例外：theme、localization、media query 這類執行期值不可硬標 `const`。

### watch-smallest-provider-scope · MUST

適用檔案：`lib/**/screens/**/*.dart`、`lib/**/widgets/**/*.dart`
安全自動檢查：manual

`ref.watch` 應放在真正需要該值的最小 widget 範圍，不要在頁面根節點一次 watch 大量 provider 再把整包 state 往下傳，造成整頁跟著重建。

✗ `HomeScreen` 一次 `watch` news、theme、run status、filters，再全部往下傳
✓ 把各區塊拆成小 widget，各自 watch 自己需要的 provider 或 selector

例外：畫面本來就只有單一主資源，而且拆分只會讓結構更難讀時。

### stable-key-for-mutable-collections · MUST

適用檔案：`lib/**/screens/**/*.dart`、`lib/**/widgets/**/*.dart`
安全自動檢查：regex（可掃描 `ListView.builder` / `SliverList` 是否帶 key）+ manual

會插入、刪除、排序或分頁追加的清單項目，必須以穩定業務 id 建 `Key`，不得用 index 當 key。

✗ `key: ValueKey(index)`
✓ `key: ValueKey(article.id)`

例外：項目數固定、順序永不變、且 item 沒有本地 state 的純展示清單。

### no-expensive-work-in-build · MUST_NOT

適用檔案：`lib/**/screens/**/*.dart`、`lib/**/widgets/**/*.dart`
安全自動檢查：manual

`build()` 不得執行排序、分組、正則、JSON 解析、資料庫查詢、圖片解碼或大型集合轉換。這些工作要前移到 provider / repository，或至少在輸入不變時快取。

✗ `final grouped = articles..sort(...);`
✓ provider 先輸出已排序資料；widget 只 render

例外：小型常數集合的輕量映射可接受，但不要讓它長成熱路徑。

### builder-for-long-scrolling-content · MUST

適用檔案：`lib/**/screens/**/*.dart`
安全自動檢查：regex（可掃描 `ListView(children: ...)`、`Column` 塞長清單）+ manual

長度未知或可能超出一屏的清單，必須使用 `ListView.builder`、`SliverList`、`GridView.builder` 等 lazy builder；不要一次建立所有 children。

✗ `ListView(children: articles.map(NewsRow.new).toList())`
✓ `ListView.builder(itemCount: articles.length, itemBuilder: ...)`

例外：數量固定且很少的靜態項目區塊。

### image-layout-reserved-upfront · SHOULD

適用檔案：`lib/**/screens/**/*.dart`、`lib/**/widgets/**/*.dart`
安全自動檢查：manual

遠端圖片或延後載入的媒體應預先保留尺寸，避免 layout shift 與反覆重排。使用 `AspectRatio`、固定高度或明確 constraints。

✗ `Image.network(url)` 直接裸放在 list item
✓ `AspectRatio(aspectRatio: 16 / 9, child: Image.network(url, fit: BoxFit.cover))`

例外：父層已給明確尺寸限制。

## 尚無定論

- `Consumer`、`HookConsumerWidget`、拆小 widget 三種方式都能縮小 rebuild 範圍。建議先選最直白、最好懂的一種，只有在 profile 真看到熱點時才再局部調整。
- 不要把 `const` 當成唯一效能策略。清單 key、provider watch 範圍、避免在 `build()` 做工作，通常比大量補 `const` 更有感。
