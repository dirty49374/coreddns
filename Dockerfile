################################################
## ETCD + COREDNS BUILDER
################################################
FROM bitnami/etcd AS etcd

RUN chmod 777 /opt/bitnami/etcd
RUN pwd
RUN wget -q https://github.com/coredns/coredns/releases/download/v1.5.0/coredns_1.5.0_linux_amd64.tgz
RUN tar zxf coredns_1.5.0_linux_amd64.tgz

################################################
## COREDDNS BUILDER
################################################
FROM golang:1.12 AS coreddns
WORKDIR /go/src/github.com/dirty49374/coreddns

COPY cmd/ cmd/
COPY pkg/ pkg/
COPY vendor/ vendor/
COPY main.go .
RUN mkdir -p /build/conf /build/data
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -a -installsuffix cgo -o /build/coreddns .

################################################
## STAGING
################################################
FROM scratch AS staging

WORKDIR /
COPY --from=coreddns build/ /
COPY --from=etcd /opt/bitnami/etcd/bin/etcd /bin/etcd
COPY --from=etcd /opt/bitnami/etcd/coredns /bin/coredns
COPY Corefile.tpl /Corefile.tpl

################################################
## COREDDNS
################################################
FROM scratch
COPY --from=staging / /

EXPOSE 53/UDP 2379 2380 12379
ENTRYPOINT [ "/coreddns" ]
