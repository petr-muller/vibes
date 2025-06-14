package main

import (
	"fmt"
	"os"

	"github.com/petr-muller/vibes/pkg/fauxinnati"
	"github.com/spf13/cobra"
)

var (
	port int
)

var rootCmd = &cobra.Command{
	Use:   "fauxinnati",
	Short: "A mock Cincinnati update graph server",
	Long:  "fauxinnati is a mock implementation of the Red Hat OpenShift Cincinnati update graph protocol",
	Run: func(cmd *cobra.Command, args []string) {
		server := fauxinnati.NewServer()
		if err := server.Start(port); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to run the server on")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
