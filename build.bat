FOR /F %%V IN ('git describe --tags') DO SET VERSION=%%V
FOR /F %%V IN ('git rev-parse --short HEAD') DO SET COMMIT=%%V
go build -o mino-%VERSION%.exe -ldflags="-s -w -H windowsgui -X 'dxkite.cn/mino.Version=%VERSION%' -X 'dxkite.cn/mino.Commit=%COMMIT%'" ./cmd/mino
