# Arbitrum 운영 리포트 - 생성 시각: 2026-03-03T00:52:28.169152+00:00 - RSS: OffchainLabs/nitro - 비교 기준(Base): v3.9.5 - 최신 커밋(Head): v3.9.6
## 한줄 요약
Docker 이미지가 업데이트되었으며, ArbOS40 체인에서 실행 중인 Validator는 별도의 설정 없이 `v3.9.6-91bf578-validator` 이미지를 사용해야 합니다. 또한 Consensus v51.1이 추가되었습니다.

## 운영자 중요 체크
*   **Validator 이미지 변경:** ArbOS40 체인을 운영 중인 Validator는 `offchainlabs/nitro-node:v3.9.6-91bf578-validator` 이미지를 사용해야 합니다. 이 이미지는 기본 엔트리포인트로 `/usr/local/bin/split-val-entry.sh`를 사용하여 v3.9.5와 v3.7.6 Validator 워커를 자동으로 실행합니다.
*   **Docker 이미지 버전:** 공식 Docker 이미지는 `offchainlabs/nitro-node:v3.9.6-91bf578`입니다.
*   **기본 플래그 확인:** 이미지 엔트리포인트 기본 플래그(`--validation.wasm.allowed-wasm-module-roots`)를 재정의할 경우 이를 반영해야 합니다.

## 주요 변경점
*   **Sequencer 수정:** 예상되는 Surplus(초과분) 문제를 수정했습니다 (#4038).
*   **Consensus 업데이트:** Consensus v51.1을 Docker 및 v3.9.x 라인에 지원을 추가했습니다 (#4370, #4420).
*   **내부 최적화:** Bold 모듈에 `assume-valid` 플래그를 추가했습니다 (#4342).

## RPC/API 영향
*   Consensus v51.1 지원으로 인해 프로토콜 호환성이 업데이트되었습니다.
*   Sequencer 수정으로 인해 예상되는 Surplus 처리 로직이 변경되었습니다.

## 아카이브 노드 영향
*   특별한 아카이브 노드 전용 변경 사항은 없으나, Consensus 업데이트로 인해 블록체인 데이터 처리 방식이 변경될 수 있습니다.

## 운영 액션 아이템
*   **Validator 재시작:** ArbOS40 체인을 운영 중인 Validator는 기존 이미지를 `v3.9.6-91bf578-validator`로 교체하고 재시작해야 합니다.
*   **Docker 업데이트:** 기존 노드를 Docker로 운영 중인 경우 `offchainlabs/nitro-node:v3.9.6-91bf578`로 이미지를 업데이트해야 합니다.
*   **플래그 확인:** 커스텀 엔트리포인트를 사용 중이라면 기본 플래그가 포함되어 있는지 확인해야 합니다.

## 마이그레이션 체크리스트
*   [ ] Docker 이미지 태그를 `v3.9.6-91bf578` 또는 `v3.9.6-91bf578-validator`로 변경
*   [ ] Validator 사용 시 `split-val-entry.sh` 기본 엔트리포인트 사용 여부 확인
*   [ ] 기존 플래그와 새로운 기본 플래그 충돌 여부 확인

## 위험/주의 사항
*   **ArbOS40 Validator:** v3.9.6 Validator 이미지를 사용하지 않으면 Validator가 작동하지 않을 수 있습니다.
*   **Split Validator:** 분할 Validator를 사용 중인 경우 v3.9.5와 v3.7.6 워커를 모두 실행해야 합니다.

## 근거(Evidence)
*   Docker 이미지: `offchainlabs/nitro-node:v3.9.6-91bf578`
*   Validator 이미지: `offchainlabs/nitro-node:v3.9.6-91bf578-validator`
*   사용자 변경 사항: Sequencer Surplus 수정 (#4038), Consensus v51.1 추가 (#4370, #4420)
*   내부 변경 사항: Bold `assume-valid` 플래그 추가 (#4342)

## 운영 메모
*   이전 버전인 v3.9.5와 호환되는 변경 사항이 포함되어 있습니다.
*   L1 이더리움 노드 연결이 필수입니다.

## 원문 커밋 내용
This release is available as a Docker Image on Docker Hub at offchainlabs/nitro-node:v3.9.6-91bf578
This Docker image specifies default flags in its entrypoint which should be replicated if you're overriding the entrypoint: /usr/local/bin/nitro --validation.wasm.allowed-wasm-module-roots /home/user/nitro-legacy/machines,/home/user/target/machines

Important for any chains still on ArbOS40:

If you're running a validator without a split validation server (this will be true of most validators), you should instead use the image offchainlabs/nitro-node:v3.9.6-91bf578-validator which has the extra script /usr/local/bin/split-val-entry.sh as the default entrypoint (no need to override the default entrypoint). This will run both v3.9.5 and v3.7.6 validator workers for you.
If you are using a split validator, you do not want to use -validator image, you need to run a validator worker on v3.7.6 as well as a worker for v3.9.5
User-facing changes
Fix expected surplus in sequencer (#4038): #4371
Add support for consensus v51.1 to Docker: #4370
Enable consensus v51.1 for v3.9.x: #4420
Internal highlights
Add assume-valid flags to bold: #4342

Full Changelog: v3.9.5...v3.9.6
