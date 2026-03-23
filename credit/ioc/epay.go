package ioc

import (
	openapiv1 "webook/api/proto/gen/openapi/v1"
	"webook/credit/api/epay"
	"webook/credit/repository"
	"webook/pkg/logger"
)

func InitEpayHandler(
	repo repository.CreditRepository,
	openapiCli openapiv1.OpenAPIServiceClient,
	l logger.LoggerV1,
) *epay.Handler {
	return epay.NewHandler(repo, openapiCli, l)
}
