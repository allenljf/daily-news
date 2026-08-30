# 09 · Widget 公開 API

**這份管**：公開 Widget 的建構子、參數形狀、值與 callback 的邊界、`super.key`、可重用 widget 不該知道哪些上層概念。
**這份不管**：Riverpod provider state、生命週期、效能與導航圖。

| 需要什麼 | 去哪 |
|---|---|
| `AsyncValue` / screen state 的建模方式 | `08-ui-state` |
| controller 擁有權、`dispose`、`mounted`、`didUpdateWidget` | `10-widget-state-and-lifecycle` |
| rebuild hot path 與 list key | `11-rendering-performance` |
| 導航 callback 與 route 參數 | `13-navigation` |
| Material 3、語意、字串資源 | `14-design-system` |

可重用 widget 的 API 必須讓人只看建構子就知道：它需要哪些值、會回傳哪些使用者意圖、以及它不依賴 provider 或 router。

---

### public-widget-uses-super-key · MUST

適用檔案：`lib/**/widgets/**/*.dart`、`lib/**/screens/**/*.dart`
安全自動檢查：regex（可掃描公開 widget 建構子是否包含 `super.key`）

所有公開 `StatelessWidget`、`StatefulWidget`、`ConsumerWidget`、`ConsumerStatefulWidget` 建構子都必須接受 `super.key`。

✗ `const NewsCard({required this.article});`
✓ `const NewsCard({super.key, required this.article});`

例外：private widget 若永遠不會被 key 控制，得省略。

### widget-api-value-plus-callback · MUST

適用檔案：`lib/**/widgets/**/*.dart`
安全自動檢查：manual

widget 參數只收 render 所需的值與使用者事件 callback，不收可變容器、service、repository、provider ref、router 實例。

✗ `NewsToolbar({required this.ref, required this.router})`
✓ `NewsToolbar({required this.selectedTagId, required this.onTagSelected, required this.onRefreshPressed})`

例外：最外層 route widget 可直接讀 provider，因為它本來就是接縫層。

### reusable-widget-no-provider-read · MUST_NOT

適用檔案：`lib/**/widgets/**/*.dart`
安全自動檢查：regex（可掃描 `WidgetRef`、`Consumer(`、`ref.watch`）+ manual

可重用 widget 與 list item widget 不得自行 `ref.watch` / `ref.read` provider。provider 讀取應留在 route / screen 容器層，再把值往下傳。

✗ `class NewsRow extends ConsumerWidget { ... ref.watch(savedNewsProvider) ... }`
✓ `class NewsRow extends StatelessWidget { final bool isSaved; final VoidCallback onSavePressed; ... }`

例外：整個 widget 的目的就是提供一個局部 provider boundary，且名稱已明確表達它是 container，例如 `NewsFeedSection`。

### widget-parameter-surface-minimized · SHOULD

適用檔案：`lib/**/widgets/**/*.dart`
安全自動檢查：manual

widget 應只接收自己真正需要的欄位，不要把整個 domain model 或整包 page state 傳進去。這能減少耦合，也讓未來調整模型欄位時影響面更小。

✗ `HeadlineText(article: article)` 但只用到 `article.title`
✓ `HeadlineText(title: article.title)`

例外：多個欄位本來就語意成組，且拆成十幾個參數反而降低可讀性時，可傳單一 UI model。

### callback-name-expresses-user-intent · SHOULD

適用檔案：`lib/**/widgets/**/*.dart`
安全自動檢查：regex（命名慣例）+ manual

callback 以 `onXxx` 命名，描述使用者意圖，不描述技術細節。

✗ `tapHandler`、`handlePressed`
✓ `onArticlePressed`、`onRetryPressed`、`onDeleteConfirmed`

例外：Flutter 既有 API 名稱如 `onChanged`、`onSubmitted` 照官方慣例。

### semantic-label-on-action-widget · MUST

適用檔案：`lib/**/widgets/**/*.dart`、`lib/**/screens/**/*.dart`
安全自動檢查：manual

沒有可見文字的互動元件必須提供可本地化的語意標籤，例如 `tooltip`、`Semantics(label: ...)` 或元件本身的無障礙屬性。只靠 icon 圖形不可接受。

✗ `IconButton(icon: Icon(Icons.delete), onPressed: onDeletePressed)`
✓ `IconButton(tooltip: context.l10n.deleteArticle, icon: const Icon(Icons.delete), onPressed: onDeletePressed)`

例外：外層已提供等價語意，且子元件明確排除重複朗讀。

## 尚無定論

- route widget 要用 `ConsumerWidget` 還是 `ConsumerStatefulWidget`：若沒有生命週期需求，建議預設 `ConsumerWidget`；只有真的需要 controller、`initState` 或訂閱外部資源時才升成 stateful。
- 對外 widget 要不要一律 `const` 建構子：若欄位全是 `final` 且父類允許，建議加上；若內含 runtime-only 預設值或 controller，先保留非 `const` 比較清楚。
