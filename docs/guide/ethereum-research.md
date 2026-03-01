---
id: ethereum-research
title: 이더리움 리서치 노트
source: openclaw
status: draft
created_at: 2026-03-01
owner: platform-team
version: 1
confidence: medium
---

# 이더리움(Ethereum) 리서치

## 핵심 개요
- 이더리움(Ethereum)은 EVM 기반의 퍼블릭 블록체인 플랫폼으로, 스마트 컨트랙트 실행 환경과 분산형 애플리케이션(디앱) 생태계를 지원한다.
- PoS 전환 이후 에너지 효율성이 개선되었고, EIP 기반 업그레이드로 확장성과 성능 개선이 지속 진행 중이다.

## 기술적 요약
- **합의 메커니즘**: 지분 증명(PoS, Casper)
- **블록 생산 구조**: proposer/builder 분리 경향
- **가스 모델**: 실행 비용을 가스 단위로 계산, 네트워크 부하 및 비용 변동이 중요
- **주요 업그레이드**: The Merge, Shanghai, Cancun(진행 중)

## 운영 관점 체크포인트
- 수수료/가스 폭등 구간 감시: `baseFee`, `priorityFee` 동향
- 노드 동기화 지연, re-org 빈도, 피어 연결 상태 모니터링
- 스테이킹/유효성검증 성능 임계치 점검
- 브리지/오라클 연동 시 지연 및 실패율 검토

## 추적해야 할 지표
- TPS(초당 처리량)
- 블록 시간/포크 발생 빈도
- 평균 가스비
- RPC 응답 지연
- 컨센서스 노드 CPU/메모리 사용률

## 참고
- 출처: Ethereum 공식 문서, 최근 네트워크 릴리즈 노트, 온체인 공개 지표
