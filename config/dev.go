//go:build !k8s

package config

var Config = config{
	DB: DBConfig{
		// 本地连接
		DSN: "root:root@tcp(localhost:13306)/webook?charset=utf8mb4&parseTime=True&loc=Local",
	},
	Redis: RedisConfig{
		// 本地连接
		Addr: "localhost:6379",
	},
	Server: ServerPort{":8081"},
}
