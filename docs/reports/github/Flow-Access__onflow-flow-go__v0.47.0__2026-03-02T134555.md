# Flow(Access) 운영 리포트 - 생성 시각: 2026-03-02T13:45:55.455917+00:00 - RSS: onflow/flow-go - 비교 기준(Base): v0.47.0-experimental-indexer-ct.9 - 최신 커밋(Head): v0.47.0
## 한줄 요약
FVM 및 FlowEVM 성능 최적화, Cadence 업데이트, 그리고 데이터 가용성(Data Availability) 관련 인덱스 추가 및 JobQueue 리팩토링이 주요 변경 사항입니다.

## 운영자 중요 체크
- **FVM/FlowEVM**: 연산 비용 절감을 위한 여러 ABI 인코딩/디코딩 최적화 기능이 추가되었습니다.
- **Cadence**: 버전 1.9.8, 1.9.9, 1.9.10으로 업데이트되었습니다.
- **Data Availability**: 계정 트랜잭션 인덱스가 추가되어 쿼리 성능이 개선될 수 있습니다.

## 주요 변경점
- **Zero-Downtime HCU**: Zero-Downtime HCU 개선 사항이 추가되었습니다.
- **FVM**: 사용되지 않는 에러 메시지가 제거되었습니다.
- **FlowEVM**: Solidity 튜플 배열의 ABI 인코딩/디코딩 구현, EOA 제한 기능 제거, 연산 비용 절감을 위한 새로운 함수 추가 및 여러 최적화 작업 수행.
- **Cadence**: 버전 1.9.8, 1.9.9, 1.9.10으로 업데이트.
- **Data Availability**: 계정 트랜잭션 인덱스 추가, JobQueue 리팩토링 (진행 소비자 초기화 요구사항 강화).

## RPC/API 영향
- **Data Availability**: 계정 트랜잭션 인덱스 추가로 인해 계정 기반 트랜잭션 조회 관련 API의 응답 속도가 향상될 수 있습니다.
- **FlowEVM**: 연산 비용 절감 기능이 추가되어 EVM 트랜잭션 처리 비용이 다소 낮아질 수 있습니다.

## 아카이브 노드 영향
- **Data Availability**: `jobqueue` 리팩토링으로 인해 진행 상태를 추적하는 소비자의 초기화가 필수적이게 되었습니다. 이는 아카이빙 프로세스의 안정성에 영향을 줄 수 있으므로 주의가 필요합니다.

## 운영 액션 아이템
- **JobQueue 확인**: 리팩토링으로 인해 JobQueue 관련 로그를 모니터링하여 진행 소비자가 정상적으로 초기화되는지 확인해야 합니다.
- **FVM/FlowEVM 테스트**: 새로운 EVM 함수 및 최적화 기능이 실제 트랜잭션 환경에서 정상적으로 작동하는지 테스트합니다.

## 마이그레이션 체크리스트
- [ ] **Cadence 업데이트**: Cadence 버전 1.9.8, 1.9.9, 1.9.10으로 업데이트된 코드를 확인하고 배포합니다.
- [ ] **Data Availability 인덱스**: 계정 트랜잭션 인덱스가 정상적으로 작동하는지 확인합니다.
- [ ] **JobQueue 리팩토링**: 리팩토링된 JobQueue 설정이 올바르게 적용되었는지 확인합니다.

## 위험/주의 사항
- **JobQueue 리팩토링**: JobQueue가 초기화되지 않으면 데이터 처리에 문제가 발생할 수 있습니다.
- **FVM/FlowEVM**: EOA 제한 기능이 제거되었으므로, 기존과 다른 트랜잭션 행동이 나타날 수 있습니다.

## 근거(Evidence)
- **FVM/FlowEVM**: `Implement ABI encoding/decoding for arrays of Solidity tuples`, `Optimize EVMDecodeABI`, `Remove EOA restriction functionality`, `Add new EVM functions that can be used to reduce computation cost`.
- **Cadence**: `Update to Cadence v1.9.8`, `Update to Cadence v1.9.9`, `Update to Cadence v1.9.10`.
- **Data Availability**: `Add index for account transactions`, `Refactor jobqueue to require initialized progress consumer`.

## 운영 메모
- v0.47.0 릴리스는 FVM 및 FlowEVM의 성능 최적화에 집중되어 있습니다.
- Data Availability 부분의 JobQueue 리팩토링은 시스템 안정성에 중요한 영향을 미칠 수 있으므로, 롤아웃 시 주의 깊게 모니터링해야 합니다.

## 원문 커밋 내용
- Zero-Downtime HCU: Add more improvements by @zhangchiqing in #8350
- POC Ledger Service: by @zhangchiqing in #8309
- FVM: Remove unused error by @janezpodhostnik in #8393
- FlowEVM: Implement ABI encoding/decoding for arrays of Solidity tuples by @m-Peter in #8371
- FlowEVM: Optimize EVMDecodeABI by removing an ArrayValue iteration by @fxamacker in #8397
- FlowEVM: Optimize EVMEncodeABI by removing an ArrayValue iteration by @fxamacker in #8398
- FlowEVM: Optimize EVMEncodeABI by creating Go reflect types at startup and reusing them by @fxamacker in #8399
- FlowEVM: Optimize EVM dryCall by removing RLP encoding/decoding by @fxamacker in #8400
- FlowEVM: Remove EOA restriction functionality from EVM by @m-Peter in #8408
- FlowEVM: Add new EVM functions that can be used to reduce computation cost of transactions by @fxamacker in #8418
- FlowEVM: Optimize and reduce computation cost of four EVM functions by @fxamacker in #8434
- FlowEVM: Add strict hex-prefix check when parsing EVM addresses from String by @m-Peter in #8437
- FlowEVM: Add proper meter and gas limit checks for EVM dry operations by @m-Peter in #8416
- Cadence: Update to Cadence v1.9.8 by @turbolent in #8395
- Cadence: Update to Cadence v1.9.9 by @turbolent in #8412
- Cadence: Update to Cadence v1.9.10 by @turbolent in #8461
- Data Availability: Add index for account transactions by @peterargue in #8381
- Data Availability: Refactor jobqueue to require initialized progress consumer - take 2 by @peterargue in #8404
- Data Availability: Fix ParseAddress() by only removing prefix
