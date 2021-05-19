package cmd

import (
	"context"
	"fmt"
	"log"
	"net"
	nethttp "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"stageai.tech/sunshine/sunshine/http"
	"stageai.tech/sunshine/sunshine/services"

	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Starts backend server",
	Run:   serve,
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func serve(_ *cobra.Command, _ []string) {
	env, err := services.NewEnv()
	if err != nil {
		log.Fatal("services.New:", err)
	}

	var url = fmt.Sprintf("http://127.0.0.1:%d", env.General.Port)
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", env.General.Port))
	if err != nil {
		log.Panicf("cannot listen: %s", err)
	}

	s := nethttp.Server{
		Addr:    fmt.Sprintf("localhost:%d", env.General.Port),
		Handler: http.New(env),

		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	var done = make(chan struct{})
	go func() {
		var c = make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		<-c

		log.Println("Got shutdown request.")
		if err := s.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP server Shutdown: %v", err)
		}
		close(done)
	}()

	go watchdog(url + "/debug/ping")

	log.Printf("Listening on %s", url)
	if err := s.Serve(l); err != nethttp.ErrServerClosed {
		// Error starting or closing listener:
		log.Printf("HTTP server Serve: %v", err)
	}
	<-done
}
