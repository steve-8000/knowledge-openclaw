# OpenClaw × ki-db 지식 활용 가이드

이 문서는 **OpenClaw 에이전트가 ki-db를 활용하여 지식을 축적, 검색, 연결하고 더 나은 판단을 내리는 방법**을 정의한다.
ki-db는 OpenClaw의 **장기 기억 저장소**이다. 모든 작업 산출물, 결정 근거, 학습 내용은 ki-db에 문서로 저장되고, 이후 작업에서 검색을 통해 재활용된다.

- **서비스 URL**: `https://kidb.clab.one`
- **대시보드**: `https://kidb.clab.one` (브라우저에서 직접 접속)
- **GitHub**: `https://github.com/steve-8000/knowledge-openclaw`

---

## 1. 핵심 원칙

### 1.1 지식 흐름

```
[작업 수행] → [문서 작성] → [ki-db 인덱싱] → [검색/RAG] → [더 나은 작업 수행]
                                                  ↑                      │
                                                  └──────────────────────┘
                                                       피드백 루프
```

### 1.2 3가지 규칙

1. **쓸 때는 풍부하게** — 문서를 저장할 때 메타데이터(tags, owners, doc_type, confidence)를 빠짐없이 채운다
2. **읽을 때는 정확하게** — 검색 결과를 맹신하지 않고 confidence와 status를 확인한 뒤 사용한다
3. **관계를 만들어라** — 문서 간 `references`, `links_to`, `supersedes` 관계를 명시적으로 생성한다

---

## 2. 문서 작성 규칙

### 2.1 Markdown 구조 (인덱싱에 최적화)

ki-db의 Chunker는 **heading 기반으로 텍스트를 분할**한다. 따라서:

```markdown
# 문서 제목 (필수)

## 개요
한 문단으로 이 문서가 무엇인지 설명.

## 상세 내용
### 하위 섹션 A
구체적 내용. 각 하위 섹션은 하나의 chunk로 분할된다.

### 하위 섹션 B
다른 주제의 내용.

## 관련 문서
- [API v2 Specification](stable_key: docs/spec/api-v2)
- [Security Baseline](stable_key: docs/policy/security-baseline)

## 결정 이력
| 날짜 | 결정 | 근거 |
|------|------|------|
| 2024-01-15 | PostgreSQL 채택 | 벡터+BM25 통합 가능 |
```

**핵심 규칙:**
- `#` 제목은 반드시 1개만 (문서 타이틀)
- `##`로 주요 섹션 구분, `###`로 하위 주제 구분
- **각 섹션은 독립적으로 이해 가능하게** 작성 — chunk 단위로 검색되므로 맥락이 섹션 내에서 완결되어야 함
- 관련 문서 링크 시 `stable_key`를 명시하면 Graph Builder가 자동으로 edge를 생성한다
- 코드 블록, 표, 목록은 정상 파싱됨

### 2.2 stable_key 네이밍 규칙

`stable_key`는 문서의 **영구 식별자**이다. 파일 경로/URL처럼 계층적으로 부여한다.

```
{카테고리}/{서브카테고리}/{슬러그}
```

| 카테고리 | 예시 | 설명 |
|----------|------|------|
| `docs/policy/` | `docs/policy/security-baseline` | 정책 문서 |
| `docs/spec/` | `docs/spec/api-v2` | 기술 스펙 |
| `docs/adr/` | `docs/adr/001-use-postgres` | 아키텍처 결정 기록 |
| `docs/guide/` | `docs/guide/onboarding` | 가이드/how-to |
| `docs/runbook/` | `docs/runbook/db-failover` | 운영 절차서 |
| `docs/incident/` | `docs/incident/2024-01-db-outage` | 장애 보고서 |
| `docs/report/` | `docs/report/q1-2024-review` | 정기 리포트 |
| `work/task/` | `work/task/implement-auth-v2` | 작업 기록 |
| `work/decision/` | `work/decision/chose-nats-over-kafka` | 작업 중 결정 |
| `learn/` | `learn/pgvector-hnsw-tuning` | 학습 메모 |

