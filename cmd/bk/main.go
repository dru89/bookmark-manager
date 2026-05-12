package main

import (
	"fmt"
	"os"

	"github.com/dru89/bookmark-manager/internal/bookmarks"
	"github.com/dru89/bookmark-manager/internal/config"
	"github.com/dru89/bookmark-manager/internal/mcp"
	"github.com/dru89/bookmark-manager/internal/picker"
	"github.com/spf13/cobra"
)

var cfgFile string

func main() {
	root := &cobra.Command{
		Use:   "bk [query]",
		Short: "Fuzzy bookmark manager",
		Long:  "A fuzzy bookmark manager backed by a markdown table.",
		Args:  cobra.ArbitraryArgs,
		RunE:  runSearch,
	}

	root.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ~/.config/bk/config.toml)")
	root.PersistentFlags().StringP("file", "f", "", "bookmark file path (overrides config)")

	root.AddCommand(addCmd())
	root.AddCommand(removeCmd())
	root.AddCommand(listCmd())
	root.AddCommand(mcpCmd())

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}

func getStore(cmd *cobra.Command) (*bookmarks.Store, error) {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}

	// Flag overrides config
	if f, _ := cmd.Flags().GetString("file"); f != "" {
		cfg.File = f
	}

	if cfg.File == "" {
		return nil, fmt.Errorf("no bookmark file configured; set 'file' in config or pass --file")
	}

	return bookmarks.NewStore(cfg.File), nil
}

func runSearch(cmd *cobra.Command, args []string) error {
	store, err := getStore(cmd)
	if err != nil {
		return err
	}

	entries, err := store.Load()
	if err != nil {
		return err
	}

	if len(entries) == 0 {
		fmt.Fprintln(os.Stderr, "No bookmarks found.")
		return nil
	}

	query := ""
	if len(args) > 0 {
		for i, a := range args {
			if i > 0 {
				query += " "
			}
			query += a
		}
	}

	selected, err := picker.Pick(entries, query)
	if err != nil {
		return err
	}

	if selected == nil {
		return nil // user cancelled
	}

	return openURL(selected.URL)
}

func addCmd() *cobra.Command {
	var name, tags string

	cmd := &cobra.Command{
		Use:   "add <url>",
		Short: "Add a bookmark",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := getStore(cmd)
			if err != nil {
				return err
			}

			url := args[0]

			if name == "" {
				fmt.Print("Name: ")
				fmt.Scanln(&name)
			}
			if tags == "" {
				fmt.Print("Tags (space-separated): ")
				fmt.Scanln(&tags)
			}

			entry := bookmarks.Entry{
				Name: name,
				Tags: tags,
				URL:  url,
			}

			if err := store.Add(entry); err != nil {
				return err
			}

			fmt.Printf("Added: %s\n", entry.Name)
			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "bookmark name")
	cmd.Flags().StringVarP(&tags, "tags", "t", "", "search tags (space-separated)")

	return cmd
}

func removeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a bookmark by name",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := getStore(cmd)
			if err != nil {
				return err
			}

			removed, err := store.Remove(args[0])
			if err != nil {
				return err
			}

			if removed {
				fmt.Printf("Removed: %s\n", args[0])
			} else {
				fmt.Fprintf(os.Stderr, "No bookmark found matching: %s\n", args[0])
			}
			return nil
		},
	}
}

func listCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all bookmarks",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := getStore(cmd)
			if err != nil {
				return err
			}

			entries, err := store.Load()
			if err != nil {
				return err
			}

			if len(entries) == 0 {
				fmt.Println("No bookmarks.")
				return nil
			}

			for _, e := range entries {
				fmt.Printf("%-30s  %s\n", e.Name, e.URL)
			}
			return nil
		},
	}
}

func mcpCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "mcp",
		Short: "Start MCP server (stdio)",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := getStore(cmd)
			if err != nil {
				return err
			}

			return mcp.Serve(store)
		},
	}
}
