FROM golang:1.20.5 AS builder

WORKDIR /jals

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" .


FROM scratch AS runtime

WORKDIR /jals

COPY --from=builder /jals/jals /usr/bin/

EXPOSE 8080

ENTRYPOINT [ "jals" ]