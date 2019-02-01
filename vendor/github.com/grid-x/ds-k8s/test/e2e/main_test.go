package e2e

import (
	"flag"
	"os"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/grid-x/ds-k8s/pkg/client/clientset/versioned"
)

var (
	GLOBAL *Framework
)

func TestMain(m *testing.M) {
	var (
		logger = log.New()

		kubeconfig       = flag.String("kubeconfig", "", "kube config path, e.g. $HOME/.kube/config")
		managerImage     = flag.String("manager.image", "", "manager image, e.g. gridx.de/ds-k8s-manager")
		managerNamespace = flag.String("manager.namespace", "system", "e2e test namespace")
	)
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		logger.Fatalf("cannot read kubeconfig: %+v", err)
	}

	// Adding a check to avoid running tests against production clusters
	if strings.Contains(config.Host, "prod") {
		logger.Fatal("DO NOT RUN AGAINST PRODUCTION CLUSTER")
	}

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		logger.Fatalf("cannot create k8s client: %+v", err)
	}

	crClient, err := versioned.NewForConfig(config)
	if err != nil {
		logger.Fatalf("cannot create CR client: %+v", err)
	}

	if *managerImage == "" {
		logger.Warn("not deploying controller")
	}

	framework, err := NewFramework(*managerImage, *managerNamespace, kubeClient, crClient)
	if err != nil {
		logger.Fatalf("fail to create framework: %v", err)
	}
	GLOBAL = framework

	if err = framework.Setup(); err != nil {
		// NOTE: we need to manually call teardown here and at the end
		// since defer does not work if we run os.Exit
		if err := framework.Teardown(); err != nil {
			logger.Errorf("failed to teardown framework: %v", err)
		}
		logger.Fatalf("failed to setup framework: %v", err)
	}

	code := m.Run()

	if code != 0 {
		if err := framework.Dump(); err != nil {
			logger.Errorf("failed to dump troubleshooting info: %+v", err)
		}
	}

	if err := framework.Teardown(); err != nil {
		logger.Errorf("failed to teardown framework: %v", err)
	}
	os.Exit(code)
}
