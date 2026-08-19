package app

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"forge-api-client/internal/curl"
	"forge-api-client/internal/httpclient"
	"forge-api-client/internal/models"
	"forge-api-client/internal/security"
	"forge-api-client/internal/soap"
	"forge-api-client/internal/storage"
	"forge-api-client/internal/variables"
	"forge-api-client/internal/wsdl"
	"forge-api-client/version"
	"github.com/google/uuid"
	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx           context.Context
	eventsEnabled bool
	store         *storage.Store
	engine        *httpclient.Engine
	protector     *security.Protector
	loadMu        sync.Mutex
	loadStops     map[string]context.CancelFunc
}

func New(store *storage.Store, protector *security.Protector) *App {
	return &App{store: store, engine: httpclient.New(), protector: protector, loadStops: map[string]context.CancelFunc{}}
}
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	a.eventsEnabled = true
	if state, err := a.store.LoadState(); err == nil {
		if state.Settings.WindowWidth >= 960 && state.Settings.WindowHeight >= 640 {
			wruntime.WindowSetSize(ctx, state.Settings.WindowWidth, state.Settings.WindowHeight)
		}
		if state.Settings.WindowX != 0 || state.Settings.WindowY != 0 {
			wruntime.WindowSetPosition(ctx, state.Settings.WindowX, state.Settings.WindowY)
		}
	}
}

func (a *App) Shutdown(ctx context.Context) {
	if state, err := a.store.LoadState(); err == nil {
		state.Settings.WindowWidth, state.Settings.WindowHeight = wruntime.WindowGetSize(ctx)
		state.Settings.WindowX, state.Settings.WindowY = wruntime.WindowGetPosition(ctx)
		_ = a.store.SaveSettings(state.Settings)
	}
	_ = a.store.Close()
}

func (a *App) Bootstrap() (models.AppState, error) {
	state, err := a.store.LoadState()
	if err != nil {
		return state, err
	}
	if len(state.Workspaces) == 0 {
		if err := a.seed(); err != nil {
			return state, err
		}
		state, err = a.store.LoadState()
		if err != nil {
			return state, err
		}
	}
	for i := range state.Environments {
		a.decryptEnvironment(&state.Environments[i])
	}
	for i := range state.Requests {
		a.decryptRequest(&state.Requests[i])
	}
	for i := range state.History {
		a.decryptRequest(&state.History[i].Request)
	}
	state.Version = version.Version
	state.Commit = version.Commit
	state.BuildDate = version.BuildDate
	state.GoVersion = runtime.Version()
	return state, nil
}

func (a *App) seed() error {
	now := time.Now().UTC()
	workspace := models.Workspace{ID: uuid.NewString(), Name: "My Workspace", CreatedAt: now, UpdatedAt: now}
	if err := a.store.SaveWorkspace(workspace); err != nil {
		return err
	}
	collection := models.Collection{ID: uuid.NewString(), WorkspaceID: workspace.ID, Name: "Getting Started", Description: "Requests locais de exemplo", Variables: map[string]string{}, SortOrder: 0}
	if err := a.store.SaveCollection(collection); err != nil {
		return err
	}
	rest := newRequest(workspace.ID, "REST", "GET")
	rest.CollectionID = collection.ID
	rest.Name = "HTTPBin GET"
	rest.URL = "https://httpbin.org/get"
	if err := a.store.SaveRequest(rest); err != nil {
		return err
	}
	soapReq := newRequest(workspace.ID, "SOAP", "POST")
	soapReq.CollectionID = collection.ID
	soapReq.Name = "SOAP example"
	soapReq.URL = "https://example.invalid/soap"
	soapReq.SOAPAction = "urn:GetCustomer"
	soapReq.Body = `<?xml version="1.0" encoding="UTF-8"?>
<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:urn="urn:example">
  <soapenv:Header/>
  <soapenv:Body><urn:GetCustomer><urn:CustomerId>{{customerId}}</urn:CustomerId></urn:GetCustomer></soapenv:Body>
</soapenv:Envelope>`
	if err := a.store.SaveRequest(soapReq); err != nil {
		return err
	}
	env := models.Environment{ID: uuid.NewString(), WorkspaceID: workspace.ID, Name: "DEV", Variables: []models.EnvironmentVariable{{ID: uuid.NewString(), Key: "customerId", InitialValue: "1000", CurrentValue: "1000", Enabled: true}}}
	if _, err := a.SaveEnvironment(env); err != nil {
		return err
	}
	settings := models.Settings{Theme: "dark", ActiveWorkspaceID: workspace.ID, ActiveEnvironmentID: env.ID, StoreHistoryBodies: true, HistoryRetentionDays: 30, MaxResponsePreviewSize: 10 * 1024 * 1024, SidebarWidth: 288, EditorSplitPercent: 52, LogLevel: "INFO", WindowWidth: 1440, WindowHeight: 900}
	return a.store.SaveSettings(settings)
}

func newRequest(workspaceID, protocol, method string) models.RequestDefinition {
	now := time.Now().UTC()
	r := models.RequestDefinition{ID: uuid.NewString(), WorkspaceID: workspaceID, Name: "Untitled request", Protocol: protocol, Method: method, URL: "", Params: []models.KeyValue{}, Headers: []models.KeyValue{}, Multipart: []models.MultipartPart{}, Variables: map[string]string{}, Auth: models.AuthDefinition{Type: "none", AddTo: "header"}, BodyType: "none", SOAPVersion: "1.1", VerifySSL: true, FollowRedirect: true, TimeoutSeconds: 30, CreatedAt: now, UpdatedAt: now}
	if protocol == "SOAP" {
		r.BodyType = "xml"
	}
	return r
}
func (a *App) NewRequest(workspaceID, protocol string) models.RequestDefinition {
	return newRequest(workspaceID, strings.ToUpper(protocol), map[bool]string{true: "POST", false: "GET"}[strings.EqualFold(protocol, "SOAP")])
}

