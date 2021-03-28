// package models stores
// structure models for
// swagger requests
package models

// SwagCredentials is the Credentials
// model for swagger requests
type SwagCredentials struct {
	Id       int64  `json:"id" example:"1" swaggerignore:"true"`
	Login    string `json:"login" example:"someLogin"`
	Password string `json:"password" example:"somePassword"`
	SiteId   int64  `json:"site_id" example:"1"`
}

// SwagBackends is the Backends
// model for swagger requests
type SwagBackends struct {
	Id      int64  `json:"id" example:"1" swaggerignore:"true"`
	Address string `json:"address" example:"127.0.0.1:80"`
	SiteId  int64  `json:"site_id" example:"1"`
}
