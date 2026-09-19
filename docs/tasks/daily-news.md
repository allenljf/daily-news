# Daily News Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` (recommended) or `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 建立可獨立複製的 Flutter 開發準則庫，並交付一個由 Flutter、Go、PostgreSQL、Cloud Run 與 GitHub Actions 組成的單一使用者每日新聞 App。

**Architecture:** Flutter 採 View/ViewModel + Repository/Service，Riverpod 只負責 composition 與 UI state，Dio 僅存在 remote service。Go 以 `net/http`、明確 composition root、`database/sql` + pgx adapter 與 Firebase Admin Go SDK 提供 versioned HTTP interface；Cloud Run service 提供 App API，Cloud Run Job 執行 ingestion，GitHub Actions 以 OIDC/WIF 觸發 scheduled Job。Python task 的完成紀錄是歷史證據；Go replacement phase 才是目前 backend 實作路徑。

**Tech Stack:** Flutter Material 3, `flutter_riverpod`, Dio, `go_router`, `json_serializable`, Firebase Google Sign-In, Go standard library (`net/http`, `context`, `encoding/json`, `errors`, `testing/httptest`), `database/sql` + pgx adapter, Firebase Admin Go SDK, golang-migrate, PostgreSQL, Cloud Run, Secret Manager, GitHub Actions OIDC/WIF, SerpApi Google News, GitHub REST API, YouTube Data API.

**Spec:** [每日新聞 App 需求與架構規格](../requirements/daily-news.md)

## Global Constraints

- 每次開始前讀 `CONTEXT.md`、`docs/requirements/daily-news.md` 與本檔；詞彙一律使用 Category、Source Setting、Article、Category Article、Ingestion Run。
- 僅實作 checkbox 尚未完成，且其 `Depends on` 的所有 task 已勾選的工作；規格改變先更新需求與本檔。
- 新增 Flutter code 必須遵守未來 `flutter-dev-guide/AGENTS.md`；Go backend code 必須遵守 `backend/AGENTS.md`、有單元或整合測試，並在完成前執行 `gofmt`、`go vet ./...`、`go test ./...`。
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
| `backend/cmd/` | Go API 與 Cloud Run Job composition roots。 |
| `backend/internal/identity/` | Firebase token 驗證與 email allowlist。 |
| `backend/internal/category/` | Category／Source Setting 的 CRUD 商業規則與 SQL repository。 |
| `backend/internal/news/` | Article read model、cursor、tag filter、永久與刪除。 |
| `backend/internal/ingestion/` | Run lifecycle、來源 adapters、dedupe、Cloud Run Job orchestration。 |
| `backend/migrations/` | golang-migrate PostgreSQL migration representation。 |
| `docs/contracts/daily-news.openapi.json` | Flutter 與 Go 共用、checked-in 的 HTTP/OpenAPI compatibility contract。 |
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

## Phase B — Python/FastAPI backend 基礎與安全 Category interface（完成歷史）

本階段記錄已完成的 Python/FastAPI／PostgreSQL 最小垂直切片，作為 Go parity 的行為證據；不回寫或重作其完成歷史。B2 與 B3 可平行；B4 需要兩者。

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

- [x] **Red:** 測試 Dio auth interceptor 在有 Firebase token 時加 Bearer header、401 時觸發 sign-out、Problem Details 轉成 typed `ApiFailure`、cursor page DTO 能 decode。
- [x] **Run Red:** `cd apps/mobile && flutter test test/core/test_dio_client.dart`，預期失敗。
- [x] **Green:** 實作 `AuthRepository`、Google Sign-In screen flow、Dio provider、cancel token、API DTO；只讓 Remote Service 使用 Dio，Repository 對上層回 domain models／typed failure。
- [x] **Run Green:** `cd apps/mobile && dart run build_runner build --delete-conflicting-outputs && flutter analyze && flutter test test/core/test_dio_client.dart`。
- [x] **Commit:** `git add apps/mobile && git commit -m "feat: add mobile auth and API client"`。

**完成紀錄：**

- 2026-08-31：Red：先新增 `test/core/test_dio_client.dart`，以真實 Dio interceptor 與受控 `HttpClientAdapter` fixture 驗證 Bearer Firebase ID token、401 sign-out、Problem Details → typed `ApiFailure`，並以 literal fixture 驗證 generic cursor page DTO；`cd apps/mobile && flutter test test/core/test_dio_client.dart` 因 `core/{auth,http,api}` 尚不存在而在 compilation 失敗，符合缺少 F2 contract 的預期。Green：新增 SDK-independent `AuthRepository`／`AuthGateway`／`AuthUser`，Firebase + Google Sign-In adapter、Riverpod auth state/controller、sign-in screen 與 `AuthGate`；新增有明確 timeout、HTTPS base URL、Bearer interceptor 與 401 session termination 的 Dio client/provider，以及不暴露 Dio 的 `ApiClient`、`ApiCancelToken`、generated `ProblemDetailsDto`／`CursorPageDto` 與 typed `ApiFailure`。token 僅由 Firebase Auth SDK 於 request 時取得，未自行持久化或 log；未加入 Firebase Admin／server secret。新增 `google_sign_in ^7.2.0`，並將既有 `json_annotation` constraint 更新為 `^4.12.0` 以符合 codegen 要求。Run Green：精確命令 chain `dart run build_runner build --delete-conflicting-outputs && flutter analyze && flutter test test/core/test_dio_client.dart` exit 0；build_runner 成功且寫入 0 個變更（該版本提示相容參數已移除並忽略），analyze 為 `No issues found!`，F2 tests 為 `4 passed`。額外 regression `flutter test` 為既有 app test `1 passed`；`python3 flutter-dev-guide/tools/check-rules.py --staged` 與 `git diff --cached --check` 均通過。Implementation commit `2bfd610`（`feat: add mobile auth and API client`）；未修改未追蹤 `android-dev-guide/`。
- Fix round 1：review 指出 absolute URL 會沿用同一 Dio interceptor、指定的 `test_dio_client.dart` 不符合 Flutter 預設 `*_test.dart` discovery、以及無有效 Problem Details 的 HTTP response 被誤分類為 network failure。Origin Red：新增 cross-origin absolute HTTPS regression 後，focused suite 失敗並顯示 untrusted request 仍帶 `Authorization: Bearer ...`；Green：interceptor 只在 resolved request `Uri.origin` 等於 configured Daily News API origin 時才取得與附加 Firebase token。Discovery：保留指定 focused 檔，新增只呼叫其 `main()` 的 `dio_client_test.dart` entrypoint，不複製 test body；plain `flutter test` 從先前只執行 `app_test.dart` 改為執行完整 F2 suite。HTTP Red：新增 503 non-Problem response test 時因 `ApiFailureKind.http` 不存在而 compilation fail；Green：新增 typed HTTP failure，保留 status code，只有沒有 HTTP response 時才分類為 network。Fresh verification：`dart run build_runner build --delete-conflicting-outputs && flutter analyze && flutter test` exit 0，codegen 寫入 0 outputs、analyze 為 `No issues found!`、plain suite `7 passed`（F2 6 + app 1）；`python3 flutter-dev-guide/tools/check-rules.py --staged` 與 `git diff --cached --check` exit 0。Fix commit `6058cb7`（`fix: harden mobile API transport`）；未修改未追蹤 `android-dev-guide/`。
- Fix round 2：review 指出 401 sign-out side effect 仍會被不同 origin 的 absolute request 觸發。Red：保留並明確命名 configured API origin 401 應 sign out 的既有測試，新增 cross-origin 401 不應 sign out 的 regression；`flutter test test/core/test_dio_client.dart` 失敗，cross-origin fixture 的 `signOutCount` actual 1、expected 0。Green：`onError` 同時要求 resolved request `Uri.origin` 等於 configured Daily News API origin 且 status 為 401，才終止 session；focused suite `7 passed`。Fresh verification：`dart run build_runner build --delete-conflicting-outputs && flutter analyze && flutter test` exit 0，codegen 寫入 0 outputs、analyze 為 `No issues found!`、plain suite `8 passed`（F2 7 + app 1）；`python3 flutter-dev-guide/tools/check-rules.py --staged` 與 `git diff --cached --check` exit 0。Fix commit `6c7c543`（`fix: scope API 401 sign-out`）；未修改未追蹤 `android-dev-guide/`。

### F3: 實作 App router、theme 與首頁更新狀態

**Depends on:** F2  
**Parallel:** no — 首頁是所有 feature 的 route entry。  
**Files:** Create `apps/mobile/lib/core/{routing/,theme/}`, `apps/mobile/lib/features/home/{application/,presentation/,data/}`, `apps/mobile/test/features/home/home_screen_test.dart`.

- [x] **Red:** Widget tests 期待首頁顯示「尚未更新」、成功 Run 的 locale-aware 時間、立即更新按鈕、以及 queued/running 時 disabled 狀態。
- [x] **Run Red:** `cd apps/mobile && flutter test test/features/home/home_screen_test.dart`，預期失敗。
- [x] **Green:** 以 `go_router` 建 route；Home controller 讀 `GET /ingestion-runs/latest`，只把 View state 暴露為 immutable Riverpod state；所有可見字串由 localization resource 提供。
- [x] **Run Green:** `cd apps/mobile && flutter analyze && flutter test test/features/home/home_screen_test.dart`。
- [x] **Commit:** `git add apps/mobile && git commit -m "feat: add home refresh status"`。

**完成紀錄：**

- 2026-08-31：Red：新增 `test/features/home/home_screen_test.dart`，以 authenticated provider 與 fake `HomeRepository` override 驗證「尚未更新」、中文 locale 時間 `2026/8/30 08:01`、立即更新 control，以及 queued/running 的 disabled 狀態；`cd apps/mobile && flutter test test/features/home/home_screen_test.dart` 因 Home contracts/provider 與 `DailyNewsApp.locale` 尚不存在而 compilation failed，符合缺少 F3 功能的預期。Green：建立單一 `GoRouter` composition root、Material 3 light/dark theme 與 spacing token；將 auth/loading/sign-in 與 Home 可見文案移入 ARB/Flutter localization resource。Home data flow 為 `HomeScreen -> HomeController -> HomeRepository -> HomeRemoteService -> ApiClient`；remote service 透過 `GET /v1/ingestion-runs/latest` 讀取 backend contract，repository 映射 DTO 與 typed failure，controller 只暴露 immutable `HomeUiState`；Widget 不直接存取 Repository、service 或 Dio。Manual Run 確認 dialog 與 POST 依 task 切分保留給 F5。新增 `flutter_localizations` SDK dependency 與 `intl ^0.20.2`。Fresh verification：`cd apps/mobile && flutter gen-l10n && flutter analyze && flutter test test/features/home/home_screen_test.dart && flutter test` exit 0，analyze 為 `No issues found!`，focused suite `4 passed`，完整 suite `12 passed`；`python3 flutter-dev-guide/tools/check-rules.py --staged` 與 `git diff --cached --check` exit 0。`integration_test/` 尚未接入，完整行程由 F6 建立。Implementation commit `b0f2e0f`（`feat: add home refresh status`）；未修改未追蹤 `android-dev-guide/`。
- Fix round 1（2026-09-01）：review 發現 production `main.dart` 未 override `apiBaseUrlProvider`，authenticated Home 組裝 `Dio` 時會觸發 provider 的預設 `StateError`。Red：擴充 `test/app_test.dart` 後，`flutter test test/app_test.dart` 因 `buildConfiguredApp`、`parseApiBaseUrl` 與 controlled socket adapter provider 不存在而 compilation failed。Green：`main.dart` 從 non-secret `DAILY_NEWS_API_BASE_URL` Dart define 讀取 endpoint，在 app bootstrap 注入 provider；只接受 HTTPS、`/v1/` path、無 credentials/query/fragment 的 URI，並將缺省值設為保留且不可路由的 `.invalid` endpoint，避免 provider construction 直接失敗。新增 production bootstrap override test，以 fake Firebase gateway 和 controlled socket adapter 保留真實 `HomeController -> HomeRepository -> HomeRemoteService -> ApiClient -> Dio` wiring，驗證 request 準確到達 `https://api.example.test/v1/ingestion-runs/latest`且首頁顯示「尚未更新」；README 記錄 `--dart-define` 用法與公開設定的安全限制。Fresh verification：`cd apps/mobile && flutter analyze && flutter test test/app_test.dart && flutter test` exit 0，analyze 為 `No issues found!`，bootstrap suite `5 passed`，完整 suite `16 passed`；`python3 flutter-dev-guide/tools/check-rules.py --staged` 與 `git diff --cached --check` exit 0。Fix commit `31cd677`（`fix: configure mobile API bootstrap`）；未修改未追蹤 `android-dev-guide/`。

### F4: 實作 Category 首頁與動態設定 Bottom Sheet

**Depends on:** F2, F3  
**Parallel:** yes — 可與 F5 並行，只共享 API contract。  
**Files:** Create `apps/mobile/lib/features/categories/{application/,data/,presentation/}`, `apps/mobile/test/features/categories/category_sheet_test.dart`.

- [x] **Red:** Widget tests：無 Category 只顯示新增方塊；有 N 個 Category 時最後一格是新增；Bottom Sheet 有名稱／關鍵字／預設來源／特殊需求；新增按鈕開 dialog 並新增可編輯 Source Setting；空 source 不送 API。
- [x] **Run Red:** `cd apps/mobile && flutter test test/features/categories/category_sheet_test.dart`，預期失敗。
- [x] **Green:** 建立 Category Repository、Riverpod controller 與 stateless content widgets；儲存成功關閉 sheet 並刷新 Categories；欄位錯誤與 submit loading 防止重複送出。
- [x] **Run Green:** `cd apps/mobile && flutter analyze && flutter test test/features/categories/category_sheet_test.dart`。
- [x] **Commit:** `git add apps/mobile && git commit -m "feat: add category configuration UI"`。

**完成紀錄：**

