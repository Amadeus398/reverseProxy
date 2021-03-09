package authorizeManager

import (
	"database/sql"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"reverseProxy/pkg/db"
	"reverseProxy/pkg/logging"
	"testing"
)

const (
	sqlFindCredential     = "SELECT COUNT(c.*) FROM credentials c JOIN sites s ON s.id = c.site_id WHERE s.host=$3 AND c.login=$1 AND c.password=$2;"
	sqlNeedsAuthorization = "SELECT c.login, c.password, s.name FROM credentials c JOIN sites s ON s.id = c.site_id WHERE s.host = $1;"
)

type fakeDbManager struct{}

func (f fakeDbManager) Connect(cfg db.DbConfig) error {
	return nil
}

func (f fakeDbManager) Close() {
	return
}

func (f fakeDbManager) Exec(query string, args ...interface{}) error {
	panic("implement me")
}

func (f fakeDbManager) QueryRow(query string, args ...interface{}) (*sql.Row, func(), error) {
	dbMock, mock, err := sqlmock.New()
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		if err := dbMock.Close(); err != nil {
			panic(err)
		}
	}()

	switch query {
	case sqlFindCredential:
		mockRow := sqlmock.NewRows([]string{"count"})
		if args[0] == "Sasha" && args[1] == "hahaha" && args[2] == "vk.com" {
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
	dbMock, mock, err := sqlmock.New()
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		if err := dbMock.Close(); err != nil {
			panic(err)
		}
	}()

	switch query {
	case sqlNeedsAuthorization:
		mockRows := sqlmock.NewRows([]string{"log", "pass", "name"})
		if args[0] == "vk.com" {
			mockRows.AddRow("Sasha", "lalala", "vk")
		}
		mock.ExpectQuery("^SELECT (.+) FROM credentials c JOIN sites s .*;$").WillReturnRows(mockRows)
		rows, err := dbMock.Query(sqlNeedsAuthorization, args[0])
		if err != nil {
			return nil, nil, err
		}
		return rows, func() {}, nil
	default:
		return nil, nil, fmt.Errorf("unrecognized sql query")
	}
}

func TestAuthorize_AuthorizeUser(t *testing.T) {
	type fields struct {
		log *logging.Logger
	}
	type args struct {
		login    string
		password string
		host     string
	}
	db.ConnManager = fakeDbManager{}
	arg1 := args{
		login:    "Sasha",
		password: "hahaha",
		host:     "vk.com",
	}
	arg2 := args{
		login:    "Petya",
		password: "pridurok",
		host:     "vk.com",
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name:   "check true user",
			fields: fields{log: nil},
			args: args{
				login:    arg1.login,
				password: arg1.password,
				host:     arg1.host,
			},
			want:    true,
			wantErr: false,
		},
		{
			name:   "check false user",
			fields: fields{log: nil},
			args: args{
				login:    arg2.login,
				password: arg2.password,
				host:     arg2.host,
			},
			want:    false,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Authorize{
				log: tt.fields.log,
			}
			got, err := a.AuthorizeUser(tt.args.login, tt.args.password, tt.args.host)
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

func TestAuthorize_NeedAuth(t *testing.T) {
	type fields struct {
		log *logging.Logger
	}
	type args struct {
		host string
	}
	db.ConnManager = fakeDbManager{}
	host1 := "vk.com"
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr bool
	}{
		{
			name:    "need auth true",
			fields:  fields{log: nil},
			args:    args{host: host1},
			want:    true,
			wantErr: false,
		},
		{
			name:    "need auth false",
			fields:  fields{log: nil},
			args:    args{host: "odnoklassniki.ru"},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &Authorize{
				log: tt.fields.log,
			}
			got, err := a.NeedAuth(tt.args.host)
			if (err != nil) != tt.wantErr {
				t.Errorf("NeedAuth() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("NeedAuth() got = %v, want %v", got, tt.want)
			}
		})
	}
}
