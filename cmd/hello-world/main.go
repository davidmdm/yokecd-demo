package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

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
	name := os.Args[0] + "-sample-app"
	labels := map[string]string{"app": name}

	replicas := flag.Int("replicas", 2, "desired replica count for deployment")

	flag.Parse()

	container := corev1.Container().
		WithName(name).
		WithImage("alpine:latest").
		WithCommand("watch", "echo", "hello", "world")

	podTemplate := corev1.PodTemplateSpec().
		WithLabels(labels).
		WithSpec(corev1.PodSpec().WithContainers(container))

	spec := appsv1.DeploymentSpec().
		WithReplicas(int32(*replicas)).
		WithSelector(metav1.LabelSelector().WithMatchLabels(labels)).
		WithTemplate(
			podTemplate,
		)

	deployment := appsv1.Deployment(name, "").
		WithLabels(labels).
		WithSpec(spec)

	return json.NewEncoder(os.Stdout).Encode(deployment)
}
