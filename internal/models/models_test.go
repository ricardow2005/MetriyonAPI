package models

import (
	"encoding/json"
	"testing"
)

func TestFlowAcceptsFractionalNodeCoordinates(t *testing.T) {
	payload := []byte(`{
		"id":"flow-1",
		"workspaceId":"workspace-1",
		"name":"LangBench",
		"nodes":[{
			"id":"node-1",
			"requestId":"request-1",
			"name":"RUST-DeleteAllClients",
			"x":89.999993896484375,
			"y":274.50000762939453,
			"mappings":[]
		}],
		"edges":[]
	}`)

	var flow Flow
	if err := json.Unmarshal(payload, &flow); err != nil {
		t.Fatalf("flow with fractional node coordinates should decode: %v", err)
	}
	if len(flow.Nodes) != 1 {
		t.Fatalf("expected one flow node, got %d", len(flow.Nodes))
	}
	if flow.Nodes[0].X != 89.999993896484375 || flow.Nodes[0].Y != 274.50000762939453 {
		t.Fatalf("fractional coordinates were not preserved: x=%v y=%v", flow.Nodes[0].X, flow.Nodes[0].Y)
	}
}

func TestFlowDecodesFunctionNodeAndBranchHandle(t *testing.T) {
	payload := []byte(`{
		"id":"flow-functions",
		"workspaceId":"workspace-1",
		"name":"Conditional",
		"nodes":[{
			"id":"if-1",
			"type":"if",
			"name":"If",
			"x":320.5,
			"y":140.25,
			"mappings":[],
			"conditions":[{"left":"status","operator":"equals","right":"200"}],
			"match":"all",
			"convertTypes":true
		}],
		"edges":[{"id":"edge-1","sourceId":"if-1","targetId":"request-2","sourceHandle":"true"}]
	}`)

	var flow Flow
	if err := json.Unmarshal(payload, &flow); err != nil {
		t.Fatalf("flow function should decode: %v", err)
	}
	if len(flow.Nodes) != 1 || flow.Nodes[0].Type != "if" || len(flow.Nodes[0].Conditions) != 1 {
		t.Fatalf("function node fields were not preserved: %#v", flow.Nodes)
	}
	if len(flow.Edges) != 1 || flow.Edges[0].SourceHandle != "true" {
		t.Fatalf("branch handle was not preserved: %#v", flow.Edges)
	}
}
