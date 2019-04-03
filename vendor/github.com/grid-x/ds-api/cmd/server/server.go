package main

import (
	"context"
	"flag"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/grid-x/ds-k8s/pkg/client/clientset/versioned"
	appsinformers "github.com/grid-x/ds-k8s/pkg/client/informers/externalversions/apps/v1beta1"
	configinformers "github.com/grid-x/ds-k8s/pkg/client/informers/externalversions/config/v1beta1"
	coreinformers "github.com/grid-x/ds-k8s/pkg/client/informers/externalversions/core/v1beta1"
	maininformers "github.com/grid-x/ds-k8s/pkg/client/informers/externalversions/maintenance/v1beta1"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	nats "github.com/nats-io/go-nats"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/grid-x/ds-api/pkg/auth/device"
	"github.com/grid-x/ds-api/pkg/auth/management"
	"github.com/grid-x/ds-api/pkg/identifier"
	"github.com/grid-x/ds-api/pkg/k8s"
	"github.com/grid-x/ds-api/pkg/messaging"
	"github.com/grid-x/ds-api/pkg/postgres"
	"github.com/grid-x/ds-api/pkg/responselog"
	"github.com/grid-x/ds-api/pkg/router"
	"github.com/grid-x/ds-api/pkg/timeout"
)

