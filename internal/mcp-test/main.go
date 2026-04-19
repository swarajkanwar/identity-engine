package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
)

// JSON-RPC structs for our client
type JSONRPCRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   interface{}     `json:"error,omitempty"`
}

func main() {
	// Start the MCP server process
	cmd := exec.Command("./bin/identity-mcp")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatalf("Error creating stdin pipe: %v", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("Error creating stdout pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		log.Fatalf("Error starting mcp server (did you run 'make build'?): %v", err)
	}

	scanner := bufio.NewScanner(stdout)

	// 1. Send the `initialize` request
	initReq := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
	}
	b, _ := json.Marshal(initReq)
	fmt.Fprintf(stdin, "%s\n", string(b))

	// Read the `initialize` response and ignore it
	if scanner.Scan() {
		// fmt.Println("Init Response:", scanner.Text())
	}

	// 2. Send the `tools/call` request to query_knowledge_base
	callReq := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name": "query_knowledge_base",
			"arguments": map[string]interface{}{
				"query": "tell me how to authenticate?",
			},
		},
	}
	b, _ = json.Marshal(callReq)
	fmt.Fprintf(stdin, "%s\n", string(b))

	// 3. Read the `tools/call` response
	if scanner.Scan() {
		var resp JSONRPCResponse
		if err := json.Unmarshal(scanner.Bytes(), &resp); err != nil {
			log.Fatalf("Failed to parse response: %v", err)
		}

		if resp.Error != nil {
			log.Fatalf("MCP Server Error: %v", resp.Error)
		}

		var result map[string]interface{}
		if err := json.Unmarshal(resp.Result, &result); err != nil {
			log.Fatalf("Failed to parse result: %v", err)
		}

		// Extract the text content from the MCP response
		content := result["content"].([]interface{})
		if len(content) > 0 {
			firstContent := content[0].(map[string]interface{})
			text := firstContent["text"].(string)

			fmt.Println("Query: 'how to authenticate?'")
			fmt.Printf("RAG Response: %s\n", text)
		}
	}

	// Close the stdin pipe to let the MCP server terminate
	stdin.Close()
	cmd.Wait()
}
