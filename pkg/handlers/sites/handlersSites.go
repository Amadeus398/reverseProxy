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

// Create godoc
// @Swagger:operation POST /sites Create
// @Summary Create new site
// @Tags Sites
// @Description Create site
// @Accept json
// @Produce json
// @Param input body sites.Site true "site info"
// @Success 200 {integer} integer 1
// @Failure 404 {object} sites.Site sites.ErrSiteNotFound
// @Failure default {string} string "error"
// @Router /sites [post]
// Create creates new site
func Create(w http.ResponseWriter, r *http.Request) {
	log := logging.NewLogs("handlersSites", "Create")
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

	site := sites.Site{}
	log.GetInfo().Msg("unmarshal request body")
	if err := json.Unmarshal(buf, &site); err != nil {
		log.GetError().Str("when", "unmarshal request body").
			Err(err).Msg("unable to unmarshal request body")
		panic(err)
	}

	log.GetInfo().Msg("create site")
	if err := sites.Create(&site); err != nil {
		if err == sites.ErrSiteNotFound {
			w.WriteHeader(http.StatusNotFound)
			if _, err := fmt.Fprint(w, "{}"); err != nil {
				log.GetError().Str("when", "create site").
					Str("when", "sites not found").Str("when", "send response").
					Err(err).Msg("unable to send response")
				panic(err)
			}
			return
		}
		log.GetError().Str("when", "create site").
			Err(err).Msg("failed to create site")
		// TODO error internal
		panic(err)
	}

	log.GetInfo().Msg("marshal created site")
	bytes, err := json.Marshal(&site)
	if err != nil {
		log.GetError().Str("when", "marshal created site").
			Err(err).Msg("unable to marshal site")
		panic(err)
	}

	log.GetInfo().Msg("send response created site")
	if _, err := formatters.WriteJsonOp(w, string(bytes), resourceName, formatters.OpCreate); err != nil {
		log.GetError().Str("when", "send response created site").
			Err(err).Msg("unable to send response")
		panic(err)
	}
	log.GetInfo().Msg("exiting handler Create")
}

// Read godoc
// @Swagger:operation GET /sites/{id} Get
// @Summary Get site based on given id
// @Tags Sites
// @Description get site
// @Accept json
// @Produce json
// @Param id path integer true "site ID"
// @Success 200 {object} sites.Site
// @Failure 404 {object} sites.Site sites.ErrSiteNotFound
// @Router /sites/{id} [get]
// Read reads sites data
func Read(w http.ResponseWriter, r *http.Request) {
	log := logging.NewLogs("handlersSites", "Read")
	log.GetInfo().Str("when", "starting processing request").Msg("start handler Update")
	w.Header().Set("Content-Type", "text/json; charset=utf-8")

	log.GetInfo().Msg("get and convert id")
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		log.GetError().Str("when", "get and convert id").
			Err(err).Msg("failed to get and convert id")
		panic(err)
	}

	site := sites.Site{Id: int64(id)}
	log.GetInfo().Msg("start read site with specified id")
	if err := sites.GetSite(&site); err != nil {
		if err == sites.ErrSiteNotFound {
			w.WriteHeader(http.StatusNotFound)
			if _, err := fmt.Fprint(w, "{}"); err != nil {
				log.GetError().Str("when", "read site").
					Str("when", "site not found").Str("when", "send response").
					Err(err).Msg("unable to send response")
				panic(err)
			}
			return
		}
		log.GetError().Str("when", "read site").
			Err(err).Msg("failed to read site")
		panic(err)
	}

	log.GetInfo().Msg("marshal read site")
	bytes, err := json.Marshal(site)
	if err != nil {
		log.GetError().Str("when", "marshal read site").
			Err(err).Msg("unable to marshal site")
		panic(err)
	}

	log.GetInfo().Msg("send response read site")
	if _, err := formatters.WriteJsonOp(w, string(bytes), resourceName, formatters.OpGet); err != nil {
		log.GetError().Str("when", "send response read site").
			Err(err).Msg("unable to send response")
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
		log.GetError().Str("when", "get and convert id").
			Err(err).Msg("failed to get and convert id")
		panic(err)
	}

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

	site := sites.Site{Id: int64(id)}
	log.GetInfo().Msg("unmarshal request body")
	if err := json.Unmarshal(buf, &site); err != nil {
		log.GetError().Str("when", "unmarshal request body").
			Err(err).Msg("unable to unmarshal body")
		panic(err)
	}

	log.GetInfo().Msg("update site")
	if err := sites.UpdateSite(&site); err != nil {
		if err == sites.ErrSiteNotFound {
			w.WriteHeader(http.StatusNotFound)
			if _, err := fmt.Fprint(w, "{}"); err != nil {
				log.GetError().Str("when", "update site").
					Str("when", "site not found").Str("when", "send response").
					Err(err).Msg("unable to send response")
				panic(err)
			}
			return
		}
		log.GetError().Str("when", "update site").
			Err(err).Msg("failed to update site")
		panic(err)
	}

	log.GetInfo().Msg("marshal update site")
	bytes, err := json.Marshal(site)
	if err != nil {
		log.GetError().Str("when", "marshal update site").
			Err(err).Msg("unable to marshal site")
		panic(err)
	}

	log.GetInfo().Msg("send response update site")
	if _, err := formatters.WriteJsonOp(w, string(bytes), resourceName, formatters.OpUpdate); err != nil {
		log.GetError().Str("when", "send response update site").
			Err(err).Msg("unable to send response")
		panic(err)
	}
	log.GetInfo().Msg("exiting handler Update")
}

// Delete godoc
// @Swagger:operation DELETE /sites/{id} Delete
// @Summary Delete site based on given id
// @Tags Sites
// @Description delete site
// @Accept json
// @Produce json
// @Param id path integer true "site ID"
// @Success 200 {object} sites.Site
// @Failure 404 {object} sites.Site sites.ErrSiteNotFound
// @Router /sites/{id} [delete]
// Delete deletes sites
func Delete(w http.ResponseWriter, r *http.Request) {
	log := logging.NewLogs("handlersSites", "delete")
	log.GetInfo().Str("when", "start processing request").Msg("start handler Delete")
	w.Header().Set("Content-Type", "text/json; charset=utf-8")

	log.GetInfo().Msg("get and convert id")
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		log.GetError().Str("when", "get and convert id").
			Err(err).Msg("failed to get and convert id")
		panic(err)
	}

	log.GetInfo().Msg("delete site with specified id")
	if err := sites.DeleteSite(int64(id)); err != nil {
		if err == sites.ErrSiteNotFound {
			w.WriteHeader(http.StatusNotFound)
			if _, err := fmt.Fprint(w, "{}"); err != nil {
				log.GetError().Str("when", "delete site").
					Str("when", "sites not found").Str("when", "send response").
					Err(err).Msg("unable to send response")
				panic(err)
			}
			return
		}
		log.GetError().Str("when", "delete site").
			Err(err).Msg("failed to delete site")
		panic(err)
	}

	log.GetInfo().Msg("marshal deleted site")
	obj, err := json.Marshal(&sites.Site{Id: int64(id)})
	if err != nil {
		log.GetError().Str("when", "marshal deleted site").
			Err(err).Msg("unable to marshal site")
		panic(err)
	}

	log.GetInfo().Msg("send response deleted site")
	if _, err := formatters.WriteJsonOp(w, string(obj), resourceName, formatters.OpDelete); err != nil {
		log.GetError().Str("when", "send response deleted site").
			Err(err).Msg("unable to send response")
		panic(err)
	}
	log.GetInfo().Msg("exiting handler Delete")
}
