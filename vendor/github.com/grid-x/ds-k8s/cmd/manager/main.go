package main

import (
	"flag"
	"log"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"

	"github.com/grid-x/ds-k8s/pkg/apis"
	"github.com/grid-x/ds-k8s/pkg/controller"
)

func main() {
	var (
		syncPeriod              = flag.Duration("k8s.sync-period", 10*time.Hour, "The k8s sync period")
		leaderElection          = flag.Bool("k8s.leader-election", false, "If leader election should be used")
		leaderElectionNamespace = flag.String("k8s.leader-election-ns", "device-services-control-plane", "The namspace to use for leader election")
		leaderElectionID        = flag.String("k8s.leader-election-config", "device-services-k8s-manager-leader", "The configmap name to use for leader election")
	)
	flag.Parse()
	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatal(err)
	}

	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := manager.New(cfg, manager.Options{
		SyncPeriod:              syncPeriod,
		LeaderElection:          *leaderElection,
		LeaderElectionNamespace: *leaderElectionNamespace,
		LeaderElectionID:        *leaderElectionID,
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Registering Components.")

	// Setup Scheme for all resources
	if err := apis.AddToScheme(mgr.GetScheme()); err != nil {
		log.Fatal(err)
	}

	// Setup all Controllers
	if err := controller.AddToManager(mgr); err != nil {
		log.Fatal(err)
	}

	log.Printf("Starting the Cmd.")

	// Start the Cmd
	log.Fatal(mgr.Start(signals.SetupSignalHandler()))
}
