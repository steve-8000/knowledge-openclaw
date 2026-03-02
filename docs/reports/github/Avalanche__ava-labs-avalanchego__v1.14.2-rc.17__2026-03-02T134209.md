# Avalanche 운영 리포트 - 생성 시각: 2026-03-02T13:42:09.161598+00:00 - RSS: ava-labs/avalanchego - 비교 기준(Base): v1.14.1 - 최신 커밋(Head): v1.14.2-rc.17

## 한줄 요약
When the benchlist is at capacity (maxPortion reached), replace canBench with tryMakeRoom which evicts the least-failing benched nodes to make room for higher-probability newcomers. This ensures the worst offenders are always benched rather than first-come-first-served. Co-Authored-By: Claude Opus 4.6 noreply@anthropic.com

## 운영자 중요 체크
- 변경 영향도 높음 영역(운영/빌드/동기화) 우선 검토
- 설정 파일/플래그 변경 유무 추적

## 주요 변경점
- 릴리스 노트에서 핵심 변경사항을 단계적으로 확인 필요
- 문서/CI/런타임 경로 변경 가능성 점검

## RPC/API 영향
- 공개 API 호환성 및 응답값 변화 여부 재검증 필요

## 아카이브 노드 영향
- 상태 저장, 압축, 인덱스 경로 변경이 있는지 배포 전 검증 필요

## 운영 액션 아이템
- 사전 스테이징에서 블록동기화/메트릭 점검
- 롤백 기준치 정의

## 마이그레이션 체크리스트
- 노드 바이너리/런타임 교체 체크리스트 재검토
- 파서/입력 스키마 변경 반영 여부 확인

## 위험/주의 사항
- 데이터 모델 변경 시 저장소 재동기화 위험
- 외부 의존성 버전 상승으로 발생하는 실행 타임 충돌 가능성

## 근거(Evidence)
When the benchlist is at capacity (maxPortion reached), replace canBench with tryMakeRoom which evicts the least-failing benched nodes to make room for higher-probability newcomers. This ensures the worst offenders are always benched rather than first-come-first-served. Co-Authored-By: Claude Opus 4.6 noreply@anthropic.com

## 운영 메모
- 상세 검증 전에는 본문 기반 추정치이므로 샌드박스 검증 권장

## 원문 커밋 내용
commit history fetch failed: fatal: fetch --all does not take a repository argument
