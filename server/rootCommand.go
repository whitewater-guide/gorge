package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/octago/sflags/gen/gpflag"
	"github.com/spf13/cobra"
	"github.com/whitewater-guide/gorge/scripts"
)

var rootCmd *cobra.Command

func init() {
	cfg := defaultConfig()
	rootCmd = &cobra.Command{
		Use:   fmt.Sprintf("%s [flags]", filepath.Base(os.Args[0])),
		Short: "Runs gorge server",
		Run: func(cmd *cobra.Command, args []string) {
			cfg.readFromEnv()
			// if cfg.Debug {
			// 	runtime.MemProfileRate = 1
			// }
			srv := newServer(cfg, scripts.Registry)
			srv.routes()
			srv.start()

			walkFunc := func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
				route = strings.Replace(route, "/*/", "/", -1)
				srv.logger.Debugf("%s %s\n", method, route)
				return nil
			}

			if err := chi.Walk(srv.router, walkFunc); err != nil {
				srv.logger.Panicf("Logging err: %s\n", err.Error())
			}

			httpSrv := &http.Server{
				Addr:    fmt.Sprintf(":%s", srv.port),
				Handler: srv.router,
			}

			// Listen in goroutine for graceful shutdown
			go func() {
				srv.logger.Fatal(httpSrv.ListenAndServe())
			}()

			// Graceful shutdown
			shutdown := make(chan os.Signal, 1)
			signal.Notify(shutdown, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
			<-shutdown
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			if err := httpSrv.Shutdown(ctx); err != nil {
				srv.logger.Infof("error shutting down server %v", err)
			} else {
				srv.logger.Infof("server gracefully stopped")
			}

			srv.shutdown()
		},
	}
	err := gpflag.ParseTo(cfg, rootCmd.Flags())
	if err != nil {
		log.Fatalf("Failed to parse flags: %v", err)
	}
}
