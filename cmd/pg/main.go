package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/davidmdm/flights/postgresql"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
}

func run() error {
	resources, err := postgresql.RenderChart(os.Args[0], os.Getenv("NAMESPACE"), &postgresql.Values{})
	if err != nil {
		return fmt.Errorf("failed to render postgres chart: %w", err)
	}
	return json.NewEncoder(os.Stdout).Encode(resources)
}
