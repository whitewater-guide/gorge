package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/octago/sflags/gen/gpflag"
	"github.com/spf13/cobra"
	"github.com/whitewater-guide/gorge/config"
	"github.com/whitewater-guide/gorge/schedule"
	"github.com/whitewater-guide/gorge/scripts"
	"github.com/whitewater-guide/gorge/storage"
	"github.com/whitewater-guide/gorge/version"

	"go.uber.org/fx"
)

func start(cfg *config.Config, srv *Server) error {
	srv.logger.Debug("starting")
	srv.routes()

	httpSrv := &http.Server{
		Addr:    fmt.Sprintf(":%s", srv.port),
		Handler: srv.router,
	}

	go func() {
		if err := httpSrv.ListenAndServe(); err != nil {
			srv.logger.Fatal(err)
		}
	}()

	srv.logger.Infof("started")

	return nil
}

func main() {
	cfg := config.NewConfig()

	rootCmd := &cobra.Command{
		Use:     fmt.Sprintf("%s [flags]", filepath.Base(os.Args[0])),
		Short:   "Runs gorge server",
		Version: version.Version,
		RunE: func(cmd *cobra.Command, args []string) error {
			app := fx.New(
				fx.Supply(cfg),
				fx.Provide(newLogger),
				scripts.Module,
				storage.Module,
				schedule.Module,
				fx.Provide(newServer),
				fx.Invoke(start),
				fx.WithLogger(newFxLogger),
			)
			app.Run()
			return nil
		},
	}

	// flags must be parsed before command execution
	err := gpflag.ParseTo(cfg, rootCmd.Flags())
	if err != nil {
		fmt.Printf("Failed to parse flags: %v", err)
		os.Exit(1)
	}

	rootCmd.Execute() //nolint:errcheck
}
