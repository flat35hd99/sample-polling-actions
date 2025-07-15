package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/ryanolee/go-chaff"
)

// これは、mcp clientを利用して、mcpをcli tool化するものです。
// --help のときには、ツール一覧からヘルプを生成します。
// --version のときには、バージョン情報を表示します。固定値0.0.1を返します。
// call <tool name> -i <tool input> のときにはツールを呼びだします。
func main() {
	c, err := client.NewStdioMCPClient(
		"npx", []string{}, "-y", "@modelcontextprotocol/server-memory",
	)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	ctx := context.Background()

	// Initialize the connection
	initRequest := mcp.InitializeRequest{}
	if _, err := c.Initialize(ctx, initRequest); err != nil {
		log.Fatal(err)
	}

	resListTools, err := c.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		log.Fatal(err)
	}

	var (
		helpFlag    bool
		versionFlag bool
		toolName    string
		toolInput   string
	)

	// global flags
	flag.BoolVar(&helpFlag, "help", false, "Display help message")
	flag.BoolVar(&versionFlag, "version", false, "Display version information")
	flag.StringVar(&toolName, "call", "", "Call an MCP tool by name")
	flag.StringVar(&toolInput, "i", "", "Input for the MCP tool (JSON string)")

	flag.Parse()

	if helpFlag {
		fmt.Println("Usage:")
		fmt.Println("  mcp -help                 Display this help message")
		fmt.Println("  mcp -version              Display version information")
		fmt.Println("  mcp call <tool name> -i <tool input>  Call an MCP tool")
		fmt.Println("\nAvailable Tools:")
		for _, tool := range resListTools.Tools {
			fmt.Printf("  - %s: %s\n", tool.Name, tool.Description)
			if tool.InputSchema.Type != "" {

				schemaBytes, err := json.MarshalIndent(tool.InputSchema, "    ", "  ")
				if err != nil {
					log.Printf("Warning: Failed to marshal input schema for tool %s: %v", tool.Name, err)
					continue
				}
				fmt.Printf("    Input Schema:\n%s\n", string(schemaBytes))

				{
					generator, err := chaff.ParseSchemaStringWithDefaults(string(schemaBytes))
					if err != nil {
						log.Printf("Warning: Failed to parse raw input schema for tool %s: %v", tool.Name, err)
						continue
					}
					result := generator.GenerateWithDefaults()
					res, err := json.MarshalIndent(result, "    ", "  ")
					if err != nil {
						log.Printf("Warning: Failed to marshal generated input for tool %s: %v", tool.Name, err)
						continue
					}
					fmt.Printf("    Example Input:\n%s\n", string(res))
				}
			}
		}
		return
	}

	if versionFlag {
		fmt.Println("0.0.1")
		return
	}

	if toolName != "" {
		if toolInput == "" {
			log.Fatal("Error: Tool input (-i) is required when calling a tool.")
		}

		inputJSON := json.RawMessage(toolInput)
		callRequest := mcp.CallToolRequest{
			Params: mcp.CallToolParams{
				Name:      toolName,
				Arguments: inputJSON,
			},
		}

		resCallTool, err := c.CallTool(ctx, callRequest)
		if err != nil {
			log.Fatalf("Error calling tool %s: %v", toolName, err)
		}

		if resCallTool.IsError {
			log.Fatalf("Tool call resulted in an error: %+v", resCallTool.Content)
		}

		for _, c := range resCallTool.Content {
			// Check if the content is TextContent and print its text
			if tc, ok := c.(mcp.TextContent); ok {
				fmt.Println(tc.Text)
			} else {
				// Handle other content types if necessary, or just print a generic message
				fmt.Printf("Tool returned non-text content of type %T: %+v\n", c, c)
			}
		}
		return
	}

	// If no flags are provided, display help
	flag.Usage()
}
