package backends

import (
	"database/sql"
	"fmt"
	"reverseProxy/pkg/db"
	"reverseProxy/pkg/repositories/sites"
)

const (
	sqlBackCreate = "INSERT INTO backends (address, site_id) VALUES ($1, $2) RETURNING id;"
	sqlGet        = "SELECT b.id, b.address, s.id, s.name, s.host FROM backends b JOIN sites s ON b.site_id = s.id WHERE b.id = $1;"
	sqlUpdate     = "UPDATE backends SET address = $1 WHERE id = $2;"
	sqlDelete     = "DELETE FROM backends WHERE id = $1;"
	sqlList       = "SELECT b.id, b.address, s.id, s.name, s.host FROM backends b JOIN sites s on s.id = b.site_id;"
)

type Backend struct {
	Id      int64       `json:"id"`
	Address string      `json:"address"`
	Site    *sites.Site `json:"site"`
}

var ErrBackendsNotFound = fmt.Errorf("backend not found")

// Create creates backend data
func Create(b *Backend) error {
	row, cancel, err := db.ConnManager.QueryRow(sqlBackCreate, b.Address, b.Site.Id)
	if err != nil {
		return err
	}
	defer cancel()
	if err := row.Scan(&b.Id); err != nil {
		return err
	}
	return nil
}

// Read reads backend data
func Read(b *Backend) error {
	row, cancel, err := db.ConnManager.QueryRow(sqlGet, b.Id)
	if err != nil {
		return err
	}
	defer cancel()
	b.Site = &sites.Site{}
	if err := row.Scan(&b.Id, &b.Address, &b.Site.Id, &b.Site.Name, &b.Site.Host); err != nil {
		if err == sql.ErrNoRows {
			return ErrBackendsNotFound
		}
		return err
	}
	return nil
}

// Update updates backend data
func Update(b *Backend) error {
	oldBack := *b
	if err := Read(&oldBack); err != nil {
		if err == sql.ErrNoRows {
			return ErrBackendsNotFound
		}
	}
	if b.Address == "" {
		b.Address = oldBack.Address
	}
	b.Site = oldBack.Site

	if err := db.ConnManager.Exec(sqlUpdate, b.Address, b.Id); err != nil {
		if err == sql.ErrNoRows {
			return ErrBackendsNotFound
		}
		return err
	}
	return nil
}

// Delete deletes backend data
func Delete(b *Backend) error {
	if err := db.ConnManager.Exec(sqlDelete, b.Id); err != nil {
		if err == db.ErrNothingDone {
			return ErrBackendsNotFound
		}
		return err
	}
	return nil
}

// List returns all backends from database
func List() ([]*Backend, error) {
	backends := []*Backend{}
	rows, cancel, err := db.ConnManager.Query(sqlList)
	if err != nil {
		if err == sql.ErrNoRows {
			return backends, nil
		}
		return nil, err
	}
	defer cancel()

	for rows.Next() {
		backend := Backend{Site: &sites.Site{}}
		if err := rows.Scan(&backend.Id, &backend.Address, &backend.Site.Id, &backend.Site.Name, &backend.Site.Host); err != nil {
			return nil, err
		}
		backends = append(backends, &backend)
	}
	return backends, nil
}
