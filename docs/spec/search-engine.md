---
title: "Search Engine Architecture"
doc_type: spec
tags: [search, bm25, vector]
owners: [search-team]
confidence: med
source:
  origin: notion
  url: https://notion.so/search-arch
---

# Search Engine Architecture

## 개요

ki-db의 하이브리드 검색 엔진은 **BM25 키워드 검색**과 **벡터 ANN 검색**을 결합하여, 정확한 키워드 매칭과 의미론적 유사도를 동시에 제공한다.

## 아키텍처

```
                    사용자 쿼리
                        │
                   Query API
                   /         \
              BM25 검색      ANN 검색
             (pg_search)    (pgvector)
                   \         /
                RRF Score 병합
                        │
                   정렬된 결과
```

## BM25 검색 (키워드)

ParadeDB의 `pg_search` 확장을 사용하여 PostgreSQL 내에서 BM25 인덱스를 유지한다.

- **인덱싱**: `chunks` 테이블의 `body` 컬럼에 BM25 인덱스 생성
- **쿼리**: 토큰화된 쿼리로 TF-IDF 기반 랭킹
- **장점**: 정확한 키워드 매칭, 구문 검색 지원

## 벡터 ANN 검색 (의미론적)

pgvector의 HNSW 인덱스를 사용하여 1024차원 임베딩 벡터의 근사 최근접 이웃 검색을 수행한다.

- **임베딩 모델**: OpenAI text-embedding-3-large (1024 dims)
- **인덱스**: HNSW (ef_construction=128, m=16)
- **거리 함수**: Cosine similarity
- **장점**: 동의어, 유사 개념 검색 가능

## RRF (Reciprocal Rank Fusion) 병합

두 검색 결과를 RRF 공식으로 병합:

```
RRF_score(d) = Σ 1 / (k + rank_i(d))
```

- `k = 60` (표준값)
- `rank_i(d)`: i번째 검색 엔진에서 문서 d의 순위
- 두 엔진 모두에서 상위에 있을수록 높은 점수

## 성능 목표

| 지표 | 목표 | 현재 |
|------|------|------|
| P50 레이턴시 | < 50ms | 측정 중 |
| P99 레이턴시 | < 200ms | 측정 중 |
| 인덱스 규모 | 100만 청크 | 데모 데이터 |

## 관련 문서
- PostgreSQL 채택 결정: `docs/adr/001-use-postgres`
- API v2 스펙: `docs/spec/api-v2`
