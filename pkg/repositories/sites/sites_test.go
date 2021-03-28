package sites

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/DATA-DOG/go-sqlmock"
	"reverseProxy/pkg/db"
	"testing"
	"time"
)

type fakeConnManager struct {
}

func (f fakeConnManager) Connect(cfg db.DbConfig) error {
	return nil
}

func (f fakeConnManager) Close() error {
	return nil
}

func (f fakeConnManager) Exec(query string, args ...interface{}) error {
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
	case sqlSiteDelete:
		mockResult := sqlmock.NewResult(5, 0)
		if args[0] == int64(1) {
			mockResult = sqlmock.NewResult(5, 1)
		}
		mock.ExpectExec("^DELETE FROM sites WHERE .*;$").WillReturnResult(mockResult)
		result, err := dbMock.ExecContext(queryCtx, sqlSiteDelete, args[0])
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
	case sqlSiteUpdate:
		mockResult := sqlmock.NewResult(5, 0)
		if args[2] == int64(1) {
			mockResult = sqlmock.NewResult(5, 1)
		}
		mock.ExpectExec("^UPDATE sites SET (.+) WHERE .*;$").WillReturnResult(mockResult)
		result, err := dbMock.ExecContext(queryCtx, sqlSiteUpdate, args[2])
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

func (f fakeConnManager) QueryRow(query string, args ...interface{}) (*sql.Row, func(), error) {
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
	case sqlSiteCreate:
		mockRows := sqlmock.NewRows([]string{"id"})
		if args[0] == "vk" && args[1] == "vk.com" {
			mockRows.AddRow(int64(1))
		}
		mock.ExpectQuery("^INSERT INTO (.+) VALUES .* RETURNING id;$").WillReturnRows(mockRows)
		row := dbMock.QueryRow(sqlSiteCreate, args[0], args[1])
		return row, func() {}, nil

	case sqlSiteGet:
		mockRows := sqlmock.NewRows([]string{"id", "name", "host"})
		if args[0] == int64(1) {
			mockRows.AddRow(int64(1), "vk", "vk.com")
		}

		mock.ExpectQuery("^SELECT .+ FROM sites .*;$").WillReturnRows(mockRows)
		row := dbMock.QueryRow(sqlSiteGet, args[0])
		return row, func() {}, nil
	default:
		return nil, nil, fmt.Errorf("unrecognized sql query")
	}
}

func (f fakeConnManager) Query(query string, args ...interface{}) (*sql.Rows, func(), error) {
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
			mockRows = mockRows.AddRow("lala", "hahaha", "vk")
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

func TestAuthorization(t *testing.T) {
	type args struct {
		hostName string
	}
	db.ConnManager = fakeConnManager{}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name:    "check valid auth",
			args:    args{hostName: "vk.com"},
			want:    true,
			wantErr: false,
		},
		{
			name:    "check invalid auth",
			args:    args{hostName: "odnokassniki.com"},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Authorization(tt.args.hostName)
			if (err != nil) != tt.wantErr {
				t.Errorf("Authorization() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Authorization() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	type args struct {
		site *Site
	}
	db.ConnManager = fakeConnManager{}
	site1 := &Site{
		Id:   1,
		Name: "vk",
		Host: "vk.com",
	}
	site2 := &Site{
		Id:   1,
		Name: "haha",
		Host: "odnoklassniki.ru",
	}
	site3 := &Site{
		Id:   2,
		Name: "vk",
		Host: "odnoklassniki.ru",
	}
	site4 := &Site{
		Id:   2,
		Name: "odnoklassniki",
		Host: "vk.com",
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "create new site with new id",
			args:    args{site: site1},
			wantErr: false,
		},
		{
			name:    "create new site with existence id",
			args:    args{site: site2},
			wantErr: true,
		},
		{
			name:    "create new site with existence site name",
			args:    args{site: site3},
			wantErr: true,
		},
		{
			name:    "create new site with existence site host",
			args:    args{site: site4},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Create(tt.args.site); (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeleteSite(t *testing.T) {
	type args struct {
		id int64
	}
	db.ConnManager = fakeConnManager{}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "delete existence site id",
			args:    args{id: int64(1)},
			wantErr: false,
		},
		{
			name:    "delete non-existence site id",
			args:    args{id: int64(3)},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := DeleteSite(tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("DeleteSite() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetSite(t *testing.T) {
	type args struct {
		site *Site
	}
	db.ConnManager = fakeConnManager{}
	site1 := &Site{
		Id:   1,
		Name: "vk",
		Host: "vk.com",
	}
	site2 := &Site{
		Id:   2,
		Name: "haha",
		Host: "odnoklassniki.ru",
	}
	tests := []struct {
		name    string
		args    args
		want    *Site
		wantErr bool
	}{
		{
			name: "get site with existence site id",
			args: args{site: site1},
			want: &Site{
				Id:   site1.Id,
				Name: site1.Name,
				Host: site1.Host,
			},
			wantErr: false,
		},
		{
			name:    "get site with non-existence site id",
			args:    args{site: site2},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := GetSite(tt.args.site); (err != nil) != tt.wantErr {
				t.Errorf("GetSite() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.want != nil {
				if tt.args.site != nil {
					condition := tt.args.site.Name != tt.want.Name || tt.args.site.Host != tt.want.Host
					condition = condition || tt.args.site.Id != tt.want.Id
					if condition {
						t.Errorf("GetSite() sites does not match, got: %v, want: %v", tt.args.site, tt.want)
					}
				} else {
					t.Errorf("GetSite() no site got, got: %v, want: %v", tt.args.site, tt.want)
				}
			}
		})
	}
}

func TestUpdateSite(t *testing.T) {
	type args struct {
		site *Site
	}
	db.ConnManager = fakeConnManager{}
	site := &Site{
		Id:   1,
		Name: "vkhui",
		Host: "vk.com",
	}
	site1 := &Site{
		Id:   1,
		Name: "vkvk",
	}
	site2 := &Site{
		Id:   1,
		Host: "hren.com",
	}
	tests := []struct {
		name    string
		args    args
		want    *Site
		wantErr bool
	}{
		{
			name: "change site name with existent id",
			args: args{site: site},
			want: &Site{
				Id:   site.Id,
				Name: site.Name,
				Host: site.Host,
			},
			wantErr: false,
		},
		{
			name: "change site name without site host",
			args: args{site: site1},
			want: &Site{
				Id:   site1.Id,
				Name: site1.Name,
				Host: "vk.com",
			},
			wantErr: false,
		},
		{
			name: "change site host with no site name specified and non-existent id",
			args: args{site: &Site{
				Id:   3,
				Host: "vk.com",
			}},
			want:    nil,
			wantErr: true,
		},
		{
			name: "change site host with no site name specified",
			args: args{site: site2},
			want: &Site{
				Id:   site2.Id,
				Name: "vk",
				Host: site2.Host,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := UpdateSite(tt.args.site); (err != nil) != tt.wantErr {
				t.Errorf("UpdateSite() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.want != nil {
				if tt.args.site != nil {
					condition := tt.args.site.Id != tt.want.Id || tt.args.site.Name != tt.want.Name
					condition = condition || tt.args.site.Host != tt.want.Host
					if condition {
						t.Errorf("UpdateSite() want: %v, got: %v", tt.want, tt.args.site)
					}
				} else {
					t.Errorf("UpdateSite() want: %v, got: %v", tt.want, tt.args.site)
				}
			}
		})
	}
}
