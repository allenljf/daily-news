# Daily News Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` (recommended) or `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 建立可獨立複製的 Flutter 開發準則庫，並交付一個由 Flutter、FastAPI、PostgreSQL、Cloud Run 與 GitHub Actions 組成的單一使用者每日新聞 App。

**Architecture:** Flutter 採 View/ViewModel + Repository/Service，Riverpod 只負責 composition 與 UI state，Dio 僅存在 remote service。FastAPI 以 identity、categories、news、ingestion 四個 module 提供 versioned HTTP interface；Cloud Run service 提供 App API，Cloud Run Job 執行 ingestion，GitHub Actions 以 OIDC/WIF 觸發 scheduled Job。

**Tech Stack:** Flutter Material 3, `flutter_riverpod`, Dio, `go_router`, `json_serializable`, Firebase Google Sign-In, FastAPI, Pydantic v2, SQLAlchemy 2, Alembic, PostgreSQL, Cloud Run, Secret Manager, GitHub Actions OIDC/WIF, Gemini Google Search grounding, GitHub REST API, YouTube Data API.

**Spec:** [每日新聞 App 需求與架構規格](../requirements/daily-news.md)

## Global Constraints

- 每次開始前讀 `CONTEXT.md`、`docs/requirements/daily-news.md` 與本檔；詞彙一律使用 Category、Source Setting、Article、Category Article、Ingestion Run。
- 僅實作 checkbox 尚未完成，且其 `Depends on` 的所有 task 已勾選的工作；規格改變先更新需求與本檔。
- 新增 Flutter code 必須遵守未來 `flutter-dev-guide/AGENTS.md`；新增 backend code 必須有單元或整合測試。
- Flutter 不得保存 server secret；LLM key、DB password、GitHub／Meta token 僅可由 Cloud Run 透過 Secret Manager 讀取。
- App API 一律驗證 Firebase ID token 與 `ALLOWED_USER_EMAIL`；Cloud Run Job 不開放給 Flutter 直接呼叫。
- `Article` 的 canonical URL hash 唯一；URL 不同時再比對 normalized title hash；刪除為全域 soft delete；預設期限 30 天。
- 列表固定最大 20 筆、以 `CategoryArticle.inserted_at DESC, article_id DESC` cursor pagination；不得使用 offset pagination。
- scheduled Run 使用 `scheduled:<taipei_date>` idempotency key；manual Run 使用 `manual:<request-id>`；系統同時最多一個 queued/running Run。
- 每一個完成的 task 都要勾選父項與已執行的子步驟，保留實際驗證指令與結果摘要在該 task 的「完成紀錄」。

---

## 使用方式

1. 尋找 `- [ ]` 的 task，檢查其 `Depends on` 是否全部為已勾選。
2. `Parallel: yes` 的 task 可以由不同 agent 同時執行；未標示者不可與同階段的 shared-file task 並行。
3. 一個 task 完成前依序完成其 Red、Green、Verify、Commit 子步驟；若 task 是純文件，Red 改為先加入可驗證的 checklist 或檢查命令。
4. 在「完成紀錄」填入日期、commit SHA（若有）、實際執行的檢查與結果；不得只寫「完成」。
5. 新對話在 `AGENTS.md` 建立前，必須由使用者訊息明確要求先讀這三份文件；Task G1 完成後由根目錄 `AGENTS.md` 強制此流程。

## 預定檔案責任

| 路徑 | 責任 |
|---|---|
| `AGENTS.md` | repo-level 任務入口與文件／task 接續規則。 |
| `flutter-dev-guide/` | 可獨立複製的 Flutter 開發準則與稽核工具。 |
| `backend/app/identity/` | Firebase token 驗證與 email allowlist。 |
| `backend/app/categories/` | Category／Source Setting 的 CRUD interface 與商業規則。 |
| `backend/app/news/` | Article read model、cursor、tag filter、永久與刪除。 |
| `backend/app/ingestion/` | Run lifecycle、來源 adapters、dedupe、Cloud Run Job entrypoint。 |
| `backend/alembic/` | PostgreSQL schema migrations。 |
| `apps/mobile/lib/core/` | App-wide routing、Dio、auth、theme、共用 types。 |
| `apps/mobile/lib/features/` | feature-first View、Riverpod controller、Repository contract。 |
| `infra/` | 非秘密的 GCP／GitHub deployment configuration 與操作文件。 |
| `.github/workflows/` | CI、deploy、每日 scheduled run workflow。 |

---

## Phase G — 任務入口與 Flutter 準則庫

本階段先建立後續 agent 一致遵循的入口與準則。G2、G3、G4 可平行，G5 整合後才驗證規則索引。

### G1: 建立 repo 級任務入口

**Depends on:** none  
**Parallel:** no — 會定義所有後續新對話的入口。  
**Files:** Create `AGENTS.md`; modify `docs/tasks/daily-news.md` 完成紀錄。

- [x] **Red:** 以 `rg --files -g 'AGENTS.md'` 確認根目錄尚無 agent 入口。
- [x] **Green:** 依 `writing-for-agents` skill 建立精簡 `AGENTS.md`，強制先讀 `CONTEXT.md`、需求規格、本 task 清單；要求挑選未阻塞 task、完成後勾選並記錄驗證；指向 `flutter-dev-guide/AGENTS.md` 與 backend／Flutter 對應文件。
- [x] **Verify:** 執行 `test -f AGENTS.md` 並人工確認它不重述需求規格、沒有秘密、沒有與本檔衝突的指示。
- [x] **Commit:** `git add AGENTS.md docs/tasks/daily-news.md && git commit -m "docs: add task execution entrypoint"`。

**Done when:** 新對話只讀 `AGENTS.md` 即能找到需求、任務和工作回寫規則。

**完成紀錄：**

- 2026-08-30：建立 repo 級入口；`rg --files -g 'AGENTS.md'` 僅列出 `android-dev-guide/AGENTS.md`，確認 root 尚無入口；`test -f AGENTS.md` 通過。已人工確認入口只含流程指引，不含秘密、需求規格副本或與本 task 清單衝突的指示。Commit SHA 見本 task 的 G1 report。

### G2: 建立 Flutter guide metadata、核心原則與架構文件

**Depends on:** G1  
**Parallel:** yes — 可與 G3、G4 並行。  
**Files:** Create `flutter-dev-guide/{README.md,AGENTS.md,rules.yaml}`, `flutter-dev-guide/guides/{00-principles.md,01-architecture.md,02-modularization.md,03-dependency-injection.md,04-domain-layer.md,05-data-layer.md,06-network.md,07-persistence.md}`.

