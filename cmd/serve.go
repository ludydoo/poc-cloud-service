package cmd

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"net"
	"net/http"
	"os"
	v1 "poc-cloud-service/gen/api/v1"
	"poc-cloud-service/log"
	"poc-cloud-service/pkg/reconciler"
	"poc-cloud-service/pkg/server"

	"github.com/spf13/cobra"
)

var (
	grpcAddr string
	httpAddr string
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use: "serve",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithCancel(cmd.Context())
		defer cancel()

		logger := log.FromContext(ctx)

		signalChan := make(chan os.Signal, 1)
		go func() {
			<-signalChan
			cancel()
		}()

		config, err := rest.InClusterConfig()
		if err != nil {
			panic(err)
		}

		client, err := kubernetes.NewForConfig(config)
		if err != nil {
			panic(err)
		}

		dynamicClient, err := dynamic.NewForConfig(config)
		if err != nil {
			panic(err)
		}

		r := reconciler.NewReconciler(client, dynamicClient)
		r.Start(ctx)

		srv := server.NewServer(client)

		listener, err := net.Listen("tcp", grpcAddr)
		if err != nil {
			return err
		}

		defer func() {
			if err := listener.Close(); err != nil {
				logger.Error("failed to close listener", zap.Error(err))
			}
		}()

		grpcServer := grpc.NewServer(grpc.Creds(insecure.NewCredentials()))
		v1.RegisterTenantServiceServer(grpcServer, srv)

		go func() {
			if err := grpcServer.Serve(listener); err != nil {
				logger.Fatal("failed to serve", zap.Error(err))
			}
		}()

		_, grpcPort, err := net.SplitHostPort(grpcAddr)
		if err != nil {
			return err
		}

		grpcClient, err := grpc.NewClient("localhost:"+grpcPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return err
		}
		defer func() {
			if err := grpcClient.Close(); err != nil {
				logger.Error("failed to close connection", zap.Error(err))
			}
		}()

		mux := runtime.NewServeMux()
		if err = v1.RegisterTenantServiceHandler(ctx, mux, grpcClient); err != nil {
			return err
		}

		gwServer := &http.Server{
			Addr:    httpAddr,
			Handler: mux,
		}

		logger.Info("starting server", zap.String("address", ":8081"))
		if err := gwServer.ListenAndServe(); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.PersistentFlags().StringVar(&grpcAddr, "grpc-addr", ":8080", "gRPC address")
	serveCmd.PersistentFlags().StringVar(&httpAddr, "http-addr", ":8081", "HTTP address")
}