func (a *App) SaveWorkspace(v models.Workspace) (models.Workspace, error) {
	if v.ID == "" {
		v.ID = uuid.NewString()
		v.CreatedAt = time.Now().UTC()
	}
	v.UpdatedAt = time.Now().UTC()
	if strings.TrimSpace(v.Name) == "" {
		return v, fmt.Errorf("informe o nome do workspace")
	}
	return v, a.store.SaveWorkspace(v)
}
func (a *App) SaveCollection(v models.Collection) (models.Collection, error) {
	if v.ID == "" {
		v.ID = uuid.NewString()
	}
	if v.Variables == nil {
		v.Variables = map[string]string{}
	}
	return v, a.store.SaveCollection(v)
}
func (a *App) SaveFolder(v models.Folder) (models.Folder, error) {
	if v.ID == "" {
		v.ID = uuid.NewString()
	}
	return v, a.store.SaveFolder(v)
}
func (a *App) SaveRequest(v models.RequestDefinition) (models.RequestDefinition, error) {
	if v.ID == "" {
		v.ID = uuid.NewString()
		v.CreatedAt = time.Now().UTC()
	}
	v.UpdatedAt = time.Now().UTC()
	if v.Name == "" {
		v.Name = "Untitled request"
	}
	plain := v
	v.Headers = append([]models.KeyValue(nil), v.Headers...)
	if err := a.encryptRequest(&v); err != nil {
		return plain, err
	}
	return plain, a.store.SaveRequest(v)
}

func (a *App) DuplicateRequest(source models.RequestDefinition) (models.RequestDefinition, error) {
	now := time.Now().UTC()
	source.ID = uuid.NewString()
	source.Name = strings.TrimSpace(source.Name) + " Copy"
	source.CreatedAt = now
	source.UpdatedAt = now
	return a.SaveRequest(source)
}
func (a *App) SaveEnvironment(v models.Environment) (models.Environment, error) {
	if v.ID == "" {
		v.ID = uuid.NewString()
	}
	for i := range v.Variables {
		if v.Variables[i].ID == "" {
			v.Variables[i].ID = uuid.NewString()
		}
	}
	plain := v
	if err := a.encryptEnvironment(&v); err != nil {
		return plain, err
	}
	return plain, a.store.SaveEnvironment(v)
}
func (a *App) SaveSettings(v models.Settings) error          { return a.store.SaveSettings(v) }
func (a *App) SaveTabs(ids []string) error                   { return a.store.SaveTabs(ids) }
func (a *App) DeleteEntity(kind, id string) error            { return a.store.Delete(kind, id) }
func (a *App) ClearHistory(workspaceID string) error         { return a.store.ClearHistory(workspaceID) }
func (a *App) ClearLoadTests(requestID string) error         { return a.store.ClearLoadTests(requestID) }
func (a *App) MoveFolder(input models.MoveFolderInput) error { return a.store.MoveFolder(input) }

func (a *App) SaveFlow(v models.Flow) (models.Flow, error) {
	if v.ID == "" {
		v.ID = uuid.NewString()
		v.CreatedAt = time.Now().UTC()
	}
	if strings.TrimSpace(v.Name) == "" {
		v.Name = "Untitled flow"
	}
	if v.Nodes == nil {
		v.Nodes = []models.FlowNode{}
	}
	if v.Edges == nil {
		v.Edges = []models.FlowEdge{}
	}
	v.UpdatedAt = time.Now().UTC()
	return v, a.store.SaveFlow(v)
}

