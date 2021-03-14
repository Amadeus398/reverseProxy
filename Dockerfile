FROM golang:1.15
WORKDIR /reverseProxy
COPY . /reverseProxy
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go mod download && go build -o reverseProxy reverseProxy/cmd

FROM alpine:3.11
WORKDIR /reverseProxy
COPY --from=0 /reverseProxy/reverseProxy /bin
ENV HOST=127.0.0.1 PORT=5432 USER=username PASSWORD=password DBNAME=dbname SSLMODE=disable REVPORT=:777 ROUTERPORT=:80 LOGLEVEL=loglevel
ENTRYPOINT ["reverseProxy"]
