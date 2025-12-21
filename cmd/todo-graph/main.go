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
		dir, output, errorsOutput, format, keywords, err := parseGenerateFlags(os.Args[2:])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(runGenerate(p, dir, output, errorsOutput, format, keywords))
	case "check":
		dir, keywords, err := parseDirFlag(os.Args[2:], "check")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		os.Exit(runCheck(p, dir, keywords))
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

func parseGenerateFlags(args []string) (string, string, string, string, []string, error) {
	dir := ""
	output := ""
	errorsOutput := ""
	format := "yaml"
	var keywords []string
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--dir":
			if i+1 >= len(args) {
				return "", "", "", "", nil, fmt.Errorf("missing value for --dir")
			}
			if dir != "" {
				return "", "", "", "", nil, fmt.Errorf("duplicate --dir flag")
			}
			dir = args[i+1]
			i++
		case "--output":
			if i+1 >= len(args) {
				return "", "", "", "", nil, fmt.Errorf("missing value for --output")
			}
			if output != "" {
				return "", "", "", "", nil, fmt.Errorf("duplicate --output flag")
			}
			output = args[i+1]
			i++
		case "--errors-output":
			if i+1 >= len(args) {
				return "", "", "", "", nil, fmt.Errorf("missing value for --errors-output")
			}
			if errorsOutput != "" {
				return "", "", "", "", nil, fmt.Errorf("duplicate --errors-output flag")
			}
			errorsOutput = args[i+1]
			i++
		case "--format":
			if i+1 >= len(args) {
				return "", "", "", "", nil, fmt.Errorf("missing value for --format")
			}
			val := strings.ToLower(args[i+1])
			if val != "yaml" && val != "json" {
				return "", "", "", "", nil, fmt.Errorf("unsupported format: %s", val)
			}
			format = val
			i++
		case "--keywords":
			if i+1 >= len(args) {
				return "", "", "", "", nil, fmt.Errorf("missing value for --keywords")
			}
			if len(keywords) != 0 {
				return "", "", "", "", nil, fmt.Errorf("duplicate --keywords flag")
			}
			keywords = parseKeywords(args[i+1])
			i++
		default:
			return "", "", "", "", nil, fmt.Errorf("unknown flag for generate: %s", args[i])
		}
	}
	return dir, output, errorsOutput, format, keywords, nil
}

func parseDirFlag(args []string, cmd string) (string, []string, error) {
	dir := ""
	var keywords []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--dir":
			if i+1 >= len(args) {
				return "", nil, fmt.Errorf("missing value for --dir")
			}
			dir = args[i+1]
			i++
		case "--keywords":
			if i+1 >= len(args) {
				return "", nil, fmt.Errorf("missing value for --keywords")
			}
			if len(keywords) != 0 {
				return "", nil, fmt.Errorf("duplicate --keywords flag")
			}
			keywords = parseKeywords(args[i+1])
			i++
		default:
			return "", nil, fmt.Errorf("unknown flag for %s: %s", cmd, arg)
		}
	}
	return dir, keywords, nil
}

func parseKeywords(raw string) []string {
	parts := strings.Split(raw, ",")
	var out []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func printHelp() {
	fmt.Printf("todo-graph CLI (version %s)\n", version)
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  todo-graph generate     Scan repository and write .todo-graph")
	fmt.Println("      --dir <path>        Run against a different root (defaults to cwd; useful in scripts)")
	fmt.Println("      --output <path>     Write .todo-graph to a different path (for CI artifacts)")
	fmt.Println("      --format <yaml|json> Output format (default yaml; json writes .todo-graph.json)")
	fmt.Println("      --keywords <list>   Comma-separated keywords to scan (default: TODO,FIXME,NOTE,WARNING,HACK,CHANGED,REVIEW)")
	fmt.Println("  todo-graph check        Validate TODO graph consistency")
	fmt.Println("      --dir <path>        Target a different root")
	fmt.Println("      --keywords <list>   Comma-separated keywords to scan")
	fmt.Println("  todo-graph version      Print the CLI version")
}
