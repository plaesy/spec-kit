// Package reach provides a capability layer for AI agents to access
// external data sources across multiple platforms.
package reach

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// MCPClient provides a client for interacting with MCP servers.
type MCPClient struct {
	serverName string
	command    string
	args       []string
	env        map[string]string
}

// NewMCPClient creates a new MCP client for a server.
func NewMCPClient(serverName, command string, args ...string) *MCPClient {
	return &MCPClient{
		serverName: serverName,
		command:    command,
		args:       args,
		env:        make(map[string]string),
	}
}

// SetEnv sets an environment variable for the MCP server process.
func (m *MCPClient) SetEnv(key, value string) {
	m.env[key] = value
}

// Call invokes a tool on the MCP server.
func (m *MCPClient) Call(ctx context.Context, tool string, params map[string]interface{}) (MCPResult, error) {
	// Build the MCP call request
	request := MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params: MCPCallParams{
			Name:      tool,
			Arguments: params,
		},
	}

	requestJSON, err := json.Marshal(request)
	if err != nil {
		return MCPResult{}, fmt.Errorf("marshaling request: %w", err)
	}

	// Execute via mcporter or direct stdio
	cmd := exec.CommandContext(ctx, m.command, m.args...)

	// Set environment
	cmd.Env = os.Environ()
	for k, v := range m.env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return MCPResult{}, fmt.Errorf("stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return MCPResult{}, fmt.Errorf("stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return MCPResult{}, fmt.Errorf("stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return MCPResult{}, fmt.Errorf("starting MCP server: %w", err)
	}

	// Send request
	if _, err := stdin.Write(requestJSON); err != nil {
		return MCPResult{}, fmt.Errorf("writing request: %w", err)
	}
	stdin.Close()

	// Read response
	responseData, err := io.ReadAll(stdout)
	if err != nil {
		return MCPResult{}, fmt.Errorf("reading response: %w", err)
	}

	// Wait for process
	if err := cmd.Wait(); err != nil {
		stderrData, _ := io.ReadAll(stderr)
		return MCPResult{}, fmt.Errorf("MCP server error: %v: %s", err, string(stderrData))
	}

	// Parse response
	var response MCPResponse
	if err := json.Unmarshal(responseData, &response); err != nil {
		return MCPResult{}, fmt.Errorf("parsing response: %w", err)
	}

	if response.Error != nil {
		return MCPResult{}, fmt.Errorf("MCP error: %s", response.Error.Message)
	}

	result := MCPResult{}
	if response.Result != nil {
		// Unmarshal the result into a map
		var resultMap map[string]interface{}
		if err := json.Unmarshal(response.Result, &resultMap); err != nil {
			return MCPResult{}, fmt.Errorf("parsing call result: %w", err)
		}

		if content, ok := resultMap["content"]; ok {
			if contentArr, ok := content.([]interface{}); ok {
				for _, item := range contentArr {
					if textItem, ok := item.(map[string]interface{}); ok {
						if text, ok := textItem["text"].(string); ok {
							result.Text += text + "\n"
						}
					}
				}
			}
		}
	}

	return result, nil
}

// ListTools lists available tools on the MCP server.
func (m *MCPClient) ListTools(ctx context.Context) ([]MCPTool, error) {
	request := MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/list",
	}

	requestJSON, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	cmd := exec.CommandContext(ctx, m.command, m.args...)
	cmd.Env = os.Environ()
	for k, v := range m.env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("starting MCP server: %w", err)
	}

	if _, err := stdin.Write(requestJSON); err != nil {
		return nil, fmt.Errorf("writing request: %w", err)
	}
	stdin.Close()

	responseData, err := io.ReadAll(stdout)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		stderrData, _ := io.ReadAll(stderr)
		return nil, fmt.Errorf("MCP server error: %v: %s", err, string(stderrData))
	}

	var response MCPResponse
	if err := json.Unmarshal(responseData, &response); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	if response.Error != nil {
		return nil, fmt.Errorf("MCP error: %s", response.Error.Message)
	}

	var tools []MCPTool
	if response.Result != nil {
		// Unmarshal the result into a map
		var resultMap map[string]interface{}
		if err := json.Unmarshal(response.Result, &resultMap); err != nil {
			return nil, fmt.Errorf("parsing tools result: %w", err)
		}

		if toolsData, ok := resultMap["tools"]; ok {
			if toolsArr, ok := toolsData.([]interface{}); ok {
				for _, t := range toolsArr {
					if toolMap, ok := t.(map[string]interface{}); ok {
						tool := MCPTool{}
						if name, ok := toolMap["name"].(string); ok {
							tool.Name = name
						}
						if desc, ok := toolMap["description"].(string); ok {
							tool.Description = desc
						}
						if schema, ok := toolMap["inputSchema"].(map[string]interface{}); ok {
							tool.InputSchema = schema
						}
						tools = append(tools, tool)
					}
				}
			}
		}
	}

	return tools, nil
}

// MCPRequest represents a JSON-RPC request to an MCP server.
type MCPRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

// MCPCallParams represents parameters for a tools/call request.
type MCPCallParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// MCPResponse represents a JSON-RPC response from an MCP server.
type MCPResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *MCPError       `json:"error,omitempty"`
}

// MCPError represents an MCP error.
type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// MCPResult represents the result of an MCP tool call.
type MCPResult struct {
	Text string
	Data interface{}
}

// MCPTool represents an MCP tool definition.
type MCPTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

// MCPManager manages multiple MCP server connections.
type MCPManager struct {
	clients map[string]*MCPClient
}

// NewMCPManager creates a new MCP manager.
func NewMCPManager() *MCPManager {
	return &MCPManager{
		clients: make(map[string]*MCPClient),
	}
}

// RegisterClient registers an MCP client.
func (m *MCPManager) RegisterClient(name string, client *MCPClient) {
	m.clients[name] = client
}

// GetClient returns an MCP client by name.
func (m *MCPManager) GetClient(name string) (*MCPClient, bool) {
	client, ok := m.clients[name]
	return client, ok
}

// Call invokes a tool on an MCP server.
func (m *MCPManager) Call(ctx context.Context, server, tool string, params map[string]interface{}) (MCPResult, error) {
	client, ok := m.GetClient(server)
	if !ok {
		return MCPResult{}, fmt.Errorf("MCP server %s not registered", server)
	}
	return client.Call(ctx, tool, params)
}

// ListTools lists tools for an MCP server.
func (m *MCPManager) ListTools(ctx context.Context, server string) ([]MCPTool, error) {
	client, ok := m.GetClient(server)
	if !ok {
		return nil, fmt.Errorf("MCP server %s not registered", server)
	}
	return client.ListTools(ctx)
}

// DefaultMCPManager returns a manager with common MCP servers pre-configured.
func DefaultMCPManager() *MCPManager {
	m := NewMCPManager()

	// Exa search
	m.RegisterClient("exa", NewMCPClient("exa", "mcporter", "exa"))

	// LinkedIn
	m.RegisterClient("linkedin", NewMCPClient("linkedin", "mcporter", "linkedin"))

	// XiaoHongShu
	m.RegisterClient("xiaohongshu", NewMCPClient("xiaohongshu", "mcporter", "xiaohongshu"))

	// Xiaoyuzhou
	m.RegisterClient("xiaoyuzhou", NewMCPClient("xiaoyuzhou", "mcporter", "xiaoyuzhou"))

	return m
}
