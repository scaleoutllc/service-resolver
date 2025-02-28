package main

import (
	"context"
	echo "github.com/labstack/echo/v4"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
)

func LivenessHandler(c echo.Context) error {
	// if the program is running, it's live
	return c.String(http.StatusOK, "live")
}

func ReadinessHandler(c echo.Context) error {
	cc := c.(*ResolverContext)
	// confirm we at least have access to list services, if we can't do that we shouldn't receive traffic
	_, err := cc.K8s.CoreV1().Services(cc.HealthCheck.Namespace).Get(context.Background(), cc.HealthCheck.Service, metav1.GetOptions{})
	if err != nil {
		return err
	}
	return cc.String(http.StatusOK, "ready")
}