**규칙:**
- 소문자, 하이픈(`-`) 구분
- 버전이 바뀌어도 `stable_key`는 변하지 않음 (내용이 바뀌면 같은 key로 재 ingest → 버전이 자동 증가)
- 삭제된 문서는 key를 재사용하지 않음

### 2.3 doc_type 선택 기준

| doc_type | 언제 사용 | confidence 기본값 |
|----------|----------|------------------|
| `policy` | 조직 규칙, 보안 기준, 프로세스 정책 | `high` |
| `spec` | API/시스템/서비스 기술 명세 | `high` |
| `adr` | 아키텍처 결정 기록 (Architecture Decision Record) | `high` |
| `guide` | 설정/사용법/운영 가이드 (how-to) | `med` |
| `runbook` | 장애 대응/운영 절차서 | `high` |
| `incident` | 장애 보고서, 포스트모템 | `med` → 검증 후 `high` |
| `report` | 정기 리포트, 분석 결과 | `med` |
| `glossary` | 용어 사전, 정의 모음 | `high` |
| `other` | 위에 해당하지 않는 기타 | `low` |

### 2.4 필수 메타데이터 체크리스트

문서를 인제스트할 때 **반드시** 아래를 채운다:

- [ ] `stable_key` — 2.2 규칙에 따라
- [ ] `title` — 명확하고 검색 가능한 제목
- [ ] `doc_type` — 2.3 기준에 따라
- [ ] `tags` — 최소 2개, 최대 8개의 관련 키워드
- [ ] `owners` — 담당 팀/사람 (최소 1개)
- [ ] `confidence` — 문서 신뢰도 (`high`/`med`/`low`)
- [ ] `source` — 원본 출처 (`{"origin": "github", "url": "..."}`)

---

### 2.5 문서 저장소 (`docs/` 디렉토리)

ki-db의 문서 SoT(Source of Truth)는 **Git 저장소의 `docs/` 디렉토리**이다.

```
knowledge-openclaw/
└── docs/                    ← 문서 SoT
    ├── policy/              정책 문서
    ├── spec/                기술 스펙
    ├── adr/                 아키텍처 결정 기록
    ├── guide/               가이드/how-to
    ├── runbook/             운영 절차서
    ├── incident/            장애 보고서
    ├── report/              정기 리포트
    └── learn/               학습 메모
```

**동기화 흐름:**
```
docs/*.md 파일 생성/수정 → git push → 서버 git pull
    → doc-sync 컨테이너가 60초마다 스캔
    → 변경된 파일만 ki-db Ingest API로 자동 인덱싱
```

**파일 포맷:** YAML frontmatter + Markdown 본문

```markdown
---
title: "문서 제목"
doc_type: adr
tags: [tag1, tag2]
owners: [team-name]
confidence: high
source:
  origin: github
  url: https://...
---

# 문서 제목

## 내용
본문 작성...
```

**규칙:**
- `stable_key`는 파일 경로에서 자동 생성: `docs/adr/001-use-postgres.md` → `adr/001-use-postgres`
- frontmatter의 `title`, `doc_type`은 필수. 나머지는 권장
- 파일을 수정하면 다음 sync 주기(60초)에 자동으로 새 버전이 인덱싱됨
- 파일을 삭제해도 ki-db에서 자동 삭제되지 않음 (수동으로 deprecated/archived 처리)

**멀티서버 동기화:**
각 서버에서 같은 Git 저장소를 `git pull`하면 동일한 `docs/`를 공유. 각 서버의 doc-sync가 독립적으로 로컬 ki-db에 인덱싱.

## 3. ki-db API 활용법

> **Base URL**: `https://kidb.clab.one`
> **모든 API 호출에 `X-Tenant-ID` 헤더 필수**: `X-Tenant-ID: 00000000-0000-0000-0000-000000000001`

### 3.1 문서 저장 (Ingest)

새 지식을 ki-db에 저장할 때:

