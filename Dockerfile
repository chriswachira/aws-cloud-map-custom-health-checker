FROM golang:1.23.2-bullseye

WORKDIR /app

COPY . .

RUN go env -w GOPROXY=direct && \
    go mod download && \
    go build -o aws-cloud-map-health-checker

CMD [ "./aws-cloud-map-health-checker" ]