func (a *App) RunFlow(input models.FlowRunInput) (models.FlowRunResult, error) {
	started := time.Now().UTC()
	result := models.FlowRunResult{
		ID:        uuid.NewString(),
		FlowID:    input.Flow.ID,
		StartedAt: started,
		Variables: map[string]string{},
		Nodes:     []models.FlowNodeResult{},
	}

	state, err := a.store.LoadState()
	if err != nil {
		return result, err
	}

	requests := map[string]models.RequestDefinition{}
	for _, request := range state.Requests {
		a.decryptRequest(&request)
		requests[request.ID] = request
	}
	collections := map[string]models.Collection{}
	for _, collection := range state.Collections {
		collections[collection.ID] = collection
	}

	nodes := map[string]models.FlowNode{}
	incomingCount := map[string]int{}
	outgoing := map[string][]models.FlowEdge{}
	for _, node := range input.Flow.Nodes {
		if strings.TrimSpace(node.ID) == "" {
			return result, fmt.Errorf("flow contém um nó sem identificador")
		}
		if _, exists := nodes[node.ID]; exists {
			return result, fmt.Errorf("flow contém nós duplicados")
		}
		nodes[node.ID] = node
		incomingCount[node.ID] = 0
	}
	for _, edge := range input.Flow.Edges {
		if _, ok := nodes[edge.SourceID]; !ok {
			return result, fmt.Errorf("conexão possui origem inválida")
		}
		if _, ok := nodes[edge.TargetID]; !ok {
			return result, fmt.Errorf("conexão possui destino inválido")
		}
		if edge.SourceID == edge.TargetID && normalizedFlowNodeType(nodes[edge.SourceID]) != "loop" {
			return result, fmt.Errorf("um nó não pode conectar a si mesmo")
		}
		outgoing[edge.SourceID] = append(outgoing[edge.SourceID], edge)
		incomingCount[edge.TargetID]++
	}

	flowValues := variables.Merge(input.Variables)
	var flowMu sync.RWMutex
	var resultMu sync.Mutex
	var failedMu sync.Mutex
	failed := false
	failureText := ""

	snapshotValues := func() map[string]string {
		flowMu.RLock()
		defer flowMu.RUnlock()
		return variables.Merge(flowValues)
	}
	setValue := func(key, value string) {
		flowMu.Lock()
		flowValues[key] = value
		flowMu.Unlock()
	}
	appendResult := func(nodeResult models.FlowNodeResult) {
		resultMu.Lock()
		result.Nodes = append(result.Nodes, nodeResult)
		resultMu.Unlock()
	}
	markFailed := func(message string) {
		failedMu.Lock()
		failed = true
		if failureText == "" {
			failureText = message
		}
		failedMu.Unlock()
	}
	isFailed := func() bool {
		failedMu.Lock()
		defer failedMu.Unlock()
		return failed
	}

	emitNode := func(nodeID, status string, iteration int) {
		if a.eventsEnabled && a.ctx != nil {
			wruntime.EventsEmit(a.ctx, "flow:progress", map[string]any{
				"flowId": input.Flow.ID, "nodeId": nodeID, "status": status, "iteration": iteration,
			})
		}
	}
	emitEdge := func(edgeID, status string, iteration int) {
		if a.eventsEnabled && a.ctx != nil {
			wruntime.EventsEmit(a.ctx, "flow:progress", map[string]any{
				"flowId": input.Flow.ID, "edgeId": edgeID, "status": status, "iteration": iteration,
			})
		}
	}

	var executeNode func(string, models.ExecuteRequestResult, map[string]bool, map[string]int) error
	var followEdges func([]models.FlowEdge, models.ExecuteRequestResult, map[string]bool, map[string]int, bool) error

	followEdges = func(edges []models.FlowEdge, response models.ExecuteRequestResult, loopReturns map[string]bool, iterations map[string]int, parallel bool) error {
		if len(edges) == 0 {
			return nil
		}
		if parallel && len(edges) > 1 {
			var wg sync.WaitGroup
			errCh := make(chan error, len(edges))
			for _, edge := range edges {
				edge := edge
				wg.Add(1)
				go func() {
					defer wg.Done()
					emitEdge(edge.ID, "running", 0)
					if loopReturns[edge.TargetID] {
						emitEdge(edge.ID, "completed", 0)
						return
					}
					childReturns := cloneBoolMap(loopReturns)
					childIterations := cloneIntMap(iterations)
					if err := executeNode(edge.TargetID, response, childReturns, childIterations); err != nil {
						errCh <- err
					}
					emitEdge(edge.ID, "completed", 0)
				}()
			}
			wg.Wait()
			close(errCh)
			for err := range errCh {
				if err != nil {
					return err
				}
			}
			return nil
		}
		for _, edge := range edges {
			emitEdge(edge.ID, "running", 0)
			if loopReturns[edge.TargetID] {
				emitEdge(edge.ID, "completed", 0)
				continue
			}
			if err := executeNode(edge.TargetID, response, cloneBoolMap(loopReturns), cloneIntMap(iterations)); err != nil {
				return err
			}
			emitEdge(edge.ID, "completed", 0)
		}
		return nil
	}

	executeNode = func(nodeID string, upstreamResponse models.ExecuteRequestResult, loopReturns map[string]bool, iterations map[string]int) error {
		if isFailed() {
			return fmt.Errorf("flow interrompido")
		}
		node, ok := nodes[nodeID]
		if !ok {
			return fmt.Errorf("nó %s não encontrado", nodeID)
		}
		nodeType := normalizedFlowNodeType(node)
		iteration := iterations["__loopIteration"]
		if iterations[nodeID] > 0 {
			iteration = iterations[nodeID]
		}
		emitNode(node.ID, "running", iteration)
		nodeResult := models.FlowNodeResult{
			NodeID: node.ID, Type: nodeType, RequestID: node.RequestID, Name: node.Name,
			StartedAt: time.Now().UTC(), Extracted: map[string]string{}, Iteration: iteration,
		}
		response := upstreamResponse
		var selected []models.FlowEdge
		parallelNext := false

		switch nodeType {
		case "request":
			request, exists := requests[node.RequestID]
			if !exists {
				nodeResult.Error = "request salva não encontrada"
				appendResult(nodeResult)
				emitNode(node.ID, "failed", iteration)
				markFailed(nodeResult.Error)
				return fmt.Errorf("%s", nodeResult.Error)
			}
			values := snapshotValues()
			request.Variables = variables.Merge(request.Variables, values)
			collectionVariables := map[string]string{}
			if collection, ok := collections[request.CollectionID]; ok {
				collectionVariables = collection.Variables
			}
			executed, executeErr := a.ExecuteRequest(models.ExecuteRequestInput{
				ExecutionID: input.ExecutionID + "-" + node.ID, Request: request, EnvironmentID: input.EnvironmentID,
				GlobalVariables: values, CollectionVariables: collectionVariables,
			})
			response = executed
			nodeResult.Response = executed
			if executeErr != nil {
				nodeResult.Error = security.SanitizeText(executeErr.Error())
				appendResult(nodeResult)
				emitNode(node.ID, "failed", iteration)
				markFailed(nodeResult.Error)
				return executeErr
			}
			for _, mapping := range node.Mappings {
				if strings.TrimSpace(mapping.Variable) == "" || strings.TrimSpace(mapping.Path) == "" {
					continue
				}
				value, extractErr := extractResponseValue(executed, mapping.Path)
				if extractErr != nil {
					nodeResult.Error = fmt.Sprintf("não foi possível extrair %s: %v", mapping.Path, extractErr)
					appendResult(nodeResult)
					emitNode(node.ID, "failed", iteration)
					markFailed(nodeResult.Error)
					return extractErr
				}
				setValue(mapping.Variable, value)
				nodeResult.Extracted[mapping.Variable] = value
			}
			if executed.StatusCode >= 400 {
				nodeResult.Error = fmt.Sprintf("HTTP %d", executed.StatusCode)
				appendResult(nodeResult)
				emitNode(node.ID, "failed", iteration)
				markFailed(nodeResult.Error)
				return fmt.Errorf("%s", nodeResult.Error)
			}
			nodeResult.Outcome = "success"
			selected = flowEdgesForHandle(outgoing[node.ID], "default")

		case "if":
			passed, conditionErr := evaluateFlowConditions(node, response, snapshotValues())
			if conditionErr != nil {
				nodeResult.Error = conditionErr.Error()
				appendResult(nodeResult)
				emitNode(node.ID, "failed", iteration)
				markFailed(nodeResult.Error)
				return conditionErr
			}
			if passed {
				nodeResult.Outcome = "true"
			} else {
				nodeResult.Outcome = "false"
			}
			selected = flowEdgesForHandle(outgoing[node.ID], nodeResult.Outcome)

		case "filter":
			passed, conditionErr := evaluateFlowConditions(node, response, snapshotValues())
			if conditionErr != nil {
				nodeResult.Error = conditionErr.Error()
				appendResult(nodeResult)
				emitNode(node.ID, "failed", iteration)
				markFailed(nodeResult.Error)
				return conditionErr
			}
			if passed {
				nodeResult.Outcome = "passed"
				selected = flowEdgesForHandle(outgoing[node.ID], "default")
			} else {
				nodeResult.Outcome = "filtered"
			}

		case "extract":
			values := snapshotValues()
			for _, mapping := range node.Mappings {
				if strings.TrimSpace(mapping.Variable) == "" || strings.TrimSpace(mapping.Path) == "" {
					continue
				}
				value, extractErr := extractFlowValue(response, values, mapping.Path)
				if extractErr != nil {
					nodeResult.Error = fmt.Sprintf("não foi possível extrair %s: %v", mapping.Path, extractErr)
					appendResult(nodeResult)
					emitNode(node.ID, "failed", iteration)
					markFailed(nodeResult.Error)
					return extractErr
				}
				setValue(mapping.Variable, value)
				nodeResult.Extracted[mapping.Variable] = value
			}
			nodeResult.Outcome = "success"
			selected = flowEdgesForHandle(outgoing[node.ID], "default")

		case "wait":
			delay, delayErr := flowWaitDuration(node)
			if delayErr != nil {
				nodeResult.Error = delayErr.Error()
				appendResult(nodeResult)
				emitNode(node.ID, "failed", iteration)
				markFailed(nodeResult.Error)
				return delayErr
			}
			time.Sleep(delay)
			nodeResult.Outcome = "waited"
			selected = flowEdgesForHandle(outgoing[node.ID], "default")

		case "switch":
			handle, switchErr := evaluateFlowSwitch(node, response, snapshotValues())
			if switchErr != nil {
				nodeResult.Error = switchErr.Error()
				appendResult(nodeResult)
				emitNode(node.ID, "failed", iteration)
				markFailed(nodeResult.Error)
				return switchErr
			}
			nodeResult.Outcome = handle
			selected = flowEdgesForHandle(outgoing[node.ID], handle)

		case "parallel":
			nodeResult.Outcome = "parallel"
			selected = outgoing[node.ID]
			parallelNext = true

		case "loop":
			count := node.LoopCount
			if count < 1 {
				count = 1
			}
			loopEdges := flowEdgesForHandle(outgoing[node.ID], "loop")
			doneEdges := flowEdgesForHandle(outgoing[node.ID], "done")
			nodeResult.Outcome = fmt.Sprintf("%d iterations", count)
			appendResult(nodeResult)
			for i := 1; i <= count; i++ {
				emitNode(node.ID, "running", i)
				bodyReturns := cloneBoolMap(loopReturns)
				bodyReturns[node.ID] = true
				bodyIterations := cloneIntMap(iterations)
				bodyIterations[node.ID] = i
				if err := followEdges(loopEdges, response, bodyReturns, bodyIterations, false); err != nil {
					return err
				}
			}
			emitNode(node.ID, "completed", count)
			return followEdges(doneEdges, response, cloneBoolMap(loopReturns), cloneIntMap(iterations), false)

		default:
			nodeResult.Error = fmt.Sprintf("tipo de nó não suportado: %s", node.Type)
			appendResult(nodeResult)
			emitNode(node.ID, "failed", iteration)
			markFailed(nodeResult.Error)
			return fmt.Errorf("%s", nodeResult.Error)
		}

		appendResult(nodeResult)
		emitNode(node.ID, "completed", iteration)
		return followEdges(selected, response, loopReturns, iterations, parallelNext)
	}

	roots := []models.FlowNode{}
	for _, node := range input.Flow.Nodes {
		// Edges returning to a Loop are feedback connections and don't disqualify it as a root.
		count := 0
		for _, edge := range input.Flow.Edges {
			if edge.TargetID == node.ID && !(normalizedFlowNodeType(node) == "loop" && strings.EqualFold(edge.SourceHandle, "loop-return")) {
				count++
			}
		}
		if count == 0 {
			roots = append(roots, node)
		}
	}
	if len(roots) == 0 && len(input.Flow.Nodes) > 0 {
		// A common Loop graph has one feedback edge into the loop. Start from the left-most node.
		roots = append(roots, input.Flow.Nodes[0])
		for _, node := range input.Flow.Nodes {
			if node.X < roots[0].X {
				roots[0] = node
			}
		}
	}
	sort.Slice(roots, func(i, j int) bool {
		if roots[i].X == roots[j].X {
			return roots[i].Y < roots[j].Y
		}
		return roots[i].X < roots[j].X
	})

	for _, root := range roots {
		if err := executeNode(root.ID, models.ExecuteRequestResult{}, map[string]bool{}, map[string]int{}); err != nil {
			break
		}
	}

	result.Successful = !isFailed()
	result.DurationMs = time.Since(started).Milliseconds()
	result.Variables = snapshotValues()
	if a.eventsEnabled && a.ctx != nil {
		wruntime.EventsEmit(a.ctx, "flow:progress", map[string]any{"flowId": input.Flow.ID, "status": "finished", "successful": result.Successful})
	}
	_ = failureText
	return result, nil
}

