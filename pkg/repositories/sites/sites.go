package sites

import (
	"database/sql"
	"fmt"
	_ "github.com/jackc/pgx/v4/stdlib"
	"reverseProxy/pkg/db"
)

const (
	sqlSiteCreate         = "INSERT INTO sites (name, host) VALUES ($1, $2) RETURNING id;"
	sqlSiteGet            = "SELECT * FROM sites WHERE id=$1;"
	sqlSiteUpdate         = "UPDATE sites SET name=$1, host=$2 WHERE id=$3"
	sqlSiteDelete         = "DELETE FROM sites WHERE id=$1;"
	sqlNeedsAuthorization = "SELECT c.login, c.password, s.name FROM credentials c JOIN sites s ON s.id = c.site_id WHERE s.host = $1;"
)

var (
	ErrSiteNotFound = fmt.Errorf("site not found")
)

type Site struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
	Host string `json:"host"`
}

func Authorization(hostName string) (bool, error) {
	rows, cancel, err := db.ConnManager.Query(sqlNeedsAuthorization, hostName)
	if err != nil {
		return false, err
	}
	defer cancel()

	defer func() {
		if err := rows.Close(); err != nil {
			// TODO log error
			return
		}
	}()
	var log, pass, name string
	for rows.Next() {
		if err := rows.Scan(&log, &pass, &name); err != nil {
			return false, err
		}
		return true, nil
	}
	// здесь всегда true
	//  не тот запрос... надо выдернуть, есть ли совпадения, если нет, то авторизация не нужна

	return false, nil
}

func Create(site *Site) error {
	row, cancel, err := db.ConnManager.QueryRow(sqlSiteCreate, site.Name, site.Host)
	if err != nil {
		if err == db.ErrNothingDone {
			return ErrSiteNotFound
		}
		return err
	}
	defer cancel()
	if err := row.Scan(&site.Id); err != nil {
		// TODO error log
		return err
	}

	return nil
}

func GetSite(site *Site) error {
	row, cancel, err := db.ConnManager.QueryRow(sqlSiteGet, site.Id)
	if err != nil {
		// TODO error log
		panic(err)
	}
	defer cancel()
	if err := row.Scan(&site.Id, &site.Name, &site.Host); err != nil {
		if err == sql.ErrNoRows {
			return ErrSiteNotFound
		}
		// TODO error log
		return err
	}
	return nil
}

func DeleteSite(id int64) error {
	if err := db.ConnManager.Exec(sqlSiteDelete, id); err != nil {
		if err == db.ErrNothingDone {
			return ErrSiteNotFound
		}
		return err
	}
	return nil
}

func UpdateSite(site *Site) error {
	oldSite := *site
	if err := GetSite(&oldSite); err != nil {
		if err == sql.ErrNoRows {
			return ErrSiteNotFound
		}
		return err
	}
	if site.Name == "" {
		site.Name = oldSite.Name
	}
	if site.Host == "" {
		site.Host = oldSite.Host
	}

	if err := db.ConnManager.Exec(sqlSiteUpdate, site.Name, site.Host, site.Id); err != nil {
		if err == db.ErrNothingDone {
			return ErrSiteNotFound
		}
		return err
	}
	return nil
}
