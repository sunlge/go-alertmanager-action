package pkg

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	"os"
)


func DeletePod(ns, podName string) {
	var err error
	var config *rest.Config
	var kubeconfig = os.Getenv("KUBECONFIG")

	// 使用 ServiceAccount 创建集群配置（InCluster模式）
	if config, err = rest.InClusterConfig(); err != nil {
		if config, err = clientcmd.BuildConfigFromFlags("", kubeconfig); err != nil {
			panic(err.Error())
		}
	}

	// 创建 clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Errorf(err.Error())
	}

	var (
		// 删除pod不夯前台，改变删除策略。执行后放到后台去删除相关的pod从资源
		deletePolicy metav1.DeletionPropagation = "Background"
		// 立即删除
		gracetime int64 = 0
		deleteOptions metav1.DeleteOptions = metav1.DeleteOptions{
			PropagationPolicy:  &deletePolicy,
			GracePeriodSeconds: &gracetime,
		}
	)

	deletepodErr := clientset.CoreV1().Pods(ns).Delete(context.TODO(), podName, deleteOptions)
	if deletepodErr != nil {
		klog.Errorf("命名空间：%s --POD名称：%s 删除失败!", ns, podName)
		klog.Errorf(deletepodErr.Error())
		//panic(deletepodErr.Error())
	} else {
		klog.Infof( "命名空间：%s --POD名称：%s 删除成功!", ns, podName)
	}

}