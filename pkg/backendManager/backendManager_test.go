package backendManager

import (
	"context"
	"reflect"
	"reverseProxy/pkg/logging"
	"reverseProxy/pkg/repositories/backends"
	"reverseProxy/pkg/repositories/sites"
	"sync"
	"testing"
	"time"
)

func TestBackendManager_CheckEndpoints(t *testing.T) {
	type fields struct {
		endPoints   map[string][]*Client
		tickBackend *time.Ticker
		tickDB      *time.Ticker
		ctx         context.Context
		mux         sync.RWMutex
		e           chan error
		log         *logging.Logger
	}
	tests := []struct {
		name   string
		fields fields
		wants  bool
	}{
		{
			name: "makes unavailable backend not alive",
			fields: fields{
				endPoints: map[string][]*Client{
					"proverka": []*Client{
						&Client{Address: "127.0.0.0:80", Alive: true},
					},
				},
			},
			wants: false,
		},
		{
			name: "makes nothing with alive client",
			fields: fields{
				endPoints: map[string][]*Client{
					"proverka": []*Client{
						&Client{Address: "93.184.216.34:80", Alive: true},
					},
				},
			},
			wants: true,
		},
		{
			name: "makes nothing with available dead client",
			fields: fields{
				endPoints: map[string][]*Client{
					"proverka": {
						{
							Address: "127.0.0.1:80",
							Alive:   false,
						},
					},
				},
			},
			wants: false,
		},
		{
			name: "makes available dead client alive",
			fields: fields{
				endPoints: map[string][]*Client{
					"proverka": {
						{
							Address: "93.184.216.34:80",
							Alive:   false,
						},
					},
				},
			},
			wants: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BackendManager{
				endPoints:   tt.fields.endPoints,
				tickBackend: tt.fields.tickBackend,
				tickDB:      tt.fields.tickDB,
				ctx:         tt.fields.ctx,
				mux:         tt.fields.mux,
				e:           tt.fields.e,
				log:         tt.fields.log,
			}
			b.CheckEndpoints()
			time.Sleep(3 * time.Second)
			if b.endPoints["proverka"][0].Alive != tt.wants {
				t.Errorf("test %s failed", tt.name)
			}
		})
	}
}

func TestBackendManager_GetClient(t *testing.T) {
	type fields struct {
		endPoints   map[string][]*Client
		tickBackend *time.Ticker
		tickDB      *time.Ticker
		ctx         context.Context
		mux         sync.RWMutex
		e           chan error
		log         *logging.Logger
	}
	type args struct {
		host string
	}
	clientExample1 := &Client{
		Alive:   true,
		Address: "1.2.3.4",
	}
	clientExample2 := &Client{
		Alive:   false,
		Address: "4.3.2.1",
	}
	clientVk1 := &Client{
		Address: "5.4.3.2",
		Alive:   true,
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Client
		wantErr bool
	}{
		{
			name: "get client available host",
			fields: fields{
				endPoints: map[string][]*Client{
					"example.com": {
						clientExample1,
						clientExample2,
					},
					"vk.com": {
						clientVk1,
					},
				},
			},
			args:    args{host: "example.com"},
			want:    clientExample1,
			wantErr: false,
		},
		{
			name: "error when no host",
			fields: fields{
				endPoints: map[string][]*Client{
					"example.com": {
						clientExample1,
						clientExample2,
					},
					"vk.com": {
						clientVk1,
					},
				},
			},
			args:    args{host: "odnoklassniki.ru"},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BackendManager{
				endPoints:   tt.fields.endPoints,
				tickBackend: tt.fields.tickBackend,
				tickDB:      tt.fields.tickDB,
				ctx:         tt.fields.ctx,
				mux:         tt.fields.mux,
				e:           tt.fields.e,
				log:         tt.fields.log,
			}
			got, err := b.GetClient(tt.args.host)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetClient() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBackendManager_syncHosts(t *testing.T) {
	type fields struct {
		endPoints   map[string][]*Client
		tickBackend *time.Ticker
		tickDB      *time.Ticker
		ctx         context.Context
		mux         sync.RWMutex
		e           chan error
		log         *logging.Logger
	}
	type args struct {
		endpoints []*backends.Backend
	}
	backends1 := []*backends.Backend{
		{
			Id:      1,
			Address: "1.2.3.4",
			Site: &sites.Site{
				Id:   1,
				Name: "example",
				Host: "example.com",
			},
		},
		{
			Id:      2,
			Address: "4.3.2.1",
			Site: &sites.Site{
				Id:   2,
				Name: "example",
				Host: "example.com",
			},
		},
	}
	clientExample1 := &Client{
		Alive:   true,
		Address: "1.2.3.4",
	}
	clientExample2 := &Client{
		Alive:   false,
		Address: "4.3.2.1",
	}
	clientVk1 := &Client{
		Address: "5.4.3.2",
		Alive:   true,
	}
	endpoint := map[string][]*Client{
		"example.com": {
			clientExample1,
			clientExample2,
		},
		"vk.com": {
			clientVk1,
		},
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[string][]*Client
		wantErr bool
	}{
		{
			name: "delete unused sites",
			fields: fields{
				endPoints: endpoint,
			},
			args: args{endpoints: backends1},
			want: map[string][]*Client{
				"example.com": {
					clientExample1,
					clientExample2,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BackendManager{
				endPoints:   tt.fields.endPoints,
				tickBackend: tt.fields.tickBackend,
				tickDB:      tt.fields.tickDB,
				ctx:         tt.fields.ctx,
				mux:         tt.fields.mux,
				e:           tt.fields.e,
				log:         tt.fields.log,
			}
			if err := b.syncHosts(tt.args.endpoints); (err != nil) != tt.wantErr {
				t.Errorf("syncHosts() error = %v, wantErr %v", err, tt.wantErr)
			}
			for host, clients := range tt.want {
				cl, ok := b.endPoints[host]
				if !ok {
					t.Errorf("hosts not in sync, want:\n%v\ngot:\n%v", tt.want, b.endPoints)
				}
				for _, client := range clients {
					match := false
					for _, c := range cl {
						if client.Address == c.Address {
							match = true
							break
						}
					}
					if !match {
						t.Errorf("hosts not in sync, want:\n%v\ngot:\n%v", tt.want, b.endPoints)
					}
				}
			}
			for host, clients := range b.endPoints {
				cl, ok := tt.want[host]
				if !ok {
					t.Errorf("hosts not in sync, want:\n%v\ngot:\n%v", tt.want, b.endPoints)
				}
				for _, client := range clients {
					match := false
					for _, c := range cl {
						if client.Address == c.Address {
							match = true
							break
						}
					}
					if !match {
						t.Errorf("hosts not in sync, want:\n%v\ngot:\n%v", tt.want, b.endPoints)
					}
				}
			}
		})
	}
}
