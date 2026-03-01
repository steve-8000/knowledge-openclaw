---
title: "OpenClaw × ki-db 지식 활용 가이드"
doc_type: guide
tags: [openclaw, kidb, knowledge, search]
owners: [platform-team]
confidence: high
source:
  origin: github
  url: "https://github.com/steve-8000/knowledge-openclaw"
status: published
---
# OpenClaw × ki-db 지식 활용 가이드

## 1. 핵심 목표와 연결 방식

OpenClaw는 작업의 입력/판단/결정 근거를 **ki-db**에 문서화하고,
ki-db 검색 결과를 참조해 판단 정확도와 재현성을 높이는 방향으로 동작한다.

- Base URL: `https://kidb.clab.one`
- 필수 헤더: `X-Tenant-ID: 00000000-0000-0000-0000-000000000001`
- Ingest: `POST /api/v1/documents`
- Search: `GET /api/v1/search?q={query}&limit={n}&doc_type={type}&status={status}`
- Graph: `GET /api/v1/graph/ego?doc_id={doc_id}&depth={1-3}`

## 2. 작업 전 검색 + 작업 후 저장 프로토콜

1) 작업 시작 전 관련 문서를 검색한다.
2) 문서 작성/ADR/런북 생성 시 문서를 `docs/` 아래에 저장하고, 동일 `stable_key`로 재인제스트한다.
3) 사용한 검색 결과는 필요 시 Feedback으로 기록한다.

### 2.1 검색 기준

- `stable_key`는 소문자/슬러그 기반으로 고정 유지.
- `status`: `inbox | published | deprecated | archived`
- `confidence`: `high | med | low`

## 3. 연결 포인트(로컬)

### 3.1 문서 SoT
`docs/` 디렉터리가 정식 SoT.
예: `docs/guide/..., docs/spec/..., docs/adr/...`

### 3.2 자동 동기화 명령

```bash
# 전체 문서 인제스트 (dry-run)
KIDB_API_BASE=https://kidb.clab.one KIDB_TENANT_ID=00000000-0000-0000-0000-000000000001 python3 scripts/kidb-ingest-docs.py docs --dryrun

# 실제 인제스트
KIDB_API_BASE=https://kidb.clab.one KIDB_TENANT_ID=00000000-0000-0000-0000-000000000001 python3 scripts/kidb-ingest-docs.py docs

# wrapper
scripts/kidb-client.sh ingest /Users/steve/.openclaw/workspace/docs
```

## 4. 운영 API 예시

### 문서 저장

```bash
curl -X POST https://kidb.clab.one/api/v1/documents \
  -H 'Content-Type: application/json' \
  -H 'X-Tenant-ID: 00000000-0000-0000-0000-000000000001' \
  -d '{"stable_key":"docs/guide/openclaw-ki-db-connection-guide","title":"OpenClaw × ki-db 지식 활용 가이드","doc_type":"guide","status":"published","confidence":"high","owners":["platform-team"],"tags":["openclaw","kidb","knowledge"],"source":{"origin":"github","url":"https://github.com/steve-8000/knowledge-openclaw"},"raw_text":"..."}'
```

### 검색

```bash
curl -H 'X-Tenant-ID: 00000000-0000-0000-0000-000000000001' \
  'https://kidb.clab.one/api/v1/search?q=knowledge%20integration&limit=10&doc_type=guide'
```

## 5. 문서 간 관계 규칙

- `references`, `links_to`, `supersedes`, `duplicates`, `contradicts` 규칙을 문서 본문/메타로 구성한다.
- deprecated된 문서는 최신성을 위해 supersedes 체인 확인 후 대체 판독.

## 6. 오퍼레이션 노트

- 운영 계열은 기본적으로 내부 바인딩(127.0.0.1)에 동작하고, 외부는 nginx를 통해 접속한다.
- 인덱싱 동기화 지연은 보통 수초~수십초 내.

