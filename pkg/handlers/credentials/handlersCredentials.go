package credentials

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"reverseProxy/pkg/formatters"
	"reverseProxy/pkg/logging"
	"reverseProxy/pkg/repositories/credentials"
	"reverseProxy/pkg/repositories/sites"
	"strconv"
)

const resourceName = "credentials"

// Create creates credentials data
func Create(w http.ResponseWriter, r *http.Request) {
	log := logging.NewLogs("handlerCredentials", "Create")
	log.GetInfo().Str("when", "start processing request").Msg("start handler Create")
	w.Header().Set("Content-Type", "text/json; charset=utf-8")

	log.GetInfo().Msg("read request body")
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.GetError().Str("when", "read body").
			Err(err).Msg("unable to read body")
		panic(err)
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			log.GetError().Str("when", "close body").
				Err(err).Msg("unable to close body")
			panic(err)
		}
	}()

	credential := credentials.Credentials{}
	log.GetInfo().Msg("unmarshal request body")
	if err := json.Unmarshal(buf, &credential); err != nil {
		log.GetError().Str("when", "unmarshal request body").
			Err(err).Msg("unable to unmarshal request body")
		panic(err)
	}

	requestRawParams := make(map[string]json.RawMessage)
	log.GetInfo().Msg("unmarshal request raw params to get site_id")
	if err := json.Unmarshal(buf, &requestRawParams); err != nil {
		log.GetError().Str("when", "unmarshal request raw params").
			Err(err).Msg("unable to unmarshal request raw params")
		panic(err)
	}

	log.GetInfo().Msg("convert site_id to integer")
	siteId, err := strconv.Atoi(string(requestRawParams["site_id"]))
	if err != nil {
		log.GetError().Str("when", "convert site_id to integer").
			Err(err).Msg("unable to convert site_id")
		panic(err)
	}
	site := sites.Site{Id: int64(siteId)}
	log.GetInfo().Msg("get site with specified id")
	if err := sites.GetSite(&site); err != nil {
		if err == sites.ErrSiteNotFound {
			w.WriteHeader(http.StatusNotFound)
			if _, err := fmt.Fprint(w, "{}"); err != nil {
				log.GetError().Str("when", "get site").
					Str("when", "site not found").Str("when", "send response").
					Err(err).Msg("unable to send response")
				panic(err)
			}
			return
		}
		log.GetError().Str("when", "get site").
			Err(err).Msg("unable to get site")
		panic(err)
	}

	credential.Site = &site
	log.GetInfo().Msg("create credential")
	if err := credentials.CreateCredentials(&credential); err != nil {
		if err == credentials.ErrCredentialsNotFound {
			w.WriteHeader(http.StatusNotFound)
			if _, err := fmt.Fprint(w, "{}"); err != nil {
				log.GetError().Str("when", "create credential").
					Str("when", "credentials not found").Str("when", "send response").
					Err(err).Msg("unable to send response")
				panic(err)
			}
		}
		log.GetError().Str("when", "create credential").
			Err(err).Msg("failed to create credential")
		panic(err)
	}

	log.GetInfo().Msg("marshal created credential")
	bytes, err := json.Marshal(credential)
	if err != nil {
		log.GetError().Str("when", "marshal created credential").
			Err(err).Msg("unable marshal credential")
		panic(err)
	}

	log.GetInfo().Msg("send response created credential")
	if _, err := formatters.WriteJsonOp(w, string(bytes), resourceName, formatters.OpCreate); err != nil {
		log.GetError().Str("when", "send response create credential").
			Err(err).Msg("unable to send response")
		panic(err)
	}
	log.GetInfo().Msg("exiting handler Create")
}

// Read reads credentials data
func Read(w http.ResponseWriter, r *http.Request) {
	log := logging.NewLogs("handlersCredentials", "read")
	log.GetInfo().Str("when", "starting processing request").Msg("start handler Update")
	w.Header().Set("Content-Type", "text/json; charset=utf-8")

	log.GetInfo().Msg("get and convert id")
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		log.GetError().Str("when", "get and convert id").
			Err(err).Msg("failed to get and convert id")
		panic(err)
	}

	credential := credentials.Credentials{Id: int64(id)}
	log.GetInfo().Msg("start read credential with specified id")
	if err := credentials.GetCredential(&credential); err != nil {
		if err == credentials.ErrCredentialsNotFound {
			w.WriteHeader(http.StatusNotFound)
			if _, err := fmt.Fprint(w, "{}"); err != nil {
				log.GetError().Str("when", "read credential").
					Str("when", "credentials not found").Str("when", "send response").
					Err(err).Msg("unable to send response")
				panic(err)
			}
			return
		}
		log.GetError().Str("when", "read credential").
			Err(err).Msg("failed to read credential")
		panic(err)
	}

	log.GetInfo().Msg("marshal read credential")
	bytes, err := json.Marshal(credential)
	if err != nil {
		log.GetError().Str("when", "marshal read credential").
			Err(err).Msg("unable to marshal credential")
		panic(err)
	}

	log.GetInfo().Msg("send response read credential")
	if _, err := formatters.WriteJsonOp(w, string(bytes), resourceName, formatters.OpGet); err != nil {
		log.GetError().Str("when", "send response read credential").
			Err(err).Msg("unable to send response")
		panic(err)
	}
	log.GetInfo().Msg("exiting handler Read")
}

