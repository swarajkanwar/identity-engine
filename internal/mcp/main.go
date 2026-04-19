package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "identity-engine/proto"
)

// JSON-RPC structs
type JSONRPCRequest struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id,omitempty"`
	Method  string           `json:"method"`
	Params  json.RawMessage  `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id"`
	Result  interface{}      `json:"result,omitempty"`
	Error   *JSONRPCError    `json:"error,omitempty"`
}

type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type InitializeResult struct {
	ProtocolVersion string `json:"protocolVersion"`
	Capabilities    struct {
		Tools map[string]interface{} `json:"tools"`
	} `json:"capabilities"`
	ServerInfo struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"serverInfo"`
}

type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"inputSchema"`
}

type InputSchema struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
	Required   []string               `json:"required,omitempty"`
}

// Sample RAG Database
var sampleRagDB = map[string]string{
	"what is identity-engine?": "Identity Engine is a fast, robust gRPC service for managing user identities.",
	"how do I authenticate?":   "You authenticate by providing a valid JWT token in the Authorization header.",
	"what is mcp?":             "MCP stands for Model Context Protocol, allowing AI models to easily call local tools.",
}

// Dummy vector search that just checks for keyword overlap
func queryRagDatabase(query string) string {
	query = strings.ToLower(query)
	var bestMatch string
	var bestScore int

	for key, value := range sampleRagDB {
		score := 0
		words := strings.Fields(query)
		for _, word := range words {
			if strings.Contains(key, word) {
				score++
			}
		}
		if score > bestScore {
			bestScore = score
			bestMatch = value
		}
	}

	if bestScore == 0 {
		return "I could not find any relevant information in the knowledge base."
	}
	return bestMatch
}

func main() {
	// Connect to the gRPC server
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewIdentityServiceClient(conn)

	// MCP communicates over stdin/stdout
	scanner := bufio.NewScanner(os.Stdin)
	// Optionally increase buffer size if huge requests are expected
	
	for scanner.Scan() {
		line := scanner.Bytes()
		var req JSONRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			continue // Skip non-JSON lines
		}

		switch req.Method {
		case "initialize":
			sendResponse(req.ID, InitializeResult{
				ProtocolVersion: "2024-11-05",
				Capabilities: struct {
					Tools map[string]interface{} `json:"tools"`
				}{Tools: map[string]interface{}{}},
				ServerInfo: struct {
					Name    string `json:"name"`
					Version string `json:"version"`
				}{Name: "identity-engine-mcp", Version: "1.0.0"},
			})
		case "notifications/initialized":
			// Ignore
		case "tools/list":
			sendResponse(req.ID, map[string]interface{}{
				"tools": []Tool{
					{
						Name:        "create_user",
						Description: "Create a new user in the Identity Service",
						InputSchema: InputSchema{
							Type: "object",
							Properties: map[string]interface{}{
								"username": map[string]string{"type": "string"},
								"email":    map[string]string{"type": "string"},
							},
							Required: []string{"username", "email"},
						},
					},
					{
						Name:        "get_user",
						Description: "Get a user by ID",
						InputSchema: InputSchema{
							Type: "object",
							Properties: map[string]interface{}{
								"id": map[string]string{"type": "string"},
							},
							Required: []string{"id"},
						},
					},
					{
						Name:        "update_user",
						Description: "Update a user's details",
						InputSchema: InputSchema{
							Type: "object",
							Properties: map[string]interface{}{
								"id":       map[string]string{"type": "string"},
								"username": map[string]string{"type": "string"},
								"email":    map[string]string{"type": "string"},
							},
							Required: []string{"id", "username", "email"},
						},
					},
					{
						Name:        "delete_user",
						Description: "Delete a user by ID",
						InputSchema: InputSchema{
							Type: "object",
							Properties: map[string]interface{}{
								"id": map[string]string{"type": "string"},
							},
							Required: []string{"id"},
						},
					},
					{
						Name:        "list_users",
						Description: "List all users",
						InputSchema: InputSchema{
							Type:       "object",
							Properties: map[string]interface{}{},
						},
					},
					{
						Name:        "query_knowledge_base",
						Description: "Query the RAG knowledge base for general information about the system",
						InputSchema: InputSchema{
							Type: "object",
							Properties: map[string]interface{}{
								"query": map[string]string{"type": "string", "description": "The question to ask the knowledge base"},
							},
							Required: []string{"query"},
						},
					},
				},
			})
		case "tools/call":
			var params struct {
				Name      string                 `json:"name"`
				Arguments map[string]interface{} `json:"arguments"`
			}
			json.Unmarshal(req.Params, &params)

			var content []map[string]interface{}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

			switch params.Name {
			case "create_user":
				username, _ := params.Arguments["username"].(string)
				email, _ := params.Arguments["email"].(string)
				res, err := client.CreateUser(ctx, &pb.CreateUserRequest{Username: username, Email: email})
				content = formatResult(res, err)
			case "get_user":
				id, _ := params.Arguments["id"].(string)
				res, err := client.GetUser(ctx, &pb.GetUserRequest{Id: id})
				content = formatResult(res, err)
			case "update_user":
				id, _ := params.Arguments["id"].(string)
				username, _ := params.Arguments["username"].(string)
				email, _ := params.Arguments["email"].(string)
				res, err := client.UpdateUser(ctx, &pb.UpdateUserRequest{
					User: &pb.User{Id: id, Username: username, Email: email},
				})
				content = formatResult(res, err)
			case "delete_user":
				id, _ := params.Arguments["id"].(string)
				res, err := client.DeleteUser(ctx, &pb.DeleteUserRequest{Id: id})
				content = formatResult(res, err)
			case "list_users":
				res, err := client.ListUsers(ctx, &pb.ListUsersRequest{})
				content = formatResult(res, err)
			case "query_knowledge_base":
				query, _ := params.Arguments["query"].(string)
				answer := queryRagDatabase(query)
				content = []map[string]interface{}{{"type": "text", "text": answer}}
			default:
				content = []map[string]interface{}{{"type": "text", "text": "unknown tool"}}
			}
			cancel()

			sendResponse(req.ID, map[string]interface{}{
				"content": content,
			})
		}
	}
}

func formatResult(res interface{}, err error) []map[string]interface{} {
	if err != nil {
		return []map[string]interface{}{{"type": "text", "text": fmt.Sprintf("Error: %v", err)}}
	}
	b, _ := json.MarshalIndent(res, "", "  ")
	return []map[string]interface{}{{"type": "text", "text": string(b)}}
}

func sendResponse(id *json.RawMessage, result interface{}) {
	if id == nil {
		return
	}
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	b, _ := json.Marshal(resp)
	fmt.Println(string(b))
}
