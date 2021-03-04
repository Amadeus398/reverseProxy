package main

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
	"net/http"
	"os"
	"os/signal"
	"reverseProxy/pkg/backendManager"
	"reverseProxy/pkg/config"
	"reverseProxy/pkg/db"
	"reverseProxy/pkg/handler"
	"reverseProxy/pkg/handlers/backends"
	"reverseProxy/pkg/handlers/credentials"
	"reverseProxy/pkg/handlers/sites"
	"reverseProxy/pkg/logging"
	"time"
)

var (
	correctExit        = fmt.Errorf("finished succesfully")
	crudServerInit     = make(chan struct{})
	reverseProxyInit   = make(chan struct{})
	backendManagerInit = make(chan struct{})
)

func main() {
	loggers := logging.NewLogs("cmd", "main")
	loggers.GetInfo().Msg("start reverse proxy server")

	cache := &config.EnvCache{}
	if err := envconfig.Process("", cache); err != nil {
		loggers.GetError().Err(err).Msg("unable to parse the environment")
		panic(err)
	}

	if err := db.ConnManager.Connect(); err != nil {
		loggers.GetError().Str("when", "connect db").Msg("error connecting to the DB")
		panic(err)
	}

	defer db.ConnManager.Close()

	router := mux.NewRouter()
	router.HandleFunc("/credentials", credentials.Create).Methods("POST")
	router.HandleFunc("/credentials/{id:[0-9]+}", credentials.Read).Methods("GET")
	router.HandleFunc("/credentials/{id:[0-9]+}", credentials.Update).Methods("PUT")
	router.HandleFunc("/credentials/{id:[0-9]+}", credentials.Delete).Methods("DELETE")

	router.HandleFunc("/sites", sites.Create).Methods("POST")
	router.HandleFunc("/sites/{id:[0-9]+}", sites.Read).Methods("GET")
	router.HandleFunc("/sites/{id:[0-9]+}", sites.Update).Methods("PUT")
	router.HandleFunc("/sites/{id:[0-9]+}", sites.Delete).Methods("DELETE")

	router.HandleFunc("/backends", backends.Create).Methods("POST")
	router.HandleFunc("/backends/{id:[0-9]+}", backends.Read).Methods("GET")
	router.HandleFunc("/backends/{id:[0-9]+}", backends.Update).Methods("PUT")
	router.HandleFunc("/backends/{id:[0-9]+}", backends.Delete).Methods("DELETE")

	srvCRUD := http.Server{
		Addr:    cache.RouterPort,
		Handler: router,
	}

	reverseProxy := http.Server{
		Addr:    cache.RevPort,
		Handler: handler.RevHandler{},
	}

	ctx := context.TODO()

	errGroup, errGroupCtx := errgroup.WithContext(ctx)

	backendManager.BackendMgr = backendManager.NewBackendManager(errGroupCtx)

	errGroup.Go(func() error {
		interruptChan := make(chan os.Signal, 1)
		signal.Notify(interruptChan, os.Interrupt, os.Kill)
		select {
		case <-interruptChan:
			shutdownBaseCtx := context.TODO()
			shutdownCtx, cancel := context.WithTimeout(shutdownBaseCtx, 5*time.Second)
			defer cancel()

			if err := srvCRUD.Shutdown(shutdownCtx); err != nil {
				loggers.GetError().Str("server", "srvCRUD").Str("when", "received os signal").
					Msg("failed shutdown srvCRUD")
				panic(err)
			}
			if err := reverseProxy.Shutdown(shutdownCtx); err != nil {
				loggers.GetError().Str("server", "reverseProxy").
					Str("when", "received os signal").Msg("failed shutdown reverseProxy")
				panic(err)
			}
			return correctExit

		case <-errGroupCtx.Done():
			shutdownBaseCtx := context.TODO()
			shutdownCtx, cancel := context.WithTimeout(shutdownBaseCtx, 5*time.Second)
			defer cancel()

			if err := srvCRUD.Shutdown(shutdownCtx); err != nil {
				panic(err)
			}
			if err := reverseProxy.Shutdown(shutdownCtx); err != nil {
				panic(err)
			}
			return errGroupCtx.Err()
		}
	})

	errGroup.Go(func() error {
		loggers.GetInfo().Msg("start reverseProxy")
		close(reverseProxyInit)

		err := reverseProxy.ListenAndServe()
		if err != http.ErrServerClosed {
			loggers.GetError().Str("server", "reverseProxy").
				Str("when", "start reverseProxy").Msg("server closed")
			return err
		}
		return nil
	})

	errGroup.Go(func() error {
		loggers.GetInfo().Msg("start srvCRUD")
		close(crudServerInit)

		err := srvCRUD.ListenAndServe()
		if err != http.ErrServerClosed {
			loggers.GetError().Str("server", "srvCRUD").
				Str("when", "start srvCRUD").Msg("server closed")
			return err
		}
		return nil
	})

	errGroup.Go(func() error {
		loggers.GetInfo().Msg("start BackendManager")
		close(backendManagerInit)

		return backendManager.BackendMgr.Serve()
	})

	<-crudServerInit
	<-reverseProxyInit
	<-backendManagerInit
	zerolog.SetGlobalLevel(cache.GetLogLevel())

	if err := errGroup.Wait(); err != nil {
		if err == correctExit {
			return
		}
		loggers.GetError().Str("when", "have error from servers").Msg("exiting")
		os.Exit(1)
	}
}
