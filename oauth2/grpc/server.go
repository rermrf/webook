package grpc

import (
	"context"
	"google.golang.org/grpc"
	oauth2v1 "webook/api/proto/gen/oauth2/v1"
	"webook/oauth2/service/wechat"
)

type Oauth2ServiceServer struct {
	svc wechat.Service
	oauth2v1.UnimplementedOauth2ServiceServer
}

func NewOauth2ServiceServer(svc wechat.Service) *Oauth2ServiceServer {
	return &Oauth2ServiceServer{svc: svc}
}

func (o *Oauth2ServiceServer) Register(server *grpc.Server) {
	oauth2v1.RegisterOauth2ServiceServer(server, o)
}

func (o *Oauth2ServiceServer) AuthURL(ctx context.Context, request *oauth2v1.AuthURLRequest) (*oauth2v1.AuthURLResponse, error) {
	url, err := o.svc.AuthURL(ctx, request.GetState())
	return &oauth2v1.AuthURLResponse{Url: url}, err
}

func (o *Oauth2ServiceServer) VerifyCode(ctx context.Context, request *oauth2v1.VerifyCodeRequest) (*oauth2v1.VerifyCodeResponse, error) {
	info, err := o.svc.VerifyCode(ctx, request.GetCode(), request.State)
	if err != nil {
		return nil, err
	}
	return &oauth2v1.VerifyCodeResponse{
		OpenId:  info.OpenId,
		UnionId: info.UnionId,
	}, nil
}
