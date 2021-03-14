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
REVPORT       string // port of reverseProxy server
ROUTERPORT    string // port of CRUDserver
LOGLEVEL      string // loglevel to display logs
```


- Environment for start PostgreSQL server:

```
HOST       string // host default value "127.0.0.1"
PORT       string // port default value "5432"
USER       string // user login
PASSWORD   string // user password
DBNAME     string // DB name
SSLMODE    string // sslmode default value "disabled"
```

To get started, run
```
$ go run main.go
```
---

## Description
**Database** includes 3 tables:
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

## Description of the CRUD server operation

In order for the reverseProxy to function correctly, you need to enter 
the data in the database tables.
```
Please note that the reverseProxy can only work with PostgreSQL, because 
the pgx driver is installed.
```

Table *Sites* stores a name and host of site, for example:

|  | id | name | host |
---|---:|:---|:---|
1| 1 | example | example.com|

Table *Backends* stores addresses of site_host, for example:

| | id | address | site_id |
---|---:|:---|:---|
1| 1|93.184.216.34:80| 1|

*Note that the site_id in the Backends corresponds to the id in the Sites*

Table *Credentials* stores a login, password and site_id of user, for example:

| | id | login | password | site_id |
---|---:|:---|:---|:---|
1| 1 | someUser | somePassword | 1|

*If there is at least 1 credential on the specified host, the reverseProxy 
requests authorization of this user.*

Let's look at the implementation of the CRUD handler using the example 
of working with the *Credentials* table.

To **create** a new credential, send a POST request:

```
POST http://localhost:8080/credentials
Content-Type: text/json; charset=utf-8

{
"login": "Vasya",
"password": "Ivanov",
"site_id": 1
}
```

To **read** the credential by id, send a GET request:

```
GET http://localhost:8080/credentials/1
Accept: text/json
```

To **update** a login or a password, send a PUT request:

```
PUT http://localhost:8080/credentials/1
Content-Type: text/json; charset=utf-8

{
"login": "Vasya",
"password": "Petrov"
}
```

To **delete** the credential by id, send a DELETE request:

```
DELETE http://localhost:8080/credentials/1
```

The implementation of the CRUD handler with tables of *Backends* 
and *Sites* follows the same principle.


---
## Used libraries

[github.com/DATA-DOG/go-sqlmock](https://github.com/DATA-DOG/go-sqlmock)

[github.com/gorilla/mux](https://github.com/gorilla/mux)

[github.com/jackc/pgx/v4](https://github.com/jackc/pgx)

[github.com/kelseyhightower/envconfig](https://github.com/kelseyhightower/envconfig)

[github.com/rs/zerolog](https://github.com/rs/zerolog)