- [x] **Red:** 在 `rules.yaml` 列出每份 guide 的 owner rule id，執行 YAML parser，確認尚未建立的 owner 會被 self-check 偵測。
- [x] **Green:** 寫入上述 8 份 guide；使用 Flutter 官方 View/ViewModel + Repository/Service 原則，規定 DTO／SQLite entity／domain／UI model 不跨層，Riverpod 作 composition root，Dio 僅可在 remote service，並記錄官方或套件一手來源。
- [x] **Verify:** `python3 -c 'import yaml; yaml.safe_load(open("flutter-dev-guide/rules.yaml"))'`；檢查每一個 rule id 只被一個 guide 宣告為 owner。
- [x] **Commit:** `git add flutter-dev-guide && git commit -m "docs: add Flutter architecture guide"`。

**Done when:** 新 Flutter feature 能依此判斷資料放置、Repository/Service 邊界、是否建立 use case 與 Riverpod wiring。

**完成紀錄：**

- 2026-08-30：建立 metadata、README、agent 入口與 00–07 guide。`python3 -c 'import yaml; yaml.safe_load(open("flutter-dev-guide/rules.yaml"))'` 通過；owner 自檢確認 89 個 rule id 均唯一且對應現有 guide。Commit `4484761`。

### G3: 建立 Flutter UI、非同步與導航指南

**Depends on:** G1  
**Parallel:** yes — 可與 G2、G4 並行。  
**Files:** Create `flutter-dev-guide/guides/{08-ui-state.md,09-widget-api.md,10-widget-state-and-lifecycle.md,11-rendering-performance.md,12-async-streams-isolates.md,13-navigation.md,14-design-system.md}`.

- [x] **Red:** 在 `rules.yaml` 先加入這 7 份文件的 rule owner 項目，確認同一 id 不重複。
- [x] **Green:** 寫入 Riverpod `AsyncValue` state、Widget public interface、dispose/lifecycle、rebuild/list keys、Future/Stream/isolate、`go_router`、Material 3/accessibility/localization 規範；每項指出適用的 Dart／Flutter 檔案和可安全自動檢查的程度。
- [x] **Verify:** `rg -n '^### [a-z0-9-]+ ·' flutter-dev-guide/guides/{08-ui-state,09-widget-api,10-widget-state-and-lifecycle,11-rendering-performance,12-async-streams-isolates,13-navigation,14-design-system}.md` 有對應條目，且 YAML owner 全部可找到。
- [x] **Commit:** `git add flutter-dev-guide && git commit -m "docs: add Flutter UI guide"`。

**Done when:** 新畫面能用指南決定 state、effect、導航、可重用 Widget 與 accessibility 寫法。

**完成紀錄：**

- 2026-08-30：建立 08–14 UI／非同步／導航 guide，並補齊非重複 owner entries。heading 檢查列出每份文件的 rule 條目；YAML 自檢確認所有 owner 存在且 id 無重複。Commit `6930b0e`。

### G4: 建立品質、採用與 agent skills 文件

**Depends on:** G1  
**Parallel:** yes — 可與 G2、G3 並行。  
**Files:** Create `flutter-dev-guide/guides/{15-testing.md,16-performance.md,17-security.md,18-analytics.md}`, `flutter-dev-guide/checklists/{new-feature.md,new-screen.md,new-api.md,refactor.md}`, `flutter-dev-guide/adoption/{getting-started.md,existing-project.md,integrating-ai-tools.md}`, `flutter-dev-guide/skills/{flutter-guide/SKILL.md,flutter-review/SKILL.md}`.

- [x] **Red:** 為每份 checklist 列出可機械驗證的最終檢查（analyse、test、rule checker）；為 skills 定義輸入與輸出格式。
- [x] **Green:** 寫入 testing、profile-mode performance、secret/PII/Firebase token、analytics adapter 指南；checklist 依「contract → data → state → view → test」安排；skills 只路由至必要 guide，不重述規則。
- [x] **Verify:** 依 `writing-skills` skill 驗證兩個 skill 的 front matter、觸發描述與連結路徑；所有 README 連結存在。
- [x] **Commit:** `git add flutter-dev-guide && git commit -m "docs: add Flutter quality and adoption guide"`。

**Done when:** guide 可用於新專案、既有專案和 AI agent 的實作／review。

**完成紀錄：**

- 2026-08-30：建立品質、採用、checklist 與兩個 router-only agent skills。front matter、description 與所有 30 個 Markdown 檔案的本地連結檢查皆通過。Commit `6d2c08c`。

### G5: 實作 Flutter guide 稽核器與 self-check

**Depends on:** G2, G3, G4  
**Parallel:** no — 需要完整 rules index。  
**Files:** Create `flutter-dev-guide/tools/check-rules.py`, `flutter-dev-guide/tools/test_check_rules.py`; modify `flutter-dev-guide/{README.md,rules.yaml}`.

- [x] **Red:** 在 `test_check_rules.py` 寫入三個失敗 fixture：UI 直接使用 Dio、Widget 直接讀 Repository、Dart 檔內 hardcoded secret；斷言輸出各自 rule id 與行號。
- [x] **Run Red:** `python3 -m unittest flutter-dev-guide.tools.test_check_rules -v`，確認因 `flutter-dev-guide/tools/check-rules.py` 尚不存在而失敗。
- [x] **Green:** 實作讀取 YAML、掃描 `.dart`／`pubspec.yaml`／`analysis_options.yaml`、支援 `--all`、`--staged`、`--diff`、`--files`、`--self-check` 和行內 `// guide-ignore: rule-id` 的 checker；並把 `fixed-delay-in-test`、`secret-in-client-code` 對齊為可單檔檢查的 regex 規則。
- [x] **Run Green:** 重跑單元測試與 `python3 flutter-dev-guide/tools/check-rules.py --self-check`，確認皆通過。
- [x] **Commit:** `git add flutter-dev-guide docs/tasks/daily-news.md && git commit -m "feat: add Flutter guide checker"`。

**Done when:** README 所列的 guide 自我檢查與 fixture test 均通過。

**完成紀錄：**

