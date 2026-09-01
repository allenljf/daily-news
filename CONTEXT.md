# Daily News

個人每日新聞 App 的核心語言。此 context 同時供 Flutter client、Go backend 與每日擷取工作使用，避免三者對資料關係有不同解讀。

## Language

**Category**:
使用者定義的一組新聞意圖，由名稱、搜尋關鍵字、來源設定與特殊需求組成。
_Avoid_: Folder, channel, topic

**Source Setting**:
隸屬於一個 Category 的單一搜尋來源指示，可代表未指定網站、公開網站、或有官方 adapter 的平台。
_Avoid_: Provider, website field

**Article**:
經 canonical URL 與標準化標題去重後的一篇新聞內容；可被多個 Category 引用。
_Avoid_: News item, post, record

**Category Article**:
將一個 Article 納入一個 Category 的關聯，保有該次發現所用的 Source Setting 作為列表 tag。
_Avoid_: Duplicate article, category news table

**Ingestion Run**:
一次每日排程或手動觸發的批次執行紀錄，記錄來源、候選、去重、寫入與失敗結果；同一時間最多一個 Run 可執行。
_Avoid_: Cron, sync

**Suppressed Article**:
被使用者刪除且不應在往後擷取中重新顯示的 Article；以 soft delete 保留去重指紋。
_Avoid_: Expired article, hidden article
