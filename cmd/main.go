package main

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
	httpSwagger "github.com/swaggo/http-swagger"
	"golang.org/x/sync/errgroup"
	"net/http"
	"os"
	"os/signal"
	_ "reverseProxy/docs"
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

// @Title CRUD server in reverseProxy
// @Version 1.0.0
// @Description How to work CRUD server in reverseProxy

// @Schemes http

// @Host localhost:80
// @BasePath /

var (
	correctExit = fmt.Errorf("finished succesfully")
)

type serverConfig interface {
	GetRevPort() string
	GetRouterPort() string
}

type loggerConfig interface {
	GetLogLevel() zerolog.Level
}

func main() {
	loggers := logging.NewLogs("cmd", "main")
	loggers.GetInfo().Msg("start reverse proxy server")

	reverseProxyInit := make(chan struct{})
	crudServerInit := make(chan struct{})
	backendManagerInit := make(chan struct{})

	cfg := &config.EnvCache{}
	if err := envconfig.Process("", cfg); err != nil {
		loggers.GetError().Err(err).Msg("unable to parse the environment")
		panic(err)
	}

	dbCfg := db.DbConfig(cfg)

	if err := db.ConnManager.Connect(dbCfg); err != nil {
		loggers.GetError().Str("when", "connect db").Err(err).Msg("error connecting to the DB")
		panic(err)
	}

	defer func() {
		if err := db.ConnManager.Close(); err != nil {
			loggers.GetError().Str("when", "close db").Err(err).Msg("unable close connection")
			panic(err)
		}
	}()

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

	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	srvCfg := serverConfig(cfg)
	srvCRUD := http.Server{
		Addr:         srvCfg.GetRouterPort(),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	reverseProxy := http.Server{
		Addr:         srvCfg.GetRevPort(),
		Handler:      handler.RevHandler{},
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
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
					Err(err).Msg("failed shutdown srvCRUD")
				panic(err)
			}
			if err := reverseProxy.Shutdown(shutdownCtx); err != nil {
				loggers.GetError().Str("server", "reverseProxy").
					Str("when", "received os signal").
					Err(err).Msg("failed shutdown reverseProxy")
				panic(err)
			}
			return correctExit

		case <-errGroupCtx.Done():
			shutdownBaseCtx := context.TODO()
			shutdownCtx, cancel := context.WithTimeout(shutdownBaseCtx, 5*time.Second)
			defer cancel()

			if err := srvCRUD.Shutdown(shutdownCtx); err != nil {
				loggers.GetError().Str("server", "srvCRUD").
					Str("when", "received context closure signal").Err(err).
					Msg("failed shutdown srvCRUD")
				panic(err)
			}
			if err := reverseProxy.Shutdown(shutdownCtx); err != nil {
				loggers.GetError().Str("server", "reverseProxy").
					Str("when", "received context closure signal").Err(err).
					Msg("failed shutdown reverseProxy")
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

	logCfg := loggerConfig(cfg)

	<-crudServerInit
	<-reverseProxyInit
	<-backendManagerInit

	zerolog.SetGlobalLevel(logCfg.GetLogLevel())

	if err := errGroup.Wait(); err != nil {
		if err == correctExit {
			return
		}
		loggers.GetError().Str("when", "have error from servers").Msg("exiting")
		os.Exit(1)
	}
}
