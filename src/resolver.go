package main

import (
	"context"
	echo "github.com/labstack/echo/v4"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"net/http"
	"strings"
)

func HeadlessResolver(c echo.Context) error {
	cc := c.(*ResolverContext)
	namespace := cc.Param("namespace")
	serviceName := cc.Param("service")

	service, err := cc.K8s.CoreV1().Services(namespace).Get(context.Background(), serviceName, metav1.GetOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			// TODO: Consider logging specific not found errors, we never want to return them to the consumer though
			return cc.String(http.StatusNotFound, "")
		}
		return err
	}

	labelSelector := metav1.LabelSelector{MatchLabels: service.Spec.Selector}
	pods, err := cc.K8s.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: labels.Set(labelSelector.MatchLabels).String(),
	})

	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			// TODO: Consider logging specific not found errors, we never want to return them to the consumer though
			return cc.String(http.StatusNotFound, "")
		}
		return err
	}

	var ips []string
	for _, pod := range pods.Items {
		ips = append(ips, pod.Status.PodIP)
	}

	return cc.String(http.StatusOK, strings.Join(ips, ","))
}

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
