package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	cmd := os.Args[1]
	switch cmd {
	case "scan":
		fmt.Println("todo-graph scan: not implemented yet")
	case "check":
		fmt.Println("todo-graph check: not implemented yet")
	case "visualize":
		fmt.Println("todo-graph visualize: not implemented yet")
	case "help", "-h", "--help":
		printHelp()
	default:
		fmt.Printf("unknown command: %s\n", cmd)
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println("todo-graph CLI")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  todo-graph scan       Scan repository and update .todo-graph")
	fmt.Println("  todo-graph check      Validate TODO graph consistency")
	fmt.Println("  todo-graph visualize  Output graph in a given format")
}
