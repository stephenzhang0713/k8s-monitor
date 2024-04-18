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
	"time"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	metricsv "k8s.io/metrics/pkg/client/clientset/versioned"
)

var podName string
var namespace string

// 初始化zap日志记录器
var logger, _ = zap.NewProduction()

func initKubernetesClient() (*metricsv.Clientset, error) {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
		}
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	metricsClient, err := metricsv.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return metricsClient, nil
}

func monitorPods(ctx context.Context, podName, namespace string) error {
	metricsClient, err := initKubernetesClient()
	if err != nil {
		logger.Error("initializing Kubernetes client", zap.Error(err))
		return fmt.Errorf("initializing Kubernetes client: %w", err)
	}

	// 无限循环，每5秒刷新一次
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Second):
			podMetrics, err := metricsClient.MetricsV1beta1().PodMetricses(namespace).Get(ctx, podName, metav1.GetOptions{})
			if err != nil {
				logger.Error("getting pod metrics", zap.Error(err))
				continue // 如果发生错误，跳过当前迭代，继续下一个循环
			}

			// 初始化总 CPU 和内存使用量
			totalCpuUsage := resource.NewQuantity(0, resource.DecimalSI)
			totalMemoryUsage := resource.NewQuantity(0, resource.DecimalSI)

			for _, container := range podMetrics.Containers {
				// 累加 CPU 和内存使用量
				totalCpuUsage.Add(*container.Usage.Cpu())
				totalMemoryUsage.Add(*container.Usage.Memory())
			}

			// 打印总 CPU 和内存使用量
			// 使用zap打印总 CPU 和内存使用量
			logger.Info("Pod metrics",
				zap.String("Pod", podMetrics.Name),
				zap.String("Total CPU Usage", totalCpuUsage.String()),
				zap.String("Total Memory Usage", totalMemoryUsage.String()),
			)
		}
	}
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
