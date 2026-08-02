FROM docker.io/library/golang:1.26-alpine

WORKDIR /app

COPY go.mod go.sum* ./
RUN go mod download

EXPOSE 8080

CMD ["go", "run", "."]
