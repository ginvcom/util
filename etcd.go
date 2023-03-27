package util

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type Client struct {
	cli *clientv3.Client
}

func NewEtcdCli(endpoints []string) (cli Client) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		logInstanceError(endpoints, err)
	}
	cli = Client{
		cli: client,
	}
	return cli
}

func (cli Client) Close() error {
	if cli.cli == nil {
		return errors.New("etcd bad cluster endpoints")
	}
	return cli.cli.Close()
}

func logInstanceError(endpoints []string, err error) {
	datasource := strings.Join(endpoints, "、")
	logx.Errorf("Error on getting etcd instance of %s: %v", datasource, err)
}

func (cli Client) Put(key, value string) error {
	if cli.cli == nil {
		return errors.New("etcd bad cluster endpoints")
	}
	timeout := 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	resp, err := cli.cli.Put(ctx, key, value)
	defer cancel()
	if err != nil {
		switch err {
		case context.Canceled:
			logx.Errorf("etcd ctx is canceled by another routine: %v", err)
		case context.DeadlineExceeded:
			logx.Errorf("etcd ctx is attached with a deadline is exceeded: %v", err)
		case rpctypes.ErrEmptyKey:
			logx.Errorf("etcd client-side error: %v", err)
		default:
			logx.Errorf("etcd bad cluster endpoints, which are not etcd servers: %v", err)
		}
		return err
	}
	if resp.PrevKv != nil {
		logx.Infof("etcd put key: %s, value: %s, prevValue: %s", key, value, resp.PrevKv.Value)
	}
	//
	// resp.PrevKv.Value
	return nil
}

func (cli Client) Get(key string) (value []byte, err error) {
	opts := []clientv3.OpOption{}
	resp, err := cli.GetWithOption(key, opts)

	if err != nil {
		return
	}

	if len(resp) == 0 {
		return nil, errors.New("etcd not exit key:" + key)
	}

	if content, ok := resp[0][key]; ok {
		value = content
		return
	}

	return nil, errors.New("etcd not exit key:" + key)
}

// 删除key
func (cli Client) Delete(key string) (err error) {
	if cli.cli == nil {
		return errors.New("etcd bad cluster endpoints")
	}
	timeout := 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	_, err = cli.cli.Delete(ctx, key)

	return
}

// 获取key前缀的kv
func (cli Client) GetByPrefixDesc(key string) (res []map[string][]byte, err error) {
	opts := []clientv3.OpOption{
		clientv3.WithPrefix(),
		clientv3.WithSort(clientv3.SortByCreateRevision, clientv3.SortDescend),
	}
	res, err = cli.GetWithOption(key, opts)

	return
}

func (cli Client) GetByPrefix(key string) (res []map[string][]byte, err error) {
	opts := []clientv3.OpOption{
		clientv3.WithPrefix(),
		clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend),
	}
	res, err = cli.GetWithOption(key, opts)

	return
}

func (cli Client) GetPageListByPrefix(key string) (res []map[string][]byte, err error) {
	opts := []clientv3.OpOption{
		clientv3.WithPrefix(),
		clientv3.WithSort(clientv3.SortByKey, clientv3.SortAscend),
	}
	res, err = cli.GetWithOption(key, opts)

	return
}

// 基于key降序获取key前缀的第一个key
func (cli Client) GetOneKeyByPrefix(keyPrefix string) (key string, err error) {
	opts := []clientv3.OpOption{
		clientv3.WithPrefix(),
		clientv3.WithKeysOnly(),
		clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend),
		clientv3.WithLimit(1),
	}
	resp, err := cli.GetWithOption(keyPrefix, opts)

	if err != nil {
		return
	}

	if len(resp) == 0 {
		return "", errors.New("etcd not exit prefix key:" + keyPrefix)
	}

	item := resp[0]
	for k := range item {
		key = k
		return
	}

	return "", errors.New("etcd not exit prefix key:" + keyPrefix)
}

// 基于key降序获取key前缀的第一个kv
func (cli Client) GetOneByPrefix(keyPrefix string) (key string, value []byte, err error) {
	opts := []clientv3.OpOption{
		clientv3.WithPrefix(),
		clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend),
		clientv3.WithLimit(1),
	}
	resp, err := cli.GetWithOption(keyPrefix, opts)

	if err != nil {
		return
	}

	if len(resp) == 0 {
		return "", nil, errors.New("etcd not exit prefix key:" + keyPrefix)
	}

	item := resp[0]
	for k, v := range item {
		key = k
		value = v
		return
	}

	return "", nil, errors.New("etcd not exit prefix key:" + keyPrefix)
}

// 获取key前缀的kv
func (cli Client) GetWithOption(key string, opts []clientv3.OpOption) (res []map[string][]byte, err error) {
	if cli.cli == nil {
		return nil, errors.New("etcd bad cluster endpoints")
	}
	timeout := 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	resp, err := cli.cli.Get(ctx, key, opts...)
	defer cancel()
	if err != nil {
		if clientv3.IsConnCanceled(err) {
			logx.Errorf("etcd gRPC client connection is closed")
			// gRPC client connection is closed
		} else {
			logx.Errorf("etcd get key %s error: %v", key, err)
		}
		return
	}

	for _, kvpair := range resp.Kvs {
		item := map[string][]byte{string(kvpair.Key): kvpair.Value}
		res = append(res, item)
	}
	return
}
