# SHRT

A tiny url service with postgres storage.

## Run
```
docker-compose up
```
Then
```
curl --url 'http://localhost:4000/register' --data '{"long":"https://www.google.fr/maps"}'
```
You should receive a json response of this format
```
{"short":"YjVTD7"}
```
You can then go to your navigator and request `http://localhost:4000/YjVTD7`

## Test
```
make test
```
You can `go test ./...` but it will skip postgres related test

## Limitations
This was done in a very short time so the scope of the features is limited