---
title: "ADR-001: Use PostgreSQL"
doc_type: adr
tags: [adr, database, postgres]
owners: [architecture]
confidence: high
source:
  origin: github
  url: https://github.com/openclaw/adr/001
---

# ADR-001: PostgreSQL을 주 데이터베이스로 채택

## 상태
Accepted (2024-01-15)

## 맥락

OpenClaw 시스템은 다음을 지원하는 데이터베이스가 필요하다:
- 구조화된 문서 메타데이터 저장
- 전문 검색 (Full-Text Search)
- 벡터 유사도 검색 (Embedding 기반 ANN)
- ACID 트랜잭션
- Row-Level Security (멀티테넌트)

## 검토한 대안

### Option A: PostgreSQL + pgvector + pg_search (ParadeDB)
- 장점: 단일 DB에서 관계형 + BM25 + 벡터 검색 통합
- 장점: RLS로 멀티테넌트 격리 네이티브 지원
- 단점: BM25 인덱스가 별도 확장 필요 (ParadeDB)

### Option B: PostgreSQL + Elasticsearch
- 장점: 검색 성능이 검증됨
- 단점: 두 시스템 간 동기화 복잡성, 운영 비용 2배

### Option C: MongoDB + Pinecone
- 장점: 각 영역에서 최적화된 서비스
- 단점: 3개 시스템 운영, 트랜잭션 일관성 보장 어려움

## 결정

**Option A: PostgreSQL + ParadeDB 확장 채택**

단일 데이터베이스에서 모든 요구사항을 충족하며, 운영 복잡성을 최소화한다.

## 근거
- CQRS 패턴에서 읽기/쓰기를 분리하더라도, 인덱스 저장소는 하나로 유지하는 것이 운영상 유리
- ParadeDB의 pg_search는 AGPL 라이선스이나, 서버 사이드 사용이므로 수용 가능
- pgvector의 HNSW 인덱스는 100만 벡터 규모까지 단일 노드에서 충분한 성능

## 관련 문서
- 검색 엔진 아키텍처: `docs/spec/search-engine`
- 이벤트 드리븐 아키텍처: `docs/adr/event-driven`
