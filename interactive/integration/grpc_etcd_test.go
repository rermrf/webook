package integration

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"go.etcd.io/etcd/client/v3/naming/resolver"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"testing"
	"time"
	intrv1 "webook/api/proto/gen/intr/v1"
	igrpc "webook/interactive/grpc"
	"webook/pkg/netx"
)

type EtcdTestSuite struct {
	suite.Suite
	client *etcdv3.Client
}

func (s *EtcdTestSuite) SetupSuite() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	client, err := etcdv3.New(etcdv3.Config{
		Context:   ctx,
		Endpoints: []string{"localhost:12379"},
	})
	require.NoError(s.T(), err)
	s.client = client
}

func (s *EtcdTestSuite) TestServer() {
	l, err := net.Listen("tcp", ":8090")
	require.NoError(s.T(), err)

	// endpoints 是以服务为维度，一个服务一个 Manager
	em, err := endpoints.NewManager(s.client, "service/interactive")
	require.NoError(s.T(), err)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// key 是指这个实例的 key
	// 如果有 instance id， 用 instacne id，如果没有，使用本机 IP + 端口
	// 端口一般从配置文件读
	addr := netx.GetOutboundIP() + ":8090"
	key := "service/interactive/" + addr

	// 这个 ctx 是控制创建租约的时间
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	// ttl 是租期
	// 单位秒
	// 1/3 就开始续约
	var ttl int64 = 30
	leaseResp, err := s.client.Grant(ctx, ttl)
	require.NoError(s.T(), err)
	cancel()
	// 在这一步之前完成所有的启动的准备工作，包括缓存预加载之类的事情
	err = em.AddEndpoint(ctx, key, endpoints.Endpoint{
		Addr: addr,
	}, etcdv3.WithLease(leaseResp.ID))
	require.NoError(s.T(), err)

	kaCtx, kaCancel := context.WithCancel(context.Background())
	go func() {
		// 在这里操作续约
		ch, err := s.client.KeepAlive(kaCtx, leaseResp.ID)
		require.NoError(s.T(), err)
		for kaResp := range ch {
			// 正常就是打印一下 DEBUG 日志啥的
			s.T().Log(kaResp.String())
		}
	}()

	// 万一，我的注册信息有变动，怎么办？
	go func() {
		ticker := time.NewTicker(time.Second)
		for now := range ticker.C {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			// AddEndpoint 应该是一个覆盖的语义，如果你这边已经有了这个 key 了，就覆盖
			// upsert
			// 更新注册信息的时候也要带上 leaseId

			err = em.AddEndpoint(ctx, key, endpoints.Endpoint{
				Addr: addr,
				// Metadata 可用放分组信息，权重信息，机房信息
				// 以及动态判定负载的信息
				Metadata: now.String(),
			}, etcdv3.WithLease(leaseResp.ID))
			require.NoError(s.T(), err)
			cancel()
			//em.Update(ctx, []*endpoints.UpdateWithOpts{
			//	{
			//		Update: endpoints.Update{
			//			// Op 只有 Add，Delete
			//			Op:  endpoints.Add,
			//			Key: key,
			//			Endpoint: endpoints.Endpoint{
			//				Addr:     addr,
			//				Metadata: now.String(),
			//			},
			//		},
			//	},
			//})
		}
	}()

	server := grpc.NewServer()
	intrv1.RegisterInteractiveServiceServer(server, &igrpc.InteractiveServiceServer{})
	err = server.Serve(l)
	s.T().Log(err)
	// 退出：正常退出
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// 退出先要取消续约
	kaCancel()
	// 退出阶段，先从注册中心删除
	err = em.DeleteEndpoint(ctx, key)
	require.NoError(s.T(), err)
	server.GracefulStop()
}

func (s *EtcdTestSuite) TestClient() {
	bd, err := resolver.NewBuilder(s.client)
	require.NoError(s.T(), err)
	// URL 的规范 scheme:///xxxx
	cc, err := grpc.NewClient("etcd:///service/interactive",
		grpc.WithResolvers(bd),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	client := intrv1.NewInteractiveServiceClient(cc)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	res, err := client.Get(ctx, &intrv1.GetRequest{
		Biz:   "test",
		BizId: 1,
		Uid:   123,
	})
	require.NoError(s.T(), err)
	s.T().Log(res.GetIntr())

}

func TestEtcd(t *testing.T) {
	suite.Run(t, new(EtcdTestSuite))
}
