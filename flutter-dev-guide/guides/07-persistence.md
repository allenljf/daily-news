# 07 · Persistence

這份管 shared_preferences、SQLite / Drift、local service 與 offline policy 的邊界。

| 想找 | 去哪 |
|---|---|
| Repository 的 SSOT 與 cache policy | [05-data-layer.md](05-data-layer.md) |
| model 邊界與 entity 不外流 | [01-architecture.md](01-architecture.md) |

---

## 一、shared_preferences 只放小型設定

`shared_preferences` 適合深色模式、首次啟動旗標、簡單 filter 這類小型設定。它不是新聞資料庫，也不是 secret store。

新版 API 優先考慮 async 介面；如果用有 cache 的 API，要知道它在多 isolate 或原生端改值時可能過期。

## 二、需要查詢、排序、關聯、批次更新時用 SQLite

當資料具有以下任一特性，就不要塞進 prefs：

1. 需要依條件查詢或排序
2. 需要局部更新而非整包覆寫
3. 有關聯或 migration
4. 需要離線列表、分頁或同步策略

這類資料走 SQLite；查詢與 transaction 複雜時可採 Drift。

## 三、local storage 也要藏在 service 後面

ViewModel 與 Widget 不直接操作 DB 或 prefs。local service 封裝單一儲存技術，Repository 再把它與 remote service 組合成產品層的 offline policy。

## 四、migration 與 offline policy

Schema 變更前先定 migration 與既有資料保留策略。offline-first 不是「先存再說」，而是先定清楚 TTL、同步、衝突與失敗回復。

---

### shared-preferences-only-for-small-settings · MUST · new-only+on-touch · manual

`shared_preferences` 只存小型設定，不存新聞清單、搜尋結果 cache、token、API key 或其他 secrets。

### local-storage-stays-behind-service · MUST_NOT · new-only+on-touch · regex

Widget 與 ViewModel 不直接持有 SQLite、Drift database 或 shared_preferences instance。這些都應藏在 local service 與 Repository 後面。

### persistence-change-needs-migration-plan · MUST · new-only+on-touch · manual

新增或修改 SQLite schema 前先寫清楚 migration 與資料保留策略。沒有 migration 計畫的 schema 變更，等同把既有資料交給運氣。

### repository-defines-offline-policy · MUST · new-only+on-touch · manual

local storage 只是能力，不是政策。真正決定 cache freshness、fallback、eviction 與 sync timing 的地方是 Repository。

---

## 主要來源

- Flutter SQLite cookbook: <https://docs.flutter.dev/cookbook/persistence/sqlite>
- Flutter SQL architecture recipe: <https://docs.flutter.dev/app-architecture/design-patterns/sql>
- Flutter offline-first architecture recipe: <https://docs.flutter.dev/app-architecture/design-patterns/offline-first>
- `shared_preferences` package page: <https://pub.dev/packages/shared_preferences>
