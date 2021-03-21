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
	sqlSiteUpdate         = "UPDATE sites SET name=$1, host=$2 WHERE id=$3;"
	sqlSiteDelete         = "DELETE FROM sites WHERE id=$1;"
	sqlNeedsAuthorization = "SELECT c.login, c.password, s.name FROM credentials c JOIN sites s ON s.id = c.site_id WHERE s.host = $1;"
)

var (
	ErrSiteNotFound = fmt.Errorf("site not found")
)

type Site struct {
	Id   int64  `json:"id" example:"1" swaggerignore:"true"`
	Name string `json:"name" example:"site"`
	Host string `json:"host" example:"site.com"`
}

// Authorization checks the received host
// in the database
func Authorization(hostName string) (bool, error) {
	rows, cancel, err := db.ConnManager.Query(sqlNeedsAuthorization, hostName)
	if err != nil {
		return false, err
	}
	defer cancel()

	defer func() {
		if err := rows.Close(); err != nil {
			fmt.Errorf("")
		}
	}()
	var log, pass, name string
	for rows.Next() {
		if err := rows.Scan(&log, &pass, &name); err != nil {
			return false, err
		}
		return true, nil
	}

	return false, nil
}

// Create creates site data
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
		return err
	}

	return nil
}

// GetSite reads site data
func GetSite(site *Site) error {
	row, cancel, err := db.ConnManager.QueryRow(sqlSiteGet, site.Id)
	if err != nil {
		return err
	}
	defer cancel()
	if err := row.Scan(&site.Id, &site.Name, &site.Host); err != nil {
		if err == sql.ErrNoRows {
			return ErrSiteNotFound
		}
		return err
	}
	return nil
}

// DeleteSite deletes site data
func DeleteSite(id int64) error {
	if err := db.ConnManager.Exec(sqlSiteDelete, id); err != nil {
		if err == db.ErrNothingDone {
			return ErrSiteNotFound
		}
		return err
	}
	return nil
}

// UpdateSite update site data
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
