package ioc

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/spf13/viper"
	"webook/bff/handler"
)

type ossConfig struct {
	Endpoint     string `mapstructure:"endpoint"`
	AccessKey    string `mapstructure:"access_key"`
	SecretKey    string `mapstructure:"secret_key"`
	Region       string `mapstructure:"region"`
	UsePathStyle bool   `mapstructure:"use_path_style"`
	AvatarBucket string `mapstructure:"avatar_bucket"`
	CoverBucket  string `mapstructure:"cover_bucket"`
	ImageBucket  string `mapstructure:"image_bucket"`
}

func InitOSSClient() (*s3.S3, handler.OSSConfig) {
	var cfg ossConfig
	err := viper.UnmarshalKey("oss", &cfg)
	if err != nil {
		panic(err)
	}
	sess, err := session.NewSession(&aws.Config{
		Endpoint:         aws.String(cfg.Endpoint),
		Region:           aws.String(cfg.Region),
		Credentials:      credentials.NewStaticCredentials(cfg.AccessKey, cfg.SecretKey, ""),
		S3ForcePathStyle: aws.Bool(cfg.UsePathStyle),
	})
	if err != nil {
		panic(err)
	}
	return s3.New(sess), handler.OSSConfig{
		Endpoint:     cfg.Endpoint,
		AvatarBucket: cfg.AvatarBucket,
		CoverBucket:  cfg.CoverBucket,
		ImageBucket:  cfg.ImageBucket,
	}
}
