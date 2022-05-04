FROM golang:1.18

RUN mkdir -p /usr/src/mino
RUN mkdir -p /usr/local/etc/mino

WORKDIR /usr/src/mino

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN VERSION=$(git describe --tags) \
    COMMIT=$(git rev-parse --short HEAD) \
    go build -v -o /usr/local/bin/mino \
    -ldflags="-s -w -X 'dxkite.cn/mino.Version=$VERSION' -X 'dxkite.cn/mino.Commit=$COMMIT'" ./cmd/mino

RUN mkdir /mino
WORKDIR /mino
VOLUME /mino

EXPOSE 1080
CMD ["mino"]