func cloneBoolMap(input map[string]bool) map[string]bool {
	out := make(map[string]bool, len(input)+1)
	for key, value := range input {
		out[key] = value
	}
	return out
}

func cloneIntMap(input map[string]int) map[string]int {
	out := make(map[string]int, len(input)+1)
	for key, value := range input {
		out[key] = value
	}
	return out
}

func flowEdgesForHandle(edges []models.FlowEdge, handle string) []models.FlowEdge {
	wanted := strings.ToLower(strings.TrimSpace(handle))
	out := []models.FlowEdge{}
	for _, edge := range edges {
		actual := strings.ToLower(strings.TrimSpace(edge.SourceHandle))
		if actual == "" {
			actual = "default"
		}
		if actual == wanted || (wanted == "default" && actual == "default") {
			out = append(out, edge)
		}
	}
	return out
}

func flowWaitDuration(node models.FlowNode) (time.Duration, error) {
	value := node.WaitValue
	if value < 0 {
		return 0, fmt.Errorf("tempo de espera não pode ser negativo")
	}
	unit := strings.ToLower(strings.TrimSpace(node.WaitUnit))
	if unit == "" {
		unit = "ms"
	}
	var multiplier time.Duration
	switch unit {
	case "ms", "millisecond", "milliseconds":
		multiplier = time.Millisecond
	case "s", "sec", "second", "seconds":
		multiplier = time.Second
	case "min", "minute", "minutes":
		multiplier = time.Minute
	default:
		return 0, fmt.Errorf("unidade de espera inválida: %s", node.WaitUnit)
	}
	return time.Duration(value * float64(multiplier)), nil
}

