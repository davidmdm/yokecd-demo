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

	dep := appsv1.Deployment(name, "").
		WithLabels(labels).
		WithSpec(
			appsv1.DeploymentSpec().
				WithReplicas(int32(*replicas)).
				WithSelector(metav1.LabelSelector().WithMatchLabels(labels)).
				WithTemplate(
					corev1.PodTemplateSpec().
						WithLabels(labels).
						WithSpec(
							corev1.PodSpec().WithContainers(
								corev1.Container().
									WithName(name).
									WithImage("alpine:latest").
									WithCommand("watch", "echo", "hello", "world"),
							)),
				),
		)

	return json.NewEncoder(os.Stdout).Encode(dep)
}