// Update updates credentials data
func Update(w http.ResponseWriter, r *http.Request) {
	log := logging.NewLogs("handlersCredentials", "update")
	log.GetInfo().Str("when", "starting processing request").Msg("start handler Update")
	w.Header().Set("Content-Type", "text/json; charset=utf-8")

	log.GetInfo().Msg("get and convert id")
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		log.GetError().Str("when", "get and convert id").
			Err(err).Msg("failed to get and convert id")
		panic(err)
	}

	credential := credentials.Credentials{Id: int64(id)}
	log.GetInfo().Msg("read request body")
	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.GetError().Str("when", "read request body").
			Err(err).Msg("unable to read body")
		panic(err)
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			log.GetError().Str("when", "close body").
				Err(err).Msg("unable to close body")
			panic(err)
		}
	}()

	log.GetInfo().Msg("unmarshal request body")
	if err := json.Unmarshal(buf, &credential); err != nil {
		log.GetError().Str("when", "unmarshal request body").
			Err(err).Msg("unable to unmarshal body ")
		panic(err)
	}

	log.GetInfo().Msg("update credential")
	if err := credentials.UpdateCredentials(&credential); err != nil {
		if err == credentials.ErrCredentialsNotFound {
			w.WriteHeader(http.StatusNotFound)
			if _, err := fmt.Fprint(w, "{}"); err != nil {
				log.GetError().Str("when", "update credential").
					Str("when", "credentials not found").Str("when", "send response").
					Err(err).Msg("unable to send response")
				panic(err)
			}
			return
		}
		log.GetError().Str("when", "update credential").
			Err(err).Msg("failed to update credential")
		panic(err)
	}

	log.GetInfo().Msg("marshal update credential")
	bytes, err := json.Marshal(credential)
	if err != nil {
		log.GetError().Str("when", "marshal update credential").
			Err(err).Msg("unable to marshal credential")
		panic(err)
	}

	log.GetInfo().Msg("send response with update credential")
	if _, err := formatters.WriteJsonOp(w, string(bytes), resourceName, formatters.OpUpdate); err != nil {
		log.GetError().Str("when", "send response with update backend").
			Err(err).Msg("unable to send response")
		panic(err)
	}
	log.GetInfo().Msg("exiting handler Update")
}

// Delete deletes credentials data
func Delete(w http.ResponseWriter, r *http.Request) {
	log := logging.NewLogs("handlersCredentials", "delete")
	log.GetInfo().Str("when", "start processing request").Msg("start handler Delete")
	w.Header().Set("Content-Type", "text/json; charset=utf-8")

	log.GetInfo().Msg("get and convert id")
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		log.GetError().Str("when", "get and convert id").
			Err(err).Msg("failed to get and convert id")
		panic(err)
	}

	credential := credentials.Credentials{Id: int64(id)}
	log.GetInfo().Msg("delete credential with specified id")
	if err := credentials.DeleteCredentials(&credential); err != nil {
		if err == credentials.ErrCredentialsNotFound {
			w.WriteHeader(http.StatusNotFound)
			if _, err := formatters.WriteJsonOp(w, "{}", resourceName, formatters.OpDelete); err != nil {
				log.GetError().Str("when", "delete credential").
					Str("when", "credential found").Str("when", "send response").
					Err(err).Msg("unable to send response")
				panic(err)
			}
			return
		}
		log.GetError().Str("when", "delete backend").
			Err(err).Msg("failed to delete backend")
		panic(err)
	}

	log.GetInfo().Msg("marshal credential")
	bytes, err := json.Marshal(&credential)
	if err != nil {
		log.GetError().Str("when", "marshal credential").
			Err(err).Msg("unable to marshal credential")
		panic(err)
	}

	log.GetInfo().Msg("send response deleted credential")
	if _, err := formatters.WriteJsonOp(w, string(bytes), resourceName, formatters.OpDelete); err != nil {
		log.GetError().Str("when", "send response deleted credential").
			Err(err).Msg("unable to send response")

		panic(err)
	}
	log.GetInfo().Msg("exiting handler Delete")
}
