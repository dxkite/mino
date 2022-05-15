FOR /F %%V IN ('git describe --tags') DO SET VERSION=%%V
FOR /F %%V IN ('git rev-parse --short HEAD') DO SET COMMIT=%%V

@echo "build client"
cd client
call npm install
call npm run build
cd ..

@echo "build x64"
SET FLAGS="-s -w -H windowsgui -X 'dxkite.cn/mino.Version=%VERSION%' -X 'dxkite.cn/mino.Commit=%COMMIT%'"
SET GOOS=windows
SET GOARCH=amd64
go build -o mino-%VERSION%-windows-amd64.exe -ldflags=%FLAGS% ./cmd/mino
7z a mino-%VERSION%-windows-amd64.exe.zip mino-%VERSION%-windows-amd64.exe

@echo "build x86"
SET GOOS=windows
SET GOARCH=386
go build -o mino-%VERSION%-windows-386.exe -ldflags=%FLAGS% ./cmd/mino
7z a mino-%VERSION%-windows-386.exe.zip mino-%VERSION%-windows-386.exe