var (
	webSocketUpgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

type mgmtMiddlewareProvider struct {
	auth router.Middleware
}

func (p *mgmtMiddlewareProvider) Auth() router.Middleware {
	return p.auth
}

type devMiddlewareProvider struct {
	auth router.Middleware
}

func (p *devMiddlewareProvider) Auth() router.Middleware {
	return p.auth
}

func main() {

	var (
		logger = log.New()

		address       = flag.String("listen", ":8080", "address to listen on")
		authRSAKey    = flag.String("auth.rsa-key", "key.pem", "")
		authDevRSAKey = flag.String("auth.dev-rsa-key", "key.pem", "")
		authJWTIssuer = flag.String("auth.dev-jwt-issuer", "api.gridx.ai", "")

		// db
		databaseURL          = flag.String("postgres.url", "postgres://postgres:pw@localhost/postgres?sslmode=disable", "postgres URL")
		databaseMaxOpenConns = flag.Int("postgres.max-open-conns", 10, "Max open connections for postgres")
		databaseMigration    = flag.Bool("postgres.migrate", false, "Whether to perform migration")

		// NATS
		natsURL = flag.String("nats.url", "nats://nats:4222", "NATS address")

		// K8s integration
		k8sConfig       = flag.String("k8s.config", "", "Path to a kubeconfig. Only required if out-of-cluster.")
		k8sAPI          = flag.String("k8s.api", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
		k8sResyncPeriod = flag.Duration("k8s.resync", 15*time.Minute, "The kubernetes resync period time")
	)
	flag.Parse()

	rsaKey, err := ioutil.ReadFile(*authRSAKey)
	if err != nil {
		logger.Fatalf("error reading rsa signing key: %+v", err)
	}
	data, err := ioutil.ReadFile(*authDevRSAKey)
	if err != nil {
		logger.Fatalf("error reading dev rsa key file: %+v", err)
	}
	devRSAKey, err := jwt.ParseRSAPrivateKeyFromPEM(data)
	if err != nil {
		logger.Fatalf("error reading Device-API rsa signing key: %+v", err)
	}

	k8sCfg, err := clientcmd.BuildConfigFromFlags(*k8sAPI, *k8sConfig)
	if err != nil {
		logger.Fatalf("Error building kubeconfig: %+v", err)
	}

	coreK8sClient, err := kubernetes.NewForConfig(k8sCfg)
	if err != nil {
		logger.Fatalf("Error building core k8s clientset: %+v", err)
	}

	k8sClient, err := versioned.NewForConfig(k8sCfg)
	if err != nil {
		logger.Fatalf("Error building gridx clientset: %+v", err)
	}

	nc, err := nats.Connect(*natsURL)
	if err != nil {
		logger.Fatalf("failed to connect to nats: %+v", err)
	}
	natsConn, err := nats.NewEncodedConn(nc, nats.DEFAULT_ENCODER)
	if err != nil {
		logger.Fatalf("failed to connect to nats: %+v", err)
	}
	defer natsConn.Close()

	db, err := sqlx.Open("postgres", *databaseURL)
	if err != nil {
		logger.Fatalf("postgres: %+v", err)
	}
	db.SetMaxOpenConns(*databaseMaxOpenConns)
	defer db.Close()
	if *databaseMigration {
		n, err := postgres.Migrate(db)
		if err != nil {
			logger.Fatalf("migrate: %+v", err)
		}
		logger.Infof("postgres: %d migrations", n)
	}
	ctx := context.Background()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	deviceInformer := coreinformers.NewDeviceInformer(k8sClient, "", *k8sResyncPeriod, cache.Indexers{})
	podInformer := coreinformers.NewDevicePodInformer(k8sClient, "", *k8sResyncPeriod, cache.Indexers{})
	deployInformer := appsinformers.NewDeviceDeploymentInformer(k8sClient, "", *k8sResyncPeriod, cache.Indexers{})
	appsInformer := appsinformers.NewDeviceApplicationInformer(k8sClient, "", *k8sResyncPeriod, cache.Indexers{})
	mainInformer := maininformers.NewMaintenanceTaskInformer(k8sClient, "", *k8sResyncPeriod, cache.Indexers{})
	dockerConfigInformer := configinformers.NewDockerConfigInformer(k8sClient, "", *k8sResyncPeriod, cache.Indexers{})

	// core/v1beta1
	devRepo, err := k8s.NewDevicesRepository(logger, k8sClient.CoreV1beta1(), deviceInformer, *k8sResyncPeriod)
	if err != nil {
		logger.Fatalf("cannot create devices repo: %+v", err)
	}
	podsRepo, err := k8s.NewPodsRepository(logger, k8sClient.CoreV1beta1(), podInformer, *k8sResyncPeriod)
	if err != nil {
		logger.Fatalf("cannot create pods repo: %+v", err)
	}
	// apps/v1beta1
	appsRepo, err := k8s.NewApplicationsRepository(logger, k8sClient.AppsV1beta1(), appsInformer, *k8sResyncPeriod)
	if err != nil {
		logger.Fatalf("cannot create apps repo: %+v", err)
	}
	deploysRepo, err := k8s.NewDeploymentsRepository(logger, k8sClient.AppsV1beta1(), deployInformer, *k8sResyncPeriod)
	if err != nil {
		logger.Fatalf("cannot create deployments repo: %+v", err)
	}
	// maintenance/v1beta1
	mainRepo, err := k8s.NewMaintenanceTasksRepository(logger, k8sClient.MaintenanceV1beta1(), mainInformer, *k8sResyncPeriod)
	if err != nil {
		logger.Fatalf("cannot create maintenance tasks repo: %+v", err)
	}
	// config/v1beta1
	dockerConfigRepo, err := k8s.NewDockerConfigRepository(logger, k8sClient.ConfigV1beta1(), dockerConfigInformer, *k8sResyncPeriod)
	if err != nil {
		logger.Fatalf("cannot create deployments repo: %+v", err)
	}

	go deviceInformer.Run(ctx.Done())
	go podInformer.Run(ctx.Done())
	go appsInformer.Run(ctx.Done())
	go deployInformer.Run(ctx.Done())
	go mainInformer.Run(ctx.Done())
	go dockerConfigInformer.Run(ctx.Done())

	for !deviceInformer.HasSynced() ||
		!podInformer.HasSynced() ||
		!appsInformer.HasSynced() ||
		!deployInformer.HasSynced() ||
		!mainInformer.HasSynced() ||
		!dockerConfigInformer.HasSynced() {
		logger.Infof("waiting for informers to sync")
		time.Sleep(time.Second)
	}

	uuidRepo, err := identifier.NewUUIDRepository(uuid.New)
	if err != nil {
		logger.Fatalf("cannot create uuid repo: %+v", err)
	}

	natsRepo, err := messaging.NewNATSRepository(natsConn)
	if err != nil {
		logger.Fatalf("cannot create nats repo: %+v", err)
	}

	timeoutRepo, err := timeout.NewDurationRepository()
	if err != nil {
		logger.Fatalf("cannot create timeout repo: %+v", err)
	}

	accRepo, err := postgres.NewAccountsRepository(db)
	if err != nil {
		logger.Fatalf("cannot create accounts repo: %+v", err)
	}
	userRepo, err := postgres.NewUsersRepository(db)
	if err != nil {
		logger.Fatalf("cannot create  users repo: %+v", err)
	}
	mgmtAp := management.NewAuthProvider(logger, rsaKey, userRepo, accRepo, coreK8sClient.CoreV1().Namespaces())
	jwtGen := device.NewJWTGenerator(*authJWTIssuer, devRSAKey)
	devAp := device.NewAuthProvider(logger, jwtGen, devRepo)

	injections := []interface{}{devRepo, podsRepo, appsRepo, deploysRepo, mainRepo, dockerConfigRepo, uuidRepo, natsRepo, timeoutRepo}
	mgmtInjections := append(injections, mgmtAp)
	devInjections := append(injections, []interface{}{devAp, jwtGen}...)

	// setup all middlewares
	globalMiddlewares := []router.Middleware{
		responselog.Middleware(logger),
	}

	deviceMiddlewareProvider := &devMiddlewareProvider{
		auth: devAp.Middleware(),
	}
	mgmtMiddlewareProvider := &mgmtMiddlewareProvider{
		auth: mgmtAp.Middleware(),
	}

	// Inject dependencies into routes and their related services.
	deviceRoutes, err := DeviceRoutes(globalMiddlewares, deviceMiddlewareProvider, devInjections)
	if err != nil {
		logger.Fatalf("DeviceRoutes: %+v", err)
	}

	managementRoutes, err := ManagementRoutes(globalMiddlewares, mgmtMiddlewareProvider, mgmtInjections)
	if err != nil {
		logger.Fatalf("ManagementRoutes: %+v", err)
	}

	r := mux.NewRouter()
	r.StrictSlash(true)

	// setup devices routes
	ds := r.PathPrefix("/api/device/").Subrouter()
	router.Populate(deviceRoutes, ds)

	// setup management routes
	ms := r.PathPrefix("/api/management/").Subrouter()
	router.Populate(managementRoutes, ms)

	logger.Infof("listening on %s", *address)
	if err := http.ListenAndServe(*address, r); err != nil {
		logger.Fatalf("ListenAndServe: %+v", err)
	}
}

// Repository is just a mock.
type Repository struct {
}

// NewRepository creates a new repository
func NewRepository() (*Repository, error) {
	return &Repository{}, nil
}

// Close closes a repository
func (r *Repository) Close() error {
	return nil
}