- 2026-08-30：先以 fixture-driven unittest 建立三個必要違規案例，並額外覆蓋 `// guide-ignore: rule-id`、只掃描支援檔型與 `--self-check`。Red：`python3 -m unittest flutter-dev-guide.tools.test_check_rules -v` 失敗，stderr 顯示 `can't open file '/Users/allen/SideProject/daily-news/flutter-dev-guide/tools/check-rules.py': [Errno 2] No such file or directory`，確認失敗原因是 checker 尚未建立。Green：完成 `flutter-dev-guide/tools/check-rules.py` 後，同一指令通過（`Ran 7 tests ... OK`）；`python3 flutter-dev-guide/tools/check-rules.py --self-check` 輸出 `Self-check passed: /Users/allen/SideProject/daily-news/flutter-dev-guide/rules.yaml`。Scope/spec 自 review：僅變更 `flutter-dev-guide/{README.md,rules.yaml,tools/check-rules.py,tools/test_check_rules.py}` 與本 task 記錄，未碰其他階段檔案。Commit `8229a0f`。Fix round 1：新增 custom `file` rule fixture，先重現 `guide-ignore` 對 `forbidden_regex` file rule 無效，再修正 checker 讓 file rule 與 regex rule 共用同一套 ignore 行為；`python3 -m unittest flutter-dev-guide.tools.test_check_rules -v` 更新為 `Ran 8 tests ... OK`。

---

## Phase B — Backend 基礎與安全 Category interface

本階段產生可本機測試的 FastAPI／PostgreSQL 最小垂直切片。B2 與 B3 可平行；B4 需要兩者。

### B1: 建立 backend 測試與設定骨架

**Depends on:** G1  
**Parallel:** yes — 可與 G2–G5 並行。  
**Files:** Create `backend/{pyproject.toml,.env.example,README.md}`, `backend/app/{__init__.py,main.py}`, `backend/app/core/{config.py,errors.py}`, `backend/tests/{conftest.py,test_health.py}`.

- [x] **Red:** 在 `test_health.py` 寫入 `GET /healthz` 必回 200 與 `{"status":"ok"}` 的 FastAPI `TestClient` 測試。
- [x] **Run Red:** `cd backend && uv run pytest tests/test_health.py -q`，預期失敗，因 app 未建立。
- [x] **Green:** 建立 Python 3.12+ 的 `pyproject.toml`（FastAPI、Pydantic v2、SQLAlchemy、Alembic、pytest、httpx、ruff），實作 `create_app()`、health router、環境設定及不含實值的 `.env.example`。
- [x] **Run Green:** `cd backend && uv run ruff check . && uv run pytest tests/test_health.py -q`，預期通過。
- [x] **Commit:** `git add backend && git commit -m "build: scaffold FastAPI backend"`。

**Interface produced:** `app.main.create_app() -> FastAPI`; `GET /healthz -> {"status": "ok"}`.

**完成紀錄：**

- 2026-08-30：先建立 health test；Red command 因尚無 `app` package 出現 `ModuleNotFoundError`。完成 FastAPI skeleton 後，`cd backend && uv run ruff check . && uv run pytest tests/test_health.py -q` 通過（1 passed）。Commit `0420e56`。

### B2: 建立 PostgreSQL schema、migration 與 repository session

**Depends on:** B1  
**Parallel:** yes — 可與 B3 並行。  
**Files:** Create `backend/alembic/`, `backend/app/db/{engine.py,models.py}`, `backend/tests/integration/test_schema.py`.

- [x] **Red:** 寫 integration test，檢查 migration 後存在 `categories`、`source_settings`、`articles`、`category_articles`、`ingestion_runs`、`ingestion_attempts` 和 spec 要求的 unique／cursor index。
- [x] **Run Red:** 使用 test PostgreSQL 執行 `cd backend && uv run pytest tests/integration/test_schema.py -q`，預期失敗。
- [x] **Green:** 以 SQLAlchemy 2 定義 model，Alembic 建立初始 migration；`ingestion_runs` 含 `trigger`、`idempotency_key` 與 status；database session 由設定注入。
- [x] **Run Green:** `cd backend && uv run alembic upgrade head && uv run pytest tests/integration/test_schema.py -q`。
- [x] **Commit:** `git add backend && git commit -m "feat: add daily news database schema"`。

**Interface produced:** `get_session() -> AsyncIterator[AsyncSession]`; initial Alembic revision.

**完成紀錄：**

- 2026-08-30：Red integration test 已在本 task 開始前建立。此輪第一次執行 `cd backend && uv run pytest tests/integration/test_schema.py -q` 為 `1 passed`，因工作樹已包含未追蹤的 B2 Green 草稿；不以 SQLite 替代 PostgreSQL。直接執行 `cd backend && uv run alembic upgrade head` 在未設定 `DATABASE_URL` 的本機環境依預期停止並提示設定連線字串。以 disposable `postgres:16-alpine` 注入 `DATABASE_URL` 執行同一 migration 指令，Alembic 成功套用 `20260830_0001`；`cd backend && uv run pytest tests/integration/test_schema.py -q` 為 `1 passed`，`cd backend && uv run ruff check .` 為 `All checks passed!`，`cd backend && uv run pytest tests/test_health.py -q` 為 `1 passed`（既有 Starlette/httpx deprecation warning），額外 `cd backend && uv run pytest -q` 為 `2 passed`。Task-scope review 未發現必要修正；只納入 backend 的 B2 檔案，保留未追蹤 `android-dev-guide/`。實作 Commit `a9eeafb`（`feat: add daily news database schema`）。

### B3: 實作 Firebase 身分驗證與單一 email allowlist

**Depends on:** B1  
**Parallel:** yes — 可與 B2 並行。  
**Files:** Create `backend/app/identity/{firebase.py,dependencies.py}`, `backend/tests/identity/test_dependencies.py`; modify `backend/app/core/{config.py,errors.py}`.

- [x] **Red:** 寫三個測試：缺少 Bearer token 回 401、invalid Firebase token 回 401、valid token 但 email 不等於 `ALLOWED_USER_EMAIL` 回 403。
- [x] **Run Red:** `cd backend && uv run pytest tests/identity/test_dependencies.py -q`，預期失敗。
- [x] **Green:** 定義 `VerifiedIdentity(uid: str, email: str)`；用 Firebase Admin SDK 的 token verifier adapter；建立 `require_allowed_identity` FastAPI dependency。測試以 fake verifier 注入，不連真 Firebase。
- [x] **Run Green:** `cd backend && uv run pytest tests/identity/test_dependencies.py -q`。
- [x] **Commit:** `git add backend && git commit -m "feat: secure API with Firebase allowlist"`。

