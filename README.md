# 钉钉告警
---
## 项目说明

该项目的宗旨在于通过获取的告警内容，然后对应用做出一个简单的处理。

该项目是当发送告警时，获取告警的标签action，然后触发一个delete pod的动作，删除pod的代码载pkg目录下。通过client-go调用。

## 构建项目
1. 构建镜像
```shell
git clone https://github.com/sunlge/go-alertmanager-action.git
cd go-alertmanager-action
docker build -t alert_action:v1 -f Dockerfile ../
```

2. 部署到集群中

前置条件
* 依赖prometheus-operator
* alertmanager可用
* 依赖KUBERNETES 1.16+
* 集群中可以访问钉钉告警地址`https://oapi.dingtalk.com`

部署到集群
```shell
cd go-alertmanager-action
kubectl apply -f yaml/*
```