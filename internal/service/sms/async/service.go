package async

//type SMSService struct {
//	svc  sms.Service
//	repo repository.SMSAsyncReqRepository
//}
//
//func NewSMSService() sms.Service {
//	return &SMSService{}
//}
//
//func (s *SMSService) StartAsync() {
//	go func() {
//		reqs := s.repo.Fin没发送出去的请求()
//		for _, req := range reqs {
//			// 在这里发送，并且控制重试
//		}
//	}()
//}
//
//func (s SMSService) Send(ctx context.Context, biz string, args []string, numbers ...string) error {
//	// 正常路径
//
//	// 异常路径
//	err := s.svc.Send(ctx, biz, args, numbers...)
//	if err != nil {
//		// 判定 是不是崩溃了
//
//		//if 崩溃了 {
//		//    s.repo.Store
//		//}
//	}
//}
