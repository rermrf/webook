package ioc

import (
	"webook/notification/service/channel"
	"webook/notification/domain"
)

func InitChannelSenders(
	inApp *channel.InAppSender,
	sms *channel.SMSSender,
	email *channel.EmailSender,
) map[domain.Channel]channel.Sender {
	return map[domain.Channel]channel.Sender{
		domain.ChannelInApp: inApp,
		domain.ChannelSMS:   sms,
		domain.ChannelEmail: email,
	}
}

func InitSMSProvider() channel.SMSProvider {
	return channel.NewAliyunSMSProvider()
}
