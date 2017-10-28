# tlstun
QUIC tunnel

# get free TLS certificates
https://letsencrypt.org

# start server on remote machine
```
quictun_server -addr :1443 -cert server.cert -key server.key -target 127.0.0.1:8888
```

# start client on local machine with domain
```
tlstun_client -addr :1083 -target quictun-server-domain.com:1443
```
# TODO
* client & server auth
* optimize bandwdith (compare with KCPTUN)