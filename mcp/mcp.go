package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/ryanolee/go-chaff"
	"github.com/spf13/cobra"
)

const version = "0.0.1"

var (
	mcpClient *client.Client
	ctx       context.Context
)

// これは、mcp clientを利用して、mcpをcli tool化するものです。
// list コマンドでツール一覧を表示
// call コマンドでツールを呼び出し
func main() {
	// Initialize context
	ctx = context.Background()

	// Initialize MCP client
	var err error
	mcpClient, err = client.NewStdioMCPClient(
		"npx", []string{}, "-y", "@modelcontextprotocol/server-memory",
	)
	if err != nil {
		log.Fatal(err)
	}
	defer mcpClient.Close()

	// Initialize the connection
	initRequest := mcp.InitializeRequest{}
	if _, err := mcpClient.Initialize(ctx, initRequest); err != nil {
		log.Fatal(err)
	}

	// Create root command
	rootCmd := &cobra.Command{
		Use:   "mcp",
		Short: "MCP CLI tool for interacting with Model Context Protocol servers",
		Long:  `A command-line interface for interacting with Model Context Protocol (MCP) servers.`,
	}

	// Add version flag to root command
	rootCmd.Version = version

	// Add subcommands
	rootCmd.AddCommand(newListCommand())
	rootCmd.AddCommand(newCallCommand())

	// Execute root command
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// newListCommand creates the list command
func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available MCP tools",
		Long:  `List all available MCP tools with their descriptions, input schemas, and example inputs.`,
		RunE:  runListCommand,
	}

	return cmd
}

// newCallCommand creates the call command
func newCallCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "call [tool-name]",
		Short: "Call an MCP tool",
		Long:  `Call an MCP tool by name with the provided input.`,
		Args:  cobra.ExactArgs(1),
		RunE:  runCallCommand,
	}

	cmd.Flags().StringP("input", "i", "", "Input for the MCP tool (JSON string)")
	cmd.Flags().BoolP("help-tool", "", false, "Show tool schema and example input")

	return cmd
}

// runListCommand executes the list command
func runListCommand(cmd *cobra.Command, args []string) error {
	resListTools, err := mcpClient.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return fmt.Errorf("failed to list tools: %v", err)
	}

	fmt.Println("Available MCP Tools:")
	fmt.Println("===================")

	for i, tool := range resListTools.Tools {
		if i > 0 {
			fmt.Println()
		}

		fmt.Printf("Tool: %s\n", tool.Name)
		fmt.Printf("Description: %s\n", tool.Description)
	}

	return nil
}

// runCallCommand executes the call command
func runCallCommand(cmd *cobra.Command, args []string) error {
	toolName := args[0]

	// Check if help-tool flag is set
	helpTool, err := cmd.Flags().GetBool("help-tool")
	if err != nil {
		return fmt.Errorf("failed to get help-tool flag: %v", err)
	}

	// If help-tool is requested, show schema and example
	if helpTool {
		return showToolHelp(toolName)
	}

	// For actual tool calls, input is required
	toolInput, err := cmd.Flags().GetString("input")
	if err != nil {
		return fmt.Errorf("failed to get input flag: %v", err)
	}

	if toolInput == "" {
		return fmt.Errorf("tool input is required (use --input flag)")
	}

	inputJSON := json.RawMessage(toolInput)
	callRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      toolName,
			Arguments: inputJSON,
		},
	}

	resCallTool, err := mcpClient.CallTool(ctx, callRequest)
	if err != nil {
		return fmt.Errorf("error calling tool %s: %v", toolName, err)
	}

	if resCallTool.IsError {
		return fmt.Errorf("tool call resulted in an error: %+v", resCallTool.Content)
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

	return nil
}

// showToolHelp displays the schema and example input for a specific tool
func showToolHelp(toolName string) error {
	resListTools, err := mcpClient.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		return fmt.Errorf("failed to list tools: %v", err)
	}

	var targetTool *mcp.Tool
	for _, tool := range resListTools.Tools {
		if tool.Name == toolName {
			targetTool = &tool
			break
		}
	}

	if targetTool == nil {
		return fmt.Errorf("tool '%s' not found", toolName)
	}

	fmt.Printf("Tool: %s\n", targetTool.Name)
	fmt.Printf("Description: %s\n", targetTool.Description)
	fmt.Println()

	if targetTool.InputSchema.Type != "" {
		schemaBytes, err := json.MarshalIndent(targetTool.InputSchema, "  ", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal input schema: %v", err)
		}
		fmt.Printf("Input Schema:\n  %s\n", string(schemaBytes))
		fmt.Println()

		// Generate example input
		generator, err := chaff.ParseSchemaStringWithDefaults(string(schemaBytes))
		if err != nil {
			return fmt.Errorf("failed to parse input schema: %v", err)
		}

		result := generator.GenerateWithDefaults()
		res, err := json.MarshalIndent(result, "  ", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal generated input: %v", err)
		}
		fmt.Printf("Example Input:\n  %s\n", string(res))
	} else {
		fmt.Println("No input schema available for this tool.")
	}

	return nil
}
