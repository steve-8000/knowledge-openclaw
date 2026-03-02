# Avalanche 운영 리포트 - 생성 시각: 2026-03-02T13:42:09.697347+00:00 - RSS: ava-labs/avalanchego - 비교 기준(Base): v1.14.1 - 최신 커밋(Head): v1.14.2-rc.17
## 한줄 요약
벤치리스트 용량 초과 시 최악의 노드를 우선 배제하여 최상의 노드를 유지하는 로직 개선.
## 운영자 중요 체크
- 벤치리스트(Benchlist) 관리 로직이 우선순위 기반에서 최악의 성능 노드 기반으로 변경됨.
- `canBench` 함수가 `tryMakeRoom`으로 대체되어, 최소 실패율을 가진 노드를 배제하는 메커니즘 적용.
## 주요 변경점
- 벤치리스트 용량 제한(`maxPortion`) 도달 시 기존 노드를 제거하는 로직 개선.
- 최상의 노드를 유지하기 위해 최소 실패율을 가진 노드를 우선 배제하는 정책 도입.
## RPC/API 영향
- P2P 네트워크 품질 개선으로 인해 RPC 응답 속도 및 안정성에 긍정적 영향 가능성.
## 아카이브 노드 영향
- P2P 네트워크 품질 개선으로 인해 아카이브 노드의 데이터 수집 효율성 및 안정성에 긍정적 영향 가능성.
## 운영 액션 아이템
- 벤치리스트 관련 메트릭을 모니터링하여 새로운 로직이 예상대로 작동하는지 확인.
- 벤치리스트 용량 제한(`maxPortion`)에 도달했을 때의 노드 배제 현황을 확인.
## 마이그레이션 체크리스트
- 기존 벤치리스트 관련 설정이 새로운 로직과 호환되는지 확인.
- 벤치리스트 관련 로그를 확인하여 새로운 배제 정책이 적용되는지 확인.
## 위험/주의 사항
- 새로운 배제 정책이 예상치 못한 결과를 초래할 수 있으므로 주의 깊게 모니터링 필요.
- 벤치리스트 관리 로직 변경으로 인해 네트워크 품질에 영향을 줄 수 있는 우려.
## 근거(Evidence)
- 벤치리스트 용량 초과 시 최소 실패율을 가진 노드를 배제하는 로직 개선.
- `canBench` 함수가 `tryMakeRoom`으로 대체되어 최상의 노드를 유지하는 정책 도입.
## 운영 메모
- 벤치리스트 관리 로직이 최악의 노드를 우선 배제하도록 변경되어, 네트워크 품질 개선에 기여할 것으로 기대됨.
- 새로운 로직이 예상대로 작동하는지 확인하기 위해 모니터링이 필요함.
## 원문 커밋 내용
When the benchlist is at capacity (maxPortion reached), replace canBench with tryMakeRoom which evicts the least-failing benched nodes to make room for higher-probability newcomers. This ensures the worst offenders are always benched rather than first-come-first-served. Co-Authored-By: Claude Opus 4.6 noreply@anthropic.com
