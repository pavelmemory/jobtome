## Test task: web-service to serve shorten URLs

### Local run

To run service locally all you need is to run
```bash
make build
```
to build binary for your platform (it requires gcc compiler because it uses SQLLite).  

Once it is completed you should have compiled binary `jobtome` in `./build/bin` directory.  
By default service uses local port `8080` to serve API requests and port `80` to serve redirects for short URLs.  
By default all data will be stored in `jobtome.dat` file.  
To start the service you need to run:
```bash
LOG_LEVEL=debug ./build/bin/jobtome
```
it will:
1. Instantiates SQLLite database instance and apply migration on top of it.
1. Run service acquiring ports `8080` and `80` to serve HTTP requests.

Once the service is ready you could try to check some basic info about it:
```bash
curl localhost:8080/-/version
```

To create a shorten please run:
```bash
curl -v -H 'Content-type: application/json' \
    -d '{"url": "https://google.com"}' \
    localhost:8080/api/shorten
```

To get newly created shorten:
```bash
curl -v localhost:8080/<Location>
```
where <Location> is the value returned in `Location` header of the previous operation result without leading slash.

To list set of shortens:
```bash
curl -v localhost:8080/api/shorten
```

To remove the shorten:
```bash
curl -v -X DELETE localhost:8080/<Location>
```

The flow described above is also available as an integration test that could be run by the command:
```bash
go test ./integration/...
```

### Not covered:

- a counter of the shortened url redirections
- API endpoint to read shortened url redirections count
- no metrics exported
- no proper README.md file with listing of configuration settings supported
- the lack of test for functionality
- caching of the shortens to reduce the load on the database
- authorization and authentication of incoming requests
- no OpenAPI specification of the endpoints
- not the best hashing function (MD5)
- ... etc.
