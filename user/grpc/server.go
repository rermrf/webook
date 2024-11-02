package grpc

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
	userv1 "webook/api/proto/gen/user/v1"
	"webook/user/domain"
	"webook/user/service"
)

type UserGRPCServer struct {
	svc service.UserService
	userv1.UnimplementedUserServiceServer
}

func NewUserGRPCServer(svc service.UserService) *UserGRPCServer {
	return &UserGRPCServer{svc: svc}
}

func (u *UserGRPCServer) Register(server *grpc.Server) {
	userv1.RegisterUserServiceServer(server, u)
}

func (u *UserGRPCServer) Signup(ctx context.Context, request *userv1.SignUpRequest) (*userv1.SignUpResponse, error) {
	err := u.svc.SignUp(ctx, u.toDTO(request.User))
	return &userv1.SignUpResponse{}, err
}

func (u *UserGRPCServer) Login(ctx context.Context, request *userv1.LoginRequest) (*userv1.LoginResponse, error) {
	user, err := u.svc.Login(ctx, request.GetEmail(), request.GetPassword())
	return &userv1.LoginResponse{
		User: u.toV(user),
	}, err
}

func (u *UserGRPCServer) Profile(ctx context.Context, request *userv1.ProfileRequest) (*userv1.ProfileResponse, error) {
	user, err := u.svc.Profile(ctx, request.GetId())
	return &userv1.ProfileResponse{
		User: u.toV(user),
	}, err
}

func (u *UserGRPCServer) EditNoSensitive(ctx context.Context, request *userv1.EditNoSensitiveRequest) (*userv1.EditNoSensitiveResponse, error) {
	err := u.svc.EditNoSensitive(ctx, u.toDTO(request.User))
	return &userv1.EditNoSensitiveResponse{}, err
}

func (u *UserGRPCServer) FindOrCreate(ctx context.Context, request *userv1.FindOrCreateRequest) (*userv1.FindOrCreateResponse, error) {
	user, err := u.svc.FindOrCreate(ctx, request.GetPhone())
	return &userv1.FindOrCreateResponse{
		User: u.toV(user),
	}, err
}

func (u *UserGRPCServer) FindOrCreateByWechat(ctx context.Context, request *userv1.FindOrCreateByWechatRequest) (*userv1.FindOrCreateByWechatResponse, error) {
	user, err := u.svc.FindOrCreateByWechat(ctx, domain.WechatInfo{
		OpenId:  request.GetInfo().GetOpenId(),
		UnionId: request.GetInfo().GetUnionId(),
	})
	return &userv1.FindOrCreateByWechatResponse{
		User: u.toV(user),
	}, err
}

func (u *UserGRPCServer) toDTO(user *userv1.User) domain.User {
	if user != nil {
		return domain.User{
			Id:       user.GetId(),
			Email:    user.GetEmail(),
			Nickname: user.GetNickName(),
			Phone:    user.GetPhone(),
			Password: user.GetPassword(),
			WechatInfo: domain.WechatInfo{
				OpenId:  user.GetWechatInfo().GetOpenId(),
				UnionId: user.GetWechatInfo().GetUnionId(),
			},
			AboutMe:  user.GetAboutMe(),
			Ctime:    user.GetCtime().AsTime(),
			Birthday: user.GetBirthday().AsTime(),
		}
	}
	return domain.User{}
}

func (u *UserGRPCServer) toV(user domain.User) *userv1.User {
	return &userv1.User{
		Id:       user.Id,
		Email:    user.Email,
		NickName: user.Nickname,
		Phone:    user.Phone,
		Password: user.Password,
		WechatInfo: &userv1.WechatInfo{
			OpenId:  user.WechatInfo.OpenId,
			UnionId: user.WechatInfo.UnionId,
		},
		AboutMe:  user.AboutMe,
		Ctime:    timestamppb.New(user.Ctime),
		Birthday: timestamppb.New(user.Birthday),
	}
}
