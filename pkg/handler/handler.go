package handler

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reverseProxy/pkg/authorizeManager"
	"reverseProxy/pkg/backendManager"
	"reverseProxy/pkg/logging"
)

const monkeys = "40000 тысяч обезьян в жопу сунули банан"

type RevHandler struct {
	log *logging.Logger
}

func (h RevHandler) getLogs() *logging.Logger {
	if h.log == nil {
		h.log = logging.NewLogs("handler", "serveHTTP")
	}
	return h.log
}

func (h RevHandler) sendAuthorizationQuery(w http.ResponseWriter) error {
	w.Header().Set("WWW-Authenticate", "Basic realm=myProxy")
	w.Header().Set("Content-Type", "text/json; charset=utf-8")
	w.WriteHeader(http.StatusUnauthorized)

	if _, err := fmt.Fprint(w, "{\"status\": \"unauthorized\"}"); err != nil {
		h.log.GetError().Str("when", "send authorization query").
			Str("when", "send response").Err(err).Msg("unable to send response")
		return err
	}
	return nil
}

func (h RevHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.getLogs().GetInfo().Str("when", "start processing request").
		Str("url", r.RequestURI).Msg("start RevHandler")

	host := r.Host
	h.getLogs().GetInfo().Msg("verifying authorization requirements")
	needAuth, err := authorizeManager.AuthorizeMnr.NeedAuth(host)
	if err != nil {
		h.getLogs().GetError().Str("when", "verifying authorization requirement").
			Err(err).Msg("failed verifying")
		panic(err)
	}

	if needAuth {
		h.getLogs().GetInfo().Msg("authorization required, user account verification required")
		login, password, ok := r.BasicAuth()
		if !ok {
			h.getLogs().GetInfo().Msg("basic authorization")
			if err := h.sendAuthorizationQuery(w); err != nil {
				h.getLogs().GetError().Str("when", "send basic authorization query").
					Err(err).Msg("unable to authorization query")
				panic(err)
			}
			return
		}

		h.getLogs().GetInfo().Msg("checking the user's data in the database")
		authorized, err := authorizeManager.AuthorizeMnr.AuthorizeUser(login, password, host)
		if err != nil {
			h.getLogs().GetError().Str("when", "checking the user's data").
				Err(err).Msg("failed check user's data")
			panic(err)
		}
		if !authorized {
			h.getLogs().GetWarn().Str("when", "entering user data").Msg("invalid user data")
			h.getLogs().GetInfo().Msg("re-attempt to enter user data")
			if err := h.sendAuthorizationQuery(w); err != nil {
				h.getLogs().GetError().Str("when", "re-attempt to enter user data").
					Err(err).Msg("unable to authorization query")
				panic(err)
			}
			return
		}
	}

	h.getLogs().GetInfo().Msg("get client for specified host")
	client, err := backendManager.BackendMgr.GetClient(host)
	if err != nil {
		switch err {
		case backendManager.ErrNoHost:
			w.Header().Set("Content-Type", "text/json; charset=utf-8")
			w.WriteHeader(http.StatusBadGateway)
			if _, err := fmt.Fprint(w, "{\"message\": \"service not found\"}"); err != nil {
				h.getLogs().GetError().Str("when", "get client").
					Str("when", "no hosts").Str("when", "send response").
					Err(err).Msg("unable to send response")
				panic(err)
			}
			return
		case backendManager.ErrClientNotFound:
			w.Header().Set("Content-Type", "text/json; charset=utf-8")
			w.WriteHeader(http.StatusServiceUnavailable)
			if _, err := fmt.Fprint(w, "{\"message\": \"service unavailable\"}"); err != nil {
				h.getLogs().GetError().Str("when", "get client").
					Str("when", "no clients").Str("when", "send response").
					Err(err).Msg("unable to send response")
				panic(err)
			}
			return
		default:
			w.Header().Set("Content-Type", "text/plain; text/json; charset=utf-8")
			w.WriteHeader(http.StatusInternalServerError)
			if _, err := fmt.Fprint(w, monkeys); err != nil {
				h.getLogs().GetError().Str("when", "get client").
					Str("when", "send response").Err(err).Msg("unable to send response")
				panic(err)
			}
			return
		}
	}

	h.getLogs().GetInfo().Msg("completed request, start response")
	ctx := context.TODO()
	req := r.Clone(ctx)
	req.URL, err = url.Parse("http://" + client.Address + req.RequestURI)
	req.RequestURI = ""
	if err != nil {
		h.log.GetError().Str("when", "parse raw url into url structure").
			Err(err).Msg("unable to parse raw url")
		panic(err)
	}
	req.Header.Del("Authorization")

	h.getLogs().GetInfo().Msg("start send HTTP request")
	resp, err := client.Cl.Do(req)
	if err != nil {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		bytesMonkeys := []byte(monkeys)
		w.Header().Set("Content-Length", fmt.Sprintf("%d", len(bytesMonkeys)))
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write(bytesMonkeys); err != nil {
			h.getLogs().GetError().Str("when", "completed request, start response").
				Str("when", "send response").Err(err).Msg("unable to send response")
			return
		}
		h.getLogs().GetError().Str("when", "completed request, start response").
			Str("url", req.RequestURI).Err(err).Msg("unable to get response")
		return
	}

	h.getLogs().GetInfo().Msg("read response body")
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		h.getLogs().GetError().Str("when", "read response body").
			Err(err).Msg("unable to read body")
		panic(err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			h.getLogs().GetError().Str("when", "close body").
				Err(err).Msg("unable to close body")
			panic(err)
		}
	}()

	h.getLogs().GetInfo().Msg("set headers")
	for header, headerVal := range resp.Header {
		for _, headerValue := range headerVal {
			w.Header().Set(header, headerValue)
		}
	}

	resp.Header.Del("Authorization")

	w.WriteHeader(resp.StatusCode)
	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		h.getLogs().GetWarn().Str("when", "status code").Msg("4xx")
	} else if resp.StatusCode > 500 {
		h.getLogs().GetWarn().Str("when", "status code").Msg("5xx")
	}

	h.getLogs().GetInfo().Msg("write response body")
	if _, err := w.Write(body); err != nil {
		h.getLogs().GetError().Str("when", "write response body").
			Err(err).Msg("unable to write body")
		panic(err)
	}
	h.getLogs().GetInfo().Msg("response complete")
}
