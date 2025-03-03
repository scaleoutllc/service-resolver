package main

import (
	"context"
	echo "github.com/labstack/echo/v4"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"strings"
)

func ServiceResolver(c echo.Context) error {
	cc := c.(*ResolverContext)

	namespace := cc.Param("namespace")
	serviceName := cc.Param("service")

	service, err := cc.K8s.CoreV1().Services(namespace).Get(context.Background(), serviceName, metav1.GetOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			// TODO: Consider logging specific not found errors, we don't want to return them to the consumer though
			return cc.String(http.StatusNotFound, "")
		}
		return err
	}

	return cc.String(http.StatusOK, strings.Join(service.Spec.ClusterIPs, ","))
}

func EndpointResolver(c echo.Context) error {
	cc := c.(*ResolverContext)
	namespace := cc.Param("namespace")
	// endpoints and services are 1-to-1 naming
	serviceName := cc.Param("service")

	endpoints, err := cc.K8s.CoreV1().Endpoints(namespace).Get(context.Background(), serviceName, metav1.GetOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			// TODO: Consider logging specific not found errors, we never want to return them to the consumer though
			return cc.String(http.StatusNotFound, "")
		}
		return err
	}

	var ips []string
	for _, address := range endpoints.Subsets[0].Addresses {
		ips = append(ips, address.IP)
	}

	return cc.String(http.StatusOK, strings.Join(ips, ","))
}
