package credentials

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"reverseProxy/pkg/db"
	"reverseProxy/pkg/repositories/sites"
	"testing"
	"time"
)

type fakeDbManager struct{}

func (f fakeDbManager) Connect(cfg db.DbConfig) error {
	return nil
}

func (f fakeDbManager) Close() {
	return
}

func (f fakeDbManager) Exec(query string, args ...interface{}) error {
	dbMock, mock, err := sqlmock.New()
	if err != nil {
		return err
	}
	defer func() {
		if err := dbMock.Close(); err != nil {
			return
		}
	}()

	ctx := context.TODO()
	queryCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	switch query {
	case sqlCredentialDelete:
		mockResult := sqlmock.NewResult(1, 0)
		if args[0] == int64(1) {
			mockResult = sqlmock.NewResult(1, 1)
		}
		mock.ExpectExec("^DELETE FROM credentials WHERE .*;$").WillReturnResult(mockResult)
		result, err := dbMock.ExecContext(queryCtx, sqlCredentialDelete, args[0])
		if err != nil {
			return err
		}
		row, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if row == 0 {
			return fmt.Errorf("no rows")
		}

	case sqlCredentialUpdate:
		mockResult := sqlmock.NewResult(5, 0)
		if args[2] == int64(1) {
			mockResult = sqlmock.NewResult(5, 1)
		}
		mock.ExpectExec("^UPDATE credentials SET .* WHERE .*;$").WillReturnResult(mockResult)
		result, err := dbMock.ExecContext(queryCtx, sqlCredentialUpdate, args[2])
		if err != nil {
			return err
		}
		row, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if row == 0 {
			return fmt.Errorf("no rows")
		}

	default:
		return fmt.Errorf("unrecognized sql query")
	}
	return nil
}

func (f fakeDbManager) QueryRow(query string, args ...interface{}) (*sql.Row, func(), error) {
	dbMock, mock, err := sqlmock.New()
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		if err := dbMock.Close(); err != nil {
			return
		}
	}()
	switch query {
	case sqlCredentialCreate:
		mockRows := sqlmock.NewRows([]string{"id"})
		if args[0] == "Petya" && args[1] == "pridurok" && args[2] == int64(2) {
			mockRows.AddRow(int64(1))
		}
		mock.ExpectQuery("^INSERT INTO (.+) VALUES .* RETURNING id;$").WillReturnRows(mockRows)
		row := dbMock.QueryRow(sqlCredentialCreate, args[0], args[1], args[2])
		return row, func() {}, err

	case sqlCredentialsGet:
		mockRows := sqlmock.NewRows([]string{"id", "login", "password", "site_id", "site_name", "site_host"})
		if args[0] == int64(1) {
			mockRows.AddRow(int64(1), "Petya", "pridurok", int64(2), "site", "example.com")
		}
		mock.ExpectQuery("^SELECT (.+) FROM credentials c JOIN sites s ON s.id=c.site_id WHERE .*;$").
			WillReturnRows(mockRows)
		row := dbMock.QueryRow(sqlCredentialsGet, args[0])
		return row, func() {}, nil

	case sqlFindCredential:
		mockRow := sqlmock.NewRows([]string{"count"})
		if args[0] == "Sasha" && args[1] == "blablabla" && args[2] == "vk.com" {
			mockRow.AddRow(int64(1))
		}
		mock.ExpectQuery("^SELECT (.+) FROM credentials c JOIN sites s ON s.id = c.site_id WHERE .*;$").WillReturnRows(mockRow)
		row := dbMock.QueryRow(sqlFindCredential, args[0], args[1], args[2])
		return row, func() {}, err
	default:
		return nil, nil, fmt.Errorf("unrecognized sql query")
	}
}

func (f fakeDbManager) Query(query string, args ...interface{}) (*sql.Rows, func(), error) {
	panic("implement me")
}

