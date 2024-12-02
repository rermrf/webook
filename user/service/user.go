package service

import (
	"context"
	"errors"
	"webook/pkg/logger"
	"webook/user/domain"
	"webook/user/events"
	"webook/user/repository"

	"golang.org/x/crypto/bcrypt"
)

var ErrUserDuplicate = repository.ErrUserDuplicate
var ErrInvalidUserOrPassword = errors.New("账号/邮箱或密码不对")
var ErrUserNotFound = repository.ErrUserNotFound

//go:generate mockgen -source=./user.go -package=svcmocks -destination=mocks/user_mock.go
type UserService interface {
	SignUp(ctx context.Context, u domain.User) error
	Login(ctx context.Context, email string, password string) (domain.User, error)
	Profile(ctx context.Context, id int64) (domain.User, error)
	EditNoSensitive(ctx context.Context, user domain.User) error
	FindOrCreate(ctx context.Context, phone string) (domain.User, error)
	FindOrCreateByWechat(ctx context.Context, wechatInfo domain.WechatInfo) (domain.User, error)
}

type UserServiceImpl struct {
	repo     repository.UserRepository
	l        logger.LoggerV1
	producer events.Producer
}

func NewUserService(repo repository.UserRepository, l logger.LoggerV1, producer events.Producer) UserService {
	return &UserServiceImpl{
		repo:     repo,
		l:        l,
		producer: producer,
	}
}

//func NewUserServiceV1(repo repository.UserRepository, l *zap.Logger) UserService {
//	return &UserServiceImpl{
//		repo: repo,
//		// 预留扩展空间
//		//logger: zap.L(),
//	}
//}

func (svc *UserServiceImpl) SignUp(ctx context.Context, u domain.User) error {
	// 你要考虑加密放在哪
	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	// 然后就是，存起来
	err = svc.repo.Create(ctx, u)
	if err == nil {
		err := svc.producer.ProduceSyncEvent(ctx, events.SyncUserEvent{
			Id:       u.Id,
			Email:    u.Email,
			Phone:    u.Phone,
			Nickname: u.Nickname,
		})
		svc.l.Error("用户注册数据写入到 kafka 失败", logger.Error(err))
	}
	return err
}

func (svc *UserServiceImpl) Login(ctx context.Context, email string, password string) (domain.User, error) {
	// 先找用户
	u, err := svc.repo.FindByEmail(ctx, email)
	if errors.Is(err, repository.ErrUserNotFound) {
		return domain.User{}, ErrInvalidUserOrPassword
	}
	if err != nil {
		return domain.User{}, err
	}
	// 然后比对密码
	err = bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	if err != nil {
		// DEBUG日志
		return domain.User{}, ErrInvalidUserOrPassword
	}
	return u, nil
}

func (svc *UserServiceImpl) Profile(ctx context.Context, id int64) (domain.User, error) {
	u, err := svc.repo.FindById(ctx, id)
	return u, err
}

func (svc *UserServiceImpl) EditNoSensitive(ctx context.Context, user domain.User) error {
	return svc.repo.UpdateNoSensitiveById(ctx, user)
}

func (svc *UserServiceImpl) FindOrCreate(ctx context.Context, phone string) (domain.User, error) {
	user, err := svc.repo.FindByPhone(ctx, phone)
	// 判断有没有这个用户
	// 快路径
	if !errors.Is(err, repository.ErrUserNotFound) {
		// 绝大部分请求会进来这里
		// nil 会进入
		// 不为 ErrUserNotFound 也会进入
		return user, err
	}

	// phone 脱敏后拿出来
	//zap.L().Info("手机用户未注册", zap.String("phone", phone))
	//svc.logger.Info("手机用户未注册", zap.String("phone", phone))
	svc.l.Info("用户未注册", logger.String("phone", phone))

	// 在系统资源不足，触发降级之后，不执行慢路径了，优先服务已经注册的用户，防止系统崩溃
	//if ctx.Value("降级") == true {
	//	return domain.User{}, errors.New("系统降级了")
	//}
	// 慢路径
	// 用户不存在注册
	err = svc.repo.Create(ctx, domain.User{
		Phone: phone,
	})
	if err != nil && !errors.Is(err, repository.ErrUserDuplicate) {
		return domain.User{Phone: phone}, err
	}
	// 主从延迟问题
	return svc.repo.FindByPhone(ctx, phone)
}

func (svc *UserServiceImpl) FindOrCreateByWechat(ctx context.Context, info domain.WechatInfo) (domain.User, error) {
	user, err := svc.repo.FindByWechat(ctx, info.OpenId)
	if !errors.Is(err, repository.ErrUserNotFound) {
		return user, err
	}

	err = svc.repo.Create(ctx, domain.User{
		WechatInfo: info,
	})
	if err != nil && !errors.Is(err, repository.ErrUserDuplicate) {
		return user, err
	}
	return svc.repo.FindByWechat(ctx, info.OpenId)
}
