# Inference Engine Fixes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fix the current inference flow so the UI only offers configured models, empty-model requests use the configured default model, failed model lookup does not leave stale tasks, and summaries are parsed/displayed as structured content.

**Architecture:** Keep the current Gin + GORM structure intact and make focused fixes in the API, engine, and template layers. Add small, pure helpers for model presentation and summary parsing so the new behavior is testable without broad refactors.

**Tech Stack:** Go, Gin, GORM, HTML templates, vanilla JavaScript, Go `testing`

---

### Task 1: Lock failing behaviors with tests

**Files:**
- Create: `internal/engine/inference_engine_test.go`
- Create: `internal/api/server_test.go`

- [ ] **Step 1: Write the failing engine tests**

```go
func TestRunInferenceDoesNotCreateTaskWhenProviderUnavailable(t *testing.T) {}
func TestParseSummaryResponseParsesStructuredSummary(t *testing.T) {}
```

- [ ] **Step 2: Run engine tests to verify they fail**

Run: `go test ./internal/engine -run 'TestRunInferenceDoesNotCreateTaskWhenProviderUnavailable|TestParseSummaryResponseParsesStructuredSummary'`
Expected: FAIL because task creation happens before provider lookup and summary parsing helper does not exist yet.

- [ ] **Step 3: Write the failing API test**

```go
func TestRunInferenceUsesConfiguredDefaultModelWhenRequestModelEmpty(t *testing.T) {}
```

- [ ] **Step 4: Run API test to verify it fails**

Run: `go test ./internal/api -run TestRunInferenceUsesConfiguredDefaultModelWhenRequestModelEmpty`
Expected: FAIL because the handler currently hard-codes `minimax`.

### Task 2: Fix backend inference flow

**Files:**
- Modify: `internal/engine/inference_engine.go`
- Modify: `internal/models/models.go`

- [ ] **Step 1: Resolve provider before persisting the task**
- [ ] **Step 2: Add a structured summary type and parsing helper**
- [ ] **Step 3: Populate summary fields from the model response with minimal fallback behavior**
- [ ] **Step 4: Run targeted engine tests**

Run: `go test ./internal/engine`
Expected: PASS

### Task 3: Fix API defaults and available-model presentation

**Files:**
- Modify: `internal/api/server.go`
- Modify: `web/templates/inference.html`

- [ ] **Step 1: Remove the hard-coded API default model and let the engine config decide**
- [ ] **Step 2: Add a deterministic model-option builder for template rendering**
- [ ] **Step 3: Render only available models in the template and mark the resolved default as selected**
- [ ] **Step 4: Run targeted API tests**

Run: `go test ./internal/api`
Expected: PASS

### Task 4: Verify end-to-end regression coverage

**Files:**
- Modify: `web/templates/inference.html`

- [ ] **Step 1: Update summary rendering to show structured sections when available**
- [ ] **Step 2: Keep string fallback rendering for older/unstructured results**
- [ ] **Step 3: Run the full test suite**

Run: `go test ./...`
Expected: PASS
