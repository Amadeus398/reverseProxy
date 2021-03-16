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

## Description of the reverseProxy operation

After the database is filled, the reverseProxy is ready to work.
For example, we have a user in the database with a login and a password, 
registered on the site "example.com". Our reverseProxy will 
listen for requests on port 8080.

When the reverseProxy receives a request:
```
GET http://localhost:8080/
Host: example.com
Accept: text/json
Authorization: Basic ...=
```
it contacts the AuthorizeManager, that checks the database to see if there
are credentials for the specified host. If it finds at least one, it
means that authorization required, and redirects the user to the authorization
page. If the user entered incorrect data (or didn't enter the data), the 
reverseProxy sends a "status: unauthorized" response, with the status code 401:

```
HTTP/1.1 401 Unauthorized
Content-Type: text/json; charset=utf-8
Www-Authenticate: Basic realm=myProxy
Date: 
Content-Length: 26

{"status": "unauthorized"}

Response code: 401 (Unauthorized); Time: 38ms; Content length: 26 bytes
```

When the user entered the correct data(or authorization is not required), 
the reverseProxy contacts the BackendManager, 
that responsible for balancing the outgoing endpoints. It searches the database 
for the specified host and the addresses of clients on this host.

_**BackendManager** responsible for backends. It gets all addresses
of each host from the database (Backends), puts them in the endpoint
map (endpoints[string]*Client), syncs them every 20 seconds, and checks
the connection with each client every 5 seconds._

If no such host exists, the reverseProxy send a "service not found" response 
with the status code 502:

```
HTTP/1.1 502 Bad Gateway
Content-Type: text/json; charset=utf-8
Date: 
Content-Length: 32

{"message": "service not found"}

Response code: 502 (Bad Gateway); Time: 141ms; Content length: 32 bytes
```

If the BackendManager finds the host, but there are no clients (or the client 
is not "alive"), the reverseProxy sends a response "service unavailable" with 
the status code 503:

```
HTTP/1.1 503 Service Unavailable
Content-Type: text/json; charset=utf-8
Date: 
Content-Length: 34

{"message": "service unavailable"}

Response code: 503 (Service Unavailable); Time: 114ms; Content length: 34 bytes
```

When the BackendManager finds a host and "alive" client in the database, the 
reverseProxy sends a response with status code 200:

```
HTTP/1.1 200 OK
Age: 540015
Cache-Control: max-age=604800
Content-Type: text/html; charset=UTF-8
Date: 
...

<!doctype html>
<html>
    some body...
</html>


Response code: 200 (OK); Time: 321ms; Content length: 1256 bytes
```

---


## Used libraries

[github.com/DATA-DOG/go-sqlmock](https://github.com/DATA-DOG/go-sqlmock)

[github.com/gorilla/mux](https://github.com/gorilla/mux)

[github.com/jackc/pgx/v4](https://github.com/jackc/pgx)

[github.com/kelseyhightower/envconfig](https://github.com/kelseyhightower/envconfig)

[github.com/rs/zerolog](https://github.com/rs/zerolog)
