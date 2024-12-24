## Launch
```
SERVER_PORT={SERVER_PORT} RAFT_NODE_ID={NODE_ID} RAFT_PORT={RAFT_PORT} RAFT_VOL_DIR={VOL_DIR} go run cmd/main.go
```


## Raft
Join
```
curl --location --request POST 'localhost:{SERVER_PORT}/raft/join' \
--header 'Content-Type: application/json' \
--data-raw '{
	"node_id": {NODE_ID}, 
	"raft_address": {RAFT_ADDR}
}'
```
Remove
```
curl -X GET 'http://localhost:{SERVER_PORT}/raft/remove' \
--header 'Content-Type: application/json' \
--data-raw '{
	"node_id": {NODE_ID}, 
	"raft_address": {RAFT_ADDR}
}'
```
Stats
```
curl -X GET 'http://localhost:{SERVER_PORT}/raft/stats' \-H 'Cotent-Type: application/json'
```

## RateLimiter
Check Quota
```
curl -X GET "http://localhost:{SERVER_PORT}/rate/check?client_id={CLIENT_ID}"
```
Increment Quota
```
curl -X POST "http://localhost:{SERVER_PORT}/rate/increment?client_id={CLIENT_ID}"
```
Reset Quota
```
curl -X POST "http://localhost:{SERVER_PORT}/rate/reset?client_id={CLIENT_ID}"
```