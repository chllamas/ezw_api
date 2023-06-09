FROM fedora:latest

RUN dnf update -y && \
    dnf install -y procps git golang net-tools && \
    dnf clean all

ENV GOROOT /usr/lib/golang
ENV GOPATH /go
ENV PATH /go/bin:$PATH

RUN mkdir -p ${GOPATH}/src/github.com/chllamas ${GOPATH}/bin

WORKDIR ${GOPATH}/src/github.com/chllamas/ezw_api

COPY . ${GOPATH}/src/github.com/chllamas/ezw_api

RUN go build -o /go/bin/ezw_server github.com/chllamas/ezw_api

EXPOSE 8000

CMD ["ezw_server"]

