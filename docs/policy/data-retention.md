---
title: "Data Retention Policy"
doc_type: policy
tags: [compliance, data, gdpr]
owners: [compliance-team]
confidence: high
source:
  origin: confluence
  url: https://wiki.internal/policy/data-retention
---

# Data Retention Policy

## 개요

본 정책은 OpenClaw 시스템에서 수집, 처리, 저장하는 모든 데이터의 보존 기간과 삭제 절차를 정의한다.

## 보존 기간

| 데이터 분류 | 보존 기간 | 삭제 방법 |
|------------|----------|----------|
| 사용자 개인정보 | 계정 삭제 후 30일 | 자동 삭제 (스케줄러) |
| 서비스 로그 | 90일 | 자동 로테이션 |
| 감사 로그 | 3년 | 수동 승인 후 삭제 |
| 백업 데이터 | 1년 | 자동 만료 |
| 분석 데이터 | 비식별화 후 영구 보존 | 해당 없음 |

## 삭제 절차

### 1단계: 삭제 요청 접수
- 사용자 요청 또는 보존 기간 만료 시 자동 트리거
- JIRA 티켓 생성 (`DATA-DELETE-*`)

### 2단계: 영향 분석
- 삭제 대상 데이터의 의존성 확인
- 관련 서비스 팀에 통보

### 3단계: 삭제 실행
- 프로덕션 DB에서 soft delete
- 30일 유예 기간 후 hard delete
- 백업에서도 제거 (다음 로테이션 시)

### 4단계: 검증
- 삭제 완료 확인 로그 기록
- 감사 보고서 생성

## GDPR 준수

- 데이터 주체의 삭제 요청(Right to Erasure)은 72시간 이내 처리
- 삭제 완료 시 데이터 주체에게 확인 이메일 발송
- 제3자 공유 데이터도 동시 삭제 요청

## 관련 문서
- 접근 제어 정책: `docs/policy/access-control`
- 보안 기준: `docs/policy/security-baseline`