func evaluateFlowSwitch(node models.FlowNode, response models.ExecuteRequestResult, values map[string]string) (string, error) {
	mode := strings.ToLower(strings.TrimSpace(node.SwitchMode))
	if mode == "" || mode == "rules" {
		for index, rule := range node.SwitchRules {
			conditionNode := node
			conditionNode.Conditions = []models.FlowCondition{{Left: rule.Left, Operator: rule.Operator, Right: rule.Right}}
			conditionNode.Match = "all"
			passed, err := evaluateFlowConditions(conditionNode, response, values)
			if err != nil {
				return "", err
			}
			if passed {
				return strconv.Itoa(index), nil
			}
		}
		return "", nil
	}
	if mode == "expression" {
		resolved, err := resolveFlowOperand(response, values, node.SwitchExpression)
		if err != nil {
			return "", err
		}
		index, err := strconv.Atoi(strings.TrimSpace(resolved))
		if err != nil {
			return "", fmt.Errorf("Switch Expression deve resultar em um índice numérico: %q", resolved)
		}
		outputs := node.SwitchOutputs
		if outputs < 2 {
			outputs = 2
		}
		if index < 0 || index >= outputs {
			return "", fmt.Errorf("índice %d fora do intervalo 0..%d", index, outputs-1)
		}
		return strconv.Itoa(index), nil
	}
	return "", fmt.Errorf("modo de Switch não suportado: %s", node.SwitchMode)
}

func normalizedFlowNodeType(node models.FlowNode) string {
	nodeType := strings.ToLower(strings.TrimSpace(node.Type))
	if nodeType == "" {
		return "request"
	}
	return nodeType
}

func activateDefaultFlowEdges(indices []int, active map[int]bool) {
	for _, index := range indices {
		active[index] = true
	}
}

func activateBranchFlowEdges(edges []models.FlowEdge, indices []int, active map[int]bool, branch bool) {
	wanted := "false"
	if branch {
		wanted = "true"
	}
	for _, index := range indices {
		handle := strings.ToLower(strings.TrimSpace(edges[index].SourceHandle))
		if handle == "" || handle == "default" {
			// A legacy/default connection from an IF is treated as the true branch.
			active[index] = branch
			continue
		}
		active[index] = handle == wanted
	}
}

func evaluateFlowConditions(node models.FlowNode, response models.ExecuteRequestResult, values map[string]string) (bool, error) {
	conditions := node.Conditions
	if len(conditions) == 0 {
		return false, fmt.Errorf("adicione ao menos uma condição ao nó %s", node.Name)
	}

	matchAny := strings.EqualFold(strings.TrimSpace(node.Match), "any")
	matched := 0
	for _, condition := range conditions {
		left, err := resolveFlowOperand(response, values, condition.Left)
		if err != nil {
			op := strings.ToLower(strings.TrimSpace(condition.Operator))
			if op == "not_exists" || op == "empty" {
				left = ""
			} else if op == "exists" || op == "not_empty" {
				left = ""
			} else {
				return false, fmt.Errorf("condição %q: %w", condition.Left, err)
			}
		}
		right, err := resolveFlowOperand(response, values, condition.Right)
		if err != nil {
			return false, fmt.Errorf("condição %q: %w", condition.Right, err)
		}
		ok, err := compareFlowValues(left, right, condition.Operator, node.ConvertTypes)
		if err != nil {
			return false, err
		}
		if ok {
			matched++
			if matchAny {
				return true, nil
			}
		} else if !matchAny {
			return false, nil
		}
	}
	if matchAny {
		return matched > 0, nil
	}
	return matched == len(conditions), nil
}

func resolveFlowOperand(response models.ExecuteRequestResult, values map[string]string, input string) (string, error) {
	value := strings.TrimSpace(input)
	if value == "" {
		return "", nil
	}
	if strings.HasPrefix(strings.ToLower(value), "variables.") {
		key := strings.TrimSpace(value[len("variables."):])
		resolved, ok := values[key]
		if !ok {
			return "", fmt.Errorf("variável %q não definida", key)
		}
		return resolved, nil
	}
	if strings.Contains(value, "{{") {
		return variables.ResolveStrict(value, values)
	}
	lower := strings.ToLower(value)
	if lower == "status" || lower == "statuscode" || lower == "body" || strings.HasPrefix(lower, "body.") || strings.HasPrefix(lower, "headers.") || strings.HasPrefix(value, "$") {
		path := value
		if strings.HasPrefix(lower, "body.") {
			path = value[len("body."):]
		}
		return extractResponseValue(response, path)
	}
	return value, nil
}

func extractFlowValue(response models.ExecuteRequestResult, values map[string]string, path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", fmt.Errorf("caminho vazio")
	}
	if strings.HasPrefix(strings.ToLower(path), "variables.") || strings.Contains(path, "{{") {
		return resolveFlowOperand(response, values, path)
	}
	lower := strings.ToLower(path)
	if strings.HasPrefix(lower, "body.") {
		path = path[len("body."):]
	}
	return extractResponseValue(response, path)
}

