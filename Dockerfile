FROM golang:1.16

WORKDIR /app
# 事前に go mod init xxx で go.mod ファイルを作成しておく
COPY go.mod .
COPY go.sum .

# go mod からパッケージをダウンロード
RUN go mod download

# /app にすべてのコードをコピー
COPY . .

# Live Reloading
RUN curl -sSfL https://raw.githubusercontent.com/cosmtrek/air/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
 
# エントリポイント(air)
CMD ["air"]

