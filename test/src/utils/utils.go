package utils

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/eks/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/aws-iam-authenticator/pkg/token"
	"sync/atomic"
	"time"
)

// NewClientSet generate a kubernetes.Clientset from an aws Cluster
func NewClientSet(cluster *types.Cluster) (*kubernetes.Clientset, error) {
	gen, err := token.NewGenerator(true, false)
	if err != nil {
		return nil, err
	}
	opts := &token.GetTokenOptions{
		ClusterID: *cluster.Name,
	}
	tok, err := gen.GetWithOptions(opts)
	if err != nil {
		return nil, err
	}
	ca, err := base64.StdEncoding.DecodeString(*cluster.CertificateAuthority.Data)
	if err != nil {
		return nil, err
	}
	clientSet, err := kubernetes.NewForConfig(
		&rest.Config{
			Host:        *cluster.Endpoint,
			BearerToken: tok.Token,
			TLSClientConfig: rest.TLSClientConfig{
				CAData: ca,
			},
		},
	)
	if err != nil {
		return nil, err
	}
	return clientSet, nil
}

// GetAwsClient returns an aws.Config client from the env variables `AWS_PROFILE` and `AWS_REGION`
func GetAwsClient() (aws.Config, error) {
	awsProfile := GetEnv("AWS_PROFILE", GetEnv("AWS_DEFAULT_PROFILE", "infex"))
	region := GetEnv("AWS_REGION", "eu-central-1")

	return config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
		config.WithSharedConfigProfile(awsProfile),
	)
}

// WaitUntilClusterIsReady waits until the kube cluster is read or returns an error
func WaitUntilClusterIsReady(cluster *types.Cluster, timeout time.Duration, expectedNodesCount uint64) error {
	// https://github.com/kubernetes/client-go
	// https://www.rushtehrani.com/post/using-kubernetes-api
	// https://rancher.com/using-kubernetes-api-go-kubecon-2017-session-recap
	// https://gianarb.it/blog/kubernetes-shared-informer
	// https://stackoverflow.com/questions/60547409/unable-to-obtain-kubeconfig-of-an-aws-eks-cluster-in-go-code/60573982#60573982

	clientSet, err := NewClientSet(cluster)
	if err != nil {
		return err
	}

	factory := informers.NewSharedInformerFactory(clientSet, 0)
	informer := factory.Core().V1().Nodes().Informer()
	stopChannel := make(chan struct{})
	var countOfWorkerNodes uint64 = 0

	_, errEventHandler := informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			node := obj.(*corev1.Node)
			fmt.Printf("Worker Node %s has joined the EKS cluster at %s\n", node.Name, node.CreationTimestamp)
			atomic.AddUint64(&countOfWorkerNodes, 1)
			if countOfWorkerNodes >= expectedNodesCount {
				stopChannel <- struct{}{} // send close signal
			}
		},
	})
	if errEventHandler != nil {
		return errEventHandler
	}

	go informer.Run(stopChannel)
	go func() {
		// wait to receive a signal to close the channel
		<-stopChannel
		close(stopChannel)
	}()

	select {
	case <-stopChannel:
		msg := "All worker nodes have joined the Kube cluster"
		fmt.Println(msg)
	case <-time.After(timeout):
		msg := "Not all worker nodes have joined the Kube cluster"
		fmt.Println(msg)
		return errors.New(msg)
	}
	return nil
}
