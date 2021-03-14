# reverseProxy

reverseProxy is reverse proxy server with additional features:

- interaction with database

- user authorization

- balancing of endpoints

---

## Start

Before you start, configure the environment, where:

- Environment for start reverseProxy:

```
RevPort   string // port of reverseProxy server
Router    string // port of CRUDserver
Loglevel  string // loglevel to display logs
```


- Environment for start PostgreSQL server:

```
Host       string // host
Port       string // port 'recomended "5432"'
User       string // user login
Password   string // user password
Dbname     string // DB name
Sslmode    string // sslmode 
```

To get started, run
```
$ go run main.go
```
---

## Description
**Database** include 3 tables:
- _Credentials_
   - it stores a login, password and site_id of user;
- _Sites_
   - it stores a name and host of site;
- _Backends_
   - it stores addresses of site_host;
    
**AuthorizeManager** responsible for user _authorization_. When reverseProxy 
receives a request, AuthorizeManager checks the database to see if there 
are credentials for the specified host. If it finds at least one, it 
means that authorization required, and redirects the user to the authorization 
page. The user enters a username and password, which the AuthorizeManager 
checks with the database (Credentials). If the username and password match 
the user's data, the AuthorizeManager open access to the site.

**BackendManager** responsible for _backends_. It gets all addresses 
of each host from the database (Backends), puts them in the endpoint 
map (endpoints[string]*Client), syncs them every 20 seconds, and checks 
the connection with each client every 5 seconds. 
After the user has passed authorization, BackendManager goes through 
all the clients by the specified host, and randomly gives "alive" client 
with the current address. 

---
## Used libraries
[github.com/DATA-DOG/go-sqlmock][1]<br>
[github.com/gorilla/mux][2]<br>
[github.com/jackc/pgx/v4][3]<br>
[github.com/kelseyhightower/envconfig][4]<br>
[github.com/rs/zerolog][5]<br>


[1]: (https://github.com/DATA-DOG/go-sqlmock)
[2]: (https://github.com/gorilla/mux)
[3]: (https://github.com/jackc/pgx)
[4]: (https://github.com/kelseyhightower/envconfig)
[5]: (https://github.com/rs/zerolog)
