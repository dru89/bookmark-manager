package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dru89/bookmark-manager/internal/bookmarks"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Serve starts the MCP server on stdio.
func Serve(store *bookmarks.Store) error {
	s := server.NewMCPServer(
		"bookmark-manager",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	// bookmark_search
	s.AddTool(
		mcp.NewTool("bookmark_search",
			mcp.WithDescription("Search bookmarks by name or tags. Returns matching entries with name, tags, and URL. If query is empty, returns all bookmarks."),
			mcp.WithString("query", mcp.Description("Search query (matched against name and tags)")),
		),
		searchHandler(store),
	)

	// bookmark_add
	s.AddTool(
		mcp.NewTool("bookmark_add",
			mcp.WithDescription("Add a new bookmark. The agent should infer a short name and relevant search tags from the user's description."),
			mcp.WithString("url", mcp.Required(), mcp.Description("The URL to bookmark")),
			mcp.WithString("name", mcp.Required(), mcp.Description("Short descriptive name for the bookmark")),
			mcp.WithString("tags", mcp.Description("Space-separated search keywords")),
		),
		addHandler(store),
	)

	// bookmark_update
	s.AddTool(
		mcp.NewTool("bookmark_update",
			mcp.WithDescription("Update an existing bookmark's name, tags, or URL. Matches by current name (case-insensitive)."),
			mcp.WithString("name", mcp.Required(), mcp.Description("Current name of the bookmark to update")),
			mcp.WithString("new_name", mcp.Description("New name (omit to keep current)")),
			mcp.WithString("new_tags", mcp.Description("New tags (omit to keep current)")),
			mcp.WithString("new_url", mcp.Description("New URL (omit to keep current)")),
		),
		updateHandler(store),
	)

	// bookmark_remove
	s.AddTool(
		mcp.NewTool("bookmark_remove",
			mcp.WithDescription("Remove a bookmark by name (case-insensitive match)."),
			mcp.WithString("name", mcp.Required(), mcp.Description("Name of the bookmark to remove")),
		),
		removeHandler(store),
	)

	// bookmark_list
	s.AddTool(
		mcp.NewTool("bookmark_list",
			mcp.WithDescription("List all bookmarks. Use this when search returns nothing or to see everything available."),
		),
		listHandler(store),
	)

	return server.ServeStdio(s)
}

func searchHandler(store *bookmarks.Store) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query := req.GetString("query", "")

		results, err := store.Search(query)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("search failed: %v", err)), nil
		}

		if len(results) == 0 {
			return mcp.NewToolResultText("No bookmarks found matching query."), nil
		}

		data, _ := json.MarshalIndent(results, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	}
}

func addHandler(store *bookmarks.Store) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		url, err := req.RequireString("url")
		if err != nil {
			return mcp.NewToolResultError("url is required"), nil
		}
		name, err := req.RequireString("name")
		if err != nil {
			return mcp.NewToolResultError("name is required"), nil
		}
		tags := req.GetString("tags", "")

		entry := bookmarks.Entry{
			Name: name,
			Tags: tags,
			URL:  url,
		}

		if err := store.Add(entry); err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("failed to add: %v", err)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Added bookmark: %s", name)), nil
	}
}

func updateHandler(store *bookmarks.Store) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, err := req.RequireString("name")
		if err != nil {
			return mcp.NewToolResultError("name is required"), nil
		}

		// Load current entry to merge changes
		entries, err := store.Search(name)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("search failed: %v", err)), nil
		}

		// Find exact match
		var current *bookmarks.Entry
		for _, e := range entries {
			if e.Name == name {
				current = &e
				break
			}
		}
		if current == nil && len(entries) > 0 {
			current = &entries[0]
		}
		if current == nil {
			return mcp.NewToolResultError(fmt.Sprintf("no bookmark found: %s", name)), nil
		}

		updated := *current
		if newName := req.GetString("new_name", ""); newName != "" {
			updated.Name = newName
		}
		if newTags := req.GetString("new_tags", ""); newTags != "" {
			updated.Tags = newTags
		}
		if newURL := req.GetString("new_url", ""); newURL != "" {
			updated.URL = newURL
		}

		ok, err := store.Update(name, updated)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("update failed: %v", err)), nil
		}
		if !ok {
			return mcp.NewToolResultError(fmt.Sprintf("no bookmark found: %s", name)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Updated bookmark: %s", updated.Name)), nil
	}
}

func removeHandler(store *bookmarks.Store) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name, err := req.RequireString("name")
		if err != nil {
			return mcp.NewToolResultError("name is required"), nil
		}

		removed, err := store.Remove(name)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("remove failed: %v", err)), nil
		}
		if !removed {
			return mcp.NewToolResultError(fmt.Sprintf("no bookmark found: %s", name)), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Removed bookmark: %s", name)), nil
	}
}

func listHandler(store *bookmarks.Store) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		entries, err := store.Load()
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("list failed: %v", err)), nil
		}

		if len(entries) == 0 {
			return mcp.NewToolResultText("No bookmarks saved."), nil
		}

		data, _ := json.MarshalIndent(entries, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	}
}