func compareFlowValues(left, right, operator string, convertTypes bool) (bool, error) {
	op := strings.ToLower(strings.TrimSpace(operator))
	if op == "" {
		op = "equals"
	}

	switch op {
	case "exists":
		return strings.TrimSpace(left) != "", nil
	case "not_exists":
		return strings.TrimSpace(left) == "", nil
	case "empty":
		return strings.TrimSpace(left) == "", nil
	case "not_empty":
		return strings.TrimSpace(left) != "", nil
	case "contains":
		return strings.Contains(left, right), nil
	case "not_contains":
		return !strings.Contains(left, right), nil
	case "starts_with":
		return strings.HasPrefix(left, right), nil
	case "ends_with":
		return strings.HasSuffix(left, right), nil
	case "regex":
		matched, err := regexp.MatchString(right, left)
		if err != nil {
			return false, fmt.Errorf("regex inválida: %w", err)
		}
		return matched, nil
	}

	if convertTypes {
		if leftNumber, leftErr := strconv.ParseFloat(strings.TrimSpace(left), 64); leftErr == nil {
			if rightNumber, rightErr := strconv.ParseFloat(strings.TrimSpace(right), 64); rightErr == nil {
				switch op {
				case "equals":
					return leftNumber == rightNumber, nil
				case "not_equals":
					return leftNumber != rightNumber, nil
				case "greater_than":
					return leftNumber > rightNumber, nil
				case "greater_or_equal":
					return leftNumber >= rightNumber, nil
				case "less_than":
					return leftNumber < rightNumber, nil
				case "less_or_equal":
					return leftNumber <= rightNumber, nil
				}
			}
		}
		leftBool, leftBoolErr := strconv.ParseBool(strings.ToLower(strings.TrimSpace(left)))
		rightBool, rightBoolErr := strconv.ParseBool(strings.ToLower(strings.TrimSpace(right)))
		if leftBoolErr == nil && rightBoolErr == nil {
			switch op {
			case "equals":
				return leftBool == rightBool, nil
			case "not_equals":
				return leftBool != rightBool, nil
			}
		}
	}

	switch op {
	case "equals":
		return left == right, nil
	case "not_equals":
		return left != right, nil
	case "greater_than":
		return left > right, nil
	case "greater_or_equal":
		return left >= right, nil
	case "less_than":
		return left < right, nil
	case "less_or_equal":
		return left <= right, nil
	default:
		return false, fmt.Errorf("operador de condição não suportado: %s", operator)
	}
}

func flowExecutionOrder(flow models.Flow) ([]models.FlowNode, error) {
	nodes := map[string]models.FlowNode{}
	indegree := map[string]int{}
	adjacent := map[string][]string{}
	for _, node := range flow.Nodes {
		if node.ID == "" {
			return nil, fmt.Errorf("flow contém um nó sem identificador")
		}
		if _, exists := nodes[node.ID]; exists {
			return nil, fmt.Errorf("flow contém nós duplicados")
		}
		nodes[node.ID] = node
		indegree[node.ID] = 0
	}
	for _, edge := range flow.Edges {
		if _, ok := nodes[edge.SourceID]; !ok {
			return nil, fmt.Errorf("conexão possui origem inválida")
		}
		if _, ok := nodes[edge.TargetID]; !ok {
			return nil, fmt.Errorf("conexão possui destino inválido")
		}
		if edge.SourceID == edge.TargetID {
			return nil, fmt.Errorf("um nó não pode conectar a si mesmo")
		}
		adjacent[edge.SourceID] = append(adjacent[edge.SourceID], edge.TargetID)
		indegree[edge.TargetID]++
	}
	ready := []models.FlowNode{}
	for id, degree := range indegree {
		if degree == 0 {
			ready = append(ready, nodes[id])
		}
	}
	sort.Slice(ready, func(i, j int) bool {
		if ready[i].X == ready[j].X {
			return ready[i].Y < ready[j].Y
		}
		return ready[i].X < ready[j].X
	})
	ordered := make([]models.FlowNode, 0, len(nodes))
	for len(ready) > 0 {
		node := ready[0]
		ready = ready[1:]
		ordered = append(ordered, node)
		for _, target := range adjacent[node.ID] {
			indegree[target]--
			if indegree[target] == 0 {
				ready = append(ready, nodes[target])
				sort.Slice(ready, func(i, j int) bool {
					if ready[i].X == ready[j].X {
						return ready[i].Y < ready[j].Y
					}
					return ready[i].X < ready[j].X
				})
			}
		}
	}
	if len(ordered) != len(nodes) {
		return nil, fmt.Errorf("o flow contém um ciclo; remova uma das conexões")
	}
	return ordered, nil
}

var flowArrayIndex = regexp.MustCompile(`\[(\d+)\]`)