```bash
curl -X POST https://kidb.clab.one/api/v1/documents \
  -H 'Content-Type: application/json' \
  -H 'X-Tenant-ID: 00000000-0000-0000-0000-000000000001' \
  -d '{
    "stable_key": "docs/adr/005-use-nats",
    "title": "ADR-005: NATS JetStream 채택",
    "doc_type": "adr",
    "status": "published",
    "confidence": "high",
    "owners": ["platform-team"],
    "tags": ["adr", "nats", "event-bus", "architecture"],
    "source": {"origin": "github", "url": "https://github.com/openclaw/adr/005"},
    "raw_text": "# ADR-005: NATS JetStream 채택\n\n## 배경\n...\n\n## 결정\n...\n\n## 근거\n..."
  }'
```

**동작 흐름:**
1. Ingest API가 문서를 DB에 저장 (기존 stable_key면 업데이트, 새 버전 생성)
2. Outbox에 `DocumentUpserted` + `VersionCreated` 이벤트 발행
3. Parser Worker → 텍스트 정규화, 링크 추출
4. Chunker Worker → heading 기반 청크 분할 + BM25 인덱스
5. Embedder Worker → 벡터 임베딩 생성 + HNSW 인덱스
6. Graph Builder → 링크/참조 기반 edge 생성

**소요 시간:** 저장 즉시 → 인덱싱 완료까지 수초~수십 초

### 3.2 지식 검색 (Search)

작업 전에 관련 지식을 검색할 때:

```bash
curl -H 'X-Tenant-ID: 00000000-0000-0000-0000-000000000001' \
  'https://kidb.clab.one/api/v1/search?q=database+failover+procedure&limit=10'
```

**응답 구조:**
```json
{
  "query": "database failover procedure",
  "results": [
    {
      "chunk_id": "uuid",
      "doc_id": "uuid",
      "title": "Database Failover Procedure",
      "chunk_text": "1단계: Primary 상태 확인...",
      "bm25_rank": 1,
      "rrf_score": 0.95
    }
  ],
  "citations": [
    {
      "doc_id": "uuid",
      "version_id": "uuid",
      "title": "Database Failover Procedure",
      "heading_path": "## 절차 > ### 1단계",
      "chunk_id": "uuid",
      "relevance": "both"
    }
  ]
}
```

**검색 결과 사용 규칙:**
- `rrf_score` 0.7 이상: 높은 신뢰도로 활용 가능
- `relevance: "both"` (키워드+의미 매칭): 가장 관련성 높음
- 결과의 원본 문서 `status`가 `deprecated`이면 → 최신 문서를 찾아야 함 (supersedes 관계 확인)
- 결과의 `confidence`가 `low`이면 → 교차 확인 필요

### 3.3 관계 탐색 (Knowledge Graph)

특정 문서와 관련된 지식 네트워크를 탐색할 때:

```bash
curl -H 'X-Tenant-ID: 00000000-0000-0000-0000-000000000001' \
  'https://kidb.clab.one/api/v1/graph/ego?doc_id={doc_id}&depth=2'
```

**응답에서 파악할 수 있는 것:**
- `references` 엣지: 이 문서가 근거로 삼는 다른 문서들
- `links_to` 엣지: 관련/연관 문서들
- `supersedes` 엣지: 이 문서를 대체하는 최신 문서 (있으면 최신 것을 사용)
- `contradicts` 엣지: 이 문서와 충돌하는 내용 (주의 필요)

### 3.4 문서 상태 관리

```bash
# 검토 완료 → 게시
curl -X PATCH https://kidb.clab.one/api/v1/documents/{doc_id}/status \
  -H 'Content-Type: application/json' \
  -H 'X-Tenant-ID: 00000000-0000-0000-0000-000000000001' \
  -d '{"status": "published"}'
```

| 상태 전이 | 의미 |
|-----------|------|
| `inbox` → `published` | 검토 완료, 신뢰할 수 있는 지식으로 등록 |
| `published` → `deprecated` | 최신 문서가 이것을 대체함 (supersedes 엣지 확인) |
| `published` → `archived` | 더 이상 유효하지 않은 지식 |
| `deprecated` → `archived` | 폐기 처리 |

