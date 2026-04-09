# Research: Project, Infra, Actions, and Makefile Setup for Reusable Workflows

## Objective
- Map the current automation surface (GitHub Actions + Makefile + infra tooling).
- Identify what is already reusable.
- Propose a practical plan to evolve into cleaner reusable workflows, without implementing changes yet.

## Repository Automation Inventory

### 1) GitHub Actions workflows
- `.github/workflows/ci.yaml`
  - Orchestrator workflow for `push` (`main`, `develop`, `feat/Sprint1`) and `pull_request` (`main`, `develop`).
  - Calls 3 local reusable workflows via `uses: ./.github/workflows/*.yaml`:
    - `build.yaml`
    - `tests.yaml`
    - `docker-build.yaml`
  - Dependency graph:
    - `build` → `unit-and-e2e-tests` → `docker-build`

- `.github/workflows/build.yaml` (already reusable)
  - Trigger: `workflow_call`.
  - Responsibilities:
    - Setup Go.
    - `go vet ./...`.
    - `go build -o cmd/api/usersvc ./cmd/api`.
    - Upload build artifact (`api-binary`) when not running in `act`.
  - Output:
    - `artifact-name`.

- `.github/workflows/tests.yaml` (already reusable)
  - Trigger: `workflow_call`.
  - Input: optional `build-artifact`.
  - Responsibilities:
    - Optional OpenAPI validation with `swagger-cli`.
    - `go test ./...` with race and coverage.
    - Upload `coverage.out` artifact.
  - Contains `act`-specific compatibility logic.

- `.github/workflows/docker-build.yaml` (already reusable)
  - Trigger: `workflow_call`.
  - Input: optional `build-artifact`.
  - Responsibilities:
    - Setup Buildx/QEMU.
    - Optional artifact download into `cmd/api`.
    - Build image with `docker/build-push-action` (no push).

- `.github/workflows/verify-copilot-contributions.yaml`
  - PR-only governance workflow for branch `main`.
  - Detects Copilot-authored commits and checks if `COPILOT_INSTRUCTIONS.md` was updated.
  - Not a CI build/test workflow, but a policy/compliance helper.

### 2) Makefile setup (local orchestration and test infra)
- Main categories:
  - Local infra lifecycle: `localstack-*`, `tflocal-*`, `infra-up`, `infra-down`, `infra-test`, `infra-debug`.
  - Cognito local lifecycle: `cognito-local-*`.
  - App containers: `docker-compose-up`, `docker-compose-down`, `swagger-only`.
  - App quality: `build`, `test`, `test-workspace`, `test-http`.
  - Production infra: `infra-prod-*`.
- Important test behavior:
  - `make test` starts DB if needed (`test-db-up`) and stops only if it started it (`test-db-down` sentinel-based).
  - This is useful behavior to preserve if CI starts consuming Make targets directly.

### 3) Infra and runtime tooling
- Terraform in `infra/` for AWS resources.
- Local emulation strategy:
  - LocalStack for S3/DynamoDB/IAM/VPC/EC2.
  - `cognito-local` as substitute for Cognito in local tests.
- Docker Compose services:
  - `db`, `cognito-local`, `api`, `swagger`.
- Dockerfile supports build arg `USE_PREBUILT=1` to reuse a previously compiled binary in CI pipelines.

## Current Reusability Status

## What is already good
- CI is already modularized into reusable local workflows (`workflow_call`).
- Artifact passing exists between build/tests/docker stages.
- Concurrency groups are configured.
- Clear separation between CI orchestration (`ci.yaml`) and task workflows.

## Gaps / duplication hotspots
- Repeated `act` compatibility steps across `build.yaml` and `tests.yaml`.
- Go setup + environment bootstrapping duplicated.
- Optional artifact handling pattern repeated across workflows.
- Workflow naming/ownership is functional but can be standardized further for future expansion (lint/security/release/deploy).

## Reusable Workflow Opportunities (Research Recommendations)

### A) Introduce shared callable “foundation” workflows
Create small reusable units in `.github/workflows/reusable-*` (or equivalent naming):
- `reusable-go-setup.yaml`
  - Handles `actions/checkout`, `actions/setup-go`, go env diagnostics, optional `act` overrides.
- `reusable-openapi-validate.yaml`
  - Owns Node install + swagger-cli validation with consistent skip logic.
- `reusable-artifact-download.yaml`
  - Encapsulates optional artifact fetching behavior.

