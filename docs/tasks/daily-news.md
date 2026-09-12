# Daily News Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` (recommended) or `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 建立可獨立複製的 Flutter 開發準則庫，並交付一個由 Flutter、Go、PostgreSQL、Cloud Run 與 GitHub Actions 組成的單一使用者每日新聞 App。

**Architecture:** Flutter 採 View/ViewModel + Repository/Service，Riverpod 只負責 composition 與 UI state，Dio 僅存在 remote service。Go 以 `net/http`、明確 composition root、`database/sql` + pgx adapter 與 Firebase Admin Go SDK 提供 versioned HTTP interface；Cloud Run service 提供 App API，Cloud Run Job 執行 ingestion，GitHub Actions 以 OIDC/WIF 觸發 scheduled Job。Python task 的完成紀錄是歷史證據；Go replacement phase 才是目前 backend 實作路徑。

**Tech Stack:** Flutter Material 3, `flutter_riverpod`, Dio, `go_router`, `json_serializable`, Firebase Google Sign-In, Go standard library (`net/http`, `context`, `encoding/json`, `errors`, `testing/httptest`), `database/sql` + pgx adapter, Firebase Admin Go SDK, golang-migrate, PostgreSQL, Cloud Run, Secret Manager, GitHub Actions OIDC/WIF, Gemini Google Search grounding, GitHub REST API, YouTube Data API.

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

**Depends on:** O3, F6, R7
**Parallel:** no — 真實外部資源驗收。  
**Files:** Create `docs/operations/staging-smoke-test.md`; modify `docs/tasks/daily-news.md` 完成紀錄。

- [ ] **Red:** 在 smoke test 文件先列出失敗判準：WIF 驗證失敗、Firebase allowlist 外帳號可讀資料、Run 沒有結束狀態、Article 沒有 citation、手動 Run 平行重複執行。
- [ ] **Green:** 由使用者設定 O1 所列 resource 與秘密後，部署 staging；以 allowlisted account 驗證登入、Category、manual Run、latest status、Article list、永久、刪除；以第二次並行 POST 驗證 active Run 合併。
- [ ] **Verify:** 檢查 Cloud Run／workflow log、`ingestion_runs`／`ingestion_attempts` 統計與 mobile integration results；記錄 pass/fail 和 run id，但不記錄 token／URL query secret。
- [ ] **Commit:** `git add docs/operations docs/tasks/daily-news.md && git commit -m "docs: record staging smoke test"`。

**Done when:** staging 以 Go image 與真正的 WIF、Cloud Run、Cloud SQL、Firebase 與至少一個公開來源完成端到端工作，且全部失敗判準均未發生。

**完成紀錄：**

- 2026-09-12：Deployment correction（O4 尚未完成）：在已確認 Service Ready、100% traffic、`ingress: all` 與 `allUsers` Invoker binding 後，公開 `/healthz` 仍由 Google edge 回傳 HTML 404，未到達 Go handler。Red：`cd backend && uv run pytest tests/test_container_contract.py -q` 因 manifest 未宣告 public invocation 而如預期失敗（1 failed, 3 passed）。Green：Service manifest 明確加入 `run.googleapis.com/invoker-iam-disabled: 'true'`，使 Cloud Run admission 與 API 內 Firebase Bearer-token allowlist 分工明確，並由 contract test 固定此部署不變量。Verify：`cd backend && uv run pytest tests/test_container_contract.py -q && go vet ./... && go test ./...` 通過（container contract 4 passed；所有 Go package passing）；`actionlint .github/workflows/deploy.yml`、manifest assertion 與 `git diff --check` 均通過。待使用者 push 並完成 GitHub Actions 部署後，以公開 `/healthz` 和 allowlisted Firebase smoke test 驗證，才可完成 O4。

- 2026-09-12：CI correction（O4 尚未完成）：GitHub Actions run `34668466412` 顯示多個 PostgreSQL integration package 在 migration 前取得 `read: connection reset by peer`；各測試僅以 container 內 Unix socket 的 `pg_isready` 判定 ready，尚未確認 runner 經 Docker published TCP port 的 pgx 連線。Green：各 integration helper 在 migration 前以相同 `platform.OpenDB`／`PingContext` 對實際 PostgreSQL URL 重試最多 30 秒；成功後立即關閉 readiness pool。Verify：`cd backend && gofmt -w internal/platform/migrations_test.go internal/category/store_integration_test.go internal/news/store_integration_test.go internal/ingestion/runs_integration_test.go && GOFLAGS=-p=1 go test -count=1 ./...` 通過，所有 Go package passing；待下一次 GitHub Actions CI 實際驗證。

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
