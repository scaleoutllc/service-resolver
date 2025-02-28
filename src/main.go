package main

import (
	"github.com/mitchellh/go-homedir"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gopkg.in/tylerb/graceful.v1"
)

type ResolverHealthCheckConfig struct {
	Namespace string
	Service   string
}
type ResolverContext struct {
	echo.Context
	K8s         kubernetes.Interface
	HealthCheck ResolverHealthCheckConfig
}

func main() {
	log.Println("service-resolver starting...")
	httpPort := os.Getenv("APP_SERVER_PORT")
	if httpPort == "" {
		httpPort = "8080"
	}

	hcNamespace := os.Getenv("HEALTH_CHECK_NAMESPACE")
	if hcNamespace == "" {
		hcNamespace = "default"
	}

	hcService := os.Getenv("HEALTH_CHECK_SERVICE")
	if hcService == "" {
		hcService = "kubernetes"
	}

	var config *rest.Config
	var err error

	kubeContext := os.Getenv("KUBE_CONTEXT")
	if kubeContext != "" {
		kubeconfigPath, expandErr := homedir.Expand("~/.kube/config")
		if expandErr != nil {
			log.Fatalf("error expanding home directory: %s", expandErr.Error())
		}
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			log.Fatalf("error loading kubeconfig: %s", err.Error())
		}
	} else {
		log.Println("detected running in cluster, building in cluster config")
		config, err = rest.InClusterConfig()
		if err != nil {
			log.Fatalf("error creating in-cluster config: %s", err.Error())
		}
	}

	k8s, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("error creating clientset: %s", err.Error())
	}

	e := echo.New()

	// setup a custom echo context so the kubernetes client does not need to be instantiated and configured on every API request
	// also use the context to pass useful configuration data to handlers that would otherwise need to be fetched from the environment
	// on each connection
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := &ResolverContext{
				Context: c,
				K8s:     k8s,
				HealthCheck: ResolverHealthCheckConfig{
					Namespace: hcNamespace,
					Service:   hcService,
				},
			}
			return next(cc)
		}
	})

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/readiness", ReadinessHandler)
	e.GET("/liveness", LivenessHandler)

	// /headless/:namespace/:service returns the actual pod IP addresses for a given :service based on that service's selector configuration
	// to build the response the handler queries the service object, then uses the service object's selector to list pods where the IPs can be parsed out
	// by building the response in this way, the :service being queried does not have to be a true 'k8s headless service' where clusterIp is set to None
	// # see: https://kubernetes.io/docs/concepts/services-networking/service/#headless-services
	// this is useful for many clustered technologies like cassandra and kafka and supporting the libraries that connect to them
	e.GET("/v1/headless/:namespace/:service", HeadlessResolver)

	// /service/:namespace/:service returns the service level clusterIp for a given :service
	// # see: https://kubernetes.io/docs/concepts/services-networking/service/#services-in-kubernetes
	e.GET("/v1/service/:namespace/:service", ServiceResolver)

	e.Server.Addr = ":" + httpPort

	e.Logger.Fatal(graceful.ListenAndServe(e.Server, 1*time.Minute))
}
