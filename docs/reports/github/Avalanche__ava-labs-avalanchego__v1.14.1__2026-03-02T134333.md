# Avalanche 운영 리포트 - 생성 시각: 2026-03-02T13:43:33.724600+00:00 - RSS: ava-labs/avalanchego - 비교 기준(Base): v1.14.2-rc.17 - 최신 커밋(Head): v1.14.1

## 한줄 요약
v1.14.1은 v1.14.0과 완전히 호환되며, 시스템 트래커 설정을 개선하고 EVM 관련 API/패키지를 정리했다. Firewood 기능을 프라이빗 서브넷에서 사용할 수 있게 되었으나, 파일 경로 변경으로 인한 재동기화가 필요하다.

## 운영자 중요 체크
*   **호환성:** v1.14.0과 완전히 호환되므로 기존 설정을 유지할 수 있다.
*   **플러그인:** 플러그인 버전 44는 v1.14.0과 호환된다.
*   **Firewood:** 프라이빗 서브넷에서 프라이빗 서브넷을 실행할 때 프라이빗 서브넷의 프라이빗 서브넷(Private Subnet of Private Subnet)을 실행할 수 있게 되었다.

## 주요 변경점
*   **시스템 트래커:** 디스크 사용량 경고/필요 공간 설정을 백분율 기반의 새로운 옵션으로 변경했다.
*   **EVM 개선:**
    *   `avax.version` API 제거 및 `customethclient` 패키지를 표준 `ethclient`로 대체했다.
    *   `blockHook` 확장 기능을 제거했다.
*   **Firewood:** 프라이빗 서브넷에서 프라이빗 서브넷을 실행할 때 프라이빗 서브넷의 프라이빗 서브넷을 실행할 수 있게 되었다.

## RPC/API 영향
*   `avax.version` API가 제거되었으므로, 이를 사용하는 애플리케이션은 업데이트가 필요하다.
*   `customethclient` 패키지가 제거되었으므로, 이를 의존하는 RPC 클라이언트는 표준 `ethclient`로 마이그레이션해야 한다.

## 아카이브 노드 영향
*   Firewood 기능을 사용 중인 노드는 파일 경로가 변경되었으므로, 재동기화가 필요하다.

## 운영 액션 아이템
*   **설정 업데이트:** 새로운 디스크 사용량 설정 옵션(`--system-tracker-disk-required-available-space-percentage`, `--system-tracker-disk-warning-available-space-percentage`)을 적용한다.
*   **재동기화:** Firewood를 사용 중인 노드는 재동기화를 수행한다.
*   **애플리케이션 업데이트:** `avax.version` API를 사용하는 애플리케이션을 업데이트한다.

## 마이그레이션 체크리스트
*   [ ] v1.14.0 호환성 확인
*   [ ] 플러그인 버전 44 호환성 확인
*   [ ] 새로운 디스크 설정 옵션 적용
*   [ ] `avax.version` API 사용 여부 확인 및 제거
*   [ ] `customethclient` 패키지 의존성 제거
*   [ ] Firewood 재동기화 수행

## 위험/주의 사항
*   **재동기화:** Firewood를 사용 중인 노드는 재동기화가 필요하므로, 네트워크 부하를 고려하여 계획된 시간에 실행해야 한다.
*   **API 제거:** `avax.version` API 제거로 인해, 이를 사용하는 애플리케이션이 중단될 수 있다.

## 근거(Evidence)
*   v1.14.1은 v1.14.0과 완전히 호환된다.
*   플러그인 버전 44는 v1.14.0과 호환된다.
*   `avax.version` API가 제거되었다.
*   `customethclient` 패키지가 제거되었다.
*   Firewood를 사용 중인 노드는 재동기화가 필요하다.

## 운영 메모
*   Coreth와 Subnet-EVM 저장소가 메인 저장소에 병합되어 변경 로그가 생략되었다.
*   Firewood 기능이 프라이빗 서브넷에서 프라이빗 서브넷을 실행할 수 있게 되었다.

## 원문 커밋 내용
This version is backwards compatible to v1.14.0. It is optional, but encouraged.

The plugin version is unchanged at 44 and is compatible with version v1.14.0.

Config
Added:
--system-tracker-disk-required-available-space-percentage
--system-tracker-disk-warning-available-space-percentage
Deprecated:
--system-tracker-disk-required-available-space
--system-tracker-disk-warning-threshold-available-space

EVM
Removed avax.version API
Removed customethclient package in favor of ethclient package and temporary type registrations (WithTempRegisteredLibEVMExtras)
Removed blockHook extension in ethclient package.
Enabled Firewood to run with pruning disabled.
This change modified the filepath of Firewood. Any nodes using Firewood will need to resync.

What's Changed

The changelog is omitted, as the Coreth and Subnet-EVM repositories were grafted into the repository.

Full Changelog: v1.14.0...v1.14.1