### 3.5 피드백 제출

검색 결과가 유용했는지 기록하여 검색 품질을 개선:

```bash
curl -X POST https://kidb.clab.one/api/v1/feedback \
  -H 'Content-Type: application/json' \
  -H 'X-Tenant-ID: 00000000-0000-0000-0000-000000000001' \
  -d '{
    "query": "database failover",
    "chunk_id": "uuid-of-used-chunk",
    "helpful": true,
    "note": "정확한 절차를 제공했음"
  }'
```

---

## 4. 작업 프로토콜 — 언제 ki-db를 사용하는가

### 4.1 작업 시작 시 (검색 → 맥락 확보)

**모든 비-사소한 작업 시작 전에** 관련 지식을 검색한다:

```
[사용자 요청 수신]
    ↓
[ki-db 검색: 관련 정책/스펙/ADR/과거 작업 기록]
    ↓
[검색 결과를 맥락으로 활용하여 작업 수행]
    ↓
[작업 완료]
```

**검색해야 하는 상황:**
- 아키텍처 결정을 해야 할 때 → 기존 ADR 검색
- 특정 시스템을 수정할 때 → 관련 spec/guide 검색
- 장애 대응 시 → 관련 incident/runbook 검색
- 보안/권한 관련 작업 시 → 관련 policy 검색

**검색 쿼리 작성 팁:**
- 핵심 키워드 2-3개로 간결하게: `"auth token rotation procedure"`
- 너무 긴 쿼리는 BM25 성능 저하: `"how do we handle authentication token rotation when secrets are compromised"` (X)
- doc_type 필터 활용: `?q=failover&doc_type=runbook`

### 4.2 작업 완료 시 (저장 → 지식 축적)

**지식 가치가 있는 산출물은 ki-db에 저장**한다:

| 산출물 유형 | doc_type | 저장 여부 |
|------------|----------|----------|
| 아키텍처 결정 | `adr` | **반드시 저장** |
| API/시스템 스펙 작성/변경 | `spec` | **반드시 저장** |
| 새 정책/규칙 수립 | `policy` | **반드시 저장** |
| 운영 절차서 작성 | `runbook` | **반드시 저장** |
| 가이드/튜토리얼 | `guide` | **반드시 저장** |
| 장애 대응/포스트모템 | `incident` | **반드시 저장** |
| 단순 버그 수정 | - | 저장 불필요 |
| 코드 포매팅/리팩터링 | - | 저장 불필요 |
| 일회성 질문 답변 | - | 저장 불필요 |

### 4.3 기존 문서 업데이트 시

기존 문서의 내용이 바뀌었을 때 — **같은 `stable_key`로 다시 ingest**:

```bash
# 기존 stable_key로 POST → 자동으로 새 버전 생성
curl -X POST https://kidb.clab.one/api/v1/documents \
  -H 'Content-Type: application/json' \
  -H 'X-Tenant-ID: 00000000-0000-0000-0000-000000000001' \
  -d '{
    "stable_key": "docs/spec/api-v2",
    "title": "API v2 Specification (Updated)",
    "doc_type": "spec",
    "raw_text": "# API v2 Specification\n\n## 변경 사항\n..."
  }'
```

기존 문서를 **완전히 대체하는 새 문서**를 만들 때:
1. 새 문서를 ingest (새 `stable_key`)
2. 이전 문서를 `deprecated`로 변경
3. (향후) `supersedes` 엣지 생성

---

## 5. 관계(Edge) 생성 가이드

### 5.1 관계 유형과 사용 기준

| 관계 | 의미 | 사용 기준 |
|------|------|----------|
| `references` | A가 B를 근거/참고로 인용 | 문서 내에서 다른 문서를 명시적으로 언급할 때 |
| `links_to` | A와 B가 관련 있음 | 같은 주제를 다루거나 함께 읽어야 할 때 |
| `supersedes` | A가 B를 대체 (A가 최신) | 정책/스펙이 개정되어 이전 버전을 대체할 때 |
| `duplicates` | A와 B가 동일한 내용 | 중복 문서를 발견했을 때 (정리 대상) |
| `contradicts` | A와 B가 모순 | 동일 주제에 대해 상반된 내용이 있을 때 (검토 필요) |