- 2026-09-01：Red：新增 `test/features/categories/category_sheet_test.dart`，以 Home/Category repository provider override 驗證空與有資料的 Category grid、最後新增方塊、完整設定 Bottom Sheet、add-source dialog 和空白來源省略；`cd apps/mobile && flutter test test/features/categories/category_sheet_test.dart` 因 Category contracts/providers/widgets 尚不存在而 compilation failed，符合缺少 F4 功能的預期。Green：新增 Category DTO/remote service/repository 與 `GET/POST /v1/categories` transport，Riverpod controller 將 domain model 轉為 immutable UI state；Home 組合 provider-backed Category section，grid 以穩定 key 顯示 Category 並固定在末尾放新增方塊。Bottom Sheet 包含必填名稱、關鍵字、預設「未指定網站」來源、可動態追加的網站與特殊需求；送出前 trim 並省略空 source，必填錯誤顯示在欄位旁，submitting 時停用儲存與新增來源，成功後關閉 sheet 並透過 Repository 重讀 Categories。Dialog-owned controller 由其 State 成對 dispose；Widget 不直接依賴 Repository、remote service 或 Dio。所有新增文案來自 ARB localization resource，未新增 package dependency。Fresh verification：`cd apps/mobile && flutter gen-l10n && flutter analyze && flutter test test/features/categories/category_sheet_test.dart` exit 0，analyze 為 `No issues found!`，F4 suite `5 passed`；`python3 flutter-dev-guide/tools/check-rules.py --staged` 與 `git diff --cached --check` exit 0。依指示未執行 F4 範圍外的完整 Flutter suite；`integration_test/` 仍由 F6 建立。未修改未追蹤 `android-dev-guide/`。

### F5: 實作新聞列表、詳情、tag filter 與手動更新 dialog

**Depends on:** F2, F3  
**Parallel:** yes — 可與 F4 並行。  
**Files:** Create `apps/mobile/lib/features/news/{application/,data/,presentation/}`, `apps/mobile/test/features/news/{news_list_test.dart,manual_refresh_dialog_test.dart}`.

- [x] **Red:** Widget tests：列表初次讀取 20 筆、scroll 使用 `next_cursor`、tag filter 重置 cursor、永久／刪除 action 更新畫面；立即更新 dialog 說明背景工作，確認後送 POST 並顯示 queued/running。
- [x] **Run Red:** `cd apps/mobile && flutter test test/features/news/news_list_test.dart test/features/news/manual_refresh_dialog_test.dart`，預期失敗。
- [x] **Green:** 實作 News Repository、cursor controller、detail view、source tag chips、Article mutation 和 ManualRun controller；不得在 Widget 直接呼叫 Dio 或 Repository。
- [x] **Run Green:** `cd apps/mobile && flutter analyze && flutter test test/features/news`。
- [x] **Commit:** `git add apps/mobile && git commit -m "feat: add news feed and manual refresh"`。

**完成紀錄：**

- 2026-09-01：Red：新增 `test/features/news/news_list_test.dart` 與 `manual_refresh_dialog_test.dart`，涵蓋首批 20 筆、`next_cursor` 捲動續載、tag filter 重置、永久／刪除 mutation，以及背景更新說明、POST 與 queued 狀態；`cd apps/mobile && flutter test test/features/news` 因 News/ManualRun contracts 與 widgets 尚不存在而 compilation failed，符合預期。Green：新增 News/ManualRun Repository、remote service、Riverpod cursor/detail/manual-run controllers 與 immutable UI state；列表提供 source tag chips 與 cursor pagination，詳情提供永久保存和刪除，Home 的立即更新 dialog 說明背景工作且 active 狀態停用 control。Widget 只使用 controller/provider，不直接呼叫 Dio 或 Repository；`ApiClient` 補上 typed PATCH/DELETE transport。Fresh verification：`cd apps/mobile && flutter analyze && flutter test test/features/news` exit 0，analyze 為 `No issues found!`，F5 suite `4 passed`；`python3 flutter-dev-guide/tools/check-rules.py --staged` 與 `git diff --cached --check` 通過。未修改未追蹤的 `android-dev-guide/` 與 `.vscode/`。

### F6: 建立 Flutter end-to-end 驗收流程

**Depends on:** F4, F5, I4  
**Parallel:** no — 需要完整 API workflow。  
**Files:** Create `apps/mobile/integration_test/daily_news_flow_test.dart`, `apps/mobile/test_support/fake_api_server.dart`.

- [x] **Red:** 寫 integration scenario：登入 → 建立 Category → 手動更新確認 → 顯示 run queued → 顯示新 Article → 設永久 → 刪除。
- [x] **Run Red:** `cd apps/mobile && flutter test integration_test/daily_news_flow_test.dart`，預期在缺 UI 元件或 fake server routes 時失敗。
- [x] **Green:** 完成 fake API transport 與所有 UI accessibility keys；測試不連 production Firebase、GCP 或 LLM。
- [x] **Run Green:** `cd apps/mobile && flutter test integration_test/daily_news_flow_test.dart`。
- [x] **Commit:** `git add apps/mobile && git commit -m "test: add daily news integration flow"`。

**完成紀錄：**

- 2026-09-01：建立單一 integration journey，由 fake Google auth 進入登入後首頁，透過 production Dio／Service／Repository／Riverpod wiring 建立 Category、確認 manual Ingestion Run 並顯示 `queued`、讀取新 Article、設為永久後全域 soft delete。fake API transport 強制 `Bearer fake-firebase-token` 並只在記憶體提供需要的 routes，未連 production Firebase、GCP、LLM 或外部網路。Red mutation：暫時移除 Article row accessibility key 後，`cd apps/mobile && flutter test integration_test/daily_news_flow_test.dart` 以 `Found 0 widgets with key ['news-item-article-1']` 如期失敗；還原 key 後 Green 為 `1 passed`。`cd apps/mobile && flutter analyze` 為 `No issues found!`。Commit message：`test: add daily news integration flow`。

### F7: 區分手動 Ingestion Run 與首頁資料刷新

**Depends on:** F3, F4, F5
**Parallel:** no — 共用首頁更新狀態與 Category 首頁資料。
**Files:** Modify `apps/mobile/lib/features/home/{application/,presentation/}`, `apps/mobile/lib/features/categories/application/`, `apps/mobile/lib/l10n/app_zh.arb`, `apps/mobile/test/features/home/home_screen_test.dart`; modify this requirements/task record.

- [x] **Red:** widget test 驗證手動 Run 為 queued/running 時「立即更新」停用，但「刷新頁面」仍可點擊；點擊後分別重新讀取首頁更新狀態與 Category。
- [x] **Run Red:** `cd apps/mobile && flutter test test/features/home/home_screen_test.dart`，因 `homeDataRefreshButtonKey` 尚不存在而 compilation failed，確認 control 尚未實作。
- [x] **Green:** 新增「刷新頁面」control；只透過既有 Home／Category controller 的 API reload 操作重新取得資料，絕不送出 `POST /v1/ingestion-runs`。所有新增可見字串放在 localization resource。
- [x] **Run Green:** `cd apps/mobile && flutter gen-l10n && flutter analyze && flutter test test/features/home/home_screen_test.dart`。
- [x] **Commit:** `git add docs apps/mobile && git commit -m "feat: add home data refresh control"`。

**完成紀錄：**

- 2026-09-13：Red：新增刷新按鈕測試後，`cd apps/mobile && flutter test test/features/home/home_screen_test.dart` 因 `homeDataRefreshButtonKey` 尚不存在而 compilation failed。Green：HomeScreen 新增「刷新頁面」control，並以既有 `HomeController.reload()` 與 `CategoriesController.reload()` 並行重讀 API 資料，不經過 ManualRun controller。focused Green command 通過（`No issues found!`、7 passed）；補強後以 queued/running 兩種 active Run 狀態驗證立即更新停用、刷新仍可點擊、Home/Category 各重讀一次且 `ManualRunRepository.requestManualRun()` 維持 0，focused suite 為 8 passed。最終驗證：`cd apps/mobile && flutter gen-l10n && flutter analyze && flutter test` 通過（31 passed）；`flutter test -d emulator-5554 integration_test/daily_news_flow_test.dart` 通過（1 passed）；`uv run --with pyyaml python flutter-dev-guide/tools/check-rules.py --files apps/mobile/lib/features/home/presentation/home_screen.dart apps/mobile/test/features/home/home_screen_test.dart` 與 `git diff --check` 通過。Commit `22ee0ff`。

---

## Phase R — Go backend replacement

本階段取代 Python/FastAPI 的實作，並把既有 Python task 視為 parity
evidence；不重寫其完成歷史。除 R0 外，每個 Go production-code task 都等待
使用者確認 [`Go migration design`](../superpowers/specs/2026-09-01-go-backend-migration-design.md)。
所有 Go task 依 `backend/AGENTS.md` 執行 context-first、明確 Problem Details
error mapping、transaction boundary、`gofmt`、`go vet ./...` 與 `go test ./...`。

### R0: 凍結 Go migration design 與 compatibility contract

**Depends on:** O3, F6
**Parallel:** no — 建立後續 replacement 的唯一規格與 baseline。
**Files:** Create `backend/AGENTS.md`, `docs/contracts/daily-news.openapi.json`, `docs/superpowers/specs/2026-09-01-go-backend-migration-design.md`; modify `CONTEXT.md`, `docs/requirements/daily-news.md`, `docs/tasks/daily-news.md`, `infra/docs/{gcp-setup.md,github-variables.md,secret-inventory.md}`.

- [x] **Red:** 將現有 FastAPI OpenAPI output 與 Flutter remote DTO endpoint use 對照，列出 `/v1` paths、status codes、Problem Details media type 與 JSON field compatibility baseline。
- [x] **Green:** 將 OpenAPI JSON 正規化後 checked in；將未來 backend 固定決策改為 Go，保留 Python task completion history；建立 Go backend agent guardrails 與完整 migration design；O4 改為等待 Go parity。
- [x] **Verify:** `cd backend && uv run python -c 'from app.main import create_app; import json; print(json.dumps(create_app().openapi(), ensure_ascii=False, indent=2, sort_keys=True))' | cmp -s - ../docs/contracts/daily-news.openapi.json`、`rg -n 'FastAPI|python -m app.jobs.daily_news|uvicorn' CONTEXT.md docs/requirements/daily-news.md infra/docs backend/AGENTS.md`、`git diff --check`。
- [x] **Commit:** `git add CONTEXT.md docs/requirements docs/tasks docs/contracts docs/superpowers/specs infra/docs backend/AGENTS.md && git commit -m "docs: plan Go backend migration"`。

**Done when:** Go design、agent rules、immutable OpenAPI artifact 和 GCP wording 已確認；未寫 Go production code，並等待使用者確認後開始 R1。

**完成紀錄：**

- 2026-09-01：以既有 FastAPI app 的 deterministic `openapi()` 輸出建立 checked-in baseline，JSON parser 確認 8 個 paths 與 14 個 schemas，且包含全部 7 個 `/v1` paths。`cd backend && uv run python -c 'from app.main import create_app; import json; print(json.dumps(create_app().openapi(), ensure_ascii=False, indent=2, sort_keys=True))' | cmp -s - ../docs/contracts/daily-news.openapi.json` exit 0；目標決策文件的 `rg -n 'FastAPI|python -m app.jobs.daily_news|uvicorn' CONTEXT.md docs/requirements/daily-news.md infra/docs backend/AGENTS.md` 無輸出；`git diff --check` 與 staged `git diff --cached --check` 通過。新增 migration design、Go backend guardrails、Go replacement phase，保留 Python 完成歷史，且 O4 現在依賴 R7 Go parity。Implementation commit `b7e7209`（`docs: plan Go backend migration`）；未寫 Go production code，未修改未追蹤 `.vscode/`、`android-dev-guide/` 或 research note。

### R1: 建立 Go module、health endpoint 與 local composition roots

**Depends on:** R0
**Parallel:** no — 建立後續 Go package 與測試的執行邊界。
**Files:** Create `backend/{go.mod,cmd/api/main.go,cmd/daily-news-job/main.go,internal/platform/,internal/httpapi/,contract/}` and Go tests; retain Python implementation until R7.

- [x] **Red:** 以 `httptest` 驗證 `GET /healthz` 回既有 200 JSON，並測試 API shutdown 與 Job 缺 `RUN_ID` 的 non-zero exit boundary。
- [x] **Green:** 建立 explicit constructors、`http.Server` graceful shutdown、`PORT` listener、config parsing 與 Job command skeleton；不得加入 framework、ORM 或 DI container。
- [x] **Verify:** `gofmt -w` changed Go files, `cd backend && go vet ./... && go test ./...`，並比對 health response 與 contract baseline。
- [x] **Commit:** verified changes only, `git commit -m "build: scaffold Go backend"`。

### R2: 移植 Firebase identity、Problem Details 與 HTTP contract harness

**Depends on:** R1
**Parallel:** no — 所有 `/v1` route 共用 authentication 與 error boundary。
**Files:** Create `backend/internal/{identity,httpapi}/` Go implementations and `backend/contract/` tests.

- [x] **Red:** handler tests 覆蓋 missing/malformed/invalid token 401、non-allowlisted email 403、strict JSON validation 422，並比對 content type/body 與 checked-in artifact。
- [x] **Green:** Firebase Admin Go SDK ADC adapter、fake verifier seam、exact email policy、Bearer middleware、typed errors、Problem Details encoder 與 JSON decoder。
- [x] **Verify:** `cd backend && go vet ./... && go test ./...`; no test contacts real Firebase.
- [x] **Commit:** verified changes only, `git commit -m "feat: add Go API identity and errors"`。

### R3: 移植 PostgreSQL access、migration parity 與 Category CRUD

**Depends on:** R2
**Parallel:** no — 需要共用 SQL pool 與 authenticated HTTP boundary。
**Files:** Create Go SQL repositories, `backend/migrations/`, schema parity and Category `httptest`/PostgreSQL tests.

