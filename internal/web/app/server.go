package app

import (
	"context"
	"errors"
	"github.com/eurofurence/reg-attendee-service/internal/repository/config"
	"github.com/eurofurence/reg-attendee-service/internal/repository/logging"
	"github.com/eurofurence/reg-attendee-service/internal/web/controller/adminctl"
	"github.com/eurofurence/reg-attendee-service/internal/web/controller/attendeectl"
	"github.com/eurofurence/reg-attendee-service/internal/web/controller/countdownctl"
	"github.com/eurofurence/reg-attendee-service/internal/web/controller/fallbackctl"
	"github.com/eurofurence/reg-attendee-service/internal/web/controller/infoctl"
	"github.com/eurofurence/reg-attendee-service/internal/web/controller/statusctl"
	"github.com/go-chi/chi/v5"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func CreateRouter(ctx context.Context) chi.Router {
	logging.Ctx(ctx).Debug("Setting up router")
	server := chi.NewRouter()
	countdownctl.Create(server)
	attendeectl.Create(server)
	adminctl.Create(server)
	statusctl.Create(server)
	infoctl.Create(server)
	fallbackctl.Create(server)
	return server
}

func newServer(ctx context.Context, router chi.Router) *http.Server {
	logging.Ctx(ctx).Debug("setting up server")
	return &http.Server{
		Addr:         config.ServerAddr(),
		Handler:      router,
		ReadTimeout:  config.ServerReadTimeout(),
		WriteTimeout: config.ServerWriteTimeout(),
		IdleTimeout:  config.ServerIdleTimeout(),
		BaseContext: func(l net.Listener) context.Context {
			return ctx
		},
	}
}

func runServerWithGracefulShutdown() error {
	ctx, cancel := context.WithCancel(context.Background())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	handler := CreateRouter(ctx)
	srv := newServer(ctx, handler)

	go func() {
		<-sig
		defer cancel()
		logging.NoCtx().Info("Stopping services now")

		tCtx, tcancel := context.WithTimeout(ctx, time.Second*5)
		defer tcancel()

		if err := srv.Shutdown(tCtx); err != nil {
			logging.NoCtx().Fatal("Couldn't shutdown server gracefully")
		}
	}()

	logging.NoCtx().Info("Running service on ", config.ServerAddr())
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		logging.NoCtx().Error("Server closed unexpectedly", err)
		return err
	}

	return nil
}
