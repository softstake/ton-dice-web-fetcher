FROM golang:1.14-alpine as builder

ARG GITHUB_TOKEN

RUN apk add --no-cache git

RUN git config --global url."https://${GITHUB_TOKEN}:@github.com/".insteadOf "https://github.com/"

RUN git clone https://github.com/tonradar/ton-api.git

WORKDIR /go/src/build
ADD . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o dice-fetcher ./cmd

FROM scratch
COPY --from=builder /go/src/build/dice-fetcher /app/
COPY --from=builder /go/src/build/trxlt.save.default /app/trxlt.save
WORKDIR /app

CMD ["./dice-fetcher"]