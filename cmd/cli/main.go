package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/xlrte/core/pkg/cmd"
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	rootCmd := cmd.BuildCommand(ctx)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("Execution cancelled by sigterm")
		cancel()
		os.Exit(1)
	}()

	err := rootCmd.Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