- [x] **Red:** 用 disposable PostgreSQL 驗證既有 tables/indexes/constraints，及 Category/Source Setting CRUD success、422、404、401 contract cases。
- [x] **Green:** 使用 pgx `database/sql` adapter、explicit SQL、golang-migrate representation、Category transaction boundaries 和 `/v1/categories` handlers；schema 結果不得變更。
- [x] **Verify:** run schema diff/index assertions, `cd backend && go vet ./... && go test ./...`, and focused Flutter contract review.
- [x] **Commit:** verified changes only, `git commit -m "feat: add Go category API and schema parity"`。

### R4: 移植 News read model、cursor、permanence 與 soft delete

**Depends on:** R3
**Parallel:** no — 建立於 Category schema and identity boundary。
**Files:** Create Go News service/repository/handlers and parity tests.

- [x] **Red:** PostgreSQL + `httptest` tests cover 20-item keyset pagination, opaque cursor continuation/400, source tag filtering, expiry exclusion, global permanent/delete, and 404.
- [x] **Green:** preserve `(inserted_at, article_id)` ordering and JSON fields for every `/v1/categories/{categoryId}/news` and `/v1/news/{newsId}` route.
- [x] **Verify:** `cd backend && go vet ./... && go test ./...`, plus contract fixture comparison.
- [x] **Commit:** verified changes only, `git commit -m "feat: add Go news API parity"`。

### R5: 移植 Ingestion Run state、manual launch 與 active-run transaction

**Depends on:** R3
**Parallel:** yes — 與 R4 共用 schema but not route implementation files。
**Files:** Create Go ingestion run repository/service/handlers and tests.

- [x] **Red:** tests cover latest status, 202 manual request, completed scheduled Run not blocking manual, and concurrent requests returning the one active Run.
- [x] **Green:** use one transaction for active-run acquisition/idempotency, retain `scheduled:<taipei_date>` and `manual:<request-id>`, and inject a Cloud Run Job launcher seam.
- [x] **Verify:** concurrent PostgreSQL test, `cd backend && go vet ./... && go test ./...`, and 202/Problem Details fixture comparison.
- [x] **Commit:** verified changes only, `git commit -m "feat: add Go ingestion run API parity"`。

### R6: 移植 source adapters、dedupe 與 Cloud Run Job orchestrator

**Depends on:** R4, R5
**Parallel:** no — 組裝 Article semantics、Run state 與 Job input。
**Files:** Create Go ingestion adapters/orchestrator/job tests and commands.

- [x] **Red:** fake-adapter PostgreSQL tests cover 10-candidate cap, URL-then-title dedupe, shared Article across Categories, soft-deleted suppression, 30-day expiry, per-source failure isolation, Attempt counters, terminal state, and `RUN_ID` handling.
- [x] **Green:** migrate adapter seams and orchestrator with context propagation; write Article/Category Article/Attempt/counters in the documented transaction boundary; Job has no HTTP listener.
- [x] **Verify:** `cd backend && go vet ./... && go test ./...`; command help/invalid `RUN_ID` tests do not use GCP, secrets, or live sources.
- [x] **Commit:** verified changes only, `git commit -m "feat: add Go ingestion job parity"`。

### R7: Cut over container, CI, and deployment manifests to verified Go image

**Depends on:** R2, R4, R5, R6
**Parallel:** no — only parity-complete implementation may replace runtime wiring.
**Files:** Modify `backend/{Dockerfile,.dockerignore,README.md}`, `.github/workflows/{ci.yml,deploy.yml}`, `infra/cloud-run/{service.yaml,job.yaml}`; remove Python runtime files only after parity verification.

- [x] **Red:** container/workflow tests assert one non-root Go image, API binary on `$PORT`, Job binary consumes `RUN_ID`, CI runs Go verification, and manifests retain all resource/identity/secret names.
- [x] **Green:** build the Go binaries and replace only image build/command wiring; retain Cloud Run Service + Job, GitHub OIDC/WIF, Secret Manager names, and Flutter contract.
- [x] **Verify:** build/run container locally, `actionlint .github/workflows/*.yml`, manifest structural checks, `cd backend && go vet ./... && go test ./...`, schema parity, Flutter integration test, and `git diff --check`.
- [x] **Commit:** verified changes only, `git commit -m "build: cut over Daily News backend to Go"`。

### R8: Correct production ingestion work planning and citation persistence

**Depends on:** R7
**Parallel:** no — restores the Go Job path required by O4.
**Files:** Modify `backend/{cmd/daily-news-job,Dockerfile,migrations}`, `backend/internal/ingestion/`; create focused Go tests and migration.

- [x] **Red:** PostgreSQL tests prove active Category/Source Setting rows produce planned work, a public RSS/Atom entry persists a non-empty citation, and an empty work slice is never passed by the Job composition root.
- [x] **Green:** Add a bounded, fakeable public RSS/Atom adapter; load active work; preserve canonical URL dedupe and store `citation_url`; apply checked-in migrations from the Job image before work starts.
- [x] **Verify:** `cd backend && gofmt -w ... && go vet ./... && go test ./...`; build the container and verify `/app/migrations/000003_article_citation.up.sql` exists.
- [x] **Commit:** verified changes only, `git commit -m "fix: restore Go ingestion staging path"`。

**完成紀錄：**

- 2026-09-12：Red：新增 RSS citation、Atom entry 與 active HTTP(S) Source Setting planner tests；Atom test 初始回空 candidates，planner test 初始錯誤納入 `ftp://` Source Setting，確認新的行為尚未實作。Green：新增 `000003_article_citation` migration、`WorkPlanner`、可注入 HTTP client 的 RSS/Atom adapter、UTM canonicalization，以及 Job migration/work planning wiring；Job composition regression 禁止 `Run(ctx, id, nil)`。Verify：`cd backend && gofmt -w cmd/daily-news-job/main.go cmd/daily-news-job/main_test.go internal/ingestion/*.go && go test ./cmd/daily-news-job ./internal/ingestion -count=1 && go vet ./... && go test ./...` 通過；`cd backend && uv run pytest tests/test_container_contract.py -q` 為 `4 passed`；`docker build -t daily-news-backend:r8 backend` 成功，runtime image 確認存在 `/app/migrations/000003_article_citation.up.sql` 與可執行 Job binary；`git diff --check` 通過。Commit 見下。

---

## Phase O — GCP、GitHub Actions 與上線驗收

O1 先把非秘密設定與權限寫成可審查文件，再容器化、部署、排程。任何真實帳號、project、token 設定均由使用者完成，不寫入 repo。

### O1: 建立 GCP／GitHub secrets 與 WIF 操作文件

**Depends on:** G1, B1  
**Parallel:** yes — 可與 B2–I4、F1–F6 並行。  
**Files:** Create `infra/docs/{gcp-setup.md,github-variables.md,secret-inventory.md}`.

- [x] **Red:** 列出必要設定並執行 `rg -n '(GEMINI_API_KEY|DB_PASSWORD|GITHUB_NEWS_TOKEN|YOUTUBE_API_KEY|ALLOWED_USER_EMAIL)' infra/docs`；確認每個值都有位置、用途、是否可選與不得放置處。
- [x] **Green:** 文件化 Cloud SQL、Secret Manager、Cloud Run service/job service accounts、Firebase project、WIF provider、GitHub Variables、最小 IAM role；明確說明 GitHub Action 使用 `id-token: write` 而非 JSON key。
- [x] **Verify:** 人工逐項對照需求規格第 9 節；文件不含真實 project id、secret 值、email 或 token。
- [x] **Commit:** `git add infra/docs && git commit -m "docs: add GCP and secret setup guide"`。

**Done when:** 使用者可在 GCP／GitHub UI 完成所有外部設定，而無需猜測 token 名稱或權限。

**完成紀錄：**

- 2026-09-02：Red：Go container/workflow contract tests 先以 `cd backend && uv run pytest tests/test_container_contract.py -q` 失敗，因 Dockerfile 仍是 Python、Job manifest 仍執行 shell script、CI 仍是 uv/pytest。Green：改為 Go 1.27 multi-stage Alpine image，建立 non-root `daily-news` user，產出 `/app/api` 與 `/app/daily-news-job`；API composition mount 全部 `/v1` route families，使用 Firebase ADC、PostgreSQL pool 與官方 Go Cloud Run v2 REST client（ADC）在 manual Run committed 後以 execution override 注入 `RUN_ID`。scheduled workflow 產生 UUID override；Job 對不存在 ID transactionally 建立或合併 `scheduled:<taipei_date>` Run。CI 改跑 Go checks，Cloud Run Job command 改為 Go binary，保留 Service/Job、OIDC/WIF、runtime service accounts 與 `ALLOWED_USER_EMAIL`/`DATABASE_URL` Secret Manager names。Flutter bootstrap test 改為記錄全部首頁並行 HTTP requests，避免以最後一個 Category request 錯誤覆蓋 ingestion-run assertion。Verification：`cd backend && gofmt -w ... && go vet ./... && go test -p 1 ./... -count=1` 通過；container contract `3 passed`，`docker build -t daily-news-backend:go-r7 backend` 成功，inspect 確認 `daily-news ["/app/api"]`，Job 缺 RUN_ID 正確 exit 2；actionlint Docker image 對三個 workflows 通過；`cd apps/mobile && flutter test` 為 25 passed；`git diff --check` 通過。未建立或操作真實 GCP/GitHub 資源。Implementation commit `ba3a88a`（`build: cut over Daily News backend to Go`）。

- 2026-09-01：Red：新增 fake-adapter PostgreSQL test 後，`cd backend && go test ./internal/ingestion -run TestOrchestratorPersistsDedupeExpiryAttemptsAndTerminalRun -v` 因 CandidateArticle、SourceWork、Orchestrator 與 Result 尚未定義而 build failed；新增 malformed `RUN_ID` test 後，Job 回傳 1 而非 2。Green：加入 context-first SourceAdapter seam、每來源上限 10 筆的 Orchestrator 與 explicit SQL persistence；每個成功來源以一個 transaction 寫入／重用 Article、Category Article、Attempt 與 Run counters，canonical URL 優先、normalized title 次之，deleted Article 保留為 suppression，new Article 設 30-day expiry。來源錯誤另以 transaction 記錄 failed Attempt 與 error counter，並繼續其他來源；Run 最後設為 succeeded/finished。Job 先驗證 UUID `RUN_ID`，再以 `DATABASE_URL` 開啟 pgx-backed database 並執行 Orchestrator，沒有 HTTP listener。Verification：`cd backend && gofmt -w internal/ingestion cmd/daily-news-job && go vet ./... && go test ./... && git diff --check` 通過；focused integration test 驗證 10-candidate cap、URL/title dedupe、跨 Category Article、suppression、expiry、Attempt counters、terminal status 和來源失敗隔離，Job tests 驗證 missing/malformed `RUN_ID` 與缺 database configuration 都不會假報成功。所有 source adapter 均為 fake，未呼叫 GCP、secrets 或真實來源。Implementation commit `252e7b0`（`feat: add Go ingestion job parity`）。

- 2026-09-01：Red：新增 Run service PostgreSQL integration tests 後，`cd backend && go test ./internal/ingestion -run 'Test(RequestManualMergesConcurrentActiveRun|RunServiceLatestAndCompletedScheduledRun)' -v` 因 `NewRunService`、`NewRunStore` 與 Run types 尚未定義而 build failed；新增 handler test 後因 `NewHandler` 尚未定義而再次 build failed。Green：加入 explicit SQL Run Store、Run service、`net/http` handlers 與 fakeable `JobLauncher` seam；`AcquireManual` 在單一 `BeginTx` 中取得 PostgreSQL transaction-scoped advisory lock、查詢 `queued`/`running` Run、建立 `manual:<uuid>` queued Run 並 commit，只有新建 Run 於 commit 後才 launch。Taipei 日期以 UTC+08 計算；latest 回傳最後 successful timestamp 與 active Run。PostgreSQL + `httptest` tests 驗證 concurrent requests 收斂為一個 active Run、已完成 scheduled Run 不阻擋 manual Run、202 JSON response，以及既有 401 `application/problem+json` body。Verification：`cd backend && gofmt -w internal/ingestion && go vet ./... && go test ./... && python3 -c '<OpenAPI ingestion-run fixture assertions>' && git diff --check` 通過；fixture 確認兩條 Ingestion Run paths、`IngestionRunResponse`/`LatestIngestionRunResponse` JSON fields 和 manual POST 202。Implementation commit `549f61c`（`feat: add Go ingestion run API parity`）；未呼叫真實 Cloud Run、Firebase、GCP 或其他外部資源。

- 2026-09-01：Red：新增 cursor tests 後，`cd backend && go test ./internal/news -run TestCursorRoundTripPreservesKeysetPosition` 因 cursor codec 尚未定義而 build failed。Green：加入以 explicit SQL 實作的 News Store 與 authenticated `net/http` handlers；list 固定回傳 20 筆、以 `(category_articles.inserted_at, article_id)` 的 DESC keyset cursor 續頁，支援 `sourceTagId` filter，排除 soft-deleted／expired Article；Article detail、permanent toggle 和 global soft delete 維持既有 JSON names、status codes 及 Problem Details boundary。PostgreSQL + `httptest` integration test 使用 disposable `postgres:16-alpine`，驗證 20+1 keyset、opaque invalid cursor 400、source filter、expiry exclusion、跨兩個 Category 的 global delete、404、GET list/detail、PATCH success/422 與 DELETE 204，並確認每個 News response 的 Flutter 欄位。Verification：`cd backend && gofmt -w internal/news && go vet ./... && go test ./... && python3 -c '<OpenAPI News fixture assertions>' && git diff --check` 通過；OpenAPI fixture 確認三條 News paths 與 `NewsListItem`/`NewsPage` fields，focused Flutter scan 確認現有 client 使用 `sourceTagId`、`next_cursor`、`canonical_url`、`first_seen_at`、`expires_at`。Implementation commit `4e8ae3e`（`feat: add Go news API parity`）；未連 Firebase、GCP 或其他實際外部資源。

