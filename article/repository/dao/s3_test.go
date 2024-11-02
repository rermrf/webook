package dao

import (
	"bytes"
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"testing"
	"time"
)

func TestS3(t *testing.T) {
	// 腾讯云中对标 s3 和 OSS 的产品叫做 COS
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
		Region:      aws.String("ap-guangzhou"),
		Endpoint:    aws.String("https://test-bucket-1258698140.cos.ap-guangzhou.myqcloud.com"),
		// 强制使用 /bucket/key 的形态
		S3ForcePathStyle: aws.Bool(true),
	})
	assert.NoError(t, err)
	client := s3.New(sess)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err = client.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:      aws.String("test-bucket-1258698140"),
		Key:         aws.String("126"),
		Body:        bytes.NewReader([]byte("测试内容 abc")),
		ContentType: aws.String("text/plain;charset=utf-8"),
	})
	assert.NoError(t, err)
	res, err := client.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: aws.String("test-bucket-1258698140"),
		Key:    aws.String("126"),
	})
	assert.NoError(t, err)
	data, err := io.ReadAll(res.Body)
	assert.NoError(t, err)
	t.Log(string(data))
}
