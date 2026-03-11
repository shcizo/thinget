FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY *.go ./
RUN CGO_ENABLED=0 go build -o /thinget .

FROM alpine:3.21
COPY --from=build /thinget /usr/local/bin/thinget
EXPOSE 5555
VOLUME /cache
CMD ["thinget"]