- 2026-09-01：Red：新增 Category handler tests 後，`cd backend && go test ./internal/category` 因 `NewHandler` 尚未定義而 build failed。Green：加入 pgx `database/sql` adapter、golang-migrate SQL representation（initial schema + Category Article soft delete）、explicit Category/Source Setting SQL Store 與 transaction-bounded create/update/delete；`/v1/categories` handler 保留 Bearer middleware、422 validation、404 soft-deleted Category 和 204 delete semantics。Go migration integration test 以 disposable `postgres:16-alpine` 建立 `articles,categories,category_articles,ingestion_attempts,ingestion_runs,schema_migrations,source_settings` 和五個 required indexes；Category lifecycle integration test 驗證來源排序／host normalization、replacement soft delete、Category Article soft delete 與 Article 保留。Verification：`cd backend && go mod tidy && gofmt -w internal/category internal/platform && go vet ./... && go test ./... && git diff --check` 通過；focused Flutter scan 確認 Category DTO/remote service 持續使用 `search_keywords`、`special_requirements`、`source_settings`、`website_input`。Implementation commit `c8cf9c9`（`feat: add Go category API and schema parity`）。

- 2026-09-01：Red：新增 HTTP handler tests 後，`cd backend && go test ./internal/identity ./internal/httpapi ./contract` 因 identity package、verifier seam、Bearer middleware、strict decoder 與 Problem encoder 尚不存在而 build failed；OpenAPI harness 已通過並確認 checked-in artifact 的每個 `/v1` operation 都宣告 `HTTPBearer`。Green：加入 Firebase Admin Go SDK（ADC `NewApp`）、`TokenVerifier` fake seam、exact email allowlist、strict single-body JSON decoder，與保留 Flutter/既有 FastAPI JSON shape 的 `{"detail":{"title","status"}}` `application/problem+json` encoder。Firebase SDK 的 `auth.IsIDTokenInvalid` 只將 invalid/expired/revoked token 映射為 401；operational failure 保留原因並映射為不暴露內部細節的 503。Verification：`cd backend && go mod tidy && gofmt -w internal/identity/firebase.go internal/identity/firebase_test.go internal/httpapi/auth.go internal/httpapi/auth_test.go internal/httpapi/problem.go internal/httpapi/request.go contract/openapi_test.go && go vet ./... && go test ./... && git diff --check` 通過；所有 test 使用 fake token client，未連 Firebase、GCP 或真實帳號。Implementation commit `9505812`（`feat: add Go API identity and errors`）。

- 2026-09-01：Red：新增 Go `httptest` health contract、`http.Server.Shutdown`、`RUN_ID` boundary 及 Cloud Run `PORT` parser tests 後，`cd backend && go test ./...` 因 `NewMux`、`NewServer`、`run`、`APIAddress` 尚未定義而如預期 build failed。Green：建立 Go module、explicit `http.ServeMux`、`http.Server` graceful shutdown composition root、`0.0.0.0:$PORT` parser，及尚未安裝 orchestrator 時不會假成功的 Job skeleton；未加入 web framework、ORM 或 DI container。Verification：`cd backend && gofmt -w cmd/api/main.go cmd/daily-news-job/main.go internal/httpapi/server.go internal/platform/config.go cmd/daily-news-job/main_test.go internal/httpapi/server_test.go internal/platform/config_test.go && go vet ./... && go test ./...` 通過（Job、HTTP API、platform test packages all `ok`；API composition root compiled）；health test assertion 保持 `200 application/json {"status":"ok"}`。Implementation commit `3e742be`（`build: scaffold Go backend`）。

- 2026-09-01：Red：`rg -n '(GEMINI_API_KEY|DB_PASSWORD|GITHUB_NEWS_TOKEN|YOUTUBE_API_KEY|ALLOWED_USER_EMAIL)' infra/docs` 在目錄尚不存在時以 exit 2 停止，確認 O1 文件尚未建立。Green：新增 Cloud SQL／Secret Manager／Firebase／Cloud Run service 與 Job identities／WIF／GitHub Variables 操作文件；runtime identities 分離 API 與 ingestion Job，GitHub deployer identity 僅具部署與 Job invocation 所需權限。文件明確要求 GitHub Actions 以 `id-token: write` + WIF，不使用 service-account JSON key，並列出 staging 前須由使用者完成的外部設定。Verify：設定名稱掃描確認所有必要值均有用途、存放位置、是否可選與禁止位置；人工對照需求規格第 9 節，且以 email／API token／OAuth client-id pattern 掃描確認沒有實值。`git diff --check` 通過。Implementation commit `e46bca4`（`docs: add GCP and secret setup guide`）。

### O2: 容器化 Python/FastAPI service 與 Cloud Run Job（完成歷史）

**Depends on:** B4, N1, N2, I4  
**Parallel:** no — 同一 image 必須同時服務 HTTP 與 Job entrypoint。  
**Files:** Create `backend/{Dockerfile,.dockerignore}`, `backend/scripts/{serve.sh,run-job.sh}`, `backend/tests/test_container_contract.py`.

- [x] **Red:** container contract test 驗證 `serve.sh` 啟動 uvicorn，`run-job.sh` 呼叫 `python -m app.jobs.daily_news`，且兩者使用同一 image／環境設定名稱。
- [x] **Run Red:** `cd backend && uv run pytest tests/test_container_contract.py -q`，預期失敗。
- [x] **Green:** 寫 multi-stage 或精簡 Python image、non-root runtime、health endpoint、Job entry script；映像不 baked-in secret。
- [x] **Run Green:** `docker build -t daily-news-backend:test backend && docker run --rm daily-news-backend:test python -m app.jobs.daily_news --help`。
- [x] **Commit:** `git add backend && git commit -m "build: containerize API and ingestion job"`。

**完成紀錄：**

- 2026-09-01：Red：新增可執行 shell entrypoint contract tests 後，`cd backend && uv run pytest tests/test_container_contract.py -q` 因 `backend/scripts/{serve.sh,run-job.sh}` 尚不存在而失敗（2 failed），確認兩種容器啟動邊界尚未實作。Green：新增同一 Python 3.12-slim image 的 Dockerfile、non-root `daily-news` runtime user、locked production dependency install、`.dockerignore`，與 API／Job entry scripts。API script 以 Cloud Run `PORT` 啟動 `uvicorn app.main:app`，Job script 以 `python -m app.jobs.daily_news` 並保留 CLI arguments；image 不複製 `.env`、虛擬環境、測試或秘密。Run Green：focused container contract 為 `2 passed`，`docker build -t daily-news-backend:test backend && docker run --rm daily-news-backend:test python -m app.jobs.daily_news --help` exit 0；`docker run --rm daily-news-backend:test id -u` 輸出 `999`。全量 `cd backend && uv run pytest -q` 為 `32 passed`（僅既有 Starlette/httpx deprecation warning），`cd backend && uv run ruff check .` 與 `git diff --check` 通過。Implementation commit `0d4bcd1`（`build: containerize API and ingestion job`）。

### O3: 建立部署與每日排程 workflows

**Depends on:** O1, O2  
**Parallel:** no — 需要已知 image、WIF 與 Cloud Run resource 名稱。  
**Files:** Create `.github/workflows/{ci.yml,deploy.yml,daily-ingestion.yml}`, `infra/cloud-run/{service.yaml,job.yaml}`.

- [x] **Red:** 以 `actionlint` 檢查 workflow；`daily-ingestion.yml` 必須包含 `cron: '0 0 * * *'`、`permissions: id-token: write` 與 Job execution command，初始檢查預期因檔案不存在而失敗。
- [x] **Green:** CI 執行 backend ruff/pytest、Flutter analyse/test、guide self-check；deploy workflow 透過 `google-github-actions/auth` 的 WIF 部署 service/job；daily workflow 僅觸發 Job，不帶 DB 或 LLM secret。
- [x] **Run Green:** `actionlint .github/workflows/*.yml` 與每個 manifest 的 schema/lint check；人工確認 GitHub Variables 與 O1 名稱一致。
- [x] **Commit:** `git add .github infra && git commit -m "ci: add Cloud Run deployment and daily job"`。

**完成紀錄：**

- 2026-09-01：Red：以 Docker image 執行 actionlint：`docker run --rm -v "$PWD":/repo -w /repo rhysd/actionlint:latest .github/workflows/daily-ingestion.yml` 在 workflow 尚不存在時以 exit 3 停止，顯示無法讀取檔案。Green：新增 CI、WIF deploy 與 UTC 00:00 daily-ingestion workflows，以及以 image／runtime service-account placeholders 為界的 Cloud Run Service／Job manifests。CI 執行 backend Ruff/pytest、Flutter analyze/test 與 guide self-check；deploy 使用 `google-github-actions/auth@v3` + `setup-gcloud@v3` 建置同一 backend image、render manifests 並部署 service/job；daily workflow 只執行既有 Cloud Run Job。Run Green：actionlint 對三個 workflow 通過；Python YAML structural check 確認 manifests 的 Cloud Run apiVersion/kind；比較 workflow 中 `vars.*` 與 O1 清冊，僅有 `CLOUD_RUN_JOB_NAME`、`GCP_PROJECT_ID`、`GCP_REGION`、`GCP_SERVICE_ACCOUNT`、`GCP_WORKLOAD_IDENTITY_PROVIDER`，且無 `secrets.*` 或 DB／LLM／adapter secret；`git diff --check` 通過。Implementation commit `3bbdad6`（`ci: add Cloud Run deployment and daily job`）。

### O4: 執行受控 staging smoke test（等待 Go parity）

**Depends on:** O3, F6, R7, R8
**Parallel:** no — 真實外部資源驗收。  
**Files:** Create `docs/operations/staging-smoke-test.md`; modify `docs/tasks/daily-news.md` 完成紀錄。

- [x] **Red:** 在 smoke test 文件先列出失敗判準：WIF 驗證失敗、Firebase allowlist 外帳號可讀資料、Run 沒有結束狀態、Article 沒有 citation、手動 Run 平行重複執行。
- [ ] **Green:** 由使用者設定 O1 所列 resource 與秘密後，部署 staging；以 allowlisted account 驗證登入、Category、manual Run、latest status、Article list、永久、刪除；以第二次並行 POST 驗證 active Run 合併。
- [ ] **Verify:** 檢查 Cloud Run／workflow log、`ingestion_runs`／`ingestion_attempts` 統計與 mobile integration results；記錄 pass/fail 和 run id，但不記錄 token／URL query secret。
- [ ] **Commit:** `git add docs/operations docs/tasks/daily-news.md && git commit -m "docs: record staging smoke test"`。

**Done when:** staging 以 Go image 與真正的 WIF、Cloud Run、Cloud SQL、Firebase 與至少一個公開來源完成端到端工作，且全部失敗判準均未發生。

**完成紀錄：**

- 2026-09-12：O4 Red/preflight 完成，Green 尚未開始。新增 `docs/operations/staging-smoke-test.md`，明列 WIF、公開 health、Firebase 401/403 allowlist、單一 active Run、terminal status、Attempt/citation、CRUD 與 mobile contract 的失敗判準、證據遮罩規則和執行順序。`git fetch origin main && git rev-list --left-right --count origin/main...main` 確認 `5f464ca` 已推送且為 `0 0`；GitHub Actions CI `34669133586` 與 deploy `34669133578` 均成功。Cloud Run control plane 顯示 revision `daily-news-api-00004-fqd` Ready、100% traffic、ingress all、`invokerIamDisabled: true`、default URI enabled；但兩個 service URL、八個 IPv4 edge、Google identity token 與 `gcloud run services proxy` 對 `/healthz` 均回 Google HTML 404，且 `run.googleapis.com/requests` 與 VPC Service Controls denial log 無對應項目。`gcloud run services update --default-url` 及一次 `--no-default-url`/`--default-url` 重建 URL 狀態後，observed generation 由 4 到 7，404 仍可重現。另以 `rg` 與 composition-root inspection 確認部署的 Go Job 呼叫 `Orchestrator.Run(ctx, id, nil)`，沒有 production Source Setting loader/concrete adapter，`CandidateArticle` 與 PostgreSQL schema 也沒有 citation 欄位；即使 ingress 恢復亦只能空跑，無法滿足 O4 公開來源/citation Done condition。故停止在任何 authenticated/data-mutating smoke step 之前，不勾選 Green/Verify/Commit。

- 2026-09-12：R8 deploy 後，revision `daily-news-api-00005-j82` 已以 commit `819e8f7` image 取得 100% traffic，public `/healthz` 仍為未進入 revision 的 Google HTML 404。為隔離 project/region 層級問題，短期部署無秘密、無資料庫的 `daily-news-edge-probe` 到同一 project/region；其 generated `run.app` URL 回 `200`。因此 404 已確認為 `daily-news-api` service hostname registration 異常，非 Cloud Run edge、region、IAM、ingress 或 R8 image。probe 已在驗證後移除。取得使用者明確授權後，精確刪除並以同一 production image、`daily-news-api` runtime service account、兩個 Secret Manager references、`ingress: all`、public invocation、1 CPU/512 MiB/maxScale 12 重建正式 service；新 revision `daily-news-api-00001-l2z` Ready 並取得 100% traffic，但 numeric 和 canonical `run.app` `/healthz` 仍均為 Google HTML 404。startup TCP probe 成功，且 `resource.type=cloud_run_revision` 沒有 health request log。故 service recreation 已排除，O4 需 Google Cloud Support/platform 修復 hostname routing 後才能繼續。

- 2026-09-12：依使用者指示嘗試建立 Google Cloud Support case。`daily-news-93f7b` 的 Support Console 顯示目前支援方案不提供 technical case，僅能「查看支援方案」；因此無法提交 case，且不會在未獲授權下購買或變更 billing/support offer。需由使用者附加合格 support plan，或由有資格的 billing/support administrator 從其帳戶建立 case，O4 才能恢復。