func extractResponseValue(response models.ExecuteRequestResult, path string) (string, error) {
	path = strings.TrimSpace(path)
	switch strings.ToLower(path) {
	case "status", "statuscode":
		return fmt.Sprint(response.StatusCode), nil
	case "body":
		return response.Body, nil
	}
	if strings.HasPrefix(strings.ToLower(path), "headers.") {
		key := strings.TrimSpace(path[len("headers."):])
		for header, values := range response.Headers {
			if strings.EqualFold(header, key) {
				return strings.Join(values, ", "), nil
			}
		}
		return "", fmt.Errorf("header não encontrado")
	}
	var current any
	if err := json.Unmarshal([]byte(response.Body), &current); err != nil {
		return "", fmt.Errorf("response não é um JSON válido")
	}
	path = strings.TrimPrefix(path, "$")
	path = strings.TrimPrefix(path, ".")
	path = flowArrayIndex.ReplaceAllString(path, ".$1")
	for _, part := range strings.Split(path, ".") {
		if part == "" {
			continue
		}
		switch value := current.(type) {
		case map[string]any:
			var exists bool
			current, exists = value[part]
			if !exists {
				return "", fmt.Errorf("campo %q não encontrado", part)
			}
		case []any:
			var index int
			if _, err := fmt.Sscanf(part, "%d", &index); err != nil || index < 0 || index >= len(value) {
				return "", fmt.Errorf("índice %q inválido", part)
			}
			current = value[index]
		default:
			return "", fmt.Errorf("não é possível acessar %q", part)
		}
	}
	if text, ok := current.(string); ok {
		return text, nil
	}
	encoded, err := json.Marshal(current)
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

func (a *App) GetOAuthToken(input models.OAuthTokenInput) (models.OAuthTokenResult, error) {
	state, err := a.store.LoadState()
	if err != nil {
		return models.OAuthTokenResult{}, err
	}
	environmentValues := map[string]string{}
	for _, environment := range state.Environments {
		if environment.ID != input.EnvironmentID {
			continue
		}
		a.decryptEnvironment(&environment)
		for _, variable := range environment.Variables {
			if variable.Enabled {
				environmentValues[variable.Key] = variable.CurrentValue
			}
		}
	}
	values := variables.Merge(input.GlobalVariables, environmentValues, input.CollectionVariables, input.RequestVariables)
	return httpclient.AcquireOAuthToken(a.ctx, input.Auth, values)
}

func (a *App) ExecuteRequest(input models.ExecuteRequestInput) (models.ExecuteRequestResult, error) {
	state, err := a.store.LoadState()
	if err != nil {
		return models.ExecuteRequestResult{}, err
	}
	environmentValues := map[string]string{}
	for _, env := range state.Environments {
		if env.ID == input.EnvironmentID {
			a.decryptEnvironment(&env)
			for _, v := range env.Variables {
				if v.Enabled {
					environmentValues[v.Key] = v.CurrentValue
				}
			}
		}
	}
	values := variables.Merge(input.GlobalVariables, environmentValues, input.CollectionVariables, input.Request.Variables)
	if _, err := httpclient.EnsureOAuthToken(a.ctx, &input.Request.Auth, values, false); err != nil {
		return models.ExecuteRequestResult{}, fmt.Errorf("não foi possível obter o token OAuth 2.0 automaticamente: %w", err)
	}
	req, err := httpclient.BuildRequest(a.ctx, input.Request, values)
	if err != nil {
		return models.ExecuteRequestResult{}, err
	}
	result, err := a.engine.Execute(a.ctx, input.ExecutionID, req, input.Request, state.Settings.MaxResponsePreviewSize)
	if err != nil {
		return result, err
	}
	if strings.EqualFold(input.Request.Protocol, "SOAP") || strings.Contains(strings.ToLower(result.ContentType), "xml") {
		result.SOAPFault = soap.DetectFault(result.Body)
	}
	historyResult := result
	if !state.Settings.StoreHistoryBodies {
		historyResult.Body = ""
		historyResult.BodyBase64 = ""
	}
	historyRequest := input.Request
	if !state.Settings.StoreHistoryBodies {
		historyRequest.Body = ""
		historyRequest.Multipart = nil
	}
	if encryptErr := a.encryptRequest(&historyRequest); encryptErr != nil {
		return result, encryptErr
	}
	entry := models.HistoryEntry{ID: uuid.NewString(), WorkspaceID: input.Request.WorkspaceID, ExecutedAt: time.Now().UTC(), Method: input.Request.Method, URL: result.ResolvedURL, Name: input.Request.Name, StatusCode: result.StatusCode, DurationMs: result.DurationMs, Size: result.Size, Request: historyRequest, Response: historyResult}
	if saveErr := a.store.SaveHistory(entry); saveErr != nil {
		return result, fmt.Errorf("resposta recebida, mas o histórico não pôde ser salvo: %w", saveErr)
	}
	_ = a.store.PruneHistory(input.Request.WorkspaceID, state.Settings.HistoryRetentionDays)
	return result, nil
}
func (a *App) CancelRequest(executionID string) bool { return a.engine.Cancel(executionID) }

func (a *App) RunLoadTest(input models.LoadTestInput) (models.LoadTestRun, error) {
	if input.Concurrency < 1 || input.Concurrency > 200 {
		return models.LoadTestRun{}, fmt.Errorf("concorrência deve estar entre 1 e 200")
	}
	if input.TotalRequests < 1 || input.TotalRequests > 100000 {
		return models.LoadTestRun{}, fmt.Errorf("total de requests deve estar entre 1 e 100000")
	}
	if input.DelayMs < 0 || input.DelayMs > 60000 {
		return models.LoadTestRun{}, fmt.Errorf("intervalo deve estar entre 0 e 60000 ms")
	}
	ctx, cancel := context.WithCancel(a.ctx)
	a.loadMu.Lock()
	a.loadStops[input.ExecutionID] = cancel
	a.loadMu.Unlock()
	defer func() { cancel(); a.loadMu.Lock(); delete(a.loadStops, input.ExecutionID); a.loadMu.Unlock() }()
	state, err := a.store.LoadState()
	if err != nil {
		return models.LoadTestRun{}, err
	}
	environmentValues := map[string]string{}
	for _, environment := range state.Environments {
		if environment.ID == input.EnvironmentID {
			a.decryptEnvironment(&environment)
			for _, variable := range environment.Variables {
				if variable.Enabled {
					environmentValues[variable.Key] = variable.CurrentValue
				}
			}
		}
	}
	values := variables.Merge(input.GlobalVariables, environmentValues, input.CollectionVariables, input.Request.Variables)
	if _, err := httpclient.EnsureOAuthToken(ctx, &input.Request.Auth, values, false); err != nil {
		return models.LoadTestRun{}, fmt.Errorf("não foi possível obter o token OAuth 2.0 automaticamente: %w", err)
	}
	started := time.Now().UTC()
	jobs := make(chan int)
	samples := make(chan models.LoadTestSample, input.TotalRequests)
	var workers sync.WaitGroup
	workerCount := input.Concurrency
	if workerCount > input.TotalRequests {
		workerCount = input.TotalRequests
	}
	for worker := 0; worker < workerCount; worker++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for index := range jobs {
				select {
				case <-ctx.Done():
					return
				default:
				}
				sample := models.LoadTestSample{Index: index, StartedAt: time.Now().UTC().Format(time.RFC3339Nano)}
				request, buildErr := httpclient.BuildRequest(ctx, input.Request, values)
				if buildErr != nil {
					sample.Error = security.SanitizeText(buildErr.Error())
				} else {
					result, executeErr := a.engine.Execute(ctx, fmt.Sprintf("%s-%d", input.ExecutionID, index), request, input.Request, 64*1024)
					sample.StatusCode = result.StatusCode
					sample.DurationMs = result.DurationMs
					sample.Size = result.Size
					if executeErr != nil {
						sample.Error = security.SanitizeText(executeErr.Error())
					}
				}
				samples <- sample
				if input.DelayMs > 0 {
					timer := time.NewTimer(time.Duration(input.DelayMs) * time.Millisecond)
					select {
					case <-ctx.Done():
						timer.Stop()
						return
					case <-timer.C:
					}
				}
			}
		}()
	}
	go func() {
		defer close(jobs)
		for index := 1; index <= input.TotalRequests; index++ {
			select {
			case <-ctx.Done():
				return
			case jobs <- index:
			}
		}
	}()
	go func() { workers.Wait(); close(samples) }()
	result := models.LoadTestResult{Samples: []models.LoadTestSample{}}
	durations := []int64{}
	for sample := range samples {
		result.Samples = append(result.Samples, sample)
		if sample.Error == "" && sample.StatusCode >= 200 && sample.StatusCode < 400 {
			result.Successful++
		} else {
			result.Failed++
		}
		if sample.DurationMs >= 0 {
			durations = append(durations, sample.DurationMs)
		}
	}
	result.Total = len(result.Samples)
	result.DurationMs = time.Since(started).Milliseconds()
	if result.DurationMs > 0 {
		result.RequestsPerSec = float64(result.Total) / (float64(result.DurationMs) / 1000)
	}
	sort.Slice(result.Samples, func(i, j int) bool { return result.Samples[i].Index < result.Samples[j].Index })
	sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
	if len(durations) > 0 {
		var sum int64
		for _, duration := range durations {
			sum += duration
		}
		result.AverageMs = float64(sum) / float64(len(durations))
		result.MinMs = durations[0]
		result.MaxMs = durations[len(durations)-1]
		result.P50Ms = percentile(durations, 50)
		result.P95Ms = percentile(durations, 95)
		result.P99Ms = percentile(durations, 99)
	}
	run := models.LoadTestRun{ID: uuid.NewString(), WorkspaceID: input.Request.WorkspaceID, RequestID: input.Request.ID, RequestName: input.Request.Name, StartedAt: started, Concurrency: input.Concurrency, TotalRequests: input.TotalRequests, DelayMs: input.DelayMs, Result: result}
	if err := a.store.SaveLoadTest(run); err != nil {
		return run, fmt.Errorf("stress test concluído, mas o histórico não pôde ser salvo: %w", err)
	}
	return run, nil
}

