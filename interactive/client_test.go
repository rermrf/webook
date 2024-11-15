package main

import (
	"context"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"testing"
	"time"
	intrv1 "webook/api/proto/gen/intr/v1"
)

func TestGRPCClient(t *testing.T) {
	cc, err := grpc.NewClient("localhost:8090", grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	client := intrv1.NewInteractiveServiceClient(cc)

	// 1s 超时
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	res, err := client.IncrReadCnt(ctx, &intrv1.IncrReadCntRequest{
		Biz:   "test",
		BizId: 123,
	})
	t.Log(res)

	resp, err := client.Get(context.Background(), &intrv1.GetRequest{
		Biz:   "test",
		BizId: 123,
		Uid:   345,
	})
	require.NoError(t, err)
	t.Log(resp)
}