- 2026-09-12：依使用者授權建立 `daily-news-staging-api` fallback，並發現原先將 Google HTML `/healthz` 404 判為 hostname routing fault 的結論不成立：`/` 已到達 Go mux，真正 `/v1/categories` 與 `/v1/ingestion-runs/latest` 均由 application 回 `401 application/problem+json`。Root cause tracing 顯示 `cmd/api` 將 `api.NewHandler` 覆蓋 `httpapi.NewMux`，但前者沒有掛載 health handler。Red：新增 `TestHandlerServesHealthzWithoutDatabaseOrAuthentication`，`cd backend && go test ./internal/api -run TestHandlerServesHealthzWithoutDatabaseOrAuthentication -count=1` 如預期 `status = 404, want 200` 失敗。Green：共用 `httpapi.Healthz` 並在 production `api.NewHandler` 掛載 `GET /healthz`；同一 focused test 及 `go test ./internal/api ./internal/httpapi -count=1` 通過。Verify：`cd backend && go vet ./... && go test ./... && uv run pytest tests/test_container_contract.py -q && git diff --check` 全部通過（container contract 4 passed）；commit `c737370` 已推送，GitHub CI `34678546839` 和 deploy `34678546828` 均成功。GitHub Linux amd64 image 已部署為 `daily-news-api-00002-brl`，public `GET /v1/categories` 確認為預期 401。Google Frontend 對精確 `/healthz` 仍回 HTML 404，即使 application handler 已正確掛載，故 O4 以 versioned API authentication response 驗證 external reachability。受控 Job execution `daily-news-ingestion-lr6jv`（Run ID `9281481e-3192-4c36-9314-45ea1368536e`）完成成功；唯讀 DB 查詢確認 schema migration version 3、Run `succeeded`、0 attempts/candidates（尚無使用者 Category/source）。Green/Verify/Commit 仍待 allowlisted Firebase session 的實際 Category、source、Article/citation、併發 Run 與 mobile journey。

- 2026-09-12：取得使用者刪除授權後，移除短期診斷 services `daily-news-api-staging` 與 `daily-news-staging-api`；`gcloud run services list` 僅保留正式 `daily-news-api`，revision `daily-news-api-00002-brl`。cleanup 後重新呼叫正式 `/v1/categories`，仍為預期 `401 application/problem+json`。O4 仍等待 allowlisted Firebase session 才能做資料驗收。

- 2026-09-12：O4 preflight 後續驗證：以 malformed Bearer 對正式 API 的 `GET /v1/categories` 與 `POST /v1/categories` 均回 `401 application/problem+json`，未造成資料寫入。mobile public base URL 的 Dart define 為 `DAILY_NEWS_API_BASE_URL=https://daily-news-api-855124405761.asia-east1.run.app/v1/`。`cd apps/mobile && flutter test` 通過（25 passed）；`flutter test -d emulator-5554 integration_test/daily_news_flow_test.dart` 通過，驗證 Android device 上的 fake-server UI journey。該 integration test 明確使用 fake API，不能取代 allowlisted Firebase 對 staging 的登入與資料流驗收。O4 仍未勾選 Green/Verify/Commit。

- 2026-09-12：使用者在 iOS Simulator 點選 Google 登入後，按鈕持續 loading、未出現帳號選擇器。Root-cause inspection：iOS target bundle ID 仍為預設 `com.example.mobile`，缺少 `GoogleService-Info.plist` 與 Google OAuth callback URL scheme；Firebase Console 的 `daily-news-93f7b` General settings 更明確顯示「專案中沒有應用程式」。因此該 Firebase project 尚未註冊 Apple/Android/Web app，也沒有可供 mobile 使用的 OAuth client，真實 Firebase allowlist／Category／ingestion staging 驗收無法開始。需使用者選定 bundle ID 並授權註冊 Firebase Apple app、下載 app config、啟用/配置 Google Sign-In 後才能恢復 O4。

- 2026-09-12：Deployment correction（O4 尚未完成）：在已確認 Service Ready、100% traffic、`ingress: all` 與 `allUsers` Invoker binding 後，公開 `/healthz` 仍由 Google edge 回傳 HTML 404，未到達 Go handler。Red：`cd backend && uv run pytest tests/test_container_contract.py -q` 因 manifest 未宣告 public invocation 而如預期失敗（1 failed, 3 passed）。Green：Service manifest 明確加入 `run.googleapis.com/invoker-iam-disabled: 'true'`，使 Cloud Run admission 與 API 內 Firebase Bearer-token allowlist 分工明確，並由 contract test 固定此部署不變量。Verify：`cd backend && uv run pytest tests/test_container_contract.py -q && go vet ./... && go test ./...` 通過（container contract 4 passed；所有 Go package passing）；`actionlint .github/workflows/deploy.yml`、manifest assertion 與 `git diff --check` 均通過。待使用者 push 並完成 GitHub Actions 部署後，以公開 `/healthz` 和 allowlisted Firebase smoke test 驗證，才可完成 O4。

- 2026-09-12：CI correction（O4 尚未完成）：GitHub Actions run `34668466412` 顯示多個 PostgreSQL integration package 在 migration 前取得 `read: connection reset by peer`；各測試僅以 container 內 Unix socket 的 `pg_isready` 判定 ready，尚未確認 runner 經 Docker published TCP port 的 pgx 連線。Green：各 integration helper 在 migration 前以相同 `platform.OpenDB`／`PingContext` 對實際 PostgreSQL URL 重試最多 30 秒；成功後立即關閉 readiness pool。Verify：`cd backend && gofmt -w internal/platform/migrations_test.go internal/category/store_integration_test.go internal/news/store_integration_test.go internal/ingestion/runs_integration_test.go && GOFLAGS=-p=1 go test -count=1 ./...` 通過，所有 Go package passing；待下一次 GitHub Actions CI 實際驗證。

- 2026-09-12：使用者確認正式 iOS bundle ID `com.allenljf.dailynews` 與 Firebase Google Sign-In 支援信箱後，已在 `daily-news-93f7b` 註冊 Apple app，啟用 Google provider，並將公開名稱設為 `Daily News`。依 Firebase 產生的設定加入 `GoogleService-Info.plist`、Runner resource、Google OAuth callback URL scheme，且將 Debug／Profile／Release 的 Runner bundle ID 改為正式值；Flutter bootstrap 現在於 `runApp` 前執行 `Firebase.initializeApp()`。驗證：`cd apps/mobile && dart format lib/main.dart && plutil -lint ios/Runner/Info.plist ios/Runner/GoogleService-Info.plist && flutter test`（25 passed）；`flutter build ios --simulator --debug` 與 `flutter run -d 0A818455-5B3B-44A3-A7CB-9AF6681297A6 --no-resident --dart-define=DAILY_NEWS_API_BASE_URL=https://daily-news-api-855124405761.asia-east1.run.app/v1/` 成功。iPhone 17 Simulator 顯示可點擊的「使用 Google 登入」，不再卡在 loading。O4 仍待使用者完成 allowlisted Google 帳戶登入，以及真實 Category、source、ingestion、Article/citation 和併發 Run 驗收。

- 2026-09-12：allowlisted iOS session 登入後，Cloud Run 對 `GET /v1/categories` 連續回 `200`，但 mobile 顯示「無法讀取新聞類別」。根因是空資料庫的 Go `Store.List` 回傳 nil slice，JSON 為 `null`，違反 OpenAPI 的 array response 並被 Flutter decoder 拒絕。Red：新增真實 PostgreSQL + authenticated handler test，`cd backend && go test ./internal/category -run TestHandlerListsNoCategoriesAsJSONArray -count=1` 如預期失敗：`"null\\n"`，預期 `"[]\\n"`。Green：初始化空 `Response` slice，使空清單編碼為 `[]`。Verify：`cd backend && go vet ./... && go test ./... && git diff --check` 通過；commit `b7bb1fd` 已推送，GitHub Deploy `34695495956` 成功。部署後 iPhone 17 Simulator 點選「重試」，原錯誤改為正常的「新增新聞類別」空狀態。O4 仍待實際建立 Category/source、Run、Article/citation、併發 Run 驗收。

---

### Task 1: Persist and expose Category content language

**Depends on:** R8
**Parallel:** no — establishes the stored Category contract before ingestion and Flutter consume it.
**Files:** Create `backend/migrations/000004_category_content_language.{up,down}.sql`; modify `backend/internal/category/{category.go,handler_test.go,store_integration_test.go}`, `docs/contracts/daily-news.openapi.json`, `docs/requirements/daily-news.md`, and this task record.

- [x] **Step 1: Write the failing test** — prove an omitted request gives `zh-Hant`, explicit `en` gives 422, and the migrated table has a default/check constraint.
- [x] **Step 2: Verify RED** — `cd backend && go test ./internal/category -run 'ContentLanguage|StorePreserves' -count=1` failed to compile because `Response.ContentLanguage` did not exist.
- [x] **Step 3: Implement GREEN** — migration 4 adds `text NOT NULL DEFAULT 'zh-Hant' CHECK (content_language = 'zh-Hant')`; Category requests normalise omissions, reject explicit unsupported values, and Store list/create/update persists the field.
- [x] **Step 4: Verify GREEN** — ran `cd backend && gofmt -w internal/category && go test ./internal/category -count=1`; all category tests passed.
- [x] **Step 5: Commit** — intentionally not committed because this shared workspace contains unrelated user changes and the delegated task explicitly prohibits a commit.

**完成紀錄：**

- 2026-09-15：Red：新增 handler/store integration contracts後，`cd backend && go test ./internal/category -run 'ContentLanguage|StorePreserves' -count=1` 以 `Response.ContentLanguage undefined` 編譯失敗，確認 field、migration 與 SQL 尚未實作。Green：新增 PostgreSQL migration `000004_category_content_language`，其 `NOT NULL DEFAULT 'zh-Hant'` 回填既有 rows 並以 check constraint 限制唯一支援值；Category `Request`/`Response`、strict JSON validation 和 Store list/create/update 現在攜帶欄位。HTTP omission 正規化為 `zh-Hant`，明確傳送 `en` 回既有 422 Problem Details。OpenAPI response 將欄位標為 required，create/update request 將其列為可省略且 default 為 `zh-Hant`，以支援舊 Flutter client。驗證：focused Red 如上；`cd backend && go test ./internal/category -run 'ContentLanguage|StorePreserves' -count=1` 為 `ok`；`cd backend && gofmt -w internal/category && go test ./internal/category -count=1 && go vet ./... && go test ./... && git diff --check` 全部通過。PostgreSQL tests 以 disposable `postgres:16-alpine` 驗證 Store omission/response/list persistence、database default 和 unsupported `en` 的 constraint rejection。因共享工作樹已有不相關 mobile、staging 文件和 task-record 修改，依指示未 commit。

### Task 2: Carry language into ingestion and reject known non-Traditional candidates

**Depends on:** Task 1
**Parallel:** no — consumes the persisted Category value.
**Files:** Modify `backend/internal/ingestion/{orchestrator.go,planner.go,planner_integration_test.go,rss.go,rss_test.go}` and this task record.

- [x] **Step 1: Write failing tests** — assert planned work contains `zh-Hant`; explicit `zh-Hans` RSS entry metadata is rejected while `zh-Hant` is retained.
- [x] **Step 2: Verify RED** — `cd backend && go test ./internal/ingestion -run 'WorkPlanner|Traditional' -count=1` first failed to compile because `SourceWork.ContentLanguage` did not exist; after the plumbing-only change, it failed with two candidates instead of one because the adapter did not gate language.
- [x] **Step 3: Implement GREEN** — select/scan `categories.content_language` into `SourceWork`; RSS/Atom accepts unlabelled entries but rejects entries whose explicit language metadata differs from the requested value. No source text is transformed.
- [x] **Step 4: Verify GREEN** — `cd backend && gofmt -w internal/ingestion && go test ./internal/ingestion -count=1` passed; `cd backend && go vet ./... && go test ./... && git diff --check` passed.
- [x] **Step 5: Commit** — intentionally not committed because this shared dirty workspace and delegated task prohibit commits.

**完成紀錄：**

- 2026-09-15：Red：planner contract 因缺少 `SourceWork.ContentLanguage` 失敗；新增欄位/SQL select 後，RSS contract 以 `article count = 2, want 1` 失敗，證明明確 `zh-Hans` metadata 尚未被過濾。Green：WorkPlanner 現在選取並 scan migration 4 持久化的 `categories.content_language`；RSS/Atom 在 candidate/item、channel 或 feed 層找到第一個明確語言 metadata 時，僅接受與 `SourceWork.ContentLanguage` case-insensitively 相符的值。metadata 缺失保持 eligible，因現有 adapter 沒有值得信賴的繁體中文文字 classifier；沒有做文字轉換。驗證：`cd backend && gofmt -w internal/ingestion && go test ./internal/ingestion -count=1` 通過；`cd backend && go vet ./... && go test ./... && git diff --check` 通過。未修改 Flutter、Category contract 或 migrations；未 commit。

### Task 4: Category content-language full verification record

**Depends on:** Task 1, Task 2, and the linked Flutter implementation.
**Files:** Modify this task record only.

- [x] **Step 1: Run backend verification** — `cd backend && go vet ./... && go test ./...` passed.
- [ ] **Step 2: Run Flutter verification** — `flutter analyze`, `flutter test`, and the feature-scoped `--diff HEAD` rules check passed, but the required full command did not complete because the host `python3` has no `yaml` module for `check-rules.py`.
- [x] **Step 3: Validate contract and diff** — `python3 -m json.tool docs/contracts/daily-news.openapi.json >/dev/null && git diff --check` passed.
- [x] **Step 4: Record the verification evidence** — recorded below; the language feature is not marked fully verified while the required rules check remains blocked.
- [x] **Step 5: Commit** — intentionally not committed in the shared dirty workspace.

**完成紀錄：**

