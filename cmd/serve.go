package cmd

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/cors"
	"github.com/spf13/cobra"
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
	"path/filepath"
	v1 "poc-cloud-service/gen/api/v1"
	"poc-cloud-service/internal/reconciler"
	"poc-cloud-service/internal/server"
	"poc-cloud-service/internal/store"
	"poc-cloud-service/log"
)

var (
	grpcAddr string
	httpAddr string
	dsn      string
)

const (
	defaultDsn = "postgresql://cloud-service:cloud-service@localhost:15432/postgres?sslmode=disable"
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
				logger.Fatal("failed to get user home directory", zap.Error(err))
			}
			kubeConfigPath := path.Join(home, ".kube", "config")
			config, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
			if err != nil {
				logger.Fatal("failed to create kubernetes config", zap.Error(err)	)
			}
		}

		client, err := kubernetes.NewForConfig(config)
		if err != nil {
			logger.Fatal("failed to create kubernetes client", zap.Error(err))
		}

		if len(dsn) == 0 {
			dsn = os.Getenv("DB_DSN")
		}
		if len(dsn) == 0 {
			dsn = defaultDsn
		}

		if err := store.Migrate(ctx, dsn); err != nil {
			logger.Fatal("failed to migrate database", zap.Error(err))
		}

		pool, err := pgxpool.New(ctx, dsn)
		if err != nil {
			logger.Fatal("failed to create pgx pool", zap.Error(err))
		}

		storeObj := store.New(pool)

		dynamicClient, err := dynamic.NewForConfig(config)
		if err != nil {
			logger.Fatal("failed to create dynamic client", zap.Error(err))
		}

		r := reconciler.NewReconciler(client, dynamicClient, storeObj)
		go func() {
			r.Start(ctx)
		}()

		srv, err := server.NewServer(ctx, client, dynamicClient, storeObj)
		if err != nil {
			logger.Fatal("failed to create server", zap.Error(err))
		}

		listener, err := net.Listen("tcp", grpcAddr)
		if err != nil {
			logger.Fatal("failed to create listener", zap.Error(err))
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
			logger.Fatal("failed to split host and port", zap.Error(err))
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
			logger.Fatal("failed to register gateway TenantServiceHandler", zap.Error(err))
		}

		spa := spaHandler{staticPath: "ui/dist", indexPath: "index.html"}

		httpMux := http.NewServeMux()
		httpMux.Handle("/v1/", mux)
		httpMux.Handle("/", spa)

		handler := cors.AllowAll().Handler(httpMux)

		gwServer := &http.Server{
			Addr:    httpAddr,
			Handler: handler,
		}

		logger.Info("starting server", zap.String("address", httpAddr))
		if err := gwServer.ListenAndServe(); err != nil {
			logger.Fatal("failed to serve", zap.Error(err))
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.PersistentFlags().StringVar(&grpcAddr, "grpc-addr", ":8080", "gRPC address")
	serveCmd.PersistentFlags().StringVar(&httpAddr, "http-addr", ":8081", "HTTP address")
	serveCmd.PersistentFlags().StringVar(&dsn, "dsn", "", "PostgreSQL DSN")
}

type spaHandler struct {
	staticPath string
	indexPath  string
}

// ServeHTTP inspects the URL path to locate a file within the static dir
// on the SPA handler. If a file is found, it will be served. If not, the
// file located at the index path on the SPA handler will be served. This
// is suitable behavior for serving an SPA (single page application).
func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Join internally call path.Clean to prevent directory traversal
	p := filepath.Join(h.staticPath, r.URL.Path)

	// check whether a file exists or is a directory at the given path
	fi, err := os.Stat(p)
	if os.IsNotExist(err) || fi.IsDir() {
		// file does not exist or path is a directory, serve index.html
		http.ServeFile(w, r, filepath.Join(h.staticPath, h.indexPath))
		return
	}

	if err != nil {
		// if we got an error (that wasn't that the file doesn't exist) stating the
		// file, return a 500 internal server error and stop
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// otherwise, use http.FileServer to serve the static file
	http.FileServer(http.Dir(h.staticPath)).ServeHTTP(w, r)
}
