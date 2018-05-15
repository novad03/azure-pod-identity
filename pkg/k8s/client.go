package k8s

import (
	"fmt"
	"os"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	aadpodid "github.com/Azure/aad-pod-identity/pkg/apis/aadpodidentity/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Client api client
type Client interface {
	// GetNodeIP return the node ip
	GetNodeIP(nodename string) (nodeip string, err error)
	// GetPodCidr return the pod cidr for the node
	GetPodCidr(nodename string) (podcidr string, err error)
	// GetPodName return the matching azure identity or nil
	GetPodName(podip string) (podns, podname string, err error)
	// GetAzureAssignedIdentity return the matching azure identity or nil
	GetAzureAssignedIdentity(podns, podname string) (azID *aadpodid.AzureIdentity, err error)
}

// KubeClient k8s client
type KubeClient struct {
	// Main Kubernetes client
	ClientSet *kubernetes.Clientset
	// Crd client used to access our CRD resources.
	CrdClient *aadpodid.CrdClient
}

// NewKubeClient new kubernetes api client
func NewKubeClient() (Client, error) {

	config, err := buildConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := getkubeclient(config)
	if err != nil {
		return nil, err
	}

	crdclient, err := aadpodid.NewCRDClient(config)
	if err != nil {
		return nil, err
	}

	kubeClient := &KubeClient{ClientSet: clientset, CrdClient: crdclient}

	return kubeClient, nil
}

// GetNodeIP get node ip from apiserver
func (c *KubeClient) GetNodeIP(nodename string) (nodeip string, err error) {
	return "127.0.0.1", nil
}

// GetPodName get pod ns,name from apiserver
func (c *KubeClient) GetPodName(podip string) (podns, poddname string, err error) {
	if c == nil {
		return "", "", fmt.Errorf("kubeclinet is nil")
	}

	if podip == "" {
		return "", "", fmt.Errorf("podip is nil")
	}

	podipFieldSel := fmt.Sprintf("status.podIp=%s", podip)
	podList, err := c.ClientSet.CoreV1().Pods("default").List(metav1.ListOptions{FieldSelector: podipFieldSel})
	if err != nil {
		return "", "", err
	}

	if len(podList.Items) != 1 {
		return "", "", fmt.Errorf("Expected 1 item in podList, got %d", len(podList.Items))
	}

	return podList.Items[0].Namespace, podList.Items[0].Name, nil
}

// GetPodCidr get node pod cidr from apiserver
func (c *KubeClient) GetPodCidr(nodename string) (podcidr string, err error) {
	if c == nil {
		return "", fmt.Errorf("kubeclinet is nil")
	}

	if nodename == "" {
		return "", fmt.Errorf("nodename is nil")
	}

	n, err := c.ClientSet.CoreV1().Nodes().Get(nodename, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	if n.Spec.PodCIDR == "" {
		return "", fmt.Errorf("podcidr is nil or empty, nodename: %s", nodename)
	}

	return n.Spec.PodCIDR, nil
}

// GetAzureAssignedIdentity return the matching azure identity or nil
func (c *KubeClient) GetAzureAssignedIdentity(podns, podname string) (azID *aadpodid.AzureIdentity, err error) {
	return c.CrdClient.GetAzureAssignedIdentity(podns, podname)
}

func getkubeclient(config *rest.Config) (*kubernetes.Clientset, error) {
	// creates the clientset
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return kubeClient, err
}

// Create the client config. Use kubeconfig if given, otherwise assume in-cluster.
func buildConfig() (*rest.Config, error) {
	kubeconfigPath := os.Getenv("KUBECONFIG")
	if kubeconfigPath != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	}
	return rest.InClusterConfig()
}
