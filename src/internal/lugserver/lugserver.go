package lugserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/LuckyGuessServices/planets/internal/api/router"
	"github.com/LuckyGuessServices/planets/internal/env"
	"github.com/LuckyGuessServices/planets/internal/luglog"
)

// Serve starts HTTP server and blocks execution.
//
// If a critical failure happens or an OS signal is caught, the server is shut down gracefully.
// If there was a panic, it will be "re-panicked" after the server graceful shutdown.
//
// All endpoint handlers' panics are recovered internally ([http.ListenAndServe] ensures that),
// and the server continues listening.
func Serve() {
	var (
		wg                    sync.WaitGroup
		shutdownSignal        os.Signal
		osSignalChan          = make(chan os.Signal, 1)
		panicServingGoroutine any
	)

	envConfig := env.Config()
	server := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", envConfig.ServerListenHost(), envConfig.ServerListenPort()),
		Handler:           router.Create(),
		ReadHeaderTimeout: 2 * time.Second,
	}

	defer func() {
		panicMain := recover()

		var anyPanic struct {
			source     string
			panicValue any
		}
		if panicMain != nil {
			anyPanic.source = "main()"
			anyPanic.panicValue = panicMain
			luglog.Printf("main() temporarily recovered from panic --> %#v", panicMain)
		} else if panicServingGoroutine != nil {
			anyPanic.source = "goroutine with ListenAndServe"
			anyPanic.panicValue = panicServingGoroutine
			luglog.Printf("main() temporarily recovered from serving goroutine panic --> %#v", panicServingGoroutine)
		} else {
			luglog.Printf("Signal caught: '%v' (%d)", shutdownSignal, shutdownSignal)
		}

		shutdownCtx, shutdownCtxCancel := context.WithTimeout(context.Background(), 20*time.Second)
		// We want to cancel the context ASAP:
		errShutdown := func() error {
			defer shutdownCtxCancel()

			return server.Shutdown(shutdownCtx)
		}()

		if errShutdown != nil {
			luglog.Print("HTTP server shutdown failure --> ", errShutdown, "\nForcing HTTP server to be closed...")
			if err := server.Close(); err != nil {
				luglog.Panic("Failed to close HTTP server forcibly --> ", err)
			}
			luglog.Print("HTTP server forcibly closed all listeners and connections.")
		}

		wg.Wait()
		luglog.Print("HTTP server is shut down gracefully.")

		if anyPanic.panicValue != nil {
			luglog.Printf(
				"Resuming panicking after previous recovery (initial source of panic: '%s')...",
				anyPanic.source,
			)
			panic(anyPanic.panicValue)
		}
	}()

	wg.Go(func() {
		defer func() {
			panicServingGoroutine = recover()
			close(osSignalChan)
		}()

		//goland:noinspection HttpUrlsUsage
		luglog.Print("HTTP server's address: http://", server.Addr)
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			luglog.Panic("HTTP server listening failure --> ", err)
		}
		luglog.Print("HTTP server has stopped listening.")
	})

	signal.Notify(osSignalChan, syscall.SIGHUP, os.Interrupt, syscall.SIGTERM)
	shutdownSignal = <-osSignalChan
}
