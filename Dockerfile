FROM golang:1.22 as build

RUN mkdir -p build

WORKDIR /build

COPY go.mod .
COPY go.sum .

RUN go mod download

WORKDIR shrt
COPY . .

RUN CGO_ENABLED=0 GOOS=linux make build
#RUN make build

FROM ubuntu:18.04
COPY --from=build build/shrt/bin/shrt .

EXPOSE 4000
ENTRYPOINT ["./shrt"]