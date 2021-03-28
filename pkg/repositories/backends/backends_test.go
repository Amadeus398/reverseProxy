package backends

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

func (f fakeDbManager) Close() error {
	return nil
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
	case sqlUpdate:
		mockResult := sqlmock.NewResult(5, 0)
		if args[1] == int64(1) {
			mockResult = sqlmock.NewResult(5, 1)
		}
		mock.ExpectExec("^UPDATE backends SET .* WHERE .*;$").WillReturnResult(mockResult)
		result, err := dbMock.ExecContext(queryCtx, sqlUpdate, args[1])
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
		return nil

	case sqlDelete:
		mockResult := sqlmock.NewResult(5, 0)
		if args[0] == int64(1) {
			mockResult = sqlmock.NewResult(5, 1)
		}
		mock.ExpectExec("^DELETE FROM backends WHERE .*;$").WillReturnResult(mockResult)
		result, err := dbMock.ExecContext(queryCtx, sqlDelete, args[0])
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
	case sqlBackCreate:
		mockRow := mock.NewRows([]string{"id"})
		if args[0] == "127.0.0.1:80" && args[1] == int64(1) {
			mockRow.AddRow(int64(1))
		}
		mock.ExpectQuery("^INSERT INTO backends (.+) VALUES .* RETURNING id;$").WillReturnRows(mockRow)
		row := dbMock.QueryRow(sqlBackCreate, args[0], args[1])
		return row, func() {}, nil

	case sqlGet:
		mockRow := mock.NewRows([]string{"id", "address", "site_id", "site_name", "site_host"})
		if args[0] == int64(1) {
			mockRow.AddRow(int64(1), "127.0.0.1:80", int64(1), "vk", "vk.com")
		}

		mock.ExpectQuery("^SELECT (.+) FROM backends b JOIN sites s ON b.site_id = s.id WHERE .*;$").
			WillReturnRows(mockRow)
		row := dbMock.QueryRow(sqlGet, args[0])
		return row, func() {}, nil
	default:
		return nil, nil, fmt.Errorf("unrecognized sql query")
	}
}

func (f fakeDbManager) Query(query string, args ...interface{}) (*sql.Rows, func(), error) {
	panic("implement me")
}

func TestCreate(t *testing.T) {
	type args struct {
		b *Backend
	}
	db.ConnManager = fakeDbManager{}
	site := &sites.Site{
		Id:   1,
		Name: "vk",
		Host: "vk.com",
	}
	backend := &Backend{
		Id:      1,
		Address: "127.0.0.1:80",
		Site:    site,
	}
	backend1 := &Backend{
		Id:      1,
		Address: "127.0.0.1:80",
		Site: &sites.Site{
			Id:   2,
			Name: "vk",
			Host: "vk.com",
		},
	}
	backend2 := &Backend{
		Id:      1,
		Address: "127.0.0.1:80",
		Site:    &sites.Site{},
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "create new Backend",
			args:    args{b: backend},
			wantErr: false,
		},
		{
			name:    "create new backend with non-existence site_id",
			args:    args{b: backend1},
			wantErr: true,
		},
		{
			name:    "create new backend without site_id",
			args:    args{b: backend2},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Create(tt.args.b); (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	type args struct {
		b *Backend
	}
	db.ConnManager = fakeDbManager{}
	backend1 := &Backend{
		Id:      1,
		Address: "127.0.0.1:80",
		Site:    &sites.Site{},
	}
	backend2 := &Backend{
		Id:      3,
		Address: "127.0.0.1:80",
		Site:    &sites.Site{},
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "delete backend with existence id",
			args:    args{b: backend1},
			wantErr: false,
		},
		{
			name:    "delete backend with non-existence id",
			args:    args{b: backend2},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Delete(tt.args.b); (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRead(t *testing.T) {
	type args struct {
		b *Backend
	}
	db.ConnManager = fakeDbManager{}
	site := &sites.Site{
		Id:   1,
		Name: "vk",
		Host: "vk.com",
	}
	backend1 := &Backend{
		Id:      1,
		Address: "127.0.0.1:80",
		Site:    site,
	}
	backend2 := &Backend{
		Id:      3,
		Address: "127.0.0.1:80",
		Site:    site,
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "read existence backend",
			args:    args{b: backend1},
			wantErr: false,
		},
		{
			name:    "read backend with non-existence id",
			args:    args{b: backend2},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Read(tt.args.b); (err != nil) != tt.wantErr {
				t.Errorf("Read() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	type args struct {
		b *Backend
	}
	db.ConnManager = fakeDbManager{}
	backend1 := &Backend{
		Id:      1,
		Address: "127.0.0.1:80",
		Site:    &sites.Site{},
	}
	backend2 := &Backend{
		Id:      2,
		Address: "127.0.0.1:80",
		Site:    &sites.Site{},
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "update backend with existence id",
			args:    args{b: backend1},
			wantErr: false,
		},
		{
			name:    "update backend with non-existence id",
			args:    args{b: backend2},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Update(tt.args.b); (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
