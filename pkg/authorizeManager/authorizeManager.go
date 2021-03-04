package authorizeManager

import (
	"reverseProxy/pkg/repositories/credentials"
	"reverseProxy/pkg/repositories/sites"
)

type Authorize struct {
	needAuth bool
	authUser bool
}

func (a *Authorize) NeedAuth(host string) (bool, error) {
	auth, err := sites.Authorization(host)
	if err != nil {
		return false, err
	}

	if !auth {
		a.needAuth = false
		return a.needAuth, nil
	}
	a.needAuth = true

	return a.needAuth, nil
}

func (a *Authorize) AuthorizeUser(login, password, host string) (bool, error) {
	user, err := credentials.AuthorizeUser(login, password, host)
	if err != nil {
		return false, err
	}

	if !user {
		a.authUser = false
		return a.authUser, nil
	}
	a.authUser = true

	return a.authUser, nil
}
