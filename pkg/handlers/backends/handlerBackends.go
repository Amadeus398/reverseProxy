package backends

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"reverseProxy/pkg/formatters"
	"reverseProxy/pkg/logging"
	"reverseProxy/pkg/repositories/backends"
	"reverseProxy/pkg/repositories/sites"
	"strconv"
)

const resourceName = "backends"

// Create creates backends data
func Create(w http.ResponseWriter, r *http.Request) {
	log := logging.NewLogs("handlerBackend", "create")
	log.GetInfo().Str("when", "starting processing request").Msg("start handler Create")

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
			return
		}
	}()

	backend := backends.Backend{}
	log.GetInfo().Msg("unmarshal request body")
	if err := json.Unmarshal(buf, &backend); err != nil {
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
			if _, err := formatters.WriteJsonOp(w, "{}", resourceName, formatters.OpCreate); err != nil {
				log.GetError().Str("when", "site not found").Str("when", "send response").
					Err(err).Msg("unable to send response")
				panic(err)
			}
			return
		}
		log.GetError().Str("when", "get site").Err(err).Msg("unable to get site")
		panic(err)
	}

	backend.Site = &site
	log.GetInfo().Msg("create backend")
	if err := backends.Create(&backend); err != nil {
		if err == backends.ErrBackendsNotFound {
			w.WriteHeader(http.StatusNotFound)
			if _, err := formatters.WriteJsonOp(w, "{}", resourceName, formatters.OpCreate); err != nil {
				log.GetError().Str("when", "backends not found").
					Str("when", "send response").
					Err(err).Msg("unable to send response")
				panic(err)
			}
			return
		}
		log.GetError().Str("when", "create backend").
			Err(err).Msg("failed to create backend")
		panic(err)
	}

	log.GetInfo().Msg("marshal created backend")
	bytes, err := json.Marshal(&backend)
	if err != nil {
		log.GetError().Str("when", "marshal created backend").
			Err(err).Msg("unable marshal created backend")
		panic(err)
	}

	log.GetInfo().Msg("send response created backend")
	if _, err := formatters.WriteJsonOp(w, string(bytes), resourceName, formatters.OpCreate); err != nil {
		log.GetError().Str("when", "send response created backend").
			Err(err).Msg("unable to send response")
		panic(err)
	}
	log.GetInfo().Msg("exiting handler Create")
}

// Read reads backends data
func Read(w http.ResponseWriter, r *http.Request) {
	log := logging.NewLogs("handlerBackends", "read")
	log.GetInfo().Str("when", "starting processing request").Msg("start handler Read")

	w.Header().Set("Content-Type", "text/json; charset=utf-8")
	log.GetInfo().Msg("get and convert id")
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		log.GetError().Str("when", "get and convert id").
			Err(err).Msg("failed get and convert id")
		panic(err)
	}
	backend := backends.Backend{Id: int64(id)}
	log.GetInfo().Msg("start read backend with specified id")
	if err := backends.Read(&backend); err != nil {
		if err == backends.ErrBackendsNotFound {
			w.WriteHeader(http.StatusNotFound)
			if _, err := formatters.WriteJsonOp(w, "{}", resourceName, formatters.OpGet); err != nil {
				log.GetError().Str("when", "read backend").
					Str("when", "backends not found").Str("when", "send response").
					Err(err).Msg("unable to send response")
				panic(err)
			}
			return
		}
		log.GetError().Str("when", "read backend").
			Err(err).Msg("failed to read backends")
		panic(err)
	}

	log.GetInfo().Msg("marshal read backend")
	bytes, err := json.Marshal(&backend)
	if err != nil {
		log.GetError().Str("when", "marshal read backend").
			Err(err).Msg("unable to marshal backend")
		panic(err)
	}

	log.GetInfo().Msg("send response read backend")
	if _, err := formatters.WriteJsonOp(w, string(bytes), resourceName, formatters.OpGet); err != nil {
		log.GetError().Str("when", "send response read backend").
			Err(err).Msg("unable to send response")
		panic(err)
	}
	log.GetInfo().Msg("exiting handler Read")
}

