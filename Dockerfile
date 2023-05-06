FROM fedora:38

RUN dnf update -y && \
    dnf install -y git golang && \
    dnf clean all

ENV GOROOT /usr/lib/golang
ENV GOPATH /go
ENV GOPATH /go/bin:$PATH

RUN mkdir -p ${GOPATH}/src/github.com ${GOPATH}/bin

WORKDIR /ezw_api

COPY . .

RUN go build -o ezw_api .

EXPOSE 8080

CMD ["./ezw_api"]