func percentile(sorted []int64, value int) int64 {
	if len(sorted) == 0 {
		return 0
	}
	index := (value*len(sorted)+99)/100 - 1
	if index < 0 {
		index = 0
	}
	if index >= len(sorted) {
		index = len(sorted) - 1
	}
	return sorted[index]
}

func (a *App) CancelLoadTest(executionID string) bool {
	a.loadMu.Lock()
	cancel, ok := a.loadStops[executionID]
	a.loadMu.Unlock()
	if ok {
		cancel()
	}
	return ok
}
func (a *App) ImportCurl(command, workspaceID string) (models.RequestDefinition, error) {
	r, err := curl.Import(command)
	if err != nil {
		return r, err
	}
	r.WorkspaceID = workspaceID
	r.CreatedAt = time.Now().UTC()
	r.UpdatedAt = r.CreatedAt
	return r, nil
}
func (a *App) ExportCurl(r models.RequestDefinition) string { return curl.Export(r) }
func (a *App) ImportWSDL(input models.WSDLImportInput) (models.WSDLImportResult, error) {
	return wsdl.Import(input.Source, input.FromURL)
}

func (a *App) PickFile(title string) (string, error) {
	return wruntime.OpenFileDialog(a.ctx, wruntime.OpenDialogOptions{Title: title})
}
func (a *App) SaveResponse(result models.ExecuteRequestResult) error {
	path, err := wruntime.SaveFileDialog(a.ctx, wruntime.SaveDialogOptions{Title: "Save response", DefaultFilename: "response.bin"})
	if err != nil || path == "" {
		return err
	}
	var body []byte
	if result.Binary {
		body, err = base64.StdEncoding.DecodeString(result.BodyBase64)
	} else {
		body = []byte(result.Body)
	}
	if err != nil {
		return err
	}
	return os.WriteFile(path, body, 0600)
}

func (a *App) WindowMinimise() { wruntime.WindowMinimise(a.ctx) }
func (a *App) WindowToggleMaximise() {
	if wruntime.WindowIsMaximised(a.ctx) {
		wruntime.WindowUnmaximise(a.ctx)
	} else {
		wruntime.WindowMaximise(a.ctx)
	}
}
func (a *App) WindowClose() { wruntime.Quit(a.ctx) }

func (a *App) encryptEnvironment(env *models.Environment) error {
	for i := range env.Variables {
		if !env.Variables[i].Secret {
			continue
		}
		var err error
		env.Variables[i].InitialValue, err = a.protector.Encrypt(env.Variables[i].InitialValue)
		if err != nil {
			return err
		}
		env.Variables[i].CurrentValue, err = a.protector.Encrypt(env.Variables[i].CurrentValue)
		if err != nil {
			return err
		}
	}
	return nil
}
func (a *App) decryptEnvironment(env *models.Environment) {
	for i := range env.Variables {
		if !env.Variables[i].Secret {
			continue
		}
		if value, err := a.protector.Decrypt(env.Variables[i].InitialValue); err == nil {
			env.Variables[i].InitialValue = value
		}
		if value, err := a.protector.Decrypt(env.Variables[i].CurrentValue); err == nil {
			env.Variables[i].CurrentValue = value
		}
	}
}

func (a *App) encryptRequest(request *models.RequestDefinition) error {
	values := []*string{&request.Auth.Password, &request.Auth.Token, &request.Auth.Value, &request.Auth.ClientSecret, &request.Auth.OAuthPassword, &request.Auth.AccessToken, &request.Auth.RefreshToken}
	for _, value := range values {
		encrypted, err := a.protector.Encrypt(*value)
		if err != nil {
			return err
		}
		*value = encrypted
	}
	for i := range request.Headers {
		if security.IsSensitive(request.Headers[i].Key) {
			encrypted, err := a.protector.Encrypt(request.Headers[i].Value)
			if err != nil {
				return err
			}
			request.Headers[i].Value = encrypted
		}
	}
	return nil
}

func (a *App) decryptRequest(request *models.RequestDefinition) {
	values := []*string{&request.Auth.Password, &request.Auth.Token, &request.Auth.Value, &request.Auth.ClientSecret, &request.Auth.OAuthPassword, &request.Auth.AccessToken, &request.Auth.RefreshToken}
	for _, value := range values {
		if plain, err := a.protector.Decrypt(*value); err == nil {
			*value = plain
		}
	}
	for i := range request.Headers {
		if security.IsSensitive(request.Headers[i].Key) {
			if plain, err := a.protector.Decrypt(request.Headers[i].Value); err == nil {
				request.Headers[i].Value = plain
			}
		}
	}
}
