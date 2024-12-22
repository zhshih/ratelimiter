# ratelimiter

test
Check Quota: curl "http://localhost:8080/rate/check?client_id=test-client"
Increment Quota: curl -X POST "http://localhost:8080/rate/increment?client_id=test-client"
Reset Quota: curl -X POST "http://localhost:8080/rate/reset?client_id=test-client"