# clonos-ws

## Build

```
go run main.go -c channels.conf
```

 or + go.mod:

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

## Install

FreeBSD:
```
cp -a clonos-channel.conf /usr/local/etc/clonos-channel.conf
cp -a files/rc.d/clonos-ws /usr/local/etc/rc.d/clonos-ws
service clonos-ws enable
service clonos-ws start
```

Linux:
```
cp -a clonos-channel.conf /etc/clonos-channel.conf
cp -a files/systemd/clonos-ws.service /etc/systemd/system/clonos-ws.service
systemctl daemon-reload
systemctl enable clonos-ws
systemctl start clonos-ws
```