Why:
- Removes repeated boilerplate.
- Keeps task workflows focused on business intent (build/test/docker).

### B) Split domain workflows by intent
Keep `ci.yaml` as orchestrator, but standardize called workflows into intent-driven blocks:
- `reusable-build-go.yaml`
- `reusable-test-go.yaml`
- `reusable-build-docker.yaml`
- (future) `reusable-security-scan.yaml` (CodeQL, dependency checks).
- (future) `reusable-release.yaml` (tag/release packaging).

Why:
- Better long-term maintainability and discoverability.
- Easier to reuse from future workflows (e.g., nightly, release, hotfix).

### C) Align Makefile and Actions contracts
Define a deliberate contract:
- Local developer flow remains `make build`, `make test`, `make infra-*`.
- CI can either:
  1) keep direct commands (`go vet`, `go test`, `docker build`), or
  2) call Make targets to unify logic.

Recommended direction:
- Keep CI commands explicit for core build/test speed and clarity.
- Use Make targets only for integration environments that mirror local infra (`infra-test` subsets), where Makefile already encodes complex orchestration.

### D) Add path-aware execution strategy (future optimization)
- Use path filters at orchestration level to skip unchanged domains:
  - Go code changes → build/test.
  - Dockerfile/compose changes → docker-build.
  - `openapi/**` changes → openapi validation + tests.
  - `infra/**` changes → terraform checks (future).

Why:
- Faster feedback and lower CI cost.

## Proposed Target Architecture (No Implementation Yet)

- `ci.yaml` (orchestrator only)
  - Calls reusable workflows with clear inputs/outputs.
- Reusable task workflows
  - Build, test, docker, security, release.
- Shared helper workflows
  - Go setup, OpenAPI validation, artifact utilities.
- Optional governance workflows
  - Keep `verify-copilot-contributions.yaml` independent from CI quality gates.

## Suggested Migration Plan (Phased)

### Phase 1 — Consolidate and standardize
- Normalize naming conventions (`reusable-*` pattern).
- Extract repeated setup blocks (Go + `act`) into one callable unit.
- Keep behavior identical (no semantic changes).

### Phase 2 — Expand reusable coverage
- Add reusable lint/security workflows.
- Standardize artifacts and outputs contracts across all reusable workflows.

### Phase 3 — Optimize execution
- Add path-based conditional execution.
- Introduce optional matrix strategies (Go versions, OS) if needed.

### Phase 4 — Release/deploy reuse
- Add reusable release and image publish workflows (with environment approvals and secrets strategy).

## Risks and Design Considerations
- Over-modularization can make debugging harder if split too aggressively.
- `act` compatibility should remain centralized to avoid drift.
- Artifact coupling should remain minimal; only pass what downstream jobs truly need.
- Keep governance workflows separate from core CI gates to avoid accidental delivery blocking.

## Practical Next Steps (when implementation is requested)
1. Define workflow naming and folder convention.
2. Extract common `act` + Go setup logic first.
3. Refactor `build.yaml` and `tests.yaml` to consume shared setup.
4. Keep `ci.yaml` behavior unchanged while migrating internals.
5. Add a short “Workflow Topology” section to `README.md` after migration.

## Source files reviewed
- `/home/runner/work/Const-Software-25-02/Const-Software-25-02/.github/workflows/ci.yaml`
- `/home/runner/work/Const-Software-25-02/Const-Software-25-02/.github/workflows/build.yaml`
- `/home/runner/work/Const-Software-25-02/Const-Software-25-02/.github/workflows/tests.yaml`
- `/home/runner/work/Const-Software-25-02/Const-Software-25-02/.github/workflows/docker-build.yaml`
- `/home/runner/work/Const-Software-25-02/Const-Software-25-02/.github/workflows/verify-copilot-contributions.yaml`
- `/home/runner/work/Const-Software-25-02/Const-Software-25-02/Makefile`
- `/home/runner/work/Const-Software-25-02/Const-Software-25-02/README.md`
- `/home/runner/work/Const-Software-25-02/Const-Software-25-02/infra/README.md`
- `/home/runner/work/Const-Software-25-02/Const-Software-25-02/docker-compose.yaml`
- `/home/runner/work/Const-Software-25-02/Const-Software-25-02/docker-compose.cognito-local.yaml`
- `/home/runner/work/Const-Software-25-02/Const-Software-25-02/Dockerfile`
- `/home/runner/work/Const-Software-25-02/Const-Software-25-02/.env.example`
