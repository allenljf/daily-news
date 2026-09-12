# Ingestion staging correction Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Restore real public-source ingestion and citation persistence for staging.

**Architecture:** A planner loads active source rows and binds them to a bounded RSS/Atom adapter; the Job applies migrations then runs planned work.

**Tech Stack:** Go, database/sql/pgx, encoding/xml, net/http, golang-migrate, PostgreSQL.

**Spec:** `docs/superpowers/specs/2026-09-12-ingestion-staging-correction-design.md`

## Task 1: R8 ingestion staging correction

- [ ] Add a failing PostgreSQL test for active source work and citation storage.
- [ ] Add migration `000003`, `WorkPlanner`, URL normalization, and a fakeable RSS/Atom adapter.
- [ ] Change the Job to apply `/app/migrations`, plan source work, then orchestrate it.
- [ ] Copy migrations into the runtime image; run `gofmt`, `go vet ./...`, `go test ./...`, container build, and `git diff --check`.
