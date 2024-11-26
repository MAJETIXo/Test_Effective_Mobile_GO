FROM golang:1.22.5
WORKDIR /app
COPY go.mod ./
COPY . .
RUN go mod tidy
RUN go get -u github.com/swaggo/swag/cmd/swag
RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN go get -u github.com/swaggo/http-swagger
RUN go get -u github.com/swaggo/swag
RUN go build -o server .
RUN swag init
EXPOSE 8000
CMD ["./server"]