- 2026-09-15：`cd backend && go vet ./... && go test ./...` exit 0；`go test` 的所有 package 通過（含 `internal/category` 與 `internal/ingestion`）。要求的完整 Flutter command `cd apps/mobile && flutter analyze && flutter test && cd ../.. && python3 flutter-dev-guide/tools/check-rules.py --all` 先完成 `flutter analyze`（`No issues found!`）與 `flutter test`（`33` passed），但最後的 checker 因 host `python3` 缺少 `yaml` module 以 `ModuleNotFoundError` exit 1。以暫時的 PyYAML environment 重試 `uv run --with pyyaml python flutter-dev-guide/tools/check-rules.py --all` 後，checker 會遞迴掃描 gitignored 的 `apps/mobile/build/ios/SourcePackages/` Firebase package fixtures，並報出其測試 delay／fixture key 規則；這些不是本功能檔案，且未修改或刪除。Feature-scoped `uv run --with pyyaml python flutter-dev-guide/tools/check-rules.py --diff HEAD` exit 0、無 violations；它只涵蓋 HEAD 以來的 tracked 變更，不能取代仍失敗的 required `--all` command。`python3 -m json.tool docs/contracts/daily-news.openapi.json >/dev/null && git diff --check` exit 0。未執行 staging 或外部驗證，未 commit，並保留共享工作樹既有未提交變更。

### Task 3: Delete a Category from its news list

**Depends on:** F4, F5
**Parallel:** no — reuses the Category deletion contract and adds its Category-detail entry point.
**Files:** Modify `apps/mobile/lib/features/categories/{application/category_controller.dart,data/category_{remote_service,repository}.dart}`, `apps/mobile/lib/features/news/presentation/news_list_screen.dart`, `apps/mobile/lib/l10n/app_zh.arb`, `apps/mobile/test/features/news/news_list_test.dart`, `docs/requirements/daily-news.md`, and this task record.

- [x] **Step 1: Write the failing test** — prove the news-list app bar exposes an accessible Category delete action; cancellation performs no deletion, while confirmation calls the Category repository, returns to the home route, and explains that shared Articles are retained.
- [x] **Step 2: Verify RED** — `cd apps/mobile && flutter test test/features/news/news_list_test.dart` fails because the Category delete command and app-bar confirmation flow do not exist.
- [x] **Step 3: Implement GREEN** — wire `DELETE /v1/categories/{id}` through the existing Category remote service, repository, and controller; add localized confirmation/failure strings and the app-bar delete icon. On successful deletion, refresh Category state and navigate home; on failure, retain the list and show its error state.
- [x] **Step 4: Verify GREEN** — `cd apps/mobile && flutter analyze && flutter test test/features/news/news_list_test.dart && flutter test`.
- [x] **Step 5: Verify guide rules and diff** — `uv run --with pyyaml python flutter-dev-guide/tools/check-rules.py --files apps/mobile/lib/features/categories/application/category_controller.dart apps/mobile/lib/features/categories/data/category_remote_service.dart apps/mobile/lib/features/categories/data/category_repository.dart apps/mobile/lib/features/news/presentation/news_list_screen.dart apps/mobile/test/features/news/news_list_test.dart && git diff --check`.
- [x] **Step 6: Commit** — did not commit because the shared worktree contains unrelated changes.

**完成紀錄：**

- 2026-09-16：需求確認為刪除 Category 時停用其 Source Setting／後續擷取並移除 Category Article 關聯，保留仍被其他 Category 引用的 Article；更新需求規格以將入口定在 Category 新聞列表右上角。Red：新增 Router-backed widget test 後，`cd apps/mobile && flutter test test/features/news/news_list_test.dart` 以找不到 `Icons.delete_outline` 失敗，確認入口尚不存在。Green：Category remote service、repository 與 controller 現在沿用既有 `DELETE /v1/categories/{id}`；新聞列表 AppBar 提供有本地化 tooltip 的刪除按鈕、確認 dialog、取消分支、失敗 SnackBar，以及成功後重新讀取 Category state 並導航至首頁。測試覆蓋文章保留說明、取消不刪除、確認刪除、state reload 與回首頁；更新其他 fake repository 以符合擴充的 contract。驗證：`cd apps/mobile && flutter analyze && flutter test test/features/news/news_list_test.dart && flutter test` 通過（focused 6 passed，full 34 passed）；`uv run --with pyyaml python flutter-dev-guide/tools/check-rules.py --files ... && git diff --check` exit 0、無 violations。未 commit，因 shared worktree 包含既有無關修改。

### Task 5: Route Category deletion through the API composition root

**Depends on:** Task 3
**Parallel:** no — repairs the deployed API path used by the new Flutter control.
**Files:** Modify `backend/internal/api/{server.go,server_test.go}` and this task record.

- [x] **Step 1: Reproduce RED** — add an authenticated-route contract for `DELETE /v1/categories/{id}`; `cd backend && go test ./internal/api -run TestHandlerRoutesEveryV1FamilyThroughSharedAuthentication -count=1` returns 404 instead of 401.
- [x] **Step 2: Implement GREEN** — route every `/v1/categories/{id}` request to the Category handler, preserving `/news` delegation to the Article handler.
- [x] **Step 3: Verify GREEN** — `cd backend && gofmt -w internal/api/server.go internal/api/server_test.go && go test ./internal/api -run TestHandlerRoutesEveryV1FamilyThroughSharedAuthentication -count=1 && go vet ./... && go test ./... && git diff --check`.
- [ ] **Step 4: Deploy and verify production** — publish the isolated backend fix through the existing deploy workflow, then confirm a Category DELETE no longer returns 404.

**完成紀錄：**

- 2026-09-16：Cloud Run request logs supplied the production repro: `DELETE /v1/categories/{id}` repeatedly returned 404 while `GET /v1/categories/{id}/news` for the same id returned 200. The deployed revision `daily-news-api-00007-4xp` uses image commit `37b7f5b`, which contains the Category DELETE handler. Red added the DELETE request to the composition-root authentication contract; it failed with `status = 404, want 401`, proving the outer router never reached that handler. Green broadens only the Category path dispatch; `/news` remains delegated to the Article handler. Focused test, `go vet ./...`, full `go test ./...`, and `git diff --check` all passed. Production deployment and the post-deploy DELETE check are pending.

### Task 6: Add English as a Category content language

**Depends on:** Task 1, Task 2
**Parallel:** no — expands the persisted Category contract and its Flutter selector together.
**Files:** Create `backend/migrations/000005_category_english_content_language.{up,down}.sql`; modify `backend/internal/category/{category.go,handler_test.go,store_integration_test.go}`, `backend/internal/ingestion/rss_test.go`, `apps/mobile/lib/features/categories/{data/category.dart,presentation/category_settings_sheet.dart}`, `apps/mobile/lib/l10n/{app_zh.arb,app_localizations.dart,app_localizations_zh.dart}`, `apps/mobile/test/features/categories/category_sheet_test.dart`, `docs/contracts/daily-news.openapi.json`, `docs/requirements/daily-news.md`, and this task record.

- [x] **Step 1: Write failing tests** — prove the Category validation/store and Flutter selector accept `en` for a new Category.
- [x] **Step 2: Verify RED** — focused Category test did not compile because `EnglishContentLanguage` was absent; the Flutter widget test failed because no `英文` menu item existed.
- [x] **Step 3: Implement GREEN** — allow only `zh-Hant` and `en`, expand the database constraint in migration 5, expose both OpenAPI enum values, and add the localized English selector option; keep `zh-Hant` as the omission default.
- [x] **Step 4: Verify GREEN** — focused and full backend/Flutter checks, guide rules, contract validation, and diff check passed.
- [x] **Step 5: Commit** — intentionally not committed because the shared worktree contains unrelated changes.

**完成紀錄：**

- 2026-09-16：Red：`cd backend && go test ./internal/category -run 'ValidateAcceptsEnglishContentLanguage|StorePreservesDefaultContentLanguageAndDatabaseConstraint' -count=1` 因 `EnglishContentLanguage` 尚未定義而編譯失敗；`cd apps/mobile && flutter test test/features/categories/category_sheet_test.dart --plain-name 'English content language can be selected for a new Category'` 因沒有 `英文` 下拉選項失敗。Green：新增 migration `000005_category_english_content_language`，將 `categories.content_language` 約束擴為 `zh-Hant` 與 `en`，維持 omitted request／既有資料的 `zh-Hant` 預設；Category validation、OpenAPI contract、Flutter value object 與本地化 selector 一併支援 `en`。RSS regression test 確認英文 Category 只保留標示 `en` 的候選。驗證：`cd backend && go vet ./... && go test ./... -count=1` 通過；`cd apps/mobile && flutter analyze && flutter test` 通過（35 passed）；`uv run --with pyyaml python flutter-dev-guide/tools/check-rules.py --files apps/mobile/lib/features/categories/data/category.dart apps/mobile/lib/features/categories/presentation/category_settings_sheet.dart apps/mobile/test/features/categories/category_sheet_test.dart`、`python3 -m json.tool docs/contracts/daily-news.openapi.json >/dev/null` 與 `git diff --check` 均 exit 0。未 commit，避免混入共享工作樹的無關變更。

### Task 7: Support homepage Source Settings with feed discovery and web search

**Depends on:** R8
**Parallel:** no — changes the Job's source-adapter composition while retaining the existing Article write boundary.
**Files:** Create `backend/internal/ingestion/{safehttp.go,web.go,googlesearch.go,meta.go,router.go}` and their `_test.go` files; modify `backend/internal/ingestion/rss.go`, `backend/cmd/daily-news-job/{main.go,main_test.go}`, `backend/.env.example`, `infra/cloud-run/job.yaml`, `infra/docs/secret-inventory.md`, `docs/requirements/daily-news.md`, `docs/superpowers/specs/2026-09-16-web-source-adapters-design.md`, and this task record.

- [x] **Step 1: Write failing tests** — prove an HTML homepage discovers a relative same-host RSS/Atom feed; cross-host, loopback/private, invalid and oversized discovery responses are rejected; a homepage without a feed delegates to a fake search adapter; off-host, malformed and over-limit Google Custom Search candidates are rejected; a missing search credential is isolated; YouTube sends a composed query with `type=video`, `maxResults=10` and `relevanceLanguage` and maps public video URLs to canonical URL + citation; the Job composition routes hosts through a composite adapter instead of treating every HTTP(S) URL as RSS.
- [x] **Step 2: Verify RED** — `cd backend && go test ./internal/ingestion ./cmd/daily-news-job -run 'WebSource|GoogleSearch|YouTube|HostRouter|JobComposition' -count=1` fails because the composite, discovery and platform adapters do not exist.
- [x] **Step 3: Implement GREEN** — add a bounded, SSRF-safe fetcher; compose direct RSS, same-host feed discovery and Google Custom Search JSON API behind `SourceAdapter`; add the YouTube `search.list` keyword adapter; route Facebook/Instagram/Threads to a restricted Meta boundary that never reaches the general search adapter.
- [x] **Step 4: Verify GREEN locally** — `cd backend && gofmt -w ... && go vet ./... && go test ./...`, container build verification, and `git diff --check`.
- [x] **Step 5: Add Secret Manager runtime configuration** — read `GOOGLE_CSE_API_KEY`/`GOOGLE_CSE_ID`/`GOOGLE_CSE_DATE_RESTRICT`/`YOUTUBE_API_KEY` from the Job environment, declare the Secret Manager references in `infra/cloud-run/job.yaml`, and document the variable names and acquisition locations without values.
- [ ] **Step 6: Redacted staging homepage Source Setting Run** — after the secrets exist and the image is deployed, run one homepage-only disposable Source Setting and record only IDs and aggregate counts. Deferred until the user creates the secrets.
- [x] **Step 7: Record and commit** — check off the steps, add exact evidence below, and commit only verified task files if the shared worktree permits.

**完成紀錄：**

- 2026-09-16：Red：新增 `safehttp_test.go`、`web_test.go`、`googlesearch_test.go`、`youtube_test.go`、`router_test.go` 及 Job composition 與 container contract 斷言後，`cd backend && go test ./internal/ingestion ./cmd/daily-news-job -run 'WebSource|GoogleSearch|YouTube|HostRouter|JobComposition' -count=1` 因 `NewGoogleSearchAdapter`、`ErrWebSearchNotConfigured`、`NewHostRouter` 等尚未定義而 build failed，`TestJobCompositionUsesCompositeSourceAdapter` 亦以 `job composition does not build the composite web source adapter` 失敗。Green：新增 `safehttp.go`（15s timeout、2 MiB limit、redirect cap 5、dial 前拒絕 loopback/link-local/private/ULA/multicast）、`web.go`（直接 feed → `rel="alternate"` 同 host/subdomain 發現 → web search fallback）、`googlesearch.go`（Custom Search JSON API、`key`/`cx`/`q`/`num=10`/`dateRestrict`、同 host 過濾、`pagemap.metatags` 取 `published_at`、缺憑證回 `web search is not configured`）、`meta.go`（Facebook/Instagram/Threads 未來邊界，回 `ErrMetaAdapterUnavailable` 且絕不 fallback 至一般搜尋）、`router.go`（YouTube → YouTube adapter、Meta hosts → 受限邊界、其餘 → web），並重構 `rss.go` 抽出 `parseFeed`（只接受 root 為 rss/feed/rdf）。`main.go` 改以 `NewHostRouter`/`NewWebSourceAdapter`/`NewGoogleSearchAdapter` 組合。Verify：`cd backend && gofmt -l . && go vet ./...` 無輸出；`cd backend && GOFLAGS=-p=1 go test -count=1 ./...` 全部 package `ok`（含 PostgreSQL integration）；`cd backend && uv run pytest tests/test_container_contract.py -q` 為 `5 passed`；container build 成功，`docker run --rm daily-news-backend:task7 /app/daily-news-job` 在缺 `RUN_ID` 時 exit 2、`id -u` 為 100；`git diff --check` exit 0。需求變更：使用者決定以 Google Custom Search JSON API 取代 Gemini（回應舊資料問題）、且 Threads keyword search 先不做，因此 `gemini*.go` 與 `threads*.go` 已刪除，`GEMINI_API_KEY`、`GEMINI_MODEL`、`THREADS_ACCESS_TOKEN` 已自部署設定與文件移除；Threads/Facebook/Instagram host 改由受限 Meta 邊界處理。環境限制：本機與 Docker build network 被 ISP 攔截 `proxy.golang.org`（TLS 憑證不符），因此以同一 Dockerfile 內容加 `ENV GOPROXY=https://goproxy.cn,direct` 的臨時檔案完成容器建置驗證；checked-in `backend/Dockerfile` 未修改。Step 6 需先由使用者建立 Secret Manager secrets 並部署，尚未執行；未 commit，保留共享工作樹既有無關變更。

