package framework

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/apimachinery/pkg/util/wait"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/apimachinery/pkg/labels"
	"strings"
	"errors"
	appsv1 "k8s.io/api/apps/v1"
	//shell "github.com/codeskyblue/go-sh"
)

const (
	// How often to poll for conditions
	Poll = 2 * time.Second

	// Default time to wait for operations to complete
	defaultTimeout = 5 * time.Minute

	// Default time to wait for an endpoint to register
	EndpointRegisterTimeout = time.Minute
)

func nowStamp() string {
	return time.Now().Format(time.StampMilli)
}

func log(level string, format string, args ...interface{}) {
	fmt.Fprintf(GinkgoWriter, nowStamp()+": "+level+": "+format+"\n", args...)
}

func Logf(format string, args ...interface{}) {
	log("INFO", format, args...)
}

func Failf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	log("INFO", msg)
	Fail(nowStamp()+": "+msg, 1)
}

func Skipf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	log("INFO", msg)
	Skip(nowStamp() + ": " + msg)
}

type ClientConfigGetter func() (*rest.Config, error)

// unique identifier of the e2e run
var RunId = uuid.NewUUID()

//func InstallKubedb(scriptPath string) error {
//	sh := shell.NewSession()
//	//args := []interface{}{"--namespace", f.Namespace()}
//	//if !f.WebhookEnabled {
//	//	args = append(args, "--enable-webhook=false")
//	//}
//	//SetupServer := filepath.Join("..", "..", "hack", "dev", "setup-server.sh")
//
//	//By("Creating API server and webhook stuffs")
//	return sh.Command(SetupServer).Run()
//}

func CreateKubeNamespace(name string, c kubernetes.Interface) (*corev1.Namespace, error) {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-", name),
		},
	}
	Logf("namespace: %v", ns)
	// Be robust about making the namespace creation call.
	var got *corev1.Namespace
	err := wait.PollImmediate(Poll, defaultTimeout, func() (bool, error) {
		var err error
		got, err = c.CoreV1().Namespaces().Create(ns)
		if err != nil {
			Logf("Unexpected error while creating namespace: %v", err)
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return nil, err
	}
	return got, nil
}

func DeleteKubeNamespace(c kubernetes.Interface, namespace string) error {
	return c.CoreV1().Namespaces().Delete(namespace, nil)
}

func ExpectNoError(err error, explain ...interface{}) {
	if err != nil {
		Logf("Unexpected error occurred: %v", err)
	}
	ExpectWithOffset(1, err).NotTo(HaveOccurred(), explain...)
}

func WaitForCreatingSecret(c kubernetes.Interface, secretName, secretNamespace string) error {
	return wait.PollImmediate(Poll, defaultTimeout, func() (bool, error) {
		_, err := c.CoreV1().Secrets(secretNamespace).Get(secretName, metav1.GetOptions{})
		if err == nil {
			return true, nil
		}

		return false, nil
	})
}

func GetBrokerPod(c kubernetes.Interface, deploy *appsv1.Deployment) (*corev1.Pod, error) {
	var pod *corev1.Pod
	wait.PollImmediate(Poll, defaultTimeout, func() (bool, error) {
		pods, err := c.CoreV1().Pods(deploy.Namespace).List(metav1.ListOptions{
			LabelSelector: labels.SelectorFromSet(map[string]string{
				"app": deploy.Name,
			}).String(),
		})
		if err != nil {
			return false, nil
		}
		for _, p := range pods.Items {
			if strings.HasPrefix(p.Name, deploy.Name) {
				pod = &p
				return true, nil
			}
		}

		return false, nil
	})

	if pod == nil {
		return nil, errors.New("No broker pod")
	}

	return pod, nil
}

// Waits default amount of time (PodStartTimeout) for the specified pod to become running.
// Returns an error if timeout occurs first, or pod goes in to failed state.
func WaitForPodRunningInNamespace(c kubernetes.Interface, pod *corev1.Pod) error {

	//
	//pods, err := c.CoreV1().Pods(deploy.Namespace).List(metav1.ListOptions{
	//	LabelSelector: labels.SelectorFromSet(map[string]string{
	//		"app": deploy.Name,
	//	}).String(),
	//})
	//if err != nil {
	//	return err
	//}
	//var pod *corev1.Pod
	//for _, p := range pods.Items {
	//	if strings.HasPrefix(p.Name, deploy.Name) {
	//		pod = &p
	//		break
	//	}
	//}
	if pod.Status.Phase == corev1.PodRunning {
		return nil
	}
	return waitTimeoutForPodRunningInNamespace(c, pod.Name, pod.Namespace, defaultTimeout)
}

func waitTimeoutForPodRunningInNamespace(c kubernetes.Interface, podName, namespace string, timeout time.Duration) error {
	return wait.PollImmediate(Poll, defaultTimeout, podRunning(c, podName, namespace))
}

func WaitForEndpoint(c kubernetes.Interface, namespace, name string) error {
	return wait.PollImmediate(Poll, EndpointRegisterTimeout, endpointAvailable(c, namespace, name))
}

func endpointAvailable(c kubernetes.Interface, namespace, name string) wait.ConditionFunc {
	return func() (bool, error) {
		endpoint, err := c.CoreV1().Endpoints(namespace).Get(name, metav1.GetOptions{})
		if err != nil {
			if apierrs.IsNotFound(err) {
				return false, nil
			}
			return false, err
		}

		if len(endpoint.Subsets) == 0 || len(endpoint.Subsets[0].Addresses) == 0 {
			return false, nil
		}

		return true, nil
	}
}

func podRunning(c kubernetes.Interface, podName, namespace string) wait.ConditionFunc {
	return func() (bool, error) {
		pod, err := c.CoreV1().Pods(namespace).Get(podName, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		switch pod.Status.Phase {
		case corev1.PodRunning:
			return true, nil
		case corev1.PodFailed, corev1.PodSucceeded:
			return false, fmt.Errorf("pod ran to completion")
		}
		return false, nil
	}
}