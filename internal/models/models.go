package models

import "time"

type Workspace struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Collection struct {
	ID          string            `json:"id"`
	WorkspaceID string            `json:"workspaceId"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Variables   map[string]string `json:"variables"`
	SortOrder   int               `json:"sortOrder"`
}

type Folder struct {
	ID           string `json:"id"`
	CollectionID string `json:"collectionId"`
	ParentID     string `json:"parentId"`
	Name         string `json:"name"`
	SortOrder    int    `json:"sortOrder"`
}

type KeyValue struct {
	ID          string `json:"id"`
	Enabled     bool   `json:"enabled"`
	Key         string `json:"key"`
	Value       string `json:"value"`
	Description string `json:"description"`
}

type AuthDefinition struct {
	Type          string `json:"type"`
	Username      string `json:"username"`
	Password      string `json:"password"`
	Token         string `json:"token"`
	Key           string `json:"key"`
	Value         string `json:"value"`
	AddTo         string `json:"addTo"`
	OAuthFlow     string `json:"oauthFlow"`
	TokenURL      string `json:"tokenUrl"`
	ClientID      string `json:"clientId"`
	ClientSecret  string `json:"clientSecret"`
	ClientAuth    string `json:"clientAuth"`
	Scope         string `json:"scope"`
	OAuthUsername string `json:"oauthUsername"`
	OAuthPassword string `json:"oauthPassword"`
	AccessToken   string `json:"accessToken"`
	RefreshToken  string `json:"refreshToken"`
	ExpiresAt     string `json:"expiresAt"`
}

type OAuthTokenInput struct {
	Auth                AuthDefinition    `json:"auth"`
	EnvironmentID       string            `json:"environmentId"`
	GlobalVariables     map[string]string `json:"globalVariables"`
	CollectionVariables map[string]string `json:"collectionVariables"`
	RequestVariables    map[string]string `json:"requestVariables"`
}

type OAuthTokenResult struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	TokenType    string `json:"tokenType"`
	ExpiresAt    string `json:"expiresAt"`
	Scope        string `json:"scope"`
}

type MultipartPart struct {
	ID      string `json:"id"`
	Enabled bool   `json:"enabled"`
	Type    string `json:"type"`
	Key     string `json:"key"`
	Value   string `json:"value"`
}

type RequestDefinition struct {
	ID             string            `json:"id"`
	WorkspaceID    string            `json:"workspaceId"`
	CollectionID   string            `json:"collectionId"`
	FolderID       string            `json:"folderId"`
	Name           string            `json:"name"`
	Description    string            `json:"description"`
	Protocol       string            `json:"protocol"`
	Method         string            `json:"method"`
	URL            string            `json:"url"`
	Params         []KeyValue        `json:"params"`
	Headers        []KeyValue        `json:"headers"`
	Auth           AuthDefinition    `json:"auth"`
	BodyType       string            `json:"bodyType"`
	Body           string            `json:"body"`
	Multipart      []MultipartPart   `json:"multipart"`
	Variables      map[string]string `json:"variables"`
	SOAPVersion    string            `json:"soapVersion"`
	SOAPAction     string            `json:"soapAction"`
	WSDLService    string            `json:"wsdlService"`
	WSDLOperation  string            `json:"wsdlOperation"`
	VerifySSL      bool              `json:"verifySsl"`
	FollowRedirect bool              `json:"followRedirects"`
	TimeoutSeconds int               `json:"timeoutSeconds"`
	CreatedAt      time.Time         `json:"createdAt"`
	UpdatedAt      time.Time         `json:"updatedAt"`
}

type EnvironmentVariable struct {
	ID           string `json:"id"`
	Key          string `json:"key"`
	InitialValue string `json:"initialValue"`
	CurrentValue string `json:"currentValue"`
	Secret       bool   `json:"secret"`
	Enabled      bool   `json:"enabled"`
}

type Environment struct {
	ID          string                `json:"id"`
	WorkspaceID string                `json:"workspaceId"`
	Name        string                `json:"name"`
	Variables   []EnvironmentVariable `json:"variables"`
}

type Timing struct {
	DNS      int64 `json:"dns"`
	Connect  int64 `json:"connect"`
	TLS      int64 `json:"tls"`
	TTFB     int64 `json:"ttfb"`
	Download int64 `json:"download"`
	Total    int64 `json:"total"`
}

type Redirect struct {
	StatusCode int    `json:"statusCode"`
	URL        string `json:"url"`
}

type SOAPFault struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ExecuteRequestInput struct {
	ExecutionID         string            `json:"executionId"`
	Request             RequestDefinition `json:"request"`
	EnvironmentID       string            `json:"environmentId"`
	GlobalVariables     map[string]string `json:"globalVariables"`
	CollectionVariables map[string]string `json:"collectionVariables"`
}

type ExecuteRequestResult struct {
	StatusCode     int                 `json:"statusCode"`
	Status         string              `json:"status"`
	Headers        map[string][]string `json:"headers"`
	Cookies        []string            `json:"cookies"`
	Body           string              `json:"body"`
	BodyBase64     string              `json:"bodyBase64"`
	ContentType    string              `json:"contentType"`
	DurationMs     int64               `json:"durationMs"`
	Size           int64               `json:"size"`
	Truncated      bool                `json:"truncated"`
	Binary         bool                `json:"binary"`
	Timings        Timing              `json:"timings"`
	Redirects      []Redirect          `json:"redirects"`
	SOAPFault      *SOAPFault          `json:"soapFault,omitempty"`
	ResolvedURL    string              `json:"resolvedUrl"`
	TechnicalError string              `json:"technicalError,omitempty"`
}

type MoveFolderInput struct {
	FolderID     string `json:"folderId"`
	CollectionID string `json:"collectionId"`
	SortOrder    int    `json:"sortOrder"`
}

type LoadTestInput struct {
	ExecutionID         string            `json:"executionId"`
	Request             RequestDefinition `json:"request"`
	EnvironmentID       string            `json:"environmentId"`
	GlobalVariables     map[string]string `json:"globalVariables"`
	CollectionVariables map[string]string `json:"collectionVariables"`
	Concurrency         int               `json:"concurrency"`
	TotalRequests       int               `json:"totalRequests"`
	DelayMs             int               `json:"delayMs"`
}

type LoadTestSample struct {
	Index      int    `json:"index"`
	StartedAt  string `json:"startedAt"`
	StatusCode int    `json:"statusCode"`
	DurationMs int64  `json:"durationMs"`
	Size       int64  `json:"size"`
	Error      string `json:"error"`
}

type LoadTestResult struct {
	Total          int              `json:"total"`
	Successful     int              `json:"successful"`
	Failed         int              `json:"failed"`
	DurationMs     int64            `json:"durationMs"`
	RequestsPerSec float64          `json:"requestsPerSec"`
	AverageMs      float64          `json:"averageMs"`
	MinMs          int64            `json:"minMs"`
	MaxMs          int64            `json:"maxMs"`
	P50Ms          int64            `json:"p50Ms"`
	P95Ms          int64            `json:"p95Ms"`
	P99Ms          int64            `json:"p99Ms"`
	Samples        []LoadTestSample `json:"samples"`
}

type LoadTestRun struct {
	ID            string         `json:"id"`
	WorkspaceID   string         `json:"workspaceId"`
	RequestID     string         `json:"requestId"`
	RequestName   string         `json:"requestName"`
	StartedAt     time.Time      `json:"startedAt"`
	Concurrency   int            `json:"concurrency"`
	TotalRequests int            `json:"totalRequests"`
	DelayMs       int            `json:"delayMs"`
	Result        LoadTestResult `json:"result"`
}

type FlowMapping struct {
	Path     string `json:"path"`
	Variable string `json:"variable"`
}

type FlowCondition struct {
	Left     string `json:"left"`
	Operator string `json:"operator"`
	Right    string `json:"right"`
}

type FlowSwitchRule struct {
	Left       string `json:"left"`
	Operator   string `json:"operator"`
	Right      string `json:"right"`
	OutputName string `json:"outputName,omitempty"`
}

type FlowNode struct {
	ID               string           `json:"id"`
	Type             string           `json:"type,omitempty"`
	RequestID        string           `json:"requestId,omitempty"`
	Name             string           `json:"name"`
	X                float64          `json:"x"`
	Y                float64          `json:"y"`
	Mappings         []FlowMapping    `json:"mappings"`
	Conditions       []FlowCondition  `json:"conditions,omitempty"`
	Match            string           `json:"match,omitempty"`
	ConvertTypes     bool             `json:"convertTypes,omitempty"`
	WaitValue        float64          `json:"waitValue,omitempty"`
	WaitUnit         string           `json:"waitUnit,omitempty"`
	LoopCount        int              `json:"loopCount,omitempty"`
	SwitchMode       string           `json:"switchMode,omitempty"`
	SwitchRules      []FlowSwitchRule `json:"switchRules,omitempty"`
	SwitchExpression string           `json:"switchExpression,omitempty"`
	SwitchOutputs    int              `json:"switchOutputs,omitempty"`
	ParallelBranches int              `json:"parallelBranches,omitempty"`
}

type FlowEdge struct {
	ID           string `json:"id"`
	SourceID     string `json:"sourceId"`
	TargetID     string `json:"targetId"`
	SourceHandle string `json:"sourceHandle,omitempty"`
}

type Flow struct {
	ID          string     `json:"id"`
	WorkspaceID string     `json:"workspaceId"`
	Name        string     `json:"name"`
	Nodes       []FlowNode `json:"nodes"`
	Edges       []FlowEdge `json:"edges"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

type FlowRunInput struct {
	ExecutionID   string            `json:"executionId"`
	Flow          Flow              `json:"flow"`
	EnvironmentID string            `json:"environmentId"`
	Variables     map[string]string `json:"variables"`
}

type FlowNodeResult struct {
	NodeID    string               `json:"nodeId"`
	Type      string               `json:"type,omitempty"`
	RequestID string               `json:"requestId,omitempty"`
	Name      string               `json:"name"`
	StartedAt time.Time            `json:"startedAt"`
	Response  ExecuteRequestResult `json:"response"`
	Extracted map[string]string    `json:"extracted"`
	Outcome   string               `json:"outcome,omitempty"`
	Skipped   bool                 `json:"skipped,omitempty"`
	Iteration int                  `json:"iteration,omitempty"`
	Error     string               `json:"error,omitempty"`
}

type FlowRunResult struct {
	ID         string            `json:"id"`
	FlowID     string            `json:"flowId"`
	StartedAt  time.Time         `json:"startedAt"`
	DurationMs int64             `json:"durationMs"`
	Successful bool              `json:"successful"`
	Variables  map[string]string `json:"variables"`
	Nodes      []FlowNodeResult  `json:"nodes"`
}

type HistoryEntry struct {
	ID          string               `json:"id"`
	WorkspaceID string               `json:"workspaceId"`
	ExecutedAt  time.Time            `json:"executedAt"`
	Method      string               `json:"method"`
	URL         string               `json:"url"`
	Name        string               `json:"name"`
	StatusCode  int                  `json:"statusCode"`
	DurationMs  int64                `json:"durationMs"`
	Size        int64                `json:"size"`
	Request     RequestDefinition    `json:"request"`
	Response    ExecuteRequestResult `json:"response"`
}

type Settings struct {
	Theme                  string   `json:"theme"`
	ActiveWorkspaceID      string   `json:"activeWorkspaceId"`
	ActiveEnvironmentID    string   `json:"activeEnvironmentId"`
	StoreHistoryBodies     bool     `json:"storeHistoryBodies"`
	HistoryRetentionDays   int      `json:"historyRetentionDays"`
	MaxResponsePreviewSize int64    `json:"maxResponsePreviewSize"`
	SidebarWidth           int      `json:"sidebarWidth"`
	EditorSplitPercent     int      `json:"editorSplitPercent"`
	LogLevel               string   `json:"logLevel"`
	WindowWidth            int      `json:"windowWidth"`
	WindowHeight           int      `json:"windowHeight"`
	WindowX                int      `json:"windowX"`
	WindowY                int      `json:"windowY"`
	CollapsedTreeIDs       []string `json:"collapsedTreeIds"`
	ResponsePaneCollapsed  bool     `json:"responsePaneCollapsed"`
}

type AppState struct {
	Workspaces   []Workspace         `json:"workspaces"`
	Collections  []Collection        `json:"collections"`
	Folders      []Folder            `json:"folders"`
	Requests     []RequestDefinition `json:"requests"`
	Environments []Environment       `json:"environments"`
	History      []HistoryEntry      `json:"history"`
	LoadTests    []LoadTestRun       `json:"loadTests"`
	Flows        []Flow              `json:"flows"`
	OpenTabs     []string            `json:"openTabs"`
	Settings     Settings            `json:"settings"`
	Version      string              `json:"version"`
	Commit       string              `json:"commit"`
	BuildDate    string              `json:"buildDate"`
	GoVersion    string              `json:"goVersion"`
}

type CollectionPackage struct {
	Format     string              `json:"format"`
	Version    int                 `json:"version"`
	ExportedAt time.Time           `json:"exportedAt"`
	Collection Collection          `json:"collection"`
	Folders    []Folder            `json:"folders"`
	Requests   []RequestDefinition `json:"requests"`
}

type CollectionImportResult struct {
	Collection Collection          `json:"collection"`
	Folders    []Folder            `json:"folders"`
	Requests   []RequestDefinition `json:"requests"`
}

type WSDLImportInput struct {
	WorkspaceID string `json:"workspaceId"`
	Source      string `json:"source"`
	FromURL     bool   `json:"fromUrl"`
}

type WSDLImportResult struct {
	Name       string          `json:"name"`
	Endpoint   string          `json:"endpoint"`
	Operations []WSDLOperation `json:"operations"`
}

type WSDLOperation struct {
	Name       string `json:"name"`
	SOAPAction string `json:"soapAction"`
	Envelope   string `json:"envelope"`
}
