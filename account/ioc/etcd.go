package ioc

import (
	"github.com/spf13/viper"
	etcdv3 "go.etcd.io/etcd/client/v3"
)

func InitEtcd() *etcdv3.Client {
	type Config struct {
		Addr string `yaml:"addr"`
	}
	var cfg Config
	if err := viper.UnmarshalKey("etcd", &cfg); err != nil {
		panic(err)
	}
	cli, err := etcdv3.NewFromURL(cfg.Addr)
	if err != nil {
		panic(err)
	}
	return cli
}
