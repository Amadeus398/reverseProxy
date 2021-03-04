package sites

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"reverseProxy/pkg/formatters"
	"reverseProxy/pkg/logging"
	"reverseProxy/pkg/repositories/sites"
	"strconv"
)

const resourceName = "site"

func Create(w http.ResponseWriter, r *http.Request) {
	log := logging.NewLogs("handlersSites", "Create")
	log.GetInfo().Str("when", "start processing request").Msg("start handler Create")
	w.Header().Set("Content-Type", "text/json; charset=utf-8")

	log.GetInfo().Msg("read request body")
	buf, err := ioutil.ReadAll(r.Body)
	defer func() {
		if err := r.Body.Close(); err != nil {
			log.GetError().Str("when", "close body").Msg("unable to close body")
			panic(err)
		}
	}()
	if err != nil {
		log.GetError().Str("when", "read body").Msg("unable to read body")
		panic(err)
	}

	site := sites.Site{}
	log.GetInfo().Msg("unmarshal request body")
	if err := json.Unmarshal(buf, &site); err != nil {
		log.GetError().Str("when", "unmarshal request body").
			Msg("unable to unmarshal request body")
		panic(err)
	}

	log.GetInfo().Msg("create site")
	if err := sites.Create(&site); err != nil {
		if err == sites.ErrSiteNotFound {
			w.WriteHeader(http.StatusNotFound)
			if _, err := fmt.Fprint(w, "{}"); err != nil {
				log.GetError().Str("when", "create site").Str("when", "sites not found").
					Str("when", "send response").Msg("unable to send response")
				panic(err)
			}
			return
		}
		log.GetError().Str("when", "create site").Msg("failed to create site")
		panic(err)
	}

	log.GetInfo().Msg("marshal created site")
	bytes, err := json.Marshal(&site)
	if err != nil {
		log.GetError().Str("when", "marshal created site").Msg("unable to marshal site")
		panic(err)
	}

	log.GetInfo().Msg("send response created site")
	if _, err := formatters.WriteJsonOp(w, string(bytes), resourceName, formatters.OpCreate); err != nil {
		log.GetError().Str("when", "send response created site").
			Msg("unable to send response")
		panic(err)
	}
	log.GetInfo().Msg("exiting handler Create")
}

func Read(w http.ResponseWriter, r *http.Request) {
	log := logging.NewLogs("handlersSites", "Read")
	log.GetInfo().Str("when", "starting processing request").Msg("start handler Update")
	w.Header().Set("Content-Type", "text/json; charset=utf-8")

	log.GetInfo().Msg("get and convert id")
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		log.GetError().Str("when", "get and convert id").Msg("failed to get and convert id")
		panic(err)
	}

	site := sites.Site{Id: int64(id)}
	log.GetInfo().Msg("start read site with specified id")
	if err := sites.GetSite(&site); err != nil {
		if err == sites.ErrSiteNotFound {
			w.WriteHeader(http.StatusNotFound)
			if _, err := fmt.Fprint(w, "{}"); err != nil {
				log.GetError().Str("when", "read site").Str("when", "site not found").
					Str("when", "send response").Msg("unable to send response")
				panic(err)
			}
			return
		}
		log.GetError().Str("when", "read site").Msg("failed to read site")
		panic(err)
	}

	log.GetInfo().Msg("marshal read site")
	bytes, err := json.Marshal(site)
	if err != nil {
		log.GetError().Str("when", "marshal read site").Msg("unable to marshal site")
		panic(err)
	}

	log.GetInfo().Msg("send response read site")
	if _, err := formatters.WriteJsonOp(w, string(bytes), resourceName, formatters.OpGet); err != nil {
		log.GetError().Str("when", "send response read site").Msg("unable to send response")
		panic(err)
	}
	log.GetInfo().Msg("exiting handler Read")
}

func Update(w http.ResponseWriter, r *http.Request) {
	log := logging.NewLogs("handlersSites", "update")
	log.GetInfo().Str("when", "starting processing request").Msg("start handler Update")
	w.Header().Set("Content-Type", "text/json; charset=utf-8")

	log.GetInfo().Msg("get and convert id")
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		log.GetError().Str("when", "get and convert id").Msg("failed to get and convert id")
		panic(err)
	}

	log.GetInfo().Msg("read request body")
	buf, err := ioutil.ReadAll(r.Body)
	defer func() {
		if err := r.Body.Close(); err != nil {
			log.GetError().Str("when", "close body").Msg("unable to close body")
			panic(err)
		}
	}()
	if err != nil {
		log.GetError().Str("when", "read request body").Msg("unable to read body")
		panic(err)
	}

	site := sites.Site{Id: int64(id)}
	log.GetInfo().Msg("unmarshal request body")
	if err := json.Unmarshal(buf, &site); err != nil {
		log.GetError().Str("when", "unmarshal request body").Msg("unable to unmarshal body")
		panic(err)
	}

	log.GetInfo().Msg("update site")
	if err := sites.UpdateSite(&site); err != nil {
		if err == sites.ErrSiteNotFound {
			w.WriteHeader(http.StatusNotFound)
			if _, err := fmt.Fprint(w, "{}"); err != nil {
				log.GetError().Str("when", "update site").Str("when", "site not found").
					Str("when", "send response").Msg("unable to send response")
				panic(err)
			}
			return
		}
		log.GetError().Str("when", "update site").Msg("failed to update site")
		panic(err)
	}

	log.GetInfo().Msg("marshal update site")
	bytes, err := json.Marshal(site)
	if err != nil {
		log.GetError().Str("when", "marshal update site").Msg("unable to marshal site")
		panic(err)
	}

	log.GetInfo().Msg("send response update site")
	if _, err := formatters.WriteJsonOp(w, string(bytes), resourceName, formatters.OpUpdate); err != nil {
		log.GetError().Str("when", "send response update site").Msg("unable to send response")
		panic(err)
	}
	log.GetInfo().Msg("exiting handler Update")
}

func Delete(w http.ResponseWriter, r *http.Request) {
	log := logging.NewLogs("handlersSites", "delete")
	log.GetInfo().Str("when", "start processing request").Msg("start handler Delete")
	w.Header().Set("Content-Type", "text/json; charset=utf-8")

	log.GetInfo().Msg("get and convert id")
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		log.GetError().Str("when", "get and convert id").Msg("failed to get and convert id")
		panic(err)
	}

	log.GetInfo().Msg("delete site with specified id")
	if err := sites.DeleteSite(int64(id)); err != nil {
		if err == sites.ErrSiteNotFound {
			w.WriteHeader(http.StatusNotFound)
			if _, err := fmt.Fprint(w, "{}"); err != nil {
				log.GetError().Str("when", "delete site").Str("when", "sites not found").
					Str("when", "send response").Msg("unable to send response")
				panic(err)
			}
			return
		}
		log.GetError().Str("when", "delete site").Msg("failed to delete site")
		panic(err)
	}

	log.GetInfo().Msg("marshal deleted site")
	obj, err := json.Marshal(&sites.Site{Id: int64(id)})
	if err != nil {
		log.GetError().Str("when", "marshal deleted site").Msg("unable to marshal site")
		panic(err)
	}

	log.GetInfo().Msg("send response deleted site")
	if _, err := formatters.WriteJsonOp(w, string(obj), resourceName, formatters.OpDelete); err != nil {
		log.GetError().Str("when", "send response deleted site").Msg("unable to send response")
		panic(err)
	}
	log.GetInfo().Msg("exiting handler Delete")
}
