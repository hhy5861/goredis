package goredis

import (
	"fmt"
	"github.com/go-redis/redis"
	"golang.org/x/net/context"
	"pkg.poizon.com/golang/go-common/logger"
	"time"
)

type (
	GoRedisConfig struct {
		Host           string        `yaml:"host"`
		Port           int           `yaml:"port"`
		Db             int           `yaml:"db"`
		Password       string        `yaml:"password"`
		PoolSize       int           `yaml:"poolSize"`
		MaxConnAge     time.Duration `yaml:"maxConnAge"`
		IdleTimeout    time.Duration `yaml:"idleTimeout"`
		ConnectTimeout time.Duration `yaml:"connectTimeout"`
		ReadTimeout    time.Duration `yaml:"readTimeout"`
		WriteTimeout   time.Duration `yaml:"writeTimeout"`
	}
)

var (
	client *redis.Client
)

func GetRedisClient(ctx context.Context) *redis.Client {
	return WrapRedisClient(ctx, client)
}

func NewRedisClient(cfg *GoRedisConfig) {
	client = redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.Db,
		PoolSize:     cfg.PoolSize,
		MaxConnAge:   cfg.MaxConnAge * time.Millisecond,
		IdleTimeout:  cfg.IdleTimeout * time.Millisecond,
		DialTimeout:  cfg.ConnectTimeout * time.Millisecond,
		ReadTimeout:  cfg.ReadTimeout * time.Millisecond,
		WriteTimeout: cfg.ReadTimeout * time.Millisecond,
	})

	_, err := client.Ping().Result()
	if err != nil {
		logger.Fatal(err, *cfg)
	}
}
