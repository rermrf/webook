package grpcx

import (
	"context"
	etcdv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
	"google.golang.org/grpc"
	"net"
	"strconv"
	"time"
	"webook/pkg/logger"
	"webook/pkg/netx"
)

type Server struct {
	*grpc.Server
	Port      int
	EtcdAddrs []string
	Name      string
	L         logger.LoggerV1
	kaCancel  func()
	em        endpoints.Manager
	client    *etcdv3.Client
	key       string
}

func (s *Server) Serve() error {
	l, err := net.Listen("tcp", ":"+strconv.Itoa(s.Port))

	if err != nil {
		return err
	}
	err = s.register()
	if err != nil {
		return err
	}
	// 这边就是直接启动，现在要嵌入服务注册过程
	return s.Server.Serve(l)
}

func (s *Server) register() error {
	client, err := etcdv3.New(etcdv3.Config{
		Endpoints: s.EtcdAddrs,
	})
	if err != nil {
		return err
	}
	s.client = client

	// endpoints 是以服务为维度，一个服务一个 Manager
	em, err := endpoints.NewManager(client, "service/"+s.Name)
	if err != nil {
		return err
	}

	// key 是指这个实例的 key
	// 如果有 instance id， 用 instance id，如果没有，使用本机 IP + 端口
	// 端口一般从配置文件读
	addr := netx.GetOutboundIP() + ":" + strconv.Itoa(s.Port)
	s.L.Info("service will register on：" + addr)
	s.key = "service/" + s.Name + "/" + addr

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// 这个 ctx 是控制创建租约的时间
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// 做成配置
	var ttl int64 = 30
	leaseResp, err := client.Grant(ctx, ttl)
	if err != nil {
		return err
	}
	// 在这一步之前完成所有的启动的准备工作，包括缓存预加载之类的事情
	err = em.AddEndpoint(ctx, s.key, endpoints.Endpoint{
		Addr: addr,
	}, etcdv3.WithLease(leaseResp.ID))
	if err != nil {
		return err
	}

	kaCtx, kaCancel := context.WithCancel(context.Background())
	s.kaCancel = kaCancel
	ch, err := client.KeepAlive(kaCtx, leaseResp.ID)
	if err != nil {
		return err
	}
	go func() {
		// 在这里操作续约
		for kaResp := range ch {
			// 正常就是打印一下 DEBUG 日志啥的
			s.L.Debug(kaResp.String())
		}
	}()

	return nil
}

func (s *Server) Close() error {
	if s.kaCancel != nil {
		s.kaCancel()
	}
	if s.em != nil {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		err := s.em.DeleteEndpoint(ctx, s.key)
		defer cancel()
		if err != nil {
			return err
		}
	}
	if s.client != nil {
		err := s.client.Close()
		if err != nil {
			return err
		}
	}
	s.Server.GracefulStop()
	return nil
}
