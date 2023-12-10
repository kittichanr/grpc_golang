package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	proto "github.com/kittichanr/pcbook/internal/gen/proto/pcbook/v1"
	"github.com/kittichanr/pcbook/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func seedUsers(userStore service.UserStore) error {
	err := createUser(userStore, "admin1", "secret", "admin")
	if err != nil {
		return err
	}
	return createUser(userStore, "user1", "secret", "user")
}

func createUser(userStore service.UserStore, username, password, role string) error {
	user, err := service.NewUser(username, password, role)
	if err != nil {
		return err
	}
	return userStore.Save(user)
}

const (
	secretKey     = "secret"
	tokenDuration = 15 * time.Minute
)

const (
	serverCertificate = "cert/server-cert.pem"
	serverKeyFile     = "cert/server-key.pem"
	clientCACertFile  = "cert/ca-cert.pem"
)

func accessibleRoles() map[string][]string {
	const laptopServicePath = "/pcbook.v1.LaptopService/"
	return map[string][]string{
		laptopServicePath + "CreateLaptop": {"admin"},
		laptopServicePath + "UploadImage":  {"admin"},
		laptopServicePath + "RateLaptop":   {"admin", "user"},
	}
}

func loadTLSCredentials() (credentials.TransportCredentials, error) {
	pemServerCA, err := os.ReadFile(clientCACertFile)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemServerCA) {
		return nil, fmt.Errorf("failed to add client CA's certificate")
	}

	// Load server's certificate and private key
	serverCert, err := tls.LoadX509KeyPair(serverCertificate, serverKeyFile)
	if err != nil {
		return nil, err
	}

	//create credentials and return it
	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequestClientCert,
		ClientCAs:    certPool,
	}

	return credentials.NewTLS(config), nil
}

func runGRPCServer(
	authServer proto.AuthServiceServer,
	laptopServer proto.LaptopServiceServer,
	jwtManager *service.JWTManager,
	enableTLS bool,
	listener net.Listener,
) error {
	interceptor := service.NewAuthInterceptor(jwtManager, accessibleRoles())
	serverOptions := []grpc.ServerOption{
		grpc.UnaryInterceptor(interceptor.Unary()),
		grpc.StreamInterceptor(interceptor.Stream()),
	}

	if enableTLS {
		tlsCredentials, err := loadTLSCredentials()
		if err != nil {
			return fmt.Errorf("cannot load TLS credentials: %w", err)
		}

		serverOptions = append(serverOptions, grpc.Creds(tlsCredentials))
	}

	grpcServer := grpc.NewServer(serverOptions...)

	proto.RegisterAuthServiceServer(grpcServer, authServer)
	proto.RegisterLaptopServiceServer(grpcServer, laptopServer)
	reflection.Register(grpcServer)

	log.Printf("Start GRPC server at %s, TLS = %t", listener.Addr().String(), enableTLS)
	return grpcServer.Serve(listener)
}

func runRestServer(
	authServer proto.AuthServiceServer,
	laptopServer proto.LaptopServiceServer,
	jwtManager *service.JWTManager,
	enableTLS bool,
	listener net.Listener,
	grpcEndPoint string,
) error {
	mux := runtime.NewServeMux()
	dialOptions := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// err := proto.RegisterAuthServiceHandlerServer(ctx, mux, authServer)
	err := proto.RegisterAuthServiceHandlerFromEndpoint(ctx, mux, grpcEndPoint, dialOptions) // proxy because http not support streaming grpc
	if err != nil {
		return err
	}

	// err = proto.RegisterLaptopServiceHandlerServer(ctx, mux, laptopServer)
	err = proto.RegisterLaptopServiceHandlerFromEndpoint(ctx, mux, grpcEndPoint, dialOptions) // proxy because http not support streaming grpc
	if err != nil {
		return err
	}

	log.Printf("start REST server at %s, TLS = %t", listener.Addr().String(), enableTLS)
	if enableTLS {
		return http.ServeTLS(listener, mux, serverCertificate, serverKeyFile)
	}
	return http.Serve(listener, mux)
}

func main() {
	port := flag.Int("port", 0, "the server port")
	enableTLS := flag.Bool("tls", false, "enable SSL/TLS")
	serverType := flag.String("type", "grpc", "type of server (grpc/rest)")
	endPoint := flag.String("endpoint", "", "gRPC endpoint")
	flag.Parse()
	log.Printf("start server on port %d, TLS = %t", *port, *enableTLS)

	userStore := service.NewInMemoryUserStore()
	err := seedUsers(userStore)
	if err != nil {
		log.Fatal("cannot seed users")
	}

	jwtManager := service.NewJWTManager(secretKey, tokenDuration)
	authServer := service.NewAuthServer(userStore, jwtManager)

	laptopStore := service.NewInMemoryLaptopStore()
	imageStore := service.NewDiskImageStore("img")
	ratingStore := service.NewInMemoryRatingStore()
	laptopServer := service.NewLaptopServer(laptopStore, imageStore, ratingStore)

	address := fmt.Sprintf("0.0.0.0:%d", *port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal("cannot start server: ", err)
	}

	if *serverType == "grpc" {
		err = runGRPCServer(authServer, laptopServer, jwtManager, *enableTLS, listener)
	} else {
		err = runRestServer(authServer, laptopServer, jwtManager, *enableTLS, listener, *endPoint)
	}

	if err != nil {
		log.Fatal("cannot start server: ", err)
	}
}
