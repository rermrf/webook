package ioc

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/spf13/viper"
	"os"
)

func InitOss() *s3.S3 {
	cosId, ok := os.LookupEnv("COS_APP_ID")
	if !ok {
		panic("未找到COS_APP_ID环境变量")
	}
	cosKey, ok := os.LookupEnv("COS_APP_SECRET")
	if !ok {
		panic("未找到COS_APP_SECRET环境变量")
	}
	sess, err := session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials(cosId, cosKey, ""),
		Region:      aws.String(viper.GetString("oss.region")),
		Endpoint:    aws.String(viper.GetString("oss.endpoint")),
		// 强制使用 /bucket/key 的形态
		S3ForcePathStyle: aws.Bool(true),
	})
	if err != nil {
		panic(err)
	}
	client := s3.New(sess)
	return client
}
