---
title: "Database Failover Procedure"
doc_type: runbook
tags: [runbook, database, failover]
owners: [sre-team, dba]
confidence: high
source:
  origin: confluence
  url: https://wiki.internal/runbook/db-failover
---

# Database Failover Procedure

## 개요

PostgreSQL Primary 서버 장애 시 Standby로 전환하는 절차.
예상 소요 시간: 5~15분.

## 사전 조건

- [ ] Standby 서버가 스트리밍 복제 상태인지 확인
- [ ] PgBouncer 또는 커넥션 풀러 접근 가능
- [ ] 관련 팀 알림 채널 접근 가능 (Slack #incident-db)

## 절차

### 1단계: 장애 확인

```bash
# Primary 상태 확인
pg_isready -h primary.db.internal -p 5432

# Standby 복제 지연 확인
psql -h standby.db.internal -c "SELECT now() - pg_last_xact_replay_timestamp() AS replication_lag;"
```

- 복제 지연이 5초 이상이면 데이터 손실 가능성 있음 → 팀 리드 승인 필요

### 2단계: Primary 격리

```bash
# Primary 네트워크 격리 (split-brain 방지)
ssh primary.db.internal "sudo systemctl stop postgresql"

# 확인
pg_isready -h primary.db.internal -p 5432  # 실패해야 함
```

### 3단계: Standby 승격

```bash
ssh standby.db.internal "sudo pg_ctlcluster 16 main promote"

# 승격 확인
psql -h standby.db.internal -c "SELECT pg_is_in_recovery();"
# 결과: false (= Primary로 전환 완료)
```

### 4단계: 커넥션 전환

```bash
# PgBouncer 설정 변경
ssh pgbouncer.internal "sed -i 's/primary.db.internal/standby.db.internal/' /etc/pgbouncer/pgbouncer.ini"
ssh pgbouncer.internal "sudo systemctl reload pgbouncer"
```

### 5단계: 검증

```bash
# 애플리케이션 헬스 체크
curl -s https://kidb.clab.one/healthz
curl -s https://kidb.clab.one/api/v1/search?q=test -H 'X-Tenant-ID: 00000000-0000-0000-0000-000000000001'
```

## 롤백

Standby를 다시 Standby로 되돌리는 것은 불가. 기존 Primary를 새 Standby로 재구성해야 함.

## 관련 문서
- 백업 복원 절차: `docs/runbook/restore-backup`
- DB 장애 포스트모템: `docs/incident/db-outage-2024`
