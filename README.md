# clonos-ws

go run main.go -c channels.conf

# or + go.mod:

```
go mod init clonos-ws
go get github.com/gorilla/websocket
```

```
go build -o clonos-ws main.go
```

```
./clonos-ws -c channels.conf
```