**Interface produced:**

```python
async def require_allowed_identity(...) -> VerifiedIdentity: ...
```

**完成紀錄：**

- 2026-08-30：Red：`cd backend && uv run pytest tests/identity/test_dependencies.py -q` 在 identity module 尚未存在時以 `ModuleNotFoundError: No module named 'app.identity'` 失敗。Green：加入 Firebase Admin verifier adapter、`VerifiedIdentity` 和 `require_allowed_identity`，測試以 fake verifier 注入且不連 Firebase；同一指令通過。Task review 發現 Firebase certificate/network outage 被誤映射為 401，已在 fix round 將例外分類縮小為 Firebase 的 invalid-token family，並新增 certificate-fetch regression test；scoped re-review 確認修正且無新增 Critical／Important 問題。最終驗證：`cd backend && uv run pytest tests/identity/test_dependencies.py -q` 為 `5 passed`、`cd backend && uv run ruff check .` 為 `All checks passed!`、`cd backend && uv run pytest -q` 為 `7 passed`；測試僅有既有 Starlette/httpx deprecation warning。未修改或加入未追蹤 `android-dev-guide/`。Commits `69398c3`（`feat: secure API with Firebase allowlist`）與 `48d5b90`（`fix: preserve Firebase operational failures`）。

### B4: 交付 Category 與 Source Setting CRUD

**Depends on:** B2, B3  
**Parallel:** no — 寫入 schema 且每個 route 受 identity 保護。  
**Files:** Create `backend/app/categories/{schemas.py,repository.py,service.py,router.py}`, `backend/tests/categories/test_routes.py`; modify `backend/app/main.py`.

- [x] **Red:** 寫 route tests：建立含一個未指定網站與兩個額外 Source Settings 的 Category 回 201；空 name 回 422；更新可排序 Source Settings；刪除 Category 不刪除 Article；未授權為 401。
- [x] **Run Red:** `cd backend && uv run pytest tests/categories/test_routes.py -q`，預期失敗。
- [x] **Green:** 實作 `GET/POST/PATCH/DELETE /v1/categories`；`CreateCategoryRequest` 含 `name`、`search_keywords`、`special_requirements`、有序 `source_settings`。空白 Source Setting 不保存，Category name 必填。
- [x] **Run Green:** 重跑測試與 `cd backend && uv run ruff check .`。
- [x] **Commit:** `git add backend && git commit -m "feat: add category settings API"`。

**Interface produced:**

```python
class CreateCategoryRequest(BaseModel):
    name: str
    search_keywords: str | None = None
    special_requirements: str | None = None
    source_settings: list[SourceSettingInput]
```

**完成紀錄：**

- 2026-08-30：Red：新增的 route tests 在 Category router 尚未存在時以 5 failures（`404 Not Found`）確認 endpoint 缺失。Green：加入受 `require_allowed_identity` 保護的 `GET/POST/PATCH/DELETE /v1/categories`、DTO、Router → Service → Repository 分層與 PostgreSQL migration `20260830_0002`，為 `category_articles` 加上 soft-delete 欄位。Route tests 使用 disposable `postgres:16-alpine`、Alembic 與真 async session，不使用 SQLite 或 Firebase。Task review 發現 delete test 缺少 Category／Source Setting／Category Article soft-delete assertions，且與 schema test 重複 Docker fixture；fix round 抽出 `tests/postgres.py`，補上三種 `deleted_at` assertions、刪除後 GET omission 與 PATCH 404 coverage，scoped re-review 確認均已解決。最終驗證：`cd backend && uv run pytest tests/categories/test_routes.py tests/integration/test_schema.py -q` 為 `6 passed`、`cd backend && uv run ruff check .` 為 `All checks passed!`、`cd backend && uv run pytest -q` 為 `12 passed`；僅有既有 Starlette/httpx deprecation warning。未修改或加入未追蹤 `android-dev-guide/`。Commits `5694400`（`feat: add category settings API`）與 `3be1418`（`test: strengthen category PostgreSQL coverage`）。

---

## Phase N — News 讀取與手動更新 interface

本階段提供 Flutter 可先串接的 news contract；N1、N2 共用 Article schema，不可平行。

### N1: 實作 Article read model、cursor、tag filter 與保存操作

**Depends on:** B2, B3, B4  
**Parallel:** no.  
**Files:** Create `backend/app/news/{schemas.py,repository.py,service.py,router.py}`, `backend/tests/news/test_routes.py`; modify `backend/app/main.py`.

- [x] **Red:** 建立 21 個 Category Article fixture，測試第一頁恰為 20 筆、`next_cursor` 可取得第 21 筆、`sourceTagId` 只回該 tag、expired/deleted 不出現、永久與 soft delete 對所有 Category 生效。
- [x] **Run Red:** `cd backend && uv run pytest tests/news/test_routes.py -q`，預期失敗。
- [x] **Green:** 實作 spec 的三個 news GET/PATCH/DELETE routes，cursor encode `(inserted_at, article_id)`，每一 read query 固定排除 `articles.deleted_at IS NOT NULL` 和已到期非永久資料。
- [x] **Run Green:** 重跑 route tests；測試 invalid cursor 取得 400、missing Article 取得 404。
- [x] **Commit:** `git add backend && git commit -m "feat: add paginated news API"`。

**Interface produced:**

```python
class NewsPage(BaseModel):
    items: list[NewsListItem]
    next_cursor: str | None
```

**完成紀錄：**

- 2026-08-30：Red：新增 PostgreSQL-backed route contracts 後，`cd backend && uv run pytest tests/news/test_routes.py -q` 為 4 failures；所有失敗均為尚未註冊的 N1 endpoint 回 `404`，符合預期。Green：加入受 Firebase dependency 保護的 `GET /v1/categories/{category_id}/news`、`GET /v1/categories/{category_id}/news/{article_id}`、`PATCH/DELETE /v1/news/{article_id}`，以 `(inserted_at, article_id)` base64 JSON cursor 實作 keyset pagination，並在所有 read query 排除 soft-deleted 與已到期非永久 Article。Run Green：同一指定 route test 為 `4 passed`；全量 `cd backend && uv run pytest -q` 為 `16 passed`，`cd backend && uv run ruff check .` 為 `All checks passed!`。全量收集首次揭露 `test_routes.py` 與既有 Category test 同名造成 import mismatch，已把新的 `tests/news/` 設為 package，僅消除收集命名衝突。唯一 warning 為既有 Starlette/httpx deprecation warning。保留未追蹤 `android-dev-guide/` 不變。Backend commit `b72fb4b`（`feat: add paginated news API`）。

