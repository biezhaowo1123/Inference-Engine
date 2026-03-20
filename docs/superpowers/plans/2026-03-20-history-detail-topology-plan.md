# History Detail And Topology Graph Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Extend the existing history page so each inference record can expand inline to show persisted summary, full steps, and a single persisted topology graph that distinguishes reasoning nodes from fact nodes.

**Architecture:** Keep the current Gin + GORM + server-rendered HTML structure, but split the new behavior into three focused layers: graph generation in the engine, typed detail-response assembly in the API layer, and inline expand/render behavior in the history page. Persist summary and graph JSON on `InferenceTask` so history detail is fast and does not trigger new model calls.

**Tech Stack:** Go, Gin, GORM, SQLite/PostgreSQL, HTML, CSS, vanilla JavaScript, Go `testing`

---

### File Structure

**Create:**
- `internal/engine/inference_graph.go`
- `internal/api/task_detail_response.go`

**Modify:**
- `internal/models/models.go`
- `internal/engine/inference_engine.go`
- `internal/engine/inference_engine_test.go`
- `internal/api/server.go`
- `internal/api/server_test.go`
- `web/templates/history.html`
- `web/static/style.css`

**Responsibilities:**
- `internal/models/models.go`: persist summary + graph fields and define graph payload structs
- `internal/engine/inference_graph.go`: build graph prompt, parse/validate graph JSON, and normalize graph payloads
- `internal/engine/inference_engine.go`: orchestrate summary persistence and non-fatal graph generation
- `internal/api/task_detail_response.go`: convert DB task records into typed API responses with parsed `summary_data` and `graph_data`
- `internal/api/server.go`: reuse the new typed detail response in `GET /api/inference/:id`
- `web/templates/history.html`: lazy detail fetch, inline expand/collapse, and SVG graph rendering
- `web/static/style.css`: styles for detail section, graph nodes, graph empty state, and expanded-card layout

---

### Task 1: Persist Summary And Graph Data On Tasks

**Files:**
- Create: `internal/engine/inference_graph.go`
- Modify: `internal/models/models.go`
- Modify: `internal/engine/inference_engine.go`
- Test: `internal/engine/inference_engine_test.go`

- [ ] **Step 1: Write the failing persistence test**

```go
func TestRunInferencePersistsSummaryAndGraphData(t *testing.T) {
    // fake provider returns:
    // 1. step JSON
    // 2. summary JSON
    // 3. graph JSON
    // then assert the stored task has Summary, SummaryData, GraphData, GraphStatus
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/engine -run TestRunInferencePersistsSummaryAndGraphData`

Expected: FAIL because `InferenceTask` does not yet persist summary/graph fields and the engine does not save graph JSON.

- [ ] **Step 3: Add persisted summary and graph fields to the task model**

```go
type InferenceTask struct {
    Summary     string `json:"summary" gorm:"type:text"`
    SummaryData string `json:"summary_data" gorm:"type:json"`
    GraphData   string `json:"graph_data" gorm:"type:json"`
    GraphStatus string `json:"graph_status" gorm:"size:20"`
    GraphError  string `json:"graph_error" gorm:"type:text"`
}
```

- [ ] **Step 4: Add graph payload structs and parser helpers**

```go
type InferenceGraph struct {
    Nodes []InferenceGraphNode `json:"nodes"`
    Edges []InferenceGraphEdge `json:"edges"`
}

type InferenceGraphNode struct {
    ID    string `json:"id"`
    Label string `json:"label"`
    Type  string `json:"type"`
}

type InferenceGraphEdge struct {
    Source string `json:"source"`
    Target string `json:"target"`
    Label  string `json:"label"`
}
```

- [ ] **Step 5: Implement graph prompt + validation in `internal/engine/inference_graph.go`**

```go
func buildGraphPrompt(result *models.InferenceResult, req *models.InferenceRequest) string
func parseGraphResponse(raw string) (*models.InferenceGraph, error)
```

- [ ] **Step 6: Persist summary and graph data in the inference engine**

```go
result.Summary = summaryData.Summary
task.Summary = result.Summary
task.SummaryData = string(summaryJSON)
task.GraphStatus = "pending"
```

