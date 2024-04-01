// Package cmd
/*
Copyright © 2024 Stephen Zhang stephenzhang0713@outlook.com

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

var podName string
var namespace string

func initKubernetesClient() (*kubernetes.Clientset, *metricsv.Clientset, error) {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
		}
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}

	metricsClient, err := metricsv.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}

	return clientset, metricsClient, nil
}

func monitorPods(ctx context.Context, podName, namespace string) error {
	_, metricsClient, err := initKubernetesClient()
	if err != nil {
		return fmt.Errorf("initializing Kubernetes client: %w", err)
	}

	podMetrics, err := metricsClient.MetricsV1beta1().PodMetricses(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("getting pod metrics: %w", err)
	}

	slog.Info("Pod: ", podMetrics.Name)
	for _, container := range podMetrics.Containers {
		slog.Info("  Container: %s\n", container.Name)
		slog.Info("    CPU Usage: %s\n", container.Usage.Cpu().String())
		slog.Info("    Memory Usage: %s\n", container.Usage.Memory().String())
	}

	return nil
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "k8s-monitor",
	Short: "Monitor the CPU and memory usage of a Kubernetes Pod",
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.TODO() // Consider using context.WithTimeout for real applications
		if err := monitorPods(ctx, podName, namespace); err != nil {
			slog.Error("Monitor pods failed: ", slog.Any("error", err))
			os.Exit(1)
		}
	},
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.Flags().StringVarP(&podName, "pod", "p", "", "Name of the pod to monitor")
	rootCmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "Namespace of the pod (default is 'default')")
	rootCmd.MarkFlagRequired("pod")
}

func initConfig() {
	// Here you can initialize any configuration before the command execution
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