### 5.2 자동 생성 vs 수동 생성

- **자동 생성**: Graph Builder가 문서 내 링크(`stable_key` 참조)를 파싱하여 `links_to`/`references` 엣지를 자동 생성
- **수동 생성**: `supersedes`, `duplicates`, `contradicts`는 사람/에이전트가 판단하여 명시적으로 생성해야 함

### 5.3 문서 내 참조 표기법

Graph Builder가 자동으로 edge를 생성하려면, 문서 본문에서 다른 문서를 참조할 때 `stable_key`를 포함시킨다:

```markdown
## 관련 문서
- 보안 기준: `docs/policy/security-baseline`
- 인증 시스템 설계: `docs/spec/auth-system`

## 근거
이 결정은 [ADR-001: PostgreSQL 채택](docs/adr/001-use-postgres)에 기반한다.
```

---

## 6. 검색 품질 향상 — 피드백 루프

### 6.1 피드백 기록

작업 중 ki-db 검색 결과를 활용했다면:

- **유용했을 때**: `helpful: true` + 어떤 chunk를 사용했는지 기록
- **유용하지 않았을 때**: `helpful: false` + `note`에 왜 부적절했는지 기록
- **결과가 없었을 때**: 해당 주제의 문서를 새로 작성하여 인제스트

### 6.2 검색 실패 대응

검색 결과가 없거나 부적절할 때:

1. **태그 부족**: 기존 문서에 태그가 부족한 경우 → 문서 재인제스트로 태그 보강
2. **문서 부재**: 해당 지식 자체가 없는 경우 → 새 문서 작성
3. **오래된 문서**: deprecated 문서만 나오는 경우 → 최신 문서 작성 필요

---

## 7. 스킬/룰 연동

### 7.1 스킬에서 ki-db 활용 패턴

스킬(skill)이 ki-db와 연동되는 표준 패턴:

```
[스킬 실행 시작]
    │
    ├─ (1) 검색: 해당 도메인의 기존 지식 조회
    │       GET /api/v1/search?q={domain keywords}&doc_type={relevant type}
    │
    ├─ (2) 그래프 탐색: 관련 문서 네트워크 확인
    │       GET /api/v1/graph/ego?doc_id={found_doc}&depth=1
    │
    ├─ (3) 맥락 기반 실행: 검색 결과를 참고하여 스킬 로직 수행
    │
    ├─ (4) 결과 저장: 스킬 산출물 중 재사용 가능한 지식을 인제스트
    │       POST /api/v1/documents (if applicable)
    │
    └─ (5) 피드백: 사용한 검색 결과에 대한 피드백 기록
            POST /api/v1/feedback
```

### 7.2 스킬별 ki-db 활용 예시

| 스킬 종류 | 검색 시점 | 저장 대상 |
|-----------|----------|----------|
| 코드 리뷰 | 관련 코딩 표준/정책 검색 | - |
| 아키텍처 설계 | 기존 ADR/스펙 검색 | 새 ADR 저장 |
| 장애 대응 | 관련 runbook/과거 incident 검색 | 장애 보고서 저장 |
| 문서 작성 | 관련 기존 문서/용어 검색 | 작성된 문서 저장 |
| 온보딩 | 가이드/정책/스펙 검색 | - |

### 7.3 룰(Rule) 정의 시 ki-db 연동

OpenClaw의 룰에서 ki-db를 활용하는 표준 패턴:

```yaml
# 예시: 아키텍처 결정 시 반드시 기존 ADR 확인
rule:
  trigger: "architecture_decision_needed"
  actions:
    - search:
        query: "{decision_topic}"
        doc_type: "adr"
        min_results: 0
    - if_results:
        found: "기존 ADR을 참고하여 결정"
        not_found: "새 ADR을 작성하여 기록"
    - after_decision:
        ingest:
          doc_type: "adr"
          stable_key: "docs/adr/{sequential_number}-{slug}"
```

