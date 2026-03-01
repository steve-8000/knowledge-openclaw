---
title: "New Engineer Onboarding"
doc_type: guide
tags: [guide, onboarding, setup]
owners: [engineering]
confidence: high
source:
  origin: confluence
  url: https://wiki.internal/guide/onboarding
---

# New Engineer Onboarding

## 개요

OpenClaw 팀에 합류한 신규 엔지니어를 위한 온보딩 가이드.
목표: 첫 주 안에 개발 환경 설정 완료 + 첫 PR 제출.

## Day 1: 환경 설정

### 필수 도구
- Git, Docker Desktop, Go 1.23+, Node.js 20+
- IDE: VSCode 또는 JetBrains (팀 라이선스 제공)
- Slack: #engineering, #incident, #deploy 채널 가입

### 저장소 클론
```bash
git clone https://github.com/steve-8000/knowledge-openclaw.git
cd knowledge-openclaw
```

### 로컬 환경 실행
```bash
cp .env.example .env
docker compose -f infra/compose.yaml up -d
# 대시보드: http://localhost:3000
# Ingest API: http://localhost:8080
# Query API: http://localhost:8081
```

## Day 2-3: 아키텍처 이해

### 필독 문서 (ki-db에서 검색)
1. `docs/adr/001-use-postgres` — 왜 PostgreSQL인지
2. `docs/adr/event-driven` — 이벤트 드리븐 아키텍처
3. `docs/spec/search-engine` — 검색 엔진 설계
4. `docs/policy/code-review` — 코드 리뷰 기준

### 시스템 구성 이해
```
ar.md (아키텍처 SoT) → 전체 설계 명세
GUIDE.md → ki-db 활용 가이드 (이 시스템의 사용법)
```

## Day 4-5: 첫 PR

### 추천 첫 이슈
- `good-first-issue` 라벨이 붙은 이슈
- 문서 오타 수정
- 테스트 추가

### PR 제출 체크리스트
- [ ] 로컬 빌드 통과 (`make build`)
- [ ] 테스트 통과 (`make test`)
- [ ] 코드 리뷰 기준 확인 (`docs/policy/code-review`)
- [ ] 관련 문서 업데이트 (필요 시)

## 유용한 링크
- 프로덕션 대시보드: https://kidb.clab.one
- GitHub: https://github.com/steve-8000/knowledge-openclaw
- 장애 대응 정책: `docs/policy/incident-response`
