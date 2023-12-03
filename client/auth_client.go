package client

import (
	"context"
	"time"

	proto "github.com/kittichanr/pcbook/internal/gen/proto/pcbook/v1"
	"google.golang.org/grpc"
)

type AuthClient struct {
	service  proto.AuthServiceClient
	username string
	password string
}

func NewAuthClient(cc *grpc.ClientConn, username string, password string) *AuthClient {
	service := proto.NewAuthServiceClient(cc)

	return &AuthClient{
		service,
		username,
		password,
	}
}

func (client *AuthClient) Login() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &proto.LoginRequest{
		Username: client.username,
		Password: client.password,
	}

	res, err := client.service.Login(ctx, req)
	if err != nil {
		return "", err
	}
	return res.GetAccessToken(), nil
}