// Update updates backends data
func Update(w http.ResponseWriter, r *http.Request) {
	log := logging.NewLogs("handlerBackend", "update")
	log.GetInfo().Str("when", "starting processing request").Msg("start handler Update")
	w.Header().Set("Content-Type", "text/json; charset=utf-8")

	log.GetInfo().Msg("get and convert id")
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		log.GetError().Str("when", "get and convert id").
			Err(err).Msg("failed to get and convert id")
		panic(err)
	}
	backend := backends.Backend{Id: int64(id)}
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
	if err := json.Unmarshal(buf, &backend); err != nil {
		log.GetError().Str("when", "unmarshal request body").
			Err(err).Msg("unable to unmarshal body ")
		panic(err)
	}

	log.GetInfo().Msg("update backend")
	if err := backends.Update(&backend); err != nil {
		if err == backends.ErrBackendsNotFound {
			w.WriteHeader(http.StatusNotFound)
			if _, err := formatters.WriteJsonOp(w, "{}", resourceName, formatters.OpUpdate); err != nil {
				log.GetError().Str("when", "update backend").
					Str("when", "backends not found").Str("when", "send response").
					Err(err).Msg("unable to send response")
				panic(err)
			}
			return
		}
		log.GetError().Str("when", "update backend").
			Err(err).Msg("failed to update backend")
		panic(err)
	}

	log.GetInfo().Msg("marshal update backend")
	bytes, err := json.Marshal(&backend)
	if err != nil {
		log.GetError().Str("when", "marshal update backend").
			Err(err).Msg("unable to marshal backend")
		panic(err)
	}

	log.GetInfo().Msg("send response with update backend")
	if _, err := formatters.WriteJsonOp(w, string(bytes), resourceName, formatters.OpUpdate); err != nil {
		log.GetError().Str("when", "send response with update backend").
			Err(err).Msg("unable to send response")
		panic(err)
	}
	log.GetInfo().Msg("exiting handler Update")
}

// Delete deletes backends data
func Delete(w http.ResponseWriter, r *http.Request) {
	log := logging.NewLogs("handlerBackends", "delete")
	log.GetInfo().Str("when", "start processing request").Msg("start handler Delete")

	w.Header().Set("Content-Type", "text/json; charset=utf-8")

	log.GetInfo().Msg("get and convert id")
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		log.GetError().Str("when", "get and convert id").
			Err(err).Msg("failed to get and convert id")
		panic(err)
	}

	backend := backends.Backend{Id: int64(id)}
	log.GetInfo().Msg("delete backend with specified id")
	if err := backends.Delete(&backend); err != nil {
		if err == backends.ErrBackendsNotFound {
			w.WriteHeader(http.StatusNotFound)
			if _, err := formatters.WriteJsonOp(w, "{}", resourceName, formatters.OpDelete); err != nil {
				log.GetError().Str("when", "delete backend").
					Str("when", "backends not found").Str("when", "send response").
					Err(err).Msg("unable to send response")
				panic(err)
			}
			return
		}
		log.GetError().Str("when", "delete backend").
			Err(err).Msg("failed to delete backend")
		panic(err)
	}

	log.GetInfo().Msg("marshal backend")
	bytes, err := json.Marshal(&backend)
	if err != nil {
		log.GetError().Str("when", "marshal backend").
			Err(err).Msg("unable to marshal backend")
		panic(err)
	}

	log.GetInfo().Msg("send response deleted backend")
	if _, err := formatters.WriteJsonOp(w, string(bytes), resourceName, formatters.OpDelete); err != nil {
		log.GetError().Str("when", "send response deleted backend").
			Err(err).Msg("unable to send response")
		panic(err)
	}
	log.GetInfo().Msg("exiting handler Delete")
}
