. {
    etcd {{ .Domain }} {
        path /skydns
        endpoint http://127.0.0.1:2379
        upstream
    }
    cache 60 {{ .Domain }}
    loadbalance
    forward . 1.1.1.1:53 8.8.8.8:53
}