func TestAuthorizeUser(t *testing.T) {
	type args struct {
		login    string
		password string
		host     string
	}
	db.ConnManager = &fakeDbManager{}
	args1 := args{
		login:    "Sasha",
		password: "blablabla",
		host:     "vk.com",
	}
	args2 := args{
		login:    "Petya",
		password: "pridurok",
		host:     "vk.com",
	}
	args3 := args{
		login:    "Sasha",
		password: "blablabla",
		host:     "example.com",
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name:    "auth true user",
			args:    args1,
			want:    true,
			wantErr: false,
		},
		{
			name:    "auth false user with existence host",
			args:    args2,
			want:    false,
			wantErr: true,
		},
		{
			name:    "auth false user with existence login and password",
			args:    args3,
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AuthorizeUser(tt.args.login, tt.args.password, tt.args.host)
			if (err != nil) != tt.wantErr {
				t.Errorf("AuthorizeUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("AuthorizeUser() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateCredentials(t *testing.T) {
	type args struct {
		c *Credentials
	}
	db.ConnManager = fakeDbManager{}
	site := &sites.Site{
		Id:   2,
		Name: "hoho",
		Host: "example.com",
	}
	cred1 := &Credentials{
		Id:       1,
		Login:    "Petya",
		Password: "pridurok",
		Site:     site,
	}
	cred2 := &Credentials{
		Id:       2,
		Login:    "Sasha",
		Password: "blablabla",
		Site:     site,
	}
	cred3 := &Credentials{
		Id:       2,
		Login:    "Petya",
		Password: "",
		Site:     site,
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "create new cred",
			args:    args{c: cred1},
			wantErr: false,
		},
		{
			name:    "create new cred with existence id",
			args:    args{c: cred2},
			wantErr: true,
		},
		{
			name:    "create new cred without password",
			args:    args{c: cred3},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CreateCredentials(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("CreateCredentials() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeleteCredentials(t *testing.T) {
	type args struct {
		c *Credentials
	}
	db.ConnManager = fakeDbManager{}
	cred1 := &Credentials{
		Id:       1,
		Login:    "Petya",
		Password: "pridurok",
		Site:     &sites.Site{Id: 2},
	}
	cred2 := &Credentials{
		Id:       3,
		Login:    "Petya",
		Password: "pridurok",
		Site:     &sites.Site{Id: 5},
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "delete credentials with existence id",
			args:    args{c: cred1},
			wantErr: false,
		},
		{
			name:    "delete credentials without existence id",
			args:    args{c: cred2},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DeleteCredentials(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("DeleteCredentials() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetCredential(t *testing.T) {
	type args struct {
		c *Credentials
	}
	db.ConnManager = fakeDbManager{}
	site1 := &sites.Site{
		Id:   2,
		Name: "site",
		Host: "example.com",
	}
	cred1 := &Credentials{
		Id:       1,
		Login:    "Petya",
		Password: "pridurok",
		Site:     site1,
	}
	cred2 := &Credentials{
		Id:       2,
		Login:    "Sanya",
		Password: "amadeus",
		Site:     site1,
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "get credentials with existence id",
			args:    args{c: cred1},
			wantErr: false,
		},
		{
			name:    "get invalid credentials with existence id",
			args:    args{c: cred2},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := GetCredential(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("GetCredential() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUpdateCredentials(t *testing.T) {
	type args struct {
		c *Credentials
	}
	db.ConnManager = &fakeDbManager{}
	args1 := &Credentials{
		Id:       1,
		Login:    "Sasha",
		Password: "blabla",
		Site:     nil,
	}
	args3 := &Credentials{
		Id:       2,
		Login:    "Sasha",
		Password: "blabla",
		Site:     nil,
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "update true credential",
			args:    args{c: args1},
			wantErr: false,
		},
		{
			name:    "update false credential with existence login and password",
			args:    args{c: args3},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := UpdateCredentials(tt.args.c); (err != nil) != tt.wantErr {
				t.Errorf("UpdateCredentials() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