### Task 8: Keep Category Articles visible after Source Setting replacement

**Depends on:** R4, Task 7
**Parallel:** no — repairs the News read model used by the Flutter list and detail.
**Files:** Modify `backend/internal/news/{news.go,store_integration_test.go}` and this task record.

- [x] **Step 1: Write the failing test** — soft-delete a Category's Source Setting (as a Category edit does) and assert its existing Category Articles still appear in `List` and `Detail`.
- [x] **Step 2: Verify RED** — `cd backend && go test ./internal/news -run TestStoreKeepsArticlesAfterSourceSettingReplacement -count=1` failed with `items = []news.listItem{}` against the old query.
- [x] **Step 3: Implement GREEN** — the News list/detail queries no longer require `source_settings.deleted_at IS NULL`; Category Article visibility follows the Category and Article lifecycle instead of the current Source Setting row.
- [x] **Step 4: Verify GREEN** — `cd backend && gofmt -l . && go vet ./... && GOFLAGS=-p=1 go test -count=1 ./...` passed for every package.
- [x] **Step 5: Commit and deploy** — commit the verified fix and publish it so the existing Articles become visible again.

**完成紀錄：**

- 2026-09-16：使用者回報「編輯 Category 後新聞全部不見」。根因：`category.Store.Update` 會 soft delete 既有 `source_settings` 再插入新列（新 id），但文章列表／詳情的 SQL 內含 `AND ss.deleted_at IS NULL`，因此任何 Category 編輯都會讓既有 `category_articles` 從 Flutter 列表與詳情消失（`backend/internal/news/news.go:87`、`:133`）。Red：`git stash` 只還原 `news.go` 後，`cd backend && go test ./internal/news -run TestStoreKeepsArticlesAfterSourceSettingReplacement -count=1` 以 `items = []news.listItem{}` 失敗，確認查詢條件即為成因。Green：移除兩處 `ss.deleted_at IS NULL`，讓 Category Article 的可見性只受 Category、Category Article 與 Article 的 soft delete／到期影響；同一 focused test 通過。Verify：`cd backend && gofmt -l . && go vet ./... && GOFLAGS=-p=1 go test -count=1 ./...` 全部 package `ok`。未修改 OpenAPI 或 Flutter contract。Fix commit `476b8af`（`fix: keep Category Articles after Source Setting replacement`），已 push 並由 Deploy Cloud Run workflow 發佈。

### Task 9: Accept Traditional Chinese feed language variants

**Depends on:** Task 2, Task 6
**Files:** Modify `backend/internal/ingestion/{rss.go,rss_test.go}` and this task record.

- [x] **Step 1: Decide** — the user only wants Traditional Chinese content, so keep filtering but treat `zh`, `zh-TW`, `zh-HK` and `zh-MO` as the same language as `zh-Hant`; still reject `zh-Hans`/`zh-CN`/`zh-SG` and non-Chinese languages.
- [x] **Step 2: Implement and verify** — `matchesContentLanguage` now normalises the requested and found language before comparing; a new regression test keeps a `zh-TW` channel for a `zh-Hant` Category, and the existing Traditional/Simplified/English tests still pass.

**完成紀錄：**

- 2026-09-16：使用者回報「以前不限定繁體中文時還有資料」。`matchesContentLanguage` 把 channel 或 feed 層的 `<language>` 當成每個 entry 的語言，因此多個繁中站台（內部標示 `zh-TW`）對 `zh-Hant` Category 全數被丟棄。Red/Green：新增 `TestRSSAdapterAcceptsTraditionalChineseVariants`，並將比對改為先做語言正規化（`zh`/`zh-TW`/`zh-HK`/`zh-MO` → `zh-Hant`，`zh-CN`/`zh-SG`/`zh-Hans` → `zh-Hans`，`en-*` → `en`）。驗證：本機探針對 `ithome.com.tw`、`inside.com.tw`、`technews.tw`、`thenewslens.com` 由 eligible=0 恢復為 eligible=10；`cd backend && gofmt -w internal/ingestion && go test ./internal/ingestion -run 'RSSCategory|ContentLanguage|Traditional|English' -count=1` 通過。

### Task 10: Replace the general-web search fallback with SerpApi Google News

**Depends on:** Task 7, Task 9
**Parallel:** no — depends on external search API availability.
**Files:** Delete `backend/internal/ingestion/googlesearch.go` and its test; create `backend/internal/ingestion/serpapi.go` and `serpapi_test.go`; modify `backend/internal/ingestion/{web.go,planner.go,planner_integration_test.go}`, `backend/cmd/daily-news-job/main.go`, `backend/.env.example`, `backend/tests/test_container_contract.py`, `infra/cloud-run/job.yaml`, `infra/docs/secret-inventory.md`, `docs/requirements/daily-news.md`, `docs/superpowers/specs/2026-09-16-web-source-adapters-design.md`, and this task record.

- [x] **Step 1: Research** — the Google Custom Search JSON API is closed to new customers, so this project gets `403`. SerpApi exposes Google News through `engine=google_news` with `site:`, `when:`, `hl` and `gl`, which fits the fallback.
- [x] **Step 2: Implement** — replace the Google adapter with `SerpAPIAdapter`; the Web adapter uses it when a site has no feed (`site:<host>`), and the planner now also plans unspecified Source Settings for whole-web search.
- [x] **Step 3: Configure** — swap the `GOOGLE_CSE_*` secrets for `SERPAPI_API_KEY` plus the non-secret `SERPAPI_WHEN` in the Job manifest, `.env.example`, and the secret inventory.
- [x] **Step 4: Verify** — `cd backend && gofmt -l . && go vet ./... && GOFLAGS=-p=1 go test -count=1 ./...` and the container contract test pass.
- [x] **Step 5: Commit and deploy** — commit the verified change and publish it.
- [x] **Step 6: Live check** — after the user stores `SERPAPI_API_KEY`, run a homepage-only and an unspecified Source Setting to confirm the fallback returns results.

**完成紀錄：**

- 2026-09-16：Google Custom Search JSON API 對新客戶關閉（`customsearch.googleapis.com` 已啟用、key 正確仍 403）。改用 SerpApi。新增 `SerpAPIAdapter`（`engine=google_news`、`site:<host>`／全網、`when:<window>`、`hl`/`gl` 依內容語言、每來源 10 筆、缺 key 回 `web search is not configured`、錯誤不洩漏 key）；`WebSourceAdapter` 對無 feed 網站與未指定網站都走此 fallback；`WorkPlanner` 現在也納入空白 `website_input` 的 Source Setting。設定由 `GOOGLE_CSE_*` 改為 `SERPAPI_API_KEY`（Secret）與 `SERPAPI_WHEN`（預設 `7d`）。Verify：`cd backend && gofmt -l . && go vet ./... && go test ./internal/ingestion ./cmd/daily-news-job -count=1` 通過；`uv run pytest tests/test_container_contract.py -q` 通過；`GOFLAGS=-p=1 go test -count=1 ./...` 全部 `ok`。Fix commit `e33ea1c`（`feat: replace web search fallback with SerpApi Google News`）已 push，Deploy Cloud Run 成功，Job `Ready: True`、image `e33ea1c`、env 含 `SERPAPI_API_KEY` 與 `SERPAPI_WHEN=7d`。Live check：使用者建立 `SERPAPI_API_KEY` 並授權後，以真實 API 的臨時探針驗證 `site:bnext.com.tw` 回 1 筆、未指定網站的全網查詢回 10 筆繁中新聞，且錯誤不含 key；探針已刪除。未修改 OpenAPI 或 Flutter contract。

### Task 11: Remove a deleted Article from its Category news list

**Depends on:** F5
**Parallel:** no — repairs the existing News list/detail state sync.
**Files:** Modify `apps/mobile/lib/features/news/application/news_detail_controller.dart`, `apps/mobile/test/features/news/news_list_test.dart`, and this task record.

- [x] **Step 1: Write the failing test** — deleting an Article from the detail screen must also remove it from the Category news list when the user returns.
- [x] **Step 2: Verify RED** — the new Router-backed widget test failed with `Found 1 widget with text "Article 1"` after navigating back, proving the list kept the deleted row.
- [x] **Step 3: Implement GREEN** — after a successful `DELETE /v1/news/{newsId}`, invalidate `newsListControllerProvider(categoryId)` so the list reloads and drops the Article.
- [x] **Step 4: Verify GREEN** — `cd apps/mobile && flutter analyze && flutter test test/features/news && flutter test` and the guide rules check passed.

**完成紀錄：**

- 2026-09-17：使用者回報「刪除新聞後，列表還存在」。根因：`NewsDetailController.delete()` 只把 detail state 標為 `deleted`，從未通知 `newsListControllerProvider`，所以返回列表時舊的 Article row 仍在記憶體 state 中（`apps/mobile/lib/features/news/application/news_detail_controller.dart:50`）。Red：在 `test/features/news/news_list_test.dart` 新增 Router-backed 測試，讓 fake repository 依 `deletedIds` 過濾列表結果，並在刪除後按 BackButton 返回；`cd apps/mobile && flutter test test/features/news/news_list_test.dart --plain-name 'deleting an Article removes it from its Category list'` 以 `Found 1 widget with text "Article 1"` 失敗，確認列表未同步。Green：刪除成功後呼叫 `ref.invalidate(newsListControllerProvider(args.categoryId))`，使列表在返回時重新讀取 API；focused test 通過。Verify：`cd apps/mobile && flutter analyze` 為 `No issues found!`；`flutter test test/features/news` 為 9 passed；`flutter test` 為 39 passed；`uv run --with pyyaml python flutter-dev-guide/tools/check-rules.py --files apps/mobile/lib/features/news/application/news_detail_controller.dart apps/mobile/test/features/news/news_list_test.dart` 與 `git diff --check` exit 0。未 commit，因 shared worktree 仍有不相關的 Category sheet 變更。

### Task 12: Enable running the app on Android

**Depends on:** F1
**Parallel:** no — changes the Android application id and Firebase bootstrap.
**Files:** Modify `apps/mobile/android/{settings.gradle.kts,app/build.gradle.kts,app/src/main/AndroidManifest.xml}`, move `apps/mobile/android/app/src/main/kotlin/com/example/mobile/MainActivity.kt` to `.../com/allenljf/dailynews/MainActivity.kt`, `apps/mobile/README.md`, and this task record.

- [x] **Step 1: Reproduce RED** — Android still used the default `com.example.mobile` package with no `google-services.json`, so `Firebase.initializeApp()` had no options and Google sign-in could not run.
- [x] **Step 2: Implement GREEN** — rename the application id to `com.allenljf.dailynews` (matching the iOS bundle id), apply `com.google.gms.google-services` 4.5.0 in `settings.gradle.kts` and `app/build.gradle.kts`, add the runtime `INTERNET` permission and set the app label to `Daily News`.
- [x] **Step 3: Verify GREEN** — `cd apps/mobile && flutter build apk --debug` succeeded with a placeholder config, and the wiring was validated with `./gradlew :app:processDebugGoogleServices`.
- [x] **Step 4: User provides Firebase config** — add the Android app in Firebase `daily-news-93f7b` and place the real `apps/mobile/android/app/google-services.json` (with the web client id) before running on a device.

**完成紀錄：**

- 2026-09-17：使用者要求「android 端也要可以 run」。現況：Android `applicationId`／`namespace` 仍是 `com.example.mobile`，且 repo 內沒有 Android 的 Firebase 設定，`Firebase.initializeApp()` 在 Android 上因缺 options 直接失敗。Green：`app/build.gradle.kts` 與 `settings.gradle.kts` 改用 `com.allenljf.dailynews` 並套用 `com.google.gms.google-services` 4.5.0（外掛會由 `google-services.json` 產生 `default_web_client_id`，`google_sign_in` 7.x 的 Android 實作以它作為 `serverClientId`）；`MainActivity.kt` 以 `git mv` 移到 `com/allenljf/dailynews/`；main manifest 加入 `android.permission.INTERNET` 並將 label 改為 `Daily News`。Verify：以暫時 placeholder 的 `google-services.json` 執行 `./gradlew :app:processDebugGoogleServices`（BUILD SUCCESSFUL）與 `cd apps/mobile && flutter build apk --debug`（✓ Built `build/app/outputs/flutter-apk/app-debug.apk`），驗證 package 名稱、外掛 wiring 與 Kotlin 套件搬移可編譯；placeholder 已刪除。`apps/mobile/README.md` 記錄 Firebase Console 步驟與本機 debug keystore SHA-1（`50:8D:37:A4:56:1E:63:00:A8:07:3F:B6:A9:11:61:15:06:F0:F7:97`）的取得指令。
- 2026-09-17（續）：使用者完成 Firebase Console 兩步後，`apps/mobile/android/app/google-services.json` 已存在，含 Android app `1:855124405761:android:4b9f4ec31436982eb11a00` 與 web client `client_type: 3`（`855124405761-1lgr2a60usqrreo9e03lq89jud4ioncm`），因此 `default_web_client_id` 可被產生。Verify：啟動 `Pixel_10_Pro_XL`（Google Play，API 37.1）模擬器後 `flutter run -d emulator-5554 --no-resident` 成功建置、安裝並啟動；logcat 顯示 `FirebaseInitProvider: FirebaseApp initialization successful`，畫面為「每日新聞／使用 Google 登入」登入頁，無 Dart 例外。點擊登入鈕會進入 Google 帳號流程（`Checking info…` → `Sign in with ease`），證明 Credential Manager 接受了 client 設定；模擬器沒有已登入的 Google 帳號，因此實際登入仍須由使用者加入 allowlisted 帳號完成。`flutter test -d emulator-5554 integration_test/daily_news_flow_test.dart` 以新 package 名稱通過（`1 passed`）。未 commit，因 shared worktree 仍有不相關的 Category sheet 變更。

