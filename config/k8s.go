//go:build k8s

package config

var Config = config{
	DB: DBConfig{
		// 本地连接
		DSN: "root:root@tcp(webook-mysql:13309)/webook?charset=utf8mb4&parseTime=True&loc=Local",
	},
	Redis: RedisConfig{
		// 本地连接
		Addr: "webook-redis:6379",
	},
	Server: ServerPort{HTTPPort: ":8080"},
}
