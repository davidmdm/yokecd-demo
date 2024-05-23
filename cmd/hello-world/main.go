package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/davidmdm/yoke/pkg/flight"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	appsv1 "k8s.io/client-go/applyconfigurations/apps/v1"
	corev1 "k8s.io/client-go/applyconfigurations/core/v1"
	metav1 "k8s.io/client-go/applyconfigurations/meta/v1"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	var (
		release   = flight.Release()
		namespace = flight.Namespace()
		labels    = map[string]string{"app": release}
	)

	replicas := flag.Int("replicas", 3, "desired replica count for deployment")
	flag.Parse()

	resources := []any{
		CreateDeployment(DeploymentConfig{
			Name:      release,
			Namespace: namespace,
			Labels:    labels,
			Replicas:  int32(*replicas),
		}),
		CreateService(ServiceConfig{
			Name:       release,
			Namespace:  namespace,
			Labels:     labels,
			Port:       80,
			TargetPort: 80,
		}),
	}

	return json.NewEncoder(os.Stdout).Encode(resources)
}

type DeploymentConfig struct {
	Name      string
	Namespace string
	Labels    map[string]string
	Replicas  int32
}

func CreateDeployment(cfg DeploymentConfig) *appsv1.DeploymentApplyConfiguration {
	container := corev1.Container().
		WithName(cfg.Name).
		WithImage("alpine:latest").
		WithCommand("watch", "echo", "hello", "world")

	podTemplate := corev1.PodTemplateSpec().
		WithLabels(cfg.Labels).
		WithSpec(corev1.PodSpec().WithContainers(container))

	spec := appsv1.DeploymentSpec().
		WithReplicas(cfg.Replicas).
		WithSelector(metav1.LabelSelector().WithMatchLabels(cfg.Labels)).
		WithTemplate(podTemplate)

	return appsv1.Deployment(cfg.Name, cfg.Namespace).
		WithLabels(cfg.Labels).
		WithSpec(spec)
}

type ServiceConfig struct {
	Name       string
	Namespace  string
	Labels     map[string]string
	Port       int32
	TargetPort int
}

func CreateService(cfg ServiceConfig) *corev1.ServiceApplyConfiguration {
	servicePort := corev1.ServicePort().
		WithProtocol(v1.ProtocolTCP).
		WithPort(cfg.Port).
		WithTargetPort(intstr.FromInt(cfg.TargetPort))

	serviceSpec := corev1.ServiceSpec().
		WithSelector(cfg.Labels).
		WithType(v1.ServiceTypeClusterIP).
		WithPorts(servicePort)

	return corev1.Service(cfg.Name, cfg.Namespace).
		WithLabels(cfg.Labels).
		WithSpec(serviceSpec)
}