### N2: 實作 Ingestion Run status 與手動觸發 HTTP interface

**Depends on:** B2, B3  
**Parallel:** yes — 可與 N1 並行。  
**Files:** Create `backend/app/ingestion/{schemas.py,run_repository.py,run_service.py,router.py}`, `backend/tests/ingestion/test_manual_runs.py`; modify `backend/app/main.py`.

- [x] **Red:** 寫 route tests：`GET /v1/ingestion-runs/latest` 在無成功 Run 時回 `last_successful_at: null`；`POST` 回 202 與 manual Run；已有 queued/running Run 時回相同 id；已完成 scheduled Run 不阻止新的 manual Run。
- [x] **Run Red:** `cd backend && uv run pytest tests/ingestion/test_manual_runs.py -q`，預期失敗。
- [x] **Green:** 實作 Run state machine (`queued → running → succeeded|failed`) 與 transaction-safe active-run lookup；將 Job launch 包在 `JobLauncher` interface，測試注入 fake launcher。
- [x] **Run Green:** 重跑測試；斷言 `POST /v1/ingestion-runs` 只回 202，不等待 Job 完成。
- [x] **Commit:** `git add backend && git commit -m "feat: add manual ingestion runs API"`。

**Interface produced:**

```python
class JobLauncher(Protocol):
    async def launch(self, run_id: UUID) -> None: ...
```

**完成紀錄：**

- 2026-08-30：Red：`cd backend && uv run pytest tests/ingestion/test_manual_runs.py -q` 在 ingestion module 尚未存在時以 `ModuleNotFoundError: No module named 'app.ingestion'` 停於 collection，確認契約尚未實作。Green：加入 authenticated `GET /v1/ingestion-runs/latest` 與 `POST /v1/ingestion-runs`、`queued → running → succeeded|failed` state transitions、PostgreSQL transaction-scoped advisory lock 以序列化 active Run lookup，以及可由測試覆寫的 `JobLauncher` interface。route tests 以 fake launcher 與 disposable PostgreSQL 驗證無成功 Run、202 manual Run、active Run 合併與 completed scheduled Run 後建立新的 manual Run。Run Green：指定測試為 `3 passed`；全量 `cd backend && uv run pytest -q` 為 `19 passed`，`cd backend && uv run ruff check .` 為 `All checks passed!`。唯一 warning 為既有 Starlette/httpx deprecation warning；未修改未追蹤 `android-dev-guide/`。Backend commit `06a2d3b`（`feat: add manual ingestion runs API`）。

---

## Phase I — Ingestion pipeline

本階段讓 Job 真正填入 Article。I1、I2、I3 有明確 seam；I2/I3 可平行，I4 組裝它們。

### I1: 建立來源 adapter contract、URL／標題 normalizer 與 dedupe service

**Depends on:** B2  
**Parallel:** yes — 可與 N1、N2 並行。  
**Files:** Create `backend/app/ingestion/{sources.py,normalization.py,dedupe.py}`, `backend/tests/ingestion/{test_normalization.py,test_dedupe.py}`.

- [x] **Red:** 測試 URL 移除 tracking parameters 後 hash 相同；大小寫／空白不同的 title hash 相同；不同 canonical URL 但相同 title 判為 duplicate；被 soft-deleted Article 仍判為 duplicate。
- [x] **Run Red:** `cd backend && uv run pytest tests/ingestion/test_normalization.py tests/ingestion/test_dedupe.py -q`，預期失敗。
- [x] **Green:** 定義 `CandidateArticle`、`SourceAdapter.search(request) -> list[CandidateArticle]`，每次 adapter 輸出限制為 10；dedupe service 將 canonical URL 放在 title 之前檢查。
- [x] **Run Green:** 重跑 tests，並檢查 CandidateArticle 沒有 database 或 FastAPI type。
- [x] **Commit:** `git add backend && git commit -m "feat: add ingestion source and dedupe contracts"`。

**Interface produced:**

```python
class SourceAdapter(Protocol):
    async def search(self, request: SourceSearchRequest) -> list[CandidateArticle]: ...
```

**完成紀錄：**

- 2026-08-31：Red：`cd backend && uv run pytest tests/ingestion/test_normalization.py tests/ingestion/test_dedupe.py -q` 因 `app.ingestion.normalization` 與 `app.ingestion.dedupe` 尚不存在而有 2 collection errors。Green：新增 URL canonicalization（移除 fragment 與 tracking parameters）、SHA-256 URL／標題 fingerprints、純 Python `CandidateArticle`／`SourceSearchRequest`／`SourceAdapter` contract、每 Source Setting 最多 10 筆的 limiter，及以 URL 優先、title 次之的可注入 fingerprint lookup dedupe service。測試將 soft-deleted Article 視為 lookup 可見的 fingerprint，確保它仍抑制重新擷取。Run Green：指定測試為 `5 passed`；`rg -n '(FastAPI|sqlalchemy|app\\.db)' backend/app/ingestion/sources.py` 無輸出；全量 `cd backend && uv run pytest -q` 為 `24 passed`，`cd backend && uv run ruff check .` 為 `All checks passed!`。唯一 warning 為既有 Starlette/httpx deprecation warning；未修改未追蹤 `android-dev-guide/`。Backend commit `7ae6647`（`feat: add ingestion source and dedupe contracts`）。

### I2: 實作公開網站 Gemini adapter

**Depends on:** I1  
**Parallel:** yes — 可與 I3 並行。  
**Files:** Create `backend/app/ingestion/gemini_adapter.py`, `backend/tests/ingestion/test_gemini_adapter.py`.

- [x] **Red:** 使用 fake Gemini client 測試 prompt 包含 Category、keywords、Source Setting、特殊需求；拒絕沒有公開 URL 或 citation 的候選；輸出最多 10 筆。
- [x] **Run Red:** `cd backend && uv run pytest tests/ingestion/test_gemini_adapter.py -q`，預期失敗。
- [x] **Green:** 在 adapter 內注入 Gemini client，使用 Google Search grounding 與必要時 URL Context；只接收 public 可驗證 URL，保存 citations，不嘗試讀登入、付費牆或動態社群內容。
- [x] **Run Green:** 重跑 test；確認 logging 不輸出完整 prompt 或 API key。
- [x] **Commit:** `git add backend && git commit -m "feat: add grounded web news adapter"`。

