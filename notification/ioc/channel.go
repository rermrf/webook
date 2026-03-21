package ioc

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/services/dysmsapi"
	"github.com/spf13/viper"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"

	"webook/notification/domain"
	"webook/notification/service/channel"
)

func InitChannelSenders(
	inApp *channel.InAppSender,
	smsSender *channel.SMSSender,
	email *channel.EmailSender,
) map[domain.Channel]channel.Sender {
	return map[domain.Channel]channel.Sender{
		domain.ChannelInApp: inApp,
		domain.ChannelSMS:   smsSender,
		domain.ChannelEmail: email,
	}
}

func InitSMSProvider() channel.SMSProvider {
	type Config struct {
		Provider string `yaml:"provider"`
		Aliyun   struct {
			AccessKeyId     string `yaml:"accessKeyId"`
			AccessKeySecret string `yaml:"accessKeySecret"`
			SignName        string `yaml:"signName"`
			Endpoint        string `yaml:"endpoint"`
		} `yaml:"aliyun"`
		Tencent struct {
			SecretId  string `yaml:"secretId"`
			SecretKey string `yaml:"secretKey"`
			AppId     string `yaml:"appId"`
			SignName  string `yaml:"signName"`
			Region    string `yaml:"region"`
		} `yaml:"tencent"`
	}
	var cfg Config
	err := viper.UnmarshalKey("sms", &cfg)
	if err != nil {
		panic(err)
	}

	switch cfg.Provider {
	case "tencent":
		credential := common.NewCredential(cfg.Tencent.SecretId, cfg.Tencent.SecretKey)
		cpf := profile.NewClientProfile()
		region := cfg.Tencent.Region
		if region == "" {
			region = "ap-guangzhou"
		}
		client, er := sms.NewClient(credential, region, cpf)
		if er != nil {
			panic(er)
		}
		return channel.NewTencentSMSProvider(client, cfg.Tencent.AppId, cfg.Tencent.SignName)
	default:
		// 默认使用阿里云
		endpoint := cfg.Aliyun.Endpoint
		if endpoint == "" {
			endpoint = "dysmsapi.aliyuncs.com"
		}
		client, er := dysmsapi.NewClientWithAccessKey("cn-hangzhou", cfg.Aliyun.AccessKeyId, cfg.Aliyun.AccessKeySecret)
		if er != nil {
			panic(er)
		}
		return channel.NewAliyunSMSProvider(client, cfg.Aliyun.SignName)
	}
}
