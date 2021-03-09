# Jumpcloud Interview Assignment


## Run
```
# Run the server
go run .

# Run the unit tests
go test
```


## Sample endpoints
Post a given password by replacing `<password>`:
```
curl --data "password=<password>" http://localhost:8080/hash
# Returns the id 
1 # for the first request
42 # for the 42nd request
```

Get the hashed password (after waiting 5 seconds) by replacing the `<id>`:
```
curl http://localhost:8080/hash/<id>
# Returns the hash
<hash string>
```

Get statistics about the total number of requests and average time per request in microseconds:
```
curl http://localhost:8080/stats
# Returns something like
{"total": 1, "average": 123}
# where total is the total number of POST /hash requests
# and average is the average time per request in microseconds
```

Gracefully shutdown the server: 
```
curl http://localhost:8080/shutdown
```


## Next Steps
1. Improve unit testing isolation by stubbing out PasswordHash so tests interacting with it can just focus on calls and returns rather than retesting PasswordHash.
2. More unit tests for main.go
