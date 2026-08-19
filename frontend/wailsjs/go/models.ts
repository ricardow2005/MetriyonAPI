export namespace models {
	
	export class Settings {
	    theme: string;
	    activeWorkspaceId: string;
	    activeEnvironmentId: string;
	    storeHistoryBodies: boolean;
	    historyRetentionDays: number;
	    maxResponsePreviewSize: number;
	    sidebarWidth: number;
	    editorSplitPercent: number;
	    logLevel: string;
	    windowWidth: number;
	    windowHeight: number;
	    windowX: number;
	    windowY: number;
	    collapsedTreeIds: string[];
	    responsePaneCollapsed: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.theme = source["theme"];
	        this.activeWorkspaceId = source["activeWorkspaceId"];
	        this.activeEnvironmentId = source["activeEnvironmentId"];
	        this.storeHistoryBodies = source["storeHistoryBodies"];
	        this.historyRetentionDays = source["historyRetentionDays"];
	        this.maxResponsePreviewSize = source["maxResponsePreviewSize"];
	        this.sidebarWidth = source["sidebarWidth"];
	        this.editorSplitPercent = source["editorSplitPercent"];
	        this.logLevel = source["logLevel"];
	        this.windowWidth = source["windowWidth"];
	        this.windowHeight = source["windowHeight"];
	        this.windowX = source["windowX"];
	        this.windowY = source["windowY"];
	        this.collapsedTreeIds = source["collapsedTreeIds"];
	        this.responsePaneCollapsed = source["responsePaneCollapsed"];
	    }
	}
	export class FlowEdge {
	    id: string;
	    sourceId: string;
	    targetId: string;
	    sourceHandle?: string;
	
	    static createFrom(source: any = {}) {
	        return new FlowEdge(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.sourceId = source["sourceId"];
	        this.targetId = source["targetId"];
	        this.sourceHandle = source["sourceHandle"];
	    }
	}
	export class FlowCondition {
	    left: string;
	    operator: string;
	    right: string;
	
	    static createFrom(source: any = {}) {
	        return new FlowCondition(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.left = source["left"];
	        this.operator = source["operator"];
	        this.right = source["right"];
	    }
	}
	export class FlowMapping {
	    path: string;
	    variable: string;
	
	    static createFrom(source: any = {}) {
	        return new FlowMapping(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.variable = source["variable"];
	    }
	}
	export class FlowSwitchRule {
	    left: string;
	    operator: string;
	    right: string;
	    outputName?: string;

	    static createFrom(source: any = {}) { return new FlowSwitchRule(source); }
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.left = source["left"];
	        this.operator = source["operator"];
	        this.right = source["right"];
	        this.outputName = source["outputName"];
	    }
	}
	export class FlowNode {
	    id: string;
	    type?: string;
	    requestId?: string;
	    name: string;
	    x: number;
	    y: number;
	    mappings: FlowMapping[];
	    conditions?: FlowCondition[];
	    match?: string;
	    convertTypes?: boolean;
	    waitValue?: number;
	    waitUnit?: string;
	    loopCount?: number;
	    switchMode?: string;
	    switchRules?: FlowSwitchRule[];
	    switchExpression?: string;
	    switchOutputs?: number;
	    parallelBranches?: number;
	
	    static createFrom(source: any = {}) {
	        return new FlowNode(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.type = source["type"];
	        this.requestId = source["requestId"];
	        this.name = source["name"];
	        this.x = source["x"];
	        this.y = source["y"];
	        this.mappings = this.convertValues(source["mappings"], FlowMapping);
	        this.conditions = this.convertValues(source["conditions"], FlowCondition);
	        this.match = source["match"];
	        this.convertTypes = source["convertTypes"];
	        this.waitValue = source["waitValue"];
	        this.waitUnit = source["waitUnit"];
	        this.loopCount = source["loopCount"];
	        this.switchMode = source["switchMode"];
	        this.switchRules = this.convertValues(source["switchRules"], FlowSwitchRule);
	        this.switchExpression = source["switchExpression"];
	        this.switchOutputs = source["switchOutputs"];
	        this.parallelBranches = source["parallelBranches"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Flow {
	    id: string;
	    workspaceId: string;
	    name: string;
	    nodes: FlowNode[];
	    edges: FlowEdge[];
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new Flow(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.workspaceId = source["workspaceId"];
	        this.name = source["name"];
	        this.nodes = this.convertValues(source["nodes"], FlowNode);
	        this.edges = this.convertValues(source["edges"], FlowEdge);
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class LoadTestSample {
	    index: number;
	    startedAt: string;
	    statusCode: number;
	    durationMs: number;
	    size: number;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new LoadTestSample(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.index = source["index"];
	        this.startedAt = source["startedAt"];
	        this.statusCode = source["statusCode"];
	        this.durationMs = source["durationMs"];
	        this.size = source["size"];
	        this.error = source["error"];
	    }
	}
	export class LoadTestResult {
	    total: number;
	    successful: number;
	    failed: number;
	    durationMs: number;
	    requestsPerSec: number;
	    averageMs: number;
	    minMs: number;
	    maxMs: number;
	    p50Ms: number;
	    p95Ms: number;
	    p99Ms: number;
	    samples: LoadTestSample[];
	
	    static createFrom(source: any = {}) {
	        return new LoadTestResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total = source["total"];
	        this.successful = source["successful"];
	        this.failed = source["failed"];
	        this.durationMs = source["durationMs"];
	        this.requestsPerSec = source["requestsPerSec"];
	        this.averageMs = source["averageMs"];
	        this.minMs = source["minMs"];
	        this.maxMs = source["maxMs"];
	        this.p50Ms = source["p50Ms"];
	        this.p95Ms = source["p95Ms"];
	        this.p99Ms = source["p99Ms"];
	        this.samples = this.convertValues(source["samples"], LoadTestSample);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class LoadTestRun {
	    id: string;
	    workspaceId: string;
	    requestId: string;
	    requestName: string;
	    // Go type: time
	    startedAt: any;
	    concurrency: number;
	    totalRequests: number;
	    delayMs: number;
	    result: LoadTestResult;
	
	    static createFrom(source: any = {}) {
	        return new LoadTestRun(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.workspaceId = source["workspaceId"];
	        this.requestId = source["requestId"];
	        this.requestName = source["requestName"];
	        this.startedAt = this.convertValues(source["startedAt"], null);
	        this.concurrency = source["concurrency"];
	        this.totalRequests = source["totalRequests"];
	        this.delayMs = source["delayMs"];
	        this.result = this.convertValues(source["result"], LoadTestResult);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SOAPFault {
	    code: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new SOAPFault(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.code = source["code"];
	        this.message = source["message"];
	    }
	}
	export class Redirect {
	    statusCode: number;
	    url: string;
	
	    static createFrom(source: any = {}) {
	        return new Redirect(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.statusCode = source["statusCode"];
	        this.url = source["url"];
	    }
	}
	export class Timing {
	    dns: number;
	    connect: number;
	    tls: number;
	    ttfb: number;
	    download: number;
	    total: number;
	
	    static createFrom(source: any = {}) {
	        return new Timing(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.dns = source["dns"];
	        this.connect = source["connect"];
	        this.tls = source["tls"];
	        this.ttfb = source["ttfb"];
	        this.download = source["download"];
	        this.total = source["total"];
	    }
	}
	export class ExecuteRequestResult {
	    statusCode: number;
	    status: string;
	    headers: Record<string, Array<string>>;
	    cookies: string[];
	    body: string;
	    bodyBase64: string;
	    contentType: string;
	    durationMs: number;
	    size: number;
	    truncated: boolean;
	    binary: boolean;
	    timings: Timing;
	    redirects: Redirect[];
	    soapFault?: SOAPFault;
	    resolvedUrl: string;
	    technicalError?: string;
	
	    static createFrom(source: any = {}) {
	        return new ExecuteRequestResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.statusCode = source["statusCode"];
	        this.status = source["status"];
	        this.headers = source["headers"];
	        this.cookies = source["cookies"];
	        this.body = source["body"];
	        this.bodyBase64 = source["bodyBase64"];
	        this.contentType = source["contentType"];
	        this.durationMs = source["durationMs"];
	        this.size = source["size"];
	        this.truncated = source["truncated"];
	        this.binary = source["binary"];
	        this.timings = this.convertValues(source["timings"], Timing);
	        this.redirects = this.convertValues(source["redirects"], Redirect);
	        this.soapFault = this.convertValues(source["soapFault"], SOAPFault);
	        this.resolvedUrl = source["resolvedUrl"];
	        this.technicalError = source["technicalError"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class HistoryEntry {
	    id: string;
	    workspaceId: string;
	    // Go type: time
	    executedAt: any;
	    method: string;
	    url: string;
	    name: string;
	    statusCode: number;
	    durationMs: number;
	    size: number;
	    request: RequestDefinition;
	    response: ExecuteRequestResult;
	
	    static createFrom(source: any = {}) {
	        return new HistoryEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.workspaceId = source["workspaceId"];
	        this.executedAt = this.convertValues(source["executedAt"], null);
	        this.method = source["method"];
	        this.url = source["url"];
	        this.name = source["name"];
	        this.statusCode = source["statusCode"];
	        this.durationMs = source["durationMs"];
	        this.size = source["size"];
	        this.request = this.convertValues(source["request"], RequestDefinition);
	        this.response = this.convertValues(source["response"], ExecuteRequestResult);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class EnvironmentVariable {
	    id: string;
	    key: string;
	    initialValue: string;
	    currentValue: string;
	    secret: boolean;
	    enabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new EnvironmentVariable(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.key = source["key"];
	        this.initialValue = source["initialValue"];
	        this.currentValue = source["currentValue"];
	        this.secret = source["secret"];
	        this.enabled = source["enabled"];
	    }
	}
	export class Environment {
	    id: string;
	    workspaceId: string;
	    name: string;
	    variables: EnvironmentVariable[];
	
	    static createFrom(source: any = {}) {
	        return new Environment(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.workspaceId = source["workspaceId"];
	        this.name = source["name"];
	        this.variables = this.convertValues(source["variables"], EnvironmentVariable);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class MultipartPart {
	    id: string;
	    enabled: boolean;
	    type: string;
	    key: string;
	    value: string;
	
	    static createFrom(source: any = {}) {
	        return new MultipartPart(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.enabled = source["enabled"];
	        this.type = source["type"];
	        this.key = source["key"];
	        this.value = source["value"];
	    }
	}
	export class AuthDefinition {
	    type: string;
	    username: string;
	    password: string;
	    token: string;
	    key: string;
	    value: string;
	    addTo: string;
	    oauthFlow: string;
	    tokenUrl: string;
	    clientId: string;
	    clientSecret: string;
	    clientAuth: string;
	    scope: string;
	    oauthUsername: string;
	    oauthPassword: string;
	    accessToken: string;
	    refreshToken: string;
	    expiresAt: string;
	
	    static createFrom(source: any = {}) {
	        return new AuthDefinition(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.type = source["type"];
	        this.username = source["username"];
	        this.password = source["password"];
	        this.token = source["token"];
	        this.key = source["key"];
	        this.value = source["value"];
	        this.addTo = source["addTo"];
	        this.oauthFlow = source["oauthFlow"];
	        this.tokenUrl = source["tokenUrl"];
	        this.clientId = source["clientId"];
	        this.clientSecret = source["clientSecret"];
	        this.clientAuth = source["clientAuth"];
	        this.scope = source["scope"];
	        this.oauthUsername = source["oauthUsername"];
	        this.oauthPassword = source["oauthPassword"];
	        this.accessToken = source["accessToken"];
	        this.refreshToken = source["refreshToken"];
	        this.expiresAt = source["expiresAt"];
	    }
	}
	export class KeyValue {
	    id: string;
	    enabled: boolean;
	    key: string;
	    value: string;
	    description: string;
	
	    static createFrom(source: any = {}) {
	        return new KeyValue(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.enabled = source["enabled"];
	        this.key = source["key"];
	        this.value = source["value"];
	        this.description = source["description"];
	    }
	}
	export class RequestDefinition {
	    id: string;
	    workspaceId: string;
	    collectionId: string;
	    folderId: string;
	    name: string;
	    description: string;
	    protocol: string;
	    method: string;
	    url: string;
	    params: KeyValue[];
	    headers: KeyValue[];
	    auth: AuthDefinition;
	    bodyType: string;
	    body: string;
	    multipart: MultipartPart[];
	    variables: Record<string, string>;
	    soapVersion: string;
	    soapAction: string;
	    wsdlService: string;
	    wsdlOperation: string;
	    verifySsl: boolean;
	    followRedirects: boolean;
	    timeoutSeconds: number;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new RequestDefinition(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.workspaceId = source["workspaceId"];
	        this.collectionId = source["collectionId"];
	        this.folderId = source["folderId"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.protocol = source["protocol"];
	        this.method = source["method"];
	        this.url = source["url"];
	        this.params = this.convertValues(source["params"], KeyValue);
	        this.headers = this.convertValues(source["headers"], KeyValue);
	        this.auth = this.convertValues(source["auth"], AuthDefinition);
	        this.bodyType = source["bodyType"];
	        this.body = source["body"];
	        this.multipart = this.convertValues(source["multipart"], MultipartPart);
	        this.variables = source["variables"];
	        this.soapVersion = source["soapVersion"];
	        this.soapAction = source["soapAction"];
	        this.wsdlService = source["wsdlService"];
	        this.wsdlOperation = source["wsdlOperation"];
	        this.verifySsl = source["verifySsl"];
	        this.followRedirects = source["followRedirects"];
	        this.timeoutSeconds = source["timeoutSeconds"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Folder {
	    id: string;
	    collectionId: string;
	    parentId: string;
	    name: string;
	    sortOrder: number;
	
	    static createFrom(source: any = {}) {
	        return new Folder(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.collectionId = source["collectionId"];
	        this.parentId = source["parentId"];
	        this.name = source["name"];
	        this.sortOrder = source["sortOrder"];
	    }
	}
	export class Collection {
	    id: string;
	    workspaceId: string;
	    name: string;
	    description: string;
	    variables: Record<string, string>;
	    sortOrder: number;
	
	    static createFrom(source: any = {}) {
	        return new Collection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.workspaceId = source["workspaceId"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.variables = source["variables"];
	        this.sortOrder = source["sortOrder"];
	    }
	}
	export class Workspace {
	    id: string;
	    name: string;
	    // Go type: time
	    createdAt: any;
	    // Go type: time
	    updatedAt: any;
	
	    static createFrom(source: any = {}) {
	        return new Workspace(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.createdAt = this.convertValues(source["createdAt"], null);
	        this.updatedAt = this.convertValues(source["updatedAt"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class AppState {
	    workspaces: Workspace[];
	    collections: Collection[];
	    folders: Folder[];
	    requests: RequestDefinition[];
	    environments: Environment[];
	    history: HistoryEntry[];
	    loadTests: LoadTestRun[];
	    flows: Flow[];
	    openTabs: string[];
	    settings: Settings;
	    version: string;
	    commit: string;
	    buildDate: string;
	    goVersion: string;
	
	    static createFrom(source: any = {}) {
	        return new AppState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.workspaces = this.convertValues(source["workspaces"], Workspace);
	        this.collections = this.convertValues(source["collections"], Collection);
	        this.folders = this.convertValues(source["folders"], Folder);
	        this.requests = this.convertValues(source["requests"], RequestDefinition);
	        this.environments = this.convertValues(source["environments"], Environment);
	        this.history = this.convertValues(source["history"], HistoryEntry);
	        this.loadTests = this.convertValues(source["loadTests"], LoadTestRun);
	        this.flows = this.convertValues(source["flows"], Flow);
	        this.openTabs = source["openTabs"];
	        this.settings = this.convertValues(source["settings"], Settings);
	        this.version = source["version"];
	        this.commit = source["commit"];
	        this.buildDate = source["buildDate"];
	        this.goVersion = source["goVersion"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	
	
	export class ExecuteRequestInput {
	    executionId: string;
	    request: RequestDefinition;
	    environmentId: string;
	    globalVariables: Record<string, string>;
	    collectionVariables: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new ExecuteRequestInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.executionId = source["executionId"];
	        this.request = this.convertValues(source["request"], RequestDefinition);
	        this.environmentId = source["environmentId"];
	        this.globalVariables = source["globalVariables"];
	        this.collectionVariables = source["collectionVariables"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	
	
	
	export class FlowNodeResult {
	    nodeId: string;
	    type?: string;
	    requestId?: string;
	    name: string;
	    // Go type: time
	    startedAt: any;
	    response: ExecuteRequestResult;
	    extracted: Record<string, string>;
	    outcome?: string;
	    skipped?: boolean;
	    iteration?: number;
	    error?: string;
	
	    static createFrom(source: any = {}) {
	        return new FlowNodeResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.nodeId = source["nodeId"];
	        this.type = source["type"];
	        this.requestId = source["requestId"];
	        this.name = source["name"];
	        this.startedAt = this.convertValues(source["startedAt"], null);
	        this.response = this.convertValues(source["response"], ExecuteRequestResult);
	        this.extracted = source["extracted"];
	        this.outcome = source["outcome"];
	        this.skipped = source["skipped"];
	        this.iteration = source["iteration"];
	        this.error = source["error"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class FlowRunInput {
	    executionId: string;
	    flow: Flow;
	    environmentId: string;
	    variables: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new FlowRunInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.executionId = source["executionId"];
	        this.flow = this.convertValues(source["flow"], Flow);
	        this.environmentId = source["environmentId"];
	        this.variables = source["variables"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class FlowRunResult {
	    id: string;
	    flowId: string;
	    // Go type: time
	    startedAt: any;
	    durationMs: number;
	    successful: boolean;
	    variables: Record<string, string>;
	    nodes: FlowNodeResult[];
	
	    static createFrom(source: any = {}) {
	        return new FlowRunResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.flowId = source["flowId"];
	        this.startedAt = this.convertValues(source["startedAt"], null);
	        this.durationMs = source["durationMs"];
	        this.successful = source["successful"];
	        this.variables = source["variables"];
	        this.nodes = this.convertValues(source["nodes"], FlowNodeResult);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	
	export class LoadTestInput {
	    executionId: string;
	    request: RequestDefinition;
	    environmentId: string;
	    globalVariables: Record<string, string>;
	    collectionVariables: Record<string, string>;
	    concurrency: number;
	    totalRequests: number;
	    delayMs: number;
	
	    static createFrom(source: any = {}) {
	        return new LoadTestInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.executionId = source["executionId"];
	        this.request = this.convertValues(source["request"], RequestDefinition);
	        this.environmentId = source["environmentId"];
	        this.globalVariables = source["globalVariables"];
	        this.collectionVariables = source["collectionVariables"];
	        this.concurrency = source["concurrency"];
	        this.totalRequests = source["totalRequests"];
	        this.delayMs = source["delayMs"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	
	
	export class MoveFolderInput {
	    folderId: string;
	    collectionId: string;
	    sortOrder: number;
	
	    static createFrom(source: any = {}) {
	        return new MoveFolderInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.folderId = source["folderId"];
	        this.collectionId = source["collectionId"];
	        this.sortOrder = source["sortOrder"];
	    }
	}
	
	export class OAuthTokenInput {
	    auth: AuthDefinition;
	    environmentId: string;
	    globalVariables: Record<string, string>;
	    collectionVariables: Record<string, string>;
	    requestVariables: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new OAuthTokenInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.auth = this.convertValues(source["auth"], AuthDefinition);
	        this.environmentId = source["environmentId"];
	        this.globalVariables = source["globalVariables"];
	        this.collectionVariables = source["collectionVariables"];
	        this.requestVariables = source["requestVariables"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class OAuthTokenResult {
	    accessToken: string;
	    refreshToken: string;
	    tokenType: string;
	    expiresAt: string;
	    scope: string;
	
	    static createFrom(source: any = {}) {
	        return new OAuthTokenResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.accessToken = source["accessToken"];
	        this.refreshToken = source["refreshToken"];
	        this.tokenType = source["tokenType"];
	        this.expiresAt = source["expiresAt"];
	        this.scope = source["scope"];
	    }
	}
	
	
	
	
	
	export class WSDLImportInput {
	    workspaceId: string;
	    source: string;
	    fromUrl: boolean;
	
	    static createFrom(source: any = {}) {
	        return new WSDLImportInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.workspaceId = source["workspaceId"];
	        this.source = source["source"];
	        this.fromUrl = source["fromUrl"];
	    }
	}
	export class WSDLOperation {
	    name: string;
	    soapAction: string;
	    envelope: string;
	
	    static createFrom(source: any = {}) {
	        return new WSDLOperation(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.soapAction = source["soapAction"];
	        this.envelope = source["envelope"];
	    }
	}
	export class WSDLImportResult {
	    name: string;
	    endpoint: string;
	    operations: WSDLOperation[];
	
	    static createFrom(source: any = {}) {
	        return new WSDLImportResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.endpoint = source["endpoint"];
	        this.operations = this.convertValues(source["operations"], WSDLOperation);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	

}

