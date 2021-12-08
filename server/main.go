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
	rootCmd := &cobra.Command{
		Use:     fmt.Sprintf("%s [flags]", filepath.Base(os.Args[0])),
		Short:   "Runs gorge server",
		Version: version.Version,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg.ReadFromEnv()
			// if cfg.Debug {
			// 	runtime.MemProfileRate = 1
			// }
			srv.routes()

			httpSrv := &http.Server{
				Addr:    fmt.Sprintf(":%s", srv.port),
				Handler: srv.router,
			}

			return httpSrv.ListenAndServe()
		},
	}
	err := gpflag.ParseTo(cfg, rootCmd.Flags())
	if err != nil {
		return fmt.Errorf("failed to parse flags: %v", err)
	}

	return rootCmd.Execute()
}

func main() {
	app := fx.New(
		fx.Options(
			config.Module,
			fx.Provide(newLogger),
			scripts.Module,
			storage.Module,
			schedule.Module,
			fx.Provide(newServer),
		),
		fx.Invoke(start),
	)
	app.Run()
}
