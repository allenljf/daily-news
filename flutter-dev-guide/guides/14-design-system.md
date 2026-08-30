# 14 · Material 3、設計 Token、無障礙與本地化

**這份管**：Material 3 theme、設計 token 的使用邊界、深淺色模式、字串與日期本地化、基本 accessibility 底線。
**這份不管**：widget 生命週期、列表效能、provider state 建模。

| 需要什麼 | 去哪 |
|---|---|
| 公開 widget API 與語意標籤 | `09-widget-api` |
| rebuild / image layout / 長清單 | `11-rendering-performance` |
| 路由標題、deep link、shell | `13-navigation` |
| 安全與 PII | `17-security` |

Flutter UI 一律站在 Material 3 theme 與本地化系統之上。畫面程式碼不應自行發明色票、間距和文案來源。

---

### material3-theme-at-app-entry · MUST

適用檔案：`lib/app/**/*.dart`、`lib/main.dart`
安全自動檢查：manual

App entry 必須以 Material 3 theme 包住整個 `MaterialApp.router`，深淺色、字體、component theme 都從同一個 theme composition root 提供。

✗ `MaterialApp.router(theme: ThemeData())`
✓ `MaterialApp.router(theme: buildLightTheme(), darkTheme: buildDarkTheme(), themeMode: themeMode, ...)`

例外：純 widget test 可用最小測試 theme，但必須明確包裝。

### no-hardcoded-design-values-in-ui · MUST_NOT

適用檔案：`lib/**/widgets/**/*.dart`、`lib/**/screens/**/*.dart`
安全自動檢查：regex（可掃描顏色十六進位、裸 `EdgeInsets.all(16)`、裸 `Radius.circular(12)`）+ manual

畫面程式碼不得硬寫顏色、主要間距、圓角與文字樣式。優先從 `Theme.of(context)`、`ColorScheme`、`TextTheme`、或專案 token extension 取值。

✗ `color: const Color(0xFF1976D2)`
✓ `color: Theme.of(context).colorScheme.primary`

例外：`EdgeInsets.zero`、`BorderRadius.zero`、`Colors.transparent` 這類零值或透明值。

### user-facing-string-localized · MUST

適用檔案：`lib/**/widgets/**/*.dart`、`lib/**/screens/**/*.dart`、`lib/l10n/*.arb`
安全自動檢查：regex（可掃描 `Text('...')`）+ manual

使用者可見文案必須來自 Flutter localization（`AppLocalizations` / `context.l10n`），不得在正式畫面硬寫字串，也不得手動拼接完整句子。

✗ `const Text('Refresh now')`
✓ `Text(context.l10n.refreshNow)`

例外：開發者用 log、debug banner、測試資料。

### supports-dark-mode-from-theme-only · MUST

適用檔案：`lib/app/**/*.dart`、`lib/**/widgets/**/*.dart`
安全自動檢查：manual

深淺色模式差異只能透過 theme 決定，不得在 feature widget 裡直接用 `Brightness` 分支硬切一套私有顏色。

✗ `isDark ? Colors.black : Colors.white`
✓ `Theme.of(context).colorScheme.surface`

例外：theme 建立處本身。

### touch-target-text-scale-and-semantics · MUST

適用檔案：`lib/**/widgets/**/*.dart`、`lib/**/screens/**/*.dart`
安全自動檢查：manual

互動元件至少保有約 48x48 logical pixels 觸控區；文字放大到 200% 時不得裁切；重要資訊與互動控制必須可被螢幕閱讀器理解。

✗ 固定高度 40 的文字按鈕、只有 icon 沒標籤的刪除按鈕
✓ `IconButton` / 足夠 padding、`minLines` / `softWrap`、必要時 `Semantics`

例外：純裝飾元素可排除語意。

### locale-aware-formatting-in-ui-boundary · SHOULD

適用檔案：`lib/**/widgets/**/*.dart`、`lib/**/screens/**/*.dart`、`lib/core/formatters/**/*.dart`
安全自動檢查：manual

日期、時間、數字與 plural 文案必須在接近 UI 的邊界依 locale 格式化，不要把已格式化的句子塞進 repository 或 provider state。

✗ provider state 直接存 `"8/30 08:00 更新"`
✓ provider state 存 `DateTime updatedAt`，widget 用 locale-aware formatter 呈現

例外：後端明確提供、且不再由 client 重新格式化的長文內容。

## 尚無定論

- 自訂 token 要用 `ThemeExtension` 還是獨立 design system package：兩者都行。單一 app 建議先用 `ThemeExtension`；只有跨 package 重用明顯時再抽成獨立套件。
- 若專案決定 light-only，也要在 app 層明文設定與文件記錄，而不是讓 feature widget 私自假設永遠不會進 dark mode。