**完成紀錄：**

- 2026-08-31：Red：`cd backend && uv run pytest tests/ingestion/test_gemini_adapter.py -q` 因 adapter module 尚不存在而以 `ModuleNotFoundError` 停於 collection。Green：新增 SDK-neutral、可注入的 Gemini client contract；request 明確啟用 Google Search，僅對 public HTTP Source Setting 提供 URL Context，並將 Category、keywords、Source Setting、特殊需求寫入 prompt。adapter 過濾缺少 public HTTP URL 或 citation 的結果，保存 citation 並限制為 10 筆；不寫入或 log API key。Run Green：指定測試 `2 passed`；全量 `cd backend && uv run pytest -q` 為 `26 passed`，ruff 通過。唯一 warning 為既有 Starlette/httpx deprecation warning；未修改未追蹤 `android-dev-guide/`。Backend commit `9692c26`（`feat: add grounded web news adapter`）。

### I3: 實作 YouTube、GitHub 與受限 Meta adapter registry

**Depends on:** I1  
**Parallel:** yes — 可與 I2 並行。  
**Files:** Create `backend/app/ingestion/{youtube_adapter.py,github_adapter.py,meta_adapter.py,registry.py}`, `backend/tests/ingestion/test_registry.py`.

- [x] **Red:** 測試 `youtube.com` route 至 YouTube Data API adapter、`github.com` route 至 GitHub REST adapter、Meta URL 在未設定 token 時回明確「未啟用」結果而非 fallback crawler。
- [x] **Run Red:** `cd backend && uv run pytest tests/ingestion/test_registry.py -q`，預期失敗。
- [x] **Green:** adapter 各自注入 HTTP client 與 token provider；GitHub public request 允許無 token 的受限模式，YouTube／Meta 只有對應 key/token 存在才啟用；不得使用 Gemini 繞過平台權限。
- [x] **Run Green:** 重跑 registry tests；確認任何 disabled adapter 不會讓整個 Run 失敗。
- [x] **Commit:** `git add backend && git commit -m "feat: add platform source adapters"`。

**完成紀錄：**

- 2026-08-31：Red：`cd backend && uv run pytest tests/ingestion/test_registry.py -q` 因 registry module 尚不存在而以 `ModuleNotFoundError` 停於 collection。Green：新增 host-based registry、注入 HTTP client/token provider 的 YouTube、GitHub、Meta adapter boundaries；GitHub 可無 token 運行，YouTube／Meta 未設定憑證會成為明確 disabled result，Meta 不會 fallback 至 Gemini。Run Green：registry tests `2 passed`；全量 backend `28 passed` 且 ruff 通過。Backend commit `8ac6d0b`（`feat: add platform source adapters`）。

### I4: 實作 Cloud Run Job ingestion orchestrator

**Depends on:** I2, I3, N2  
**Parallel:** no — 組裝 Run state、registry 與資料寫入。  
**Files:** Create `backend/app/jobs/daily_news.py`, `backend/app/ingestion/orchestrator.py`, `backend/tests/ingestion/test_orchestrator.py`; modify `backend/app/ingestion/run_service.py`.

- [x] **Red:** 建立兩 Category、兩 Source Settings 的 fake adapters；測試每來源至多 10 candidate、重複不新增、同 Article 可關聯兩 Category、單一 adapter failure 僅標記 Attempt、Run 最後正確統計／狀態。
- [x] **Run Red:** `cd backend && uv run pytest tests/ingestion/test_orchestrator.py -q`，預期失敗。
- [x] **Green:** 實作 `run_ingestion(run_id)`；將 Run 標為 running，逐一處理設定與 transaction 寫 Article/CategoryArticle/Attempt，最後標為 succeeded 或 failed；Cloud Run Job entrypoint 接收 `RUN_ID`。
- [x] **Run Green:** 重跑 test 並執行 `cd backend && uv run python -m app.jobs.daily_news --help`。
- [x] **Commit:** `git add backend && git commit -m "feat: add daily news ingestion job"`。

**完成紀錄：**

- 2026-08-31：Red：orchestrator contract 在 module 尚不存在時以 `ModuleNotFoundError` 失敗。Green：新增 per-source isolation／10 candidate quota／global Article dedupe 的 orchestrator，以及 SQLAlchemy store，transaction 內寫入 Article、Category Article、Ingestion Attempt 與 Run counters；同一 Article 可連結兩個 Category。Job module 接收 `RUN_ID`／`--run-id`，並提供可注入 session/work composition 的 `run_job`。Run Green：`tests/ingestion/test_orchestrator.py` 為 `2 passed`，`python -m app.jobs.daily_news --help` 通過；完整 suite 以 verbose run 通過前 28/30、補跑最後 2 個為 `2 passed`，ruff 通過。唯一 warning 為既有 Starlette/httpx deprecation warning。 

---

## Phase F — Flutter App 基線與功能

F1 先產生 App，F2/F3 共享 core contracts 可順序執行；F4、F5 在 interface 穩定後可平行。

### F1: 建立 Flutter app、Firebase 與 code generation 基線

**Depends on:** G5, B3  
**Parallel:** no — 所有 mobile task 的基礎。  
**Files:** Create `apps/mobile/` via `flutter create`, `apps/mobile/lib/{app.dart,main.dart}`, `apps/mobile/analysis_options.yaml`, `apps/mobile/pubspec.yaml`, `apps/mobile/test/app_test.dart`.

- [x] **Red:** 將初始 widget test 改為期待受保護 app loading screen；`flutter test` 應因 app root 尚未建立而失敗。
- [x] **Run Red:** `cd apps/mobile && flutter test test/app_test.dart`。
- [x] **Green:** 建立 Material 3 root、`ProviderScope`、`flutter_lints`、Riverpod、Dio、go_router、Firebase core/auth、json serialization dependencies；Firebase config 僅使用 client config，不加入 service secret。
- [x] **Run Green:** `cd apps/mobile && dart run build_runner build --delete-conflicting-outputs && flutter analyze && flutter test`。
- [x] **Commit:** `git add apps/mobile && git commit -m "build: scaffold Flutter news app"`。

**完成紀錄：**

