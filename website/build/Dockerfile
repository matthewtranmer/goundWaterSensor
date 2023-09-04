FROM golang:1.20

WORKDIR /website

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o website

EXPOSE 3500

CMD ["./website"]