---

## 8. 프로덕션 인프라 (운영자용)

### 8.1 서버 정보

| 항목 | 값 |
|------|-----|
| **서버** | `219.255.103.189` (Ubuntu 24.04) |
| **도메인** | `https://kidb.clab.one` |
| **SSL** | Let's Encrypt (자동갱신, certbot) |
| **프로젝트 경로** | `/opt/kidb/` |
| **GitHub** | `https://github.com/steve-8000/knowledge-openclaw` |

### 8.2 서비스 구성 및 포트 매핑

| 서비스 | 컨테이너 | 호스트 포트 | 역할 |
|--------|----------|------------|------|
| **Ingest API** | kidb-ingest-api | `127.0.0.1:8180` | 문서 저장, 상태 관리, 운영 정보 |
| **Query API** | kidb-query-api | `127.0.0.1:8181` | 검색, 피드백, 그래프 탐색 |
| **Dashboard** | kidb-dashboard | `127.0.0.1:3100` | 시각화 (Search/Graph/Curation/Quality/Ops) |
| **PostgreSQL** | kidb-postgres | `127.0.0.1:5532` | ParadeDB (pgvector + pg_search) |
| **NATS** | kidb-nats | `127.0.0.1:4322` | JetStream 이벤트 버스 |
| **Workers** | kidb-worker-* (5개) | 내부 전용 | parser, chunker, embedder, graph, quality |
| **Outbox Relay** | kidb-outbox-relay | 내부 전용 | DB outbox → NATS 발행 |
| **Doc Sync** | kidb-doc-sync | 내부 전용 | docs/ 파일 → Ingest API 자동 동기화 (60초 주기) |

> **모든 서비스는 `127.0.0.1`에만 바인딩** — 외부 접근은 nginx 리버스 프록시를 통해서만 가능

### 8.3 nginx 라우팅 규칙

nginx (`/etc/nginx/sites-available/kidb.clab.one`)가 경로 기반으로 요청을 분배:

| 경로 패턴 | 백엔드 |
|-----------|--------|
| `/api/v1/search`, `/api/v1/graph/`, `/api/v1/feedback` | Query API (8181) |
| `/api/v1/documents`, `/api/v1/ops/` | Ingest API (8180) |
| `/` (그 외 모든 경로) | Dashboard (3100) |

### 8.4 테넌트 헤더

모든 API 호출에 `X-Tenant-ID` 헤더가 필수:
```
X-Tenant-ID: 00000000-0000-0000-0000-000000000001
```

### 8.5 인덱싱 파이프라인

```
docs/*.md → [doc-sync] → Ingest API → [Outbox] → NATS JetStream
                                                      │
                                            ┌─────────┼─────────┬─────────────┐
                                            ▼         ▼         ▼             ▼
                                         Parser → Chunker → Embedder → Graph Builder
```

각 워커는 멱등(idempotent)하게 동작하며, 실패 시 자동 재시도된다.

### 8.6 서버 운영 명령어

```bash
# SSH 접속
ssh root@219.255.103.189

# 프로젝트 디렉토리
cd /opt/kidb

# 컨테이너 상태 확인
docker compose -f infra/compose.yaml -f infra/compose.override.yaml --env-file .env ps

# 전체 재시작
docker compose -f infra/compose.yaml -f infra/compose.override.yaml --env-file .env restart

# 특정 서비스만 재시작
docker compose -f infra/compose.yaml -f infra/compose.override.yaml --env-file .env restart query-api

# 로그 확인 (실시간)
docker compose -f infra/compose.yaml -f infra/compose.override.yaml --env-file .env logs -f --tail 50

# 특정 서비스 로그
docker compose -f infra/compose.yaml -f infra/compose.override.yaml --env-file .env logs ingest-api --tail 100

# 전체 중지 후 재시작 (빌드 포함)
docker compose -f infra/compose.yaml -f infra/compose.override.yaml --env-file .env down
docker compose -f infra/compose.yaml -f infra/compose.override.yaml --env-file .env up -d --build

# GitHub에서 최신 코드 가져오기 + 재배포
cd /opt/kidb && git pull origin main
docker compose -f infra/compose.yaml -f infra/compose.override.yaml --env-file .env up -d --build

# nginx 설정 변경 후
nginx -t && systemctl reload nginx

# SSL 인증서 수동 갱신 (보통 자동갱신됨)
certbot renew
```

