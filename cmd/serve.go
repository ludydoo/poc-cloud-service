package cmd

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"net"
	"net/http"
	"os"
	"path"
	v1 "poc-cloud-service/gen/api/v1"
	"poc-cloud-service/internal/reconciler"
	"poc-cloud-service/internal/server"
	"poc-cloud-service/internal/store"
	"poc-cloud-service/log"

	"github.com/spf13/cobra"
)

var (
	grpcAddr string
	httpAddr string
	dsn      string
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
			home, err := os.UserHomeDir()
			if err != nil {
				return err
			}
			kubeConfigPath := path.Join(home, ".kube", "config")
			config, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
			if err != nil {
				logger.Error("failed to create config", zap.Error(err))
				return err
			}
		}

		client, err := kubernetes.NewForConfig(config)
		if err != nil {
			return err
		}

		if len(dsn) == 0 {
			dsn = os.Getenv("DB_DSN")
		}

		if err := store.Migrate(ctx, dsn); err != nil {
			return err
		}

		dbConn, err := pgx.Connect(ctx, dsn)
		if err != nil {
			return err
		}

		storeObj := store.New(dbConn)

		dynamicClient, err := dynamic.NewForConfig(config)
		if err != nil {
			return err
		}

		r := reconciler.NewReconciler(client, dynamicClient, storeObj)
		go func() {
			r.Start(ctx)
		}()

		srv := server.NewServer(client, storeObj)

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
	serveCmd.PersistentFlags().StringVar(&dsn, "dsn", "postgresql://cloud-service:cloud-service@localhost:15432/postgres?sslmode=disable", "PostgreSQL DSN")
}
