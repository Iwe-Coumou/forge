package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Iwe-Coumou/forge/v2/internal/config"
	"github.com/Iwe-Coumou/forge/v2/internal/forger"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var inspectCmd = &cobra.Command{
	Use:   "inspect [language/template]",
	Short: "Inspect a template",
	Long:  "Inspect a template to see the file structure and details",
	Args:  cobra.ExactArgs(1),
	RunE:  runInspect,
}

func init() {
	rootCmd.AddCommand(inspectCmd)
}

func runInspect(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	lang, name, err := forger.ParseTemplateID(cfg, args[0])
	if err != nil {
		return err
	}

	detail, err := forger.InspectTemplate(cfg, lang, name)
	if err != nil {
		return err
	}

	color.Blue("\n%s\n\n", detail.ID())

	if detail.NotImplemented != "" {
		color.Yellow("  not implemented yet: %s\n\n", detail.NotImplemented)
	}

	if detail.Short != "" {
		fmt.Printf("  %s\n\n", detail.Short)
	}
	if long := strings.TrimSpace(detail.Long); long != "" {
		fmt.Printf("  %s\n\n", long)
	}

	if len(detail.VerifyCmd) > 0 {
		color.HiBlack("  verify   %s\n", strings.Join(detail.VerifyCmd, " "))
	}
	if len(detail.Keys.Flag) > 0 {
		color.HiBlack("  --set    %s\n", strings.Join(detail.Keys.Flag, ", "))
	}
	if len(detail.Keys.Config) > 0 {
		color.HiBlack("  config   %s\n", strings.Join(detail.Keys.Config, ", "))
	}
	if detail.Source != "" && detail.Source != "embedded" {
		color.HiBlack("  source   %s\n", detail.Source)
	}
	fmt.Println()

	color.HiBlack("  files\n")
	if len(detail.Files) == 0 {
		fmt.Println("  (none)")
		return nil
	}
	printTree(buildTree(detail.Files), "  ")
	fmt.Println()

	return nil
}

// treeNode is one entry in the rendered file tree. A node with no children
// is a file.
type treeNode struct {
	name     string
	children map[string]*treeNode
}

// buildTree turns slash-separated paths into a nested tree.
func buildTree(paths []string) *treeNode {
	root := &treeNode{children: map[string]*treeNode{}}

	for _, p := range paths {
		node := root
		for _, seg := range strings.Split(p, "/") {
			child, ok := node.children[seg]
			if !ok {
				child = &treeNode{name: seg, children: map[string]*treeNode{}}
				node.children[seg] = child
			}
			node = child
		}
	}

	return root
}

// printTree writes node's children with box-drawing connectors, indenting
// each level under its parent.
func printTree(node *treeNode, prefix string) {
	names := make([]string, 0, len(node.children))
	for name := range node.children {
		names = append(names, name)
	}
	sort.Strings(names)

	for i, name := range names {
		isLast := i == len(names)-1

		connector, childPrefix := "├── ", prefix+"│   "
		if isLast {
			connector, childPrefix = "└── ", prefix+"    "
		}

		fmt.Printf("%s%s%s\n", prefix, connector, name)
		printTree(node.children[name], childPrefix)
	}
}
