package coreddns

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go.etcd.io/etcd/clientv3"
)

type ReverseRecord struct {
	Host string `json:"host,omitempty"`
}

type ARecord struct {
	Host string `json:"host,omitempty"`
	TTL  int    `json:"ttl,omitempty"`
}

type Record struct {
	Name  string
	IPV4  string
	Rev   bool
	Lease int64
}

type Entry struct {
	record  Record
	leaseID clientv3.LeaseID
}

type EtcdUpdater struct {
	ctx     context.Context
	cli     *clientv3.Client
	kv      clientv3.KV
	domain  string
	prefix  string
	entries map[string]*Entry
}

func NewUpdater(servers []string, domain string) (*EtcdUpdater, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   servers,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	updater := &EtcdUpdater{
		ctx:     context.Background(),
		cli:     cli,
		kv:      cli.KV,
		domain:  domain,
		prefix:  "",
		entries: make(map[string]*Entry),
	}

	updater.prefix = "/skydns/" + updater.toARecordKey(domain)

	return updater, nil
}

func (u *EtcdUpdater) toARecordKey(name string) string {
	paths := strings.Split(strings.Trim(name, "."), ".")

	var sb strings.Builder
	sb.WriteString(u.prefix)
	for i := len(paths) - 1; i >= 0; i-- {
		sb.WriteString(paths[i])
		sb.WriteString("/")
	}

	return sb.String()
}

func (u *EtcdUpdater) toReverseRecordKey(ip string) string {
	paths := strings.Split(ip, ".")

	var sb strings.Builder
	sb.WriteString("/skydns/arpa/in-addr")
	for i := len(paths) - 1; i >= 0; i-- {
		sb.WriteString("/")
		sb.WriteString(paths[i])
	}

	return sb.String()
}

func (u *EtcdUpdater) SetARecord(name string, ip string, lease int64) error {
	key := u.toARecordKey(name)
	value, err := json.Marshal(ARecord{
		Host: ip,
		TTL:  60,
	})
	if err != nil {
		return err
	}

	if lease == 0 {
		_, err := u.kv.Put(u.ctx, key, string(value))
		return err
	} else {
		lease, err := u.cli.Grant(u.ctx, lease)
		if err != nil {
			return err
		}
		fmt.Println("Lease", name, ip, lease)
		_, err = u.kv.Put(u.ctx, key, string(value), clientv3.WithLease(lease.ID))
		return err
	}
}

func (u *EtcdUpdater) SetReverseRecord(name string, ip string, lease int64) error {
	key := u.toReverseRecordKey(ip)
	value, err := json.Marshal(ReverseRecord{
		Host: name + "." + u.domain,
	})
	if err != nil {
		return err
	}

	if lease == 0 {
		_, err := u.kv.Put(u.ctx, key, string(value))
		return err
	} else {
		lease, err := u.cli.Grant(u.ctx, lease)
		if err != nil {
			return err
		}
		_, err = u.kv.Put(u.ctx, key, string(value), clientv3.WithLease(lease.ID))
		return err
	}
}

func (u *EtcdUpdater) UnsetARecord(name string) error {
	key := u.toARecordKey(name)
	_, err := u.kv.Delete(u.ctx, key)
	return err
}

func (u *EtcdUpdater) UnsetReverseRecord(ip string) error {
	key := u.toReverseRecordKey(ip)
	_, err := u.kv.Delete(u.ctx, key)
	return err
}
