FROM golang:1.22.5-alpine

WORKDIR /app

COPY go.* ./

RUN go mod download

COPY . .

RUN go build -o main cmd/api/main.go

EXPOSE 3001

# Add wait-for-it script
ADD https://github.com/vishnubob/wait-for-it/raw/master/wait-for-it.sh /usr/local/bin/wait-for-it
RUN chmod +x /usr/local/bin/wait-for-it
RUN apk add --no-cache bash

CMD ["wait-for-it", "db:3306", "--", "./main"]