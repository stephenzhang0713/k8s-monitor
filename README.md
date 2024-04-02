# k8s-monitor

[![Go Report Card](https://goreportcard.com/badge/github.com/stephenzhang0713/k8s-monitor)](https://goreportcard.com/report/github.com/stephenzhang0713/k8s-monitor)
![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)

k8s-monitor 是一个简单的命令行工具，用于实时监控 Kubernetes 中特定 Pod 的 CPU 和内存使用情况。该工具使用 `Kubernetes Metrics API` 来获取和展示 Pod 资源使用数据。

## 功能

- 监控指定 Pod 的 CPU 和内存使用情况。
- 支持指定 Kubernetes 命名空间中的 Pod。
- 显示每个容器的资源使用情况。

## 开始使用

### 前提条件

- 一个运行中的 Kubernetes 集群。
- 集群中已部署 Metrics Server。
- 你的机器上配置有对集群的访问权限（`~/.kube/config` 或通过 `KUBECONFIG` 环境变量指定）。

### 安装

```bash
go install github.com/stephenzhang0713/k8s-monitor@latest
```

### 使用 
运行下面的命令来监控一个 Pod：

```bash
./k8s-monitor --p POD_NAME --n NAMESPACE
```


参数说明：

```bash
--p: 要监控的 Pod 名称。
--n: Pod 所在的命名空间。如果未指定，默认为 default。
```
