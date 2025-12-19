package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var version = "dev"

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	p := newPrinter()

	cmd := os.Args[1]
	switch cmd {
	case "generate":
		dir, output, err := parseGenerateFlags(os.Args[2:])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(runGenerate(p, dir, output))
	case "check":
		dir, err := parseDirFlag(os.Args[2:], "check")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(runCheck(p, dir))
	case "deps":
		if len(os.Args) < 3 {
			fmt.Fprintln(os.Stderr, "deps requires a subcommand (e.g. set)")
			os.Exit(1)
		}
		sub := os.Args[2]
		switch sub {
		case "set":
			dir, child, parents, err := parseDepsSetFlags(os.Args[3:])
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			os.Exit(runDepsSet(p, dir, child, parents))
		case "detach":
			dir, child, target, detachAll, err := parseDepsDetachFlags(os.Args[3:])
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			os.Exit(runDepsDetach(p, dir, child, target, detachAll))
		default:
			fmt.Fprintf(os.Stderr, "unknown deps subcommand: %s\n", sub)
			os.Exit(1)
		}
	case "fix":
		dir, err := parseDirFlag(os.Args[2:], "fix")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(runFix(p, dir))
	case "help", "-h", "--help":
		printHelp()
	default:
		fmt.Printf("unknown command: %s\n", cmd)
		printHelp()
		os.Exit(1)
	}
}

func resolveRoot(dir string) (string, error) {
	if dir == "" {
		return currentRoot()
	}
	return filepath.Abs(dir)
}

func currentRoot() (string, error) {
	root, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Abs(root)
}

func parseGenerateFlags(args []string) (string, string, error) {
	dir := ""
	output := ""
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--dir":
			if i+1 >= len(args) {
				return "", "", fmt.Errorf("missing value for --dir")
			}
			if dir != "" {
				return "", "", fmt.Errorf("duplicate --dir flag")
			}
			dir = args[i+1]
			i++
		case "--output":
			if i+1 >= len(args) {
				return "", "", fmt.Errorf("missing value for --output")
			}
			if output != "" {
				return "", "", fmt.Errorf("duplicate --output flag")
			}
			output = args[i+1]
			i++
		default:
			return "", "", fmt.Errorf("unknown flag for generate: %s", args[i])
		}
	}
	return dir, output, nil
}

func parseDirFlag(args []string, cmd string) (string, error) {
	dir := ""
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--dir":
			if i+1 >= len(args) {
				return "", fmt.Errorf("missing value for --dir")
			}
			dir = args[i+1]
			i++
		default:
			return "", fmt.Errorf("unknown flag for %s: %s", cmd, arg)
		}
	}
	return dir, nil
}

func parseDepsSetFlags(args []string) (string, string, []string, error) {
	dir := ""
	id := ""
	var parents []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--dir":
			if i+1 >= len(args) {
				return "", "", nil, fmt.Errorf("missing value for --dir")
			}
			dir = args[i+1]
			i++
		case "--id":
			if i+1 >= len(args) {
				return "", "", nil, fmt.Errorf("missing value for --id")
			}
			id = args[i+1]
			i++
		case "--depends-on":
			if i+1 >= len(args) {
				return "", "", nil, fmt.Errorf("missing value for --depends-on")
			}
			list := strings.Split(args[i+1], ",")
			for _, p := range list {
				p = strings.TrimSpace(p)
				if p != "" {
					parents = append(parents, p)
				}
			}
			i++
		default:
			return "", "", nil, fmt.Errorf("unknown flag for deps set: %s", arg)
		}
	}
	if id == "" {
		return "", "", nil, fmt.Errorf("--id is required")
	}
	if len(parents) == 0 {
		return "", "", nil, fmt.Errorf("--depends-on requires at least one parent id")
	}
	return dir, id, parents, nil
}

func parseDepsDetachFlags(args []string) (string, string, string, bool, error) {
	dir := ""
	id := ""
	target := ""
	detachAll := false
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--dir":
			if i+1 >= len(args) {
				return "", "", "", false, fmt.Errorf("missing value for --dir")
			}
			dir = args[i+1]
			i++
		case "--id":
			if i+1 >= len(args) {
				return "", "", "", false, fmt.Errorf("missing value for --id")
			}
			id = args[i+1]
			i++
		case "--target":
			if i+1 >= len(args) {
				return "", "", "", false, fmt.Errorf("missing value for --target")
			}
			target = args[i+1]
			i++
		case "--all":
			detachAll = true
		default:
			return "", "", "", false, fmt.Errorf("unknown flag for deps detach: %s", arg)
		}
	}
	if id == "" {
		return "", "", "", false, fmt.Errorf("--id is required")
	}
	if target == "" && !detachAll {
		return "", "", "", false, fmt.Errorf("--target is required unless --all")
	}
	return dir, id, target, detachAll, nil
}

func printHelp() {
	fmt.Printf("todo-graph CLI (version %s)\n", version)
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  todo-graph generate     Scan repository and write .todo-graph")
	fmt.Println("      --dir <path>        Run against a different root (defaults to cwd; useful in scripts)")
	fmt.Println("      --output <path>     Write .todo-graph to a different path (for CI artifacts)")
	fmt.Println("  todo-graph check        Validate TODO graph consistency")
	fmt.Println("      --dir <path>        Target a different root")
	fmt.Println("  todo-graph deps set     Update @todo-deps for a TODO id")
	fmt.Println("      --id <id>           Target TODO id to update")
	fmt.Println("      --depends-on <ids>  Comma-separated parent TODO ids")
	fmt.Println("      --dir <path>        Target a different root")
	fmt.Println("  todo-graph deps detach  Remove a parent from a TODO's @todo-deps")
	fmt.Println("      --id <id>           Target TODO id to update")
	fmt.Println("      --target <id>       Parent TODO id to remove")
	fmt.Println("      --dir <path>        Target a different root")
	fmt.Println("  todo-graph fix          Auto-add @todo-id placeholders for missing TODO ids")
	fmt.Println("      --dir <path>        Target a different root")
	fmt.Println("  todo-graph view         Show the graph as an indented tree")
	fmt.Println("      --dir <path>        Target a different root (runs generate first)")
	fmt.Println("      --roots-only        Show only root TODOs (no descendants)")
}
