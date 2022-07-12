FROM golang

COPY . /app

WORKDIR /app

ENV GOPROXY=https://proxy.golang.com.cn,direct \
    GO111MODULE=on

RUN go mod tidy \
    && go build main.go

CMD [ "./main" ]