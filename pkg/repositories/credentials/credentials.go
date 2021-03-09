package credentials

import (
	"database/sql"
	"fmt"
	"reverseProxy/pkg/db"
	"reverseProxy/pkg/repositories/sites"
)

type Credentials struct {
	Id       int64       `json:"id"`
	Login    string      `json:"login"`
	Password string      `json:"password"`
	Site     *sites.Site `json:"site"`
}

const (
	sqlCredentialCreate = "INSERT INTO credentials (login, password, site_id) VALUES ($1, $2, $3) RETURNING id;"
	sqlCredentialsGet   = "SELECT c.id, c.login, c.password, s.id, s.name, s.host FROM credentials c JOIN sites s ON s.id=c.site_id WHERE c.id=$1;"
	sqlCredentialUpdate = "UPDATE credentials SET login=$1, password=$2 WHERE id=$3;"
	sqlCredentialDelete = "DELETE FROM credentials WHERE id=$1;"
	sqlFindCredential   = "SELECT COUNT(c.*) FROM credentials c JOIN sites s ON s.id = c.site_id WHERE s.host=$3 AND c.login=$1 AND c.password=$2;"
)

var (
	ErrCredentialsNotFound = fmt.Errorf("credential not found")
)

// AuthorizeUser compares the received user data
// with the database data
func AuthorizeUser(login, password, host string) (bool, error) {
	row, cancel, err := db.ConnManager.QueryRow(sqlFindCredential, login, password, host)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, err
		}
		return false, err
	}
	defer cancel()
	var count int64
	if err := row.Scan(&count); err != nil {
		return false, err
	}
	if count == 1 {
		return true, nil
	}

	return false, nil
}

// CreateCredentials creates credentials data
func CreateCredentials(c *Credentials) error {
	row, cancel, err := db.ConnManager.QueryRow(sqlCredentialCreate, c.Login, c.Password, c.Site.Id)
	if err != nil {
		if err == db.ErrNothingDone {
			return ErrCredentialsNotFound
		}
		return err
	}
	defer cancel()
	if err := row.Scan(&c.Id); err != nil {
		return err
	}
	return nil
}

// GetCredential reads credentials data
func GetCredential(c *Credentials) error {
	row, cancel, err := db.ConnManager.QueryRow(sqlCredentialsGet, c.Id)
	if err != nil {
		return err
	}
	defer cancel()
	c.Site = &sites.Site{}
	if err := row.Scan(&c.Id, &c.Login, &c.Password, &c.Site.Id, &c.Site.Name, &c.Site.Host); err != nil {
		if err == sql.ErrNoRows {
			return ErrCredentialsNotFound
		}
		return err
	}
	return nil
}

// UpdateCredentials updates credentials data
func UpdateCredentials(c *Credentials) error {
	oldCredential := *c
	if err := GetCredential(&oldCredential); err != nil {
		if err == sql.ErrNoRows {
			return ErrCredentialsNotFound
		}
		return err
	}

	if c.Login == "" {
		c.Login = oldCredential.Login
	}
	if c.Password == "" {
		c.Password = oldCredential.Password
	}

	c.Site = oldCredential.Site

	if err := db.ConnManager.Exec(sqlCredentialUpdate, c.Login, c.Password, c.Id); err != nil {
		if err == db.ErrNothingDone {
			return ErrCredentialsNotFound
		}
		return err
	}

	return nil
}

// DeleteCredentials deletes credentials data
func DeleteCredentials(c *Credentials) error {
	if err := db.ConnManager.Exec(sqlCredentialDelete, c.Id); err != nil {
		if err == db.ErrNothingDone {
			return ErrCredentialsNotFound
		}
		return err
	}
	return nil
}
