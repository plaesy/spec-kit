package reach

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fakeMCPServer makes the fake tool answer one JSON-RPC message with response.
// It returns the path where the request the client sent is recorded.
func fakeMCPServer(t *testing.T, response string, exit int) string {
	t.Helper()
	fakeToolExit(t, response, exit)
	stdinPath := filepath.Join(t.TempDir(), "request.json")
	t.Setenv("REACH_FAKE_STDIN", stdinPath)
	return stdinPath
}

func TestMCPClient_CallReturnsTextContent(t *testing.T) {
	stdin := fakeMCPServer(t,
		`{"jsonrpc":"2.0","id":1,"result":{"content":[{"type":"text","text":"hello"},{"type":"image"},{"type":"text","text":"world"}]}}`, 0)
	client := NewMCPClient("exa", "exa-mcp", "--stdio")
	client.SetEnv("EXA_API_KEY", "placeholder")

	res, err := client.Call(context.Background(), "web_search", map[string]interface{}{"query": "rust"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Text != "hello\nworld\n" {
		t.Errorf("Text = %q, want only the text items, one per line", res.Text)
	}

	raw, err := os.ReadFile(stdin)
	if err != nil {
		t.Fatal(err)
	}
	var req MCPRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		t.Fatalf("the request sent is not JSON: %v (%s)", err, raw)
	}
	if req.Method != "tools/call" || req.JSONRPC != "2.0" {
		t.Errorf("request = %+v, want a JSON-RPC tools/call", req)
	}
}

func TestMCPClient_CallReportsServerErrorMessage(t *testing.T) {
	fakeMCPServer(t, `{"jsonrpc":"2.0","id":1,"error":{"code":-32601,"message":"no such tool"}}`, 0)
	_, err := NewMCPClient("exa", "exa-mcp").Call(context.Background(), "missing", nil)
	if err == nil || !strings.Contains(err.Error(), "no such tool") {
		t.Errorf("err = %v, want the server's error message", err)
	}
}

func TestMCPClient_CallReportsProcessFailureWithStderr(t *testing.T) {
	fakeMCPServer(t, "", 3)
	t.Setenv("REACH_FAKE_ERR", "server crashed")
	_, err := NewMCPClient("exa", "exa-mcp").Call(context.Background(), "x", nil)
	if err == nil || !strings.Contains(err.Error(), "server crashed") {
		t.Errorf("err = %v, want the failing process's stderr", err)
	}
}

func TestMCPClient_CallRejectsMalformedResponse(t *testing.T) {
	fakeMCPServer(t, "this is not json", 0)
	if _, err := NewMCPClient("exa", "exa-mcp").Call(context.Background(), "x", nil); err == nil {
		t.Error("a response that is not JSON must be an error")
	}
}

func TestMCPClient_CallRejectsMalformedResult(t *testing.T) {
	fakeMCPServer(t, `{"jsonrpc":"2.0","id":1,"result":"not an object"}`, 0)
	if _, err := NewMCPClient("exa", "exa-mcp").Call(context.Background(), "x", nil); err == nil {
		t.Error("a result that is not an object must be an error")
	}
}

func TestMCPClient_CallWithMissingServerBinary(t *testing.T) {
	noTools(t)
	if _, err := NewMCPClient("exa", "no-such-mcp-server-xyz").Call(context.Background(), "x", nil); err == nil {
		t.Error("a server binary that is not installed must be an error")
	}
}

func TestMCPClient_ListToolsParsesTheToolList(t *testing.T) {
	fakeMCPServer(t,
		`{"jsonrpc":"2.0","id":1,"result":{"tools":[{"name":"web_search","description":"search the web"},{"name":"fetch"}]}}`, 0)
	tools, err := NewMCPClient("exa", "exa-mcp").ListTools(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(tools) != 2 || tools[0].Name != "web_search" || tools[0].Description != "search the web" {
		t.Errorf("tools = %+v, want the two listed tools", tools)
	}
}

func TestMCPClient_ListToolsReportsServerError(t *testing.T) {
	fakeMCPServer(t, `{"jsonrpc":"2.0","id":1,"error":{"code":-1,"message":"not ready"}}`, 0)
	if _, err := NewMCPClient("exa", "exa-mcp").ListTools(context.Background()); err == nil {
		t.Error("an error response must be reported")
	}
}

func TestMCPClient_ListToolsWithMissingBinary(t *testing.T) {
	noTools(t)
	if _, err := NewMCPClient("exa", "no-such-mcp-server-xyz").ListTools(context.Background()); err == nil {
		t.Error("a missing server binary must be an error")
	}
}

func TestMCPManager_RoutesCallsByServerName(t *testing.T) {
	ctx := context.Background()
	fakeMCPServer(t, `{"jsonrpc":"2.0","id":1,"result":{"content":[{"type":"text","text":"routed"}]}}`, 0)

	m := NewMCPManager()
	if _, err := m.Call(ctx, "exa", "x", nil); err == nil || !strings.Contains(err.Error(), "not registered") {
		t.Errorf("calling an unregistered server: err = %v, want a not-registered error", err)
	}
	if _, err := m.ListTools(ctx, "exa"); err == nil || !strings.Contains(err.Error(), "not registered") {
		t.Errorf("listing an unregistered server: err = %v, want a not-registered error", err)
	}

	m.RegisterClient("exa", NewMCPClient("exa", "exa-mcp"))
	if _, ok := m.GetClient("exa"); !ok {
		t.Fatal("GetClient should find a registered client")
	}
	res, err := m.Call(ctx, "exa", "web_search", map[string]interface{}{"query": "q"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Text != "routed\n" {
		t.Errorf("Text = %q, want the registered server's answer", res.Text)
	}
}

func TestDefaultMCPManager_RegistersKnownServers(t *testing.T) {
	m := DefaultMCPManager()
	if m == nil {
		t.Fatal("DefaultMCPManager returned nil")
	}
	if _, ok := m.GetClient("exa"); !ok {
		t.Error("the default manager should register the exa server")
	}
}