### Task 13: Show the original Article page in an in-app WebView

**Depends on:** F5, Task 11
**Parallel:** no — changes the News detail presentation and its routing-facing widget tests.
**Files:** Create `apps/mobile/lib/features/news/presentation/article_web_view.dart`; modify `apps/mobile/{pubspec.yaml,pubspec.lock}`, `apps/mobile/lib/features/news/presentation/news_detail_screen.dart`, `apps/mobile/lib/l10n/{app_zh.arb,app_localizations.dart,app_localizations_zh.dart}`, `apps/mobile/test/features/news/news_list_test.dart`, `apps/mobile/integration_test/daily_news_flow_test.dart`, `docs/requirements/daily-news.md`, and this task record.

- [x] **Step 1: Write the failing test** — assert the detail loads the Article's `canonicalUrl` through an overridable WebView seam, keeps the permanent/delete controls, and renders an explicit error state instead of a WebView when the URL is not HTTPS.
- [x] **Step 2: Verify RED** — `cd apps/mobile && flutter test test/features/news/news_list_test.dart` failed because `ArticleWebView` and its provider did not exist.
- [x] **Step 3: Implement GREEN** — added `webview_flutter`; built `ArticleWebView` that validates HTTPS before loading, renders the platform WebView with JavaScript for page rendering but no file access or bridge, blocks non-HTTP(S) navigation, and shows loading/error states with retry. The detail screen embeds it below the Article summary and moves permanent/delete to localized app-bar actions.
- [x] **Step 4: Verify GREEN** — `cd apps/mobile && flutter analyze && flutter test` passed (40 passed); `flutter test -d 0A818455-5B3B-44A3-A7CB-9AF6681297A6 integration_test/daily_news_flow_test.dart` passed (1 passed) because no Android emulator could boot on this host.
- [x] **Step 5: Verify guide rules and diff** — `uv run --with pyyaml python flutter-dev-guide/tools/check-rules.py --files apps/mobile/lib/features/news/presentation/article_web_view.dart apps/mobile/lib/features/news/presentation/news_detail_screen.dart apps/mobile/test/features/news/news_list_test.dart && git diff --check` passed.
- [x] **Step 6: Commit** — intentionally not committed because the shared worktree contains unrelated changes.

**完成紀錄：**

- 2026-09-17：使用者要求「新聞詳情的內文不要只放連結，用 WebView 直接讀取網址顯示網頁內容」。Red：新增 `articleWebViewBuilderProvider` seam、`articleWebViewKey`、`articleWebViewBlockedKey` 斷言後，`cd apps/mobile && flutter test test/features/news/news_list_test.dart` 因 `ArticleWebView`、`ArticleWebViewBuilder`、provider 與 keys 尚未存在而編譯失敗（`Type 'ArticleWebViewBuilder' not found`、`Undefined name 'articleWebViewKey'`）。Green：加入 `webview_flutter ^4.14.1`；新增 `article_web_view.dart`，`ArticleWebView` 只接受 `https` 且 host 非空的 URL（否則顯示 `articleWebViewBlockedKey` 與「無法顯示原文網頁」），平台 WebView 以 `JavaScriptMode.unrestricted` 渲染公開新聞頁、不設 JavaScript channel 或檔案存取、`onNavigationRequest` 只放行 http(s)、主頁載入失敗時顯示 `articleWebViewErrorKey` 與重試；`news_detail_screen.dart` 在標題／來源 chip／摘要下方嵌入 `ArticleWebView(canonicalUrl)`，並把「設為永久／刪除新聞」移到有本地化 tooltip 的 AppBar icon actions。`flutter gen-l10n` 產生 `newsOriginalPageUnavailable`、`newsOriginalPageFailed`。Verify：focused detail tests 由 Red 轉 `2 passed`，`flutter analyze` 為 `No issues found!`，`flutter test` 為 `40 passed`；`flutter test -d 0A818455-5B3B-44A3-A7CB-9AF6681297A6 integration_test/daily_news_flow_test.dart` 為 `1 passed`（Android `Pixel_10_Pro_XL` emulator 啟動失敗，改用 iOS Simulator；integration test 以 provider override 取代 platform WebView）；`uv run --with pyyaml python flutter-dev-guide/tools/check-rules.py --files ...` 與 `git diff --check` exit 0。未 commit，因 shared worktree 仍含使用者進行中的 auth／theme／assets 變更。

### Task 14: Toggle permanence and return to the list after deleting an Article

**Depends on:** F5, Task 11, Task 13
**Parallel:** no — changes the News detail mutation flow and its presentation.
**Files:** Modify `apps/mobile/lib/features/news/{data/news.dart,data/news_repository.dart,data/news_remote_service.dart,application/news_detail_controller.dart,presentation/news_detail_screen.dart}`, `apps/mobile/lib/l10n/app_zh.arb`, `apps/mobile/test/features/news/news_list_test.dart`, `apps/mobile/test_support/fake_api_server.dart`, `apps/mobile/integration_test/daily_news_flow_test.dart`, `docs/requirements/daily-news.md`, and this task record.

- [x] **Step 1: Write the failing test** — tapping the permanence control on an Article that is already permanent must send `permanent: false` and restore its expiry; deleting an Article must return to the Category news list with the Article gone.
- [x] **Step 2: Verify RED** — `cd apps/mobile && flutter test test/features/news/news_list_test.dart --plain-name 'tapping permanence again restores the original expiry'` failed because the control was disabled when permanent and `NewsRepository.setPermanent` returned `void`.
- [x] **Step 3: Implement GREEN** — `setPermanent` returns the updated `expires_at`; the detail controller exposes a single `togglePermanent()`; the app-bar control stays enabled and swaps its tooltip between make/restore; a successful delete pops the detail route, and the list is invalidated so it drops the Article.
- [x] **Step 4: Verify GREEN** — `cd apps/mobile && flutter analyze && flutter test test/features/news && flutter test`.
- [x] **Step 5: Verify integration and rules** — `flutter test -d <ios-simulator> integration_test/daily_news_flow_test.dart`, the guide rules check on changed files, and `git diff --check`.
- [x] **Step 6: Commit** — intentionally not committed because the shared worktree contains unrelated changes.

**完成紀錄：**

- 2026-09-17：使用者要求「將新聞設為永久的按鈕，如果再點一次，應取消恢復原本的時效」與「刪除新聞後應直接返回列表」。Red：新增 Router-backed 刪除測試與 permanence toggle 測試後，`cd apps/mobile && flutter test test/features/news/news_list_test.dart` 因 `setPermanent` 仍回 `void`、control 在永久時被停用（`onPressed: null`）而失敗。Green：`NewsItem.copyWith` 接受 `expiresAt`，`NewsRepository`／`NewsRemoteService.setPermanent` 改為回傳後端的 `expires_at`；`NewsDetailController` 以單一 `togglePermanent()` 依目前狀態送出 `permanent: !isPermanent`，並用回應更新 `isPermanent`／`expiresAt`；AppBar 永久按鈕不再停用，tooltip 在「設為永久」與「恢復原本時效」間切換。刪除成功後由詳情畫面 `context.pop()` 返回列表，並保留 `newsListControllerProvider(categoryId)` invalidation 讓列表重新載入；失敗時顯示本地化 `newsDeleteFailed` SnackBar。移除不再使用的 `NewsDetailUiState.deleted`、`newsDeleted`、`savedPermanently`，新增 `restoreExpiry`、`newsDeleteFailed` 並執行 `flutter gen-l10n`。fake API PATCH 改為依 request body 設定 `articleIsPermanent`。Verify：`cd apps/mobile && flutter analyze` 為 `No issues found!`；`flutter test test/features/news` 為 11 passed；`flutter test` 為 41 passed；`flutter test -d 0A818455-5B3B-44A3-A7CB-9AF6681297A6 integration_test/daily_news_flow_test.dart` 為 `1 passed`（journey 內含 make permanent → restore → make permanent → delete → 返回列表）；`uv run --with pyyaml python flutter-dev-guide/tools/check-rules.py --files ...` 與 `git diff --check` exit 0。未 commit，因 shared worktree 仍含不相關變更。

### Task 15: Remove YouTube as an ingestion source

**Depends on:** Task 10
**Parallel:** no — changes the shared ingest adapter composition and write boundary.
**Files:** Delete `backend/internal/ingestion/{youtube.go,youtube_test.go}`; create `backend/internal/ingestion/{query.go,blocklist.go,blocklist_test.go}`; modify `backend/internal/ingestion/{router.go,router_test.go,orchestrator.go}`, `backend/cmd/daily-news-job/main.go`, `backend/{.env.example,tests/test_container_contract.py}`, `infra/cloud-run/job.yaml`, `infra/docs/secret-inventory.md`, `docs/requirements/daily-news.md`, `docs/superpowers/specs/2026-09-16-web-source-adapters-design.md`, and this task record.

- [x] **Step 1: Write the failing tests** — prove a `youtube.com`/`youtu.be` Source Setting returns an explicit unsupported error and never reaches the web/search adapter, and that `youtube.com`/`youtu.be` candidates from any adapter are dropped at the write boundary while still counted as considered candidates.
- [x] **Step 2: Verify RED** — `cd backend && go test ./internal/ingestion -run 'YouTube|BlockedSource|Capped' -count=1` failed because the blocklist and unsupported error did not exist and YouTube still routed to its own adapter.
- [x] **Step 3: Implement GREEN** — delete the YouTube adapter and its `YOUTUBE_API_KEY` wiring; reject YouTube Source Settings in `HostRouter` with `ErrYouTubeUnsupported`; move `searchQuery` to `query.go`; and filter blocked platform hosts in the orchestrator's cap step before persistence.
- [x] **Step 4: Verify GREEN** — `cd backend && gofmt -w internal/ingestion internal/cmd 2>/dev/null; gofmt -w cmd/daily-news-job && go vet ./... && go test ./... && uv run pytest tests/test_container_contract.py -q`.
- [x] **Step 5: Update docs and deploy config** — requirements source table/secret list, web source adapter design, `.env.example`, Cloud Run Job manifest, and secret inventory no longer reference YouTube.
- [x] **Step 6: Commit** — intentionally not committed because the shared worktree contains unrelated changes.

**完成紀錄：**

- 2026-09-18：使用者回報「今天的新聞有些不 AI 相關，尤其 YouTube，很像自己的搜尋紀錄；若無法避免就把 YouTube 來源去掉」。根因：`youtube.com` Source Setting 走 `YouTubeAdapter`，只以 Category 名稱＋關鍵字＋特殊需求做全 YouTube `search.list`，忽略使用者填的頻道路徑，因此回傳排序雜亂；另外未指定網站與無 feed 網站的 SerpApi 全網搜尋也可能回傳 YouTube 連結。使用者決定移除 YouTube 來源、既有 YouTube 文章先不動。Red：新增 `blocklist_test.go` 與 router YouTube 測試後，`cd backend && go test ./internal/ingestion -run 'YouTube|BlockedSource|Capped' -count=1` 因 `isBlockedSourceURL`、`ErrYouTubeUnsupported` 尚未定義而編譯失敗。Green：刪除 `youtube.go`/`youtube_test.go` 與 `buildSourceAdapter` 的 YouTube 參數；`HostRouter` 對 `youtube.com`/`youtu.be` 回 `ErrYouTubeUnsupported` 且不呼叫 web adapter；新增 `blocklist.go` 的 `isBlockedSourceURL`，並在 `SourceSearchResult.capped()` 內於寫入前丟棄 blocked host 候選（`CandidateCount` 仍計入 considered candidates，符合既有語言過濾語意）；`searchQuery` 移至 `query.go` 供 SerpApi 使用。設定面同步移除 `backend/.env.example`、`infra/cloud-run/job.yaml`、`infra/docs/secret-inventory.md` 的 `YOUTUBE_API_KEY`，並更新 `backend/tests/test_container_contract.py` 的預期 secret 清單。Verify：`cd backend && gofmt -w internal/ingestion cmd/daily-news-job && go vet ./... && go test ./...` 全部 package `ok`（含 PostgreSQL integration）；`cd backend && uv run pytest tests/test_container_contract.py -q` 為 `5 passed`；`git diff --check` 通過。需求與設計文件已改為「YouTube 不支援並於寫入前過濾」。未 commit，因 shared worktree 仍含既有未提交變更。

## Plan Self-Review

| Spec requirement | Covered by |
|---|---|
| Flutter guide 的結構、規則、skills、checker、adoption | G1–G5 |
| 單一使用者 Firebase allowlist | B3, F1–F2, O1, O4 |
| Category Bottom Sheet、動態來源、首頁方塊 | B4, F4 |
| Article 關聯、去重、30 天、永久與刪除 | B2, N1, I1, I4 |
| 來源每組最多 10、平台限制與 citations | I1–I4 |
| 20 筆 cursor、tag filter、詳情 | N1, F5 |
| 新聞詳情以 WebView 顯示原文網頁 | Task 13 |
| 首頁更新時間、立即更新 dialog、單一 active Run | N2, F3, F5 |
| GitHub 08:00、WIF、Cloud Run Job、Secret Manager | O1–O4 |
| 可跨新對話執行的文件與 checkbox protocol | G1、全檔 task metadata |

Self-review completed on 2026-08-30:

- 所有需求至少對應一項 task。
- task id、route、domain type 與 `Depends on` 名稱在全檔一致。
- 本計劃沒有未決佔位詞；外部實際值由使用者在 GCP／GitHub 設定，文件只引用變數名稱。
