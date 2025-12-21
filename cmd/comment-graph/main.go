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
		dir, format, allowErrors, err := parseGenerateFlags(os.Args[2:])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(runGenerate(p, dir, format, allowErrors))
	case "graph":
		dir, allowErrors, err := parseGraphFlags(os.Args[2:])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(runGraph(p, dir, allowErrors))
	case "check":
		dir, err := parseDirFlag(os.Args[2:], "check")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(runCheck(p, dir))
	case "version", "--version", "-v":
		fmt.Println(version)
		return
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

func parseGenerateFlags(args []string) (string, string, bool, error) {
	dir := ""
	format := "yaml"
	allowErrors := false
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--dir":
			if i+1 >= len(args) {
				return "", "", false, fmt.Errorf("missing value for --dir")
			}
			if dir != "" {
				return "", "", false, fmt.Errorf("duplicate --dir flag")
			}
			dir = args[i+1]
			i++
		case "--format":
			if i+1 >= len(args) {
				return "", "", false, fmt.Errorf("missing value for --format")
			}
			val := strings.ToLower(args[i+1])
			if val != "yaml" && val != "json" {
				return "", "", false, fmt.Errorf("unsupported format: %s", val)
			}
			format = val
			i++
		case "--allow-errors":
			if allowErrors {
				return "", "", false, fmt.Errorf("duplicate --allow-errors flag")
			}
			allowErrors = true
		default:
			return "", "", false, fmt.Errorf("unknown flag for generate: %s", args[i])
		}
	}
	return dir, format, allowErrors, nil
}

func parseGraphFlags(args []string) (string, bool, error) {
	dir := ""
	allowErrors := false
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--dir":
			if i+1 >= len(args) {
				return "", false, fmt.Errorf("missing value for --dir")
			}
			dir = args[i+1]
			i++
		case "--allow-errors":
			if allowErrors {
				return "", false, fmt.Errorf("duplicate --allow-errors flag")
			}
			allowErrors = true
		default:
			return "", false, fmt.Errorf("unknown flag for graph: %s", args[i])
		}
	}
	return dir, allowErrors, nil
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

func printHelp() {
	fmt.Printf("comment-graph CLI (version %s)\n", version)
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  comment-graph generate  Scan repository and write .comment-graph")
	fmt.Println("      --dir <path>        Run against a different root (defaults to cwd; useful in scripts)")
	fmt.Println("      --format <yaml|json> Output format (default yaml; json writes .comment-graph.json)")
	fmt.Println("      --allow-errors      Return success even if validation finds issues (report still included)")
	fmt.Println("  comment-graph graph     Scan repository and stream graph+report JSON to stdout (no files written)")
	fmt.Println("      --dir <path>        Target a different root")
	fmt.Println("      --allow-errors      Return success even if validation finds issues (payload still emitted)")
	fmt.Println("  comment-graph check     Validate comment graph consistency")
	fmt.Println("      --dir <path>        Target a different root")
	fmt.Println("  comment-graph version   Print the CLI version")
}