- [ ] **Step 7: Generate the graph after summary and save the normalized graph JSON**

```go
graph, err := generateInferenceGraph(ctx, provider, req, result)
if err != nil {
    task.GraphStatus = "failed"
    task.GraphError = err.Error()
} else {
    task.GraphStatus = "completed"
    task.GraphData = string(graphJSON)
}
```

- [ ] **Step 8: Run the persistence test to verify it passes**

Run: `go test ./internal/engine -run TestRunInferencePersistsSummaryAndGraphData`

Expected: PASS

- [ ] **Step 9: Commit the persistence slice**

```bash
git add internal/models/models.go internal/engine/inference_graph.go internal/engine/inference_engine.go internal/engine/inference_engine_test.go
git commit -m "feat: persist inference summary and graph data"
```

### Task 2: Keep Graph Generation Non-Fatal

**Files:**
- Modify: `internal/engine/inference_engine.go`
- Test: `internal/engine/inference_engine_test.go`

- [ ] **Step 1: Write the failing non-fatal test**

```go
func TestRunInferenceKeepsTaskCompletedWhenGraphGenerationFails(t *testing.T) {
    // fake provider returns valid step + summary, invalid graph JSON
    // assert returned inference succeeds
    // assert stored task.GraphStatus == "failed"
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/engine -run TestRunInferenceKeepsTaskCompletedWhenGraphGenerationFails`

Expected: FAIL because graph errors are not yet isolated from the main success path.

- [ ] **Step 3: Implement a dedicated non-fatal graph branch**

```go
if graphErr != nil {
    task.GraphStatus = "failed"
    task.GraphError = truncateGraphError(graphErr)
    // do not return an error here
}
```

- [ ] **Step 4: Keep the main task status as completed after successful steps**

```go
e.db.Model(task).Updates(map[string]interface{}{
    "status": "completed",
    "graph_status": task.GraphStatus,
    "graph_error": task.GraphError,
})
```

- [ ] **Step 5: Run the non-fatal test to verify it passes**

Run: `go test ./internal/engine -run TestRunInferenceKeepsTaskCompletedWhenGraphGenerationFails`

Expected: PASS

- [ ] **Step 6: Run the full engine package tests**

Run: `go test ./internal/engine`

Expected: PASS

- [ ] **Step 7: Commit the non-fatal graph handling**

```bash
git add internal/engine/inference_engine.go internal/engine/inference_engine_test.go
git commit -m "feat: make graph generation non-fatal"
```

### Task 3: Return Typed Detail Payloads From The Existing Detail API

**Files:**
- Create: `internal/api/task_detail_response.go`
- Modify: `internal/api/server.go`
- Test: `internal/api/server_test.go`

- [ ] **Step 1: Write the failing detail API test**

```go
func TestGetInferenceReturnsSummaryAndGraphData(t *testing.T) {
    // seed a task with Summary, SummaryData, GraphData, GraphStatus
    // GET /api/inference/:id
    // assert response includes parsed summary_data and graph_data objects
}
```

- [ ] **Step 2: Run the test to verify it fails**

Run: `go test ./internal/api -run TestGetInferenceReturnsSummaryAndGraphData`

Expected: FAIL because the current endpoint returns raw `InferenceTask` and does not parse persisted JSON fields into typed payloads.

- [ ] **Step 3: Add a typed response assembler**

```go
type inferenceTaskDetailResponse struct {
    ID          uint                     `json:"id"`
    Title       string                   `json:"title"`
    Summary     string                   `json:"summary"`
    SummaryData *models.InferenceSummary `json:"summary_data,omitempty"`
    GraphData   *models.InferenceGraph   `json:"graph_data,omitempty"`
    GraphStatus string                   `json:"graph_status"`
    GraphError  string                   `json:"graph_error,omitempty"`
    Steps       []models.InferenceStep   `json:"steps"`
}
```

- [ ] **Step 4: Parse `SummaryData` and `GraphData` safely in the new helper**

```go
func newInferenceTaskDetailResponse(task models.InferenceTask) inferenceTaskDetailResponse
```

- [ ] **Step 5: Update `GET /api/inference/:id` to return the typed response**

