package authorizeManager

import (
	"reverseProxy/pkg/logging"
	"reverseProxy/pkg/repositories/credentials"
	"reverseProxy/pkg/repositories/sites"
)

type Authorize struct {
	log *logging.Logger
}

// NeedAuth determines whether authorization is
// required on the specified host
func (a *Authorize) NeedAuth(host string) (bool, error) {
	a.log = logging.NewLogs("authorizeManager", "needAuth")

	a.log.GetInfo().Msg("check the host in the database")
	auth, err := sites.Authorization(host)
	if err != nil {
		a.log.GetError().Str("when", "check the host in the database").
			Err(err).Msg("failed check the host in the database")
		return false, err
	}

	var needAuth bool
	if !auth {
		needAuth = false
		return needAuth, err
	}
	needAuth = true

	return needAuth, nil
}

// AuthorizeUser verifying user data on the
// specified host
func (a *Authorize) AuthorizeUser(login, password, host string) (bool, error) {
	a.log = logging.NewLogs("authorizeManager", "authorizeUser")

	a.log.GetInfo().Msg("verifying user data")
	user, err := credentials.AuthorizeUser(login, password, host)
	if err != nil {
		a.log.GetError().Str("when", "verifying user data").
			Err(err).Msg("failed verified user data")
		return false, err
	}

	var authUser bool
	if !user {
		authUser = false
		return authUser, err
	}
	authUser = true

	return authUser, nil
}
