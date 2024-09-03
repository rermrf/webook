package config

type config struct {
	DB     DBConfig
	Redis  RedisConfig
	Server ServerPort
}

type ServerPort struct {
	HTTPPort string
}

type DBConfig struct {
	DSN string
}

type RedisConfig struct {
	Addr string
}