```go
c.JSON(200, gin.H{
    "success": true,
    "data":    newInferenceTaskDetailResponse(task),
})
```

- [ ] **Step 6: Run the detail API test to verify it passes**

Run: `go test ./internal/api -run TestGetInferenceReturnsSummaryAndGraphData`

Expected: PASS

- [ ] **Step 7: Run the full API package tests**

Run: `go test ./internal/api`

Expected: PASS

- [ ] **Step 8: Commit the API detail payload changes**

```bash
git add internal/api/task_detail_response.go internal/api/server.go internal/api/server_test.go
git commit -m "feat: return typed inference detail payloads"
```

### Task 4: Expand History Cards Inline And Render The Graph

**Files:**
- Modify: `web/templates/history.html`
- Modify: `web/static/style.css`

- [ ] **Step 1: Add the detail toggle control to each history card**

```html
<button class="btn-detail" onclick="toggleDetail(${task.id})">查看详情</button>
<div class="history-detail" id="task-detail-${task.id}" hidden></div>
```

- [ ] **Step 2: Add lazy detail-fetch logic with in-memory caching**

```js
const detailCache = new Map();

async function toggleDetail(id) {
  if (!detailCache.has(id)) {
    detailCache.set(id, await fetchTaskDetail(id));
  }
  renderTaskDetail(id, detailCache.get(id));
}
```

- [ ] **Step 3: Add the expanded detail markup for summary, steps, and graph**

```js
function renderTaskDetail(id, task) {
  detail.innerHTML = `
    <section class="detail-summary">...</section>
    <section class="detail-steps">...</section>
    <section class="detail-graph">${renderGraph(task.graph_data, task.graph_status, task.graph_error)}</section>
  `;
}
```

- [ ] **Step 4: Implement a lightweight SVG graph renderer**

```js
function renderGraph(graphData, graphStatus, graphError) {
  if (!graphData || graphStatus === 'failed') return renderGraphEmptyState(graphError);
  const layout = buildGraphLayout(graphData);
  return `<svg ...>${layout.edges}${layout.nodes}</svg>`;
}
```

- [ ] **Step 5: Add styles that clearly distinguish node types**

```css
.graph-node.reasoning { fill: var(--primary); }
.graph-node.fact { fill: #6bbf83; }
.history-detail[hidden] { display: none; }
```

- [ ] **Step 6: Verify the history page manually**

Run: `curl -s http://localhost:8080/history | sed -n '1,240p'`

Expected: detail controls and graph container hooks appear in the rendered HTML.

- [ ] **Step 7: Commit the history page UI slice**

```bash
git add web/templates/history.html web/static/style.css
git commit -m "feat: expand history details with topology graph"
```

### Task 5: Final Verification

**Files:**
- Modify: `internal/engine/inference_engine_test.go`
- Modify: `internal/api/server_test.go`
- Modify: `web/templates/history.html`
- Modify: `web/static/style.css`

- [ ] **Step 1: Run the full Go test suite**

Run: `go test ./...`

Expected: PASS

- [ ] **Step 2: Restart the local server**

Run: `go run ./cmd/server`

Expected: the server starts on `http://localhost:8080` with no startup errors.

- [ ] **Step 3: Manually verify one real inference end to end**

Run:

```bash
curl -sS -X POST http://localhost:8080/api/inference \
  -H 'Content-Type: application/json' \
  -d '{"title":"历史图谱测试","domain":"历史","subject":"秦朝","change_point":"扶苏继位","time_frame":{"start":"前210年","end":"前180年"},"steps_count":1,"model":"minimax"}'
```

Expected:

- response succeeds
- returned task has persisted `created_at`
- history detail later shows summary, steps, and graph state

- [ ] **Step 4: Open the history page and verify inline detail expansion**

Manual verification checklist:

- `查看详情` expands the clicked card in place
- summary appears for historical records
- steps render in order
- graph renders with distinct fact/reasoning styling or shows a graph empty-state message
- collapsing and re-expanding reuse cached data without duplicating UI blocks

- [ ] **Step 5: Commit any final polish required by verification**

```bash
git add internal/engine/inference_engine_test.go internal/api/server_test.go web/templates/history.html web/static/style.css
git commit -m "test: verify history detail and topology flow"
```