- 2026-08-31：Red：以受保護 loading screen 的 widget test 取代初始測試；修正測試本身缺少 Material 型別匯入後，`cd apps/mobile && flutter test test/app_test.dart` 依預期失敗，原因是畫面未顯示「正在載入…」。Green：以 `flutter create --empty --platforms=android,ios` 建立 App，加入 Material 3 `DailyNewsApp`、`ProviderScope`、loading screen，並加入 Riverpod、Dio、go_router、Firebase core/auth、JSON serialization 與 codegen 依賴；未加入 Firebase 設定、server secret 或真實憑證。Run Green：`cd apps/mobile && dart run build_runner build --delete-conflicting-outputs` 成功（目前 build_runner 將該相容參數標為已忽略），`flutter analyze` 為 `No issues found!`，`flutter test` 為 `1 passed`，`python3 ../../flutter-dev-guide/tools/check-rules.py --files lib/app.dart lib/main.dart test/app_test.dart` 通過。Commit `ccb4ae8`（`build: scaffold Flutter news app`）。

### F2: 實作 Flutter auth、Dio client 與 typed API transport

**Depends on:** F1, B4, N1, N2  
**Parallel:** no — 所有 feature Repository 依賴它。  
**Files:** Create `apps/mobile/lib/core/{auth/,http/,api/}`, `apps/mobile/test/core/test_dio_client.dart`.

- [ ] **Red:** 測試 Dio auth interceptor 在有 Firebase token 時加 Bearer header、401 時觸發 sign-out、Problem Details 轉成 typed `ApiFailure`、cursor page DTO 能 decode。
- [ ] **Run Red:** `cd apps/mobile && flutter test test/core/test_dio_client.dart`，預期失敗。
- [ ] **Green:** 實作 `AuthRepository`、Google Sign-In screen flow、Dio provider、cancel token、API DTO；只讓 Remote Service 使用 Dio，Repository 對上層回 domain models／typed failure。
- [ ] **Run Green:** `cd apps/mobile && dart run build_runner build --delete-conflicting-outputs && flutter analyze && flutter test test/core/test_dio_client.dart`。
- [ ] **Commit:** `git add apps/mobile && git commit -m "feat: add mobile auth and API client"`。

**完成紀錄：**

### F3: 實作 App router、theme 與首頁更新狀態

**Depends on:** F2  
**Parallel:** no — 首頁是所有 feature 的 route entry。  
**Files:** Create `apps/mobile/lib/core/{routing/,theme/}`, `apps/mobile/lib/features/home/{application/,presentation/,data/}`, `apps/mobile/test/features/home/home_screen_test.dart`.

- [ ] **Red:** Widget tests 期待首頁顯示「尚未更新」、成功 Run 的 locale-aware 時間、立即更新按鈕、以及 queued/running 時 disabled 狀態。
- [ ] **Run Red:** `cd apps/mobile && flutter test test/features/home/home_screen_test.dart`，預期失敗。
- [ ] **Green:** 以 `go_router` 建 route；Home controller 讀 `GET /ingestion-runs/latest`，只把 View state 暴露為 immutable Riverpod state；所有可見字串由 localization resource 提供。
- [ ] **Run Green:** `cd apps/mobile && flutter analyze && flutter test test/features/home/home_screen_test.dart`。
- [ ] **Commit:** `git add apps/mobile && git commit -m "feat: add home refresh status"`。

**完成紀錄：**

### F4: 實作 Category 首頁與動態設定 Bottom Sheet

**Depends on:** F2, F3  
**Parallel:** yes — 可與 F5 並行，只共享 API contract。  
**Files:** Create `apps/mobile/lib/features/categories/{application/,data/,presentation/}`, `apps/mobile/test/features/categories/category_sheet_test.dart`.

- [ ] **Red:** Widget tests：無 Category 只顯示新增方塊；有 N 個 Category 時最後一格是新增；Bottom Sheet 有名稱／關鍵字／預設來源／特殊需求；新增按鈕開 dialog 並新增可編輯 Source Setting；空 source 不送 API。
- [ ] **Run Red:** `cd apps/mobile && flutter test test/features/categories/category_sheet_test.dart`，預期失敗。
- [ ] **Green:** 建立 Category Repository、Riverpod controller 與 stateless content widgets；儲存成功關閉 sheet 並刷新 Categories；欄位錯誤與 submit loading 防止重複送出。
- [ ] **Run Green:** `cd apps/mobile && flutter analyze && flutter test test/features/categories/category_sheet_test.dart`。
- [ ] **Commit:** `git add apps/mobile && git commit -m "feat: add category configuration UI"`。

**完成紀錄：**

### F5: 實作新聞列表、詳情、tag filter 與手動更新 dialog

**Depends on:** F2, F3  
**Parallel:** yes — 可與 F4 並行。  
**Files:** Create `apps/mobile/lib/features/news/{application/,data/,presentation/}`, `apps/mobile/test/features/news/{news_list_test.dart,manual_refresh_dialog_test.dart}`.

- [ ] **Red:** Widget tests：列表初次讀取 20 筆、scroll 使用 `next_cursor`、tag filter 重置 cursor、永久／刪除 action 更新畫面；立即更新 dialog 說明背景工作，確認後送 POST 並顯示 queued/running。
- [ ] **Run Red:** `cd apps/mobile && flutter test test/features/news/news_list_test.dart test/features/news/manual_refresh_dialog_test.dart`，預期失敗。
- [ ] **Green:** 實作 News Repository、cursor controller、detail view、source tag chips、Article mutation 和 ManualRun controller；不得在 Widget 直接呼叫 Dio 或 Repository。
- [ ] **Run Green:** `cd apps/mobile && flutter analyze && flutter test test/features/news`。
- [ ] **Commit:** `git add apps/mobile && git commit -m "feat: add news feed and manual refresh"`。

**完成紀錄：**

### F6: 建立 Flutter end-to-end 驗收流程

**Depends on:** F4, F5, I4  
**Parallel:** no — 需要完整 API workflow。  
**Files:** Create `apps/mobile/integration_test/daily_news_flow_test.dart`, `apps/mobile/test_support/fake_api_server.dart`.

- [ ] **Red:** 寫 integration scenario：登入 → 建立 Category → 手動更新確認 → 顯示 run queued → 顯示新 Article → 設永久 → 刪除。
- [ ] **Run Red:** `cd apps/mobile && flutter test integration_test/daily_news_flow_test.dart`，預期在缺 UI 元件或 fake server routes 時失敗。
- [ ] **Green:** 完成 fake API transport 與所有 UI accessibility keys；測試不連 production Firebase、GCP 或 LLM。
- [ ] **Run Green:** `cd apps/mobile && flutter test integration_test/daily_news_flow_test.dart`。
- [ ] **Commit:** `git add apps/mobile && git commit -m "test: add daily news integration flow"`。