### 8.7 장애 대응

| 증상 | 확인 방법 | 조치 |
|------|----------|------|
| 사이트 접속 불가 | `curl -I https://kidb.clab.one` | nginx 상태 확인: `systemctl status nginx` |
| API 502 Bad Gateway | 컨테이너 로그 확인 | `docker compose ... logs {service} --tail 50` |
| 검색 결과 없음 | 워커 상태 확인 | `docker compose ... ps`로 워커 alive 확인 |
| DB 연결 실패 | postgres 로그 확인 | `docker compose ... logs postgres --tail 50` |
| SSL 만료 | `certbot certificates` | `certbot renew && systemctl reload nginx` |

### 8.8 주요 파일 위치

```
/opt/kidb/                              # 프로젝트 루트
/opt/kidb/.env                          # 환경변수 (패스워드, 포트 등)
/opt/kidb/infra/compose.yaml            # Docker Compose 메인 설정
/opt/kidb/infra/compose.override.yaml   # 프로덕션 포트 오버라이드
/opt/kidb/db/migrations/                # DB 마이그레이션 파일 (000001~000013)
/etc/nginx/sites-available/kidb.clab.one  # nginx 설정
/etc/nginx/sites-enabled/kidb.clab.one    # nginx 심볼릭 링크
/etc/letsencrypt/live/kidb.clab.one/      # SSL 인증서
```

### 8.9 기존 서비스와의 공존

이 서버에는 **criv** 서비스가 함께 운영 중:

| 서비스 | 포트 | 도메인 |
|--------|------|--------|
| criv-web | 3000 | clab.one |
| criv-api | 8080 | clab.one/api |
| criv-postgres | 5432 | - |
| **kidb 서비스들** | **3100, 8180, 8181, 5532, 4322** | **kidb.clab.one** |

> ki-db 포트는 모두 criv와 충돌하지 않도록 리매핑됨

### 8.10 아키텍처 SoT

- 아키텍처 명세: `ar.md`
- API 계약: `contracts/openapi/openapi.yaml`
- 이벤트 계약: `contracts/events/README.md`

---

## 9. 빠른 참조 — API 요약

> Base URL: `https://kidb.clab.one`
> 필수 헤더: `X-Tenant-ID: 00000000-0000-0000-0000-000000000001`

### 저장

```
POST /api/v1/documents
  Body: { stable_key, title, doc_type, raw_text, tags, owners, confidence, source }
  → 문서 생성/업데이트 + 인덱싱 파이프라인 자동 시작
```

### 검색

```
GET /api/v1/search?q={query}&limit={n}&doc_type={type}&status={status}
  → 하이브리드 검색 (BM25 + 벡터 ANN) + citations
```

### 그래프

```
GET /api/v1/graph/ego?doc_id={id}&depth={1-3}
  → 문서 중심 N-hop 관계 네트워크
```

### 상태 변경

```
PATCH /api/v1/documents/{docID}/status
  Body: { "status": "published" | "deprecated" | "archived" | "inbox" }
```

### 피드백

```
POST /api/v1/feedback
  Body: { query, chunk_id, helpful, note }
```

### 문서 조회

```
GET /api/v1/documents                    → 전체 목록
GET /api/v1/documents/{docID}            → 상세 + 버전 이력
```

### 운영 정보

```
GET /api/v1/ops/status                   → 파이프라인 상태
GET /api/v1/ops/jobs                     → 워커 작업 현황
GET /api/v1/ops/quality                  → 문서 품질 지표
```