**完成紀錄：**

---

## Phase O — GCP、GitHub Actions 與上線驗收

O1 先把非秘密設定與權限寫成可審查文件，再容器化、部署、排程。任何真實帳號、project、token 設定均由使用者完成，不寫入 repo。

### O1: 建立 GCP／GitHub secrets 與 WIF 操作文件

**Depends on:** G1, B1  
**Parallel:** yes — 可與 B2–I4、F1–F6 並行。  
**Files:** Create `infra/docs/{gcp-setup.md,github-variables.md,secret-inventory.md}`.

- [ ] **Red:** 列出必要設定並執行 `rg -n '(GEMINI_API_KEY|DB_PASSWORD|GITHUB_NEWS_TOKEN|YOUTUBE_API_KEY|ALLOWED_USER_EMAIL)' infra/docs`；確認每個值都有位置、用途、是否可選與不得放置處。
- [ ] **Green:** 文件化 Cloud SQL、Secret Manager、Cloud Run service/job service accounts、Firebase project、WIF provider、GitHub Variables、最小 IAM role；明確說明 GitHub Action 使用 `id-token: write` 而非 JSON key。
- [ ] **Verify:** 人工逐項對照需求規格第 9 節；文件不含真實 project id、secret 值、email 或 token。
- [ ] **Commit:** `git add infra/docs && git commit -m "docs: add GCP and secret setup guide"`。

**Done when:** 使用者可在 GCP／GitHub UI 完成所有外部設定，而無需猜測 token 名稱或權限。

**完成紀錄：**

### O2: 容器化 FastAPI service 與 Cloud Run Job

**Depends on:** B4, N1, N2, I4  
**Parallel:** no — 同一 image 必須同時服務 HTTP 與 Job entrypoint。  
**Files:** Create `backend/{Dockerfile,.dockerignore}`, `backend/scripts/{serve.sh,run-job.sh}`, `backend/tests/test_container_contract.py`.

- [ ] **Red:** container contract test 驗證 `serve.sh` 啟動 uvicorn，`run-job.sh` 呼叫 `python -m app.jobs.daily_news`，且兩者使用同一 image／環境設定名稱。
- [ ] **Run Red:** `cd backend && uv run pytest tests/test_container_contract.py -q`，預期失敗。
- [ ] **Green:** 寫 multi-stage 或精簡 Python image、non-root runtime、health endpoint、Job entry script；映像不 baked-in secret。
- [ ] **Run Green:** `docker build -t daily-news-backend:test backend && docker run --rm daily-news-backend:test python -m app.jobs.daily_news --help`。
- [ ] **Commit:** `git add backend && git commit -m "build: containerize API and ingestion job"`。

**完成紀錄：**

### O3: 建立部署與每日排程 workflows

**Depends on:** O1, O2  
**Parallel:** no — 需要已知 image、WIF 與 Cloud Run resource 名稱。  
**Files:** Create `.github/workflows/{ci.yml,deploy.yml,daily-ingestion.yml}`, `infra/cloud-run/{service.yaml,job.yaml}`.

- [ ] **Red:** 以 `actionlint` 檢查 workflow；`daily-ingestion.yml` 必須包含 `cron: '0 0 * * *'`、`permissions: id-token: write` 與 Job execution command，初始檢查預期因檔案不存在而失敗。
- [ ] **Green:** CI 執行 backend ruff/pytest、Flutter analyse/test、guide self-check；deploy workflow 透過 `google-github-actions/auth` 的 WIF 部署 service/job；daily workflow 僅觸發 Job，不帶 DB 或 LLM secret。
- [ ] **Run Green:** `actionlint .github/workflows/*.yml` 與每個 manifest 的 schema/lint check；人工確認 GitHub Variables 與 O1 名稱一致。
- [ ] **Commit:** `git add .github infra && git commit -m "ci: add Cloud Run deployment and daily job"`。

**完成紀錄：**

### O4: 執行受控 staging smoke test

**Depends on:** O3, F6  
**Parallel:** no — 真實外部資源驗收。  
**Files:** Create `docs/operations/staging-smoke-test.md`; modify `docs/tasks/daily-news.md` 完成紀錄。

- [ ] **Red:** 在 smoke test 文件先列出失敗判準：WIF 驗證失敗、Firebase allowlist 外帳號可讀資料、Run 沒有結束狀態、Article 沒有 citation、手動 Run 平行重複執行。
- [ ] **Green:** 由使用者設定 O1 所列 resource 與秘密後，部署 staging；以 allowlisted account 驗證登入、Category、manual Run、latest status、Article list、永久、刪除；以第二次並行 POST 驗證 active Run 合併。
- [ ] **Verify:** 檢查 Cloud Run／workflow log、`ingestion_runs`／`ingestion_attempts` 統計與 mobile integration results；記錄 pass/fail 和 run id，但不記錄 token／URL query secret。
- [ ] **Commit:** `git add docs/operations docs/tasks/daily-news.md && git commit -m "docs: record staging smoke test"`。

**Done when:** staging 以真正的 WIF、Cloud Run、Cloud SQL、Firebase 與至少一個公開來源完成端到端工作，且全部失敗判準均未發生。

**完成紀錄：**

---

## Plan Self-Review

| Spec requirement | Covered by |
|---|---|
| Flutter guide 的結構、規則、skills、checker、adoption | G1–G5 |
| 單一使用者 Firebase allowlist | B3, F1–F2, O1, O4 |
| Category Bottom Sheet、動態來源、首頁方塊 | B4, F4 |
| Article 關聯、去重、30 天、永久與刪除 | B2, N1, I1, I4 |
| 來源每組最多 10、平台限制與 citations | I1–I4 |
| 20 筆 cursor、tag filter、詳情 | N1, F5 |
| 首頁更新時間、立即更新 dialog、單一 active Run | N2, F3, F5 |
| GitHub 08:00、WIF、Cloud Run Job、Secret Manager | O1–O4 |
| 可跨新對話執行的文件與 checkbox protocol | G1、全檔 task metadata |

Self-review completed on 2026-08-30:

- 所有需求至少對應一項 task。
- task id、route、domain type 與 `Depends on` 名稱在全檔一致。
- 本計劃沒有未決佔位詞；外部實際值由使用者在 GCP／GitHub 設定，文件只引用變數名稱。
