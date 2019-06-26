// Copyright 2019-present Open Networking Foundation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package runner

import (
	"bufio"
	"errors"
	atomixk8s "github.com/atomix/atomix-k8s-controller/pkg/client/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	apiextension "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	log "k8s.io/klog"
	"time"
)

// ClusterController manages a single cluster in Kubernetes
type ClusterController struct {
	ClusterId        string
	kubeclient       *kubernetes.Clientset
	atomixclient     *atomixk8s.Clientset
	extensionsclient *apiextension.Clientset
	config           *ClusterConfig
}

// Setup sets up a test cluster with the given configuration
func (c *ClusterController) Setup() error {
	log.Infof("Setting up test cluster %s", c.ClusterId)
	if err := c.setupAtomixController(); err != nil {
		return err
	}
	if err := c.setupPartitions(); err != nil {
		return err
	}
	if err := c.setupOnosConfig(); err != nil {
		return err
	}
	return nil
}

// AddSimulator adds a device simulator with the given configuration
func (c *ClusterController) AddSimulator(name string, config *SimulatorConfig) error {
	log.Infof("Setting up simulator %s/%s", name, c.ClusterId)
	if err := c.setupSimulator(name, config); err != nil {
		return err
	}

	log.Infof("Waiting for simulator %s/%s to become ready", name, c.ClusterId)
	if err := c.awaitSimulatorReady(name); err != nil {
		return err
	}
	return c.redeployOnosConfig()
}

// RunTests runs the given tests on Kubernetes
func (c *ClusterController) RunTests(testId string, tests []string, timeout time.Duration) (string, int, error) {
	// Default the test timeout to 10 minutes
	if timeout == 0 {
		timeout = 10 * time.Minute
	}

	// Start the test job
	pod, err := c.startTests(testId, tests, timeout)
	if err != nil {
		return "", 0, err
	}

	// Stream the logs to stdout
	if err = c.streamLogs(pod); err != nil {
		return "", 0, err
	}

	// Get the exit message and code
	return c.getStatus(pod)
}

// GetLogs returns the logs for a test resource
func (c *ClusterController) GetLogs(resourceId string) ([][]string, error) {
	pod, err := c.kubeclient.CoreV1().Pods(c.ClusterId).Get(resourceId, metav1.GetOptions{})
	if err == nil {
		return c.getAllLogs([]corev1.Pod{*pod})
	} else if !k8serrors.IsNotFound(err) {
		return nil, err
	}

	pods, err := c.kubeclient.CoreV1().Pods(c.ClusterId).List(metav1.ListOptions{
		LabelSelector: "resource=" + resourceId,
	})
	if err != nil {
		return nil, err
	} else if len(pods.Items) == 0 {
		return nil, errors.New("unknown test resource " + resourceId)
	} else {
		return c.getAllLogs(pods.Items)
	}
}

// getAllLogs gets the logs from all of the given pods
func (c *ClusterController) getAllLogs(pods []corev1.Pod) ([][]string, error) {
	allLogs := make([][]string, len(pods))
	for i, pod := range pods {
		logs, err := c.getLogs(pod)
		if err != nil {
			return nil, err
		}
		allLogs[i] = logs
	}
	return allLogs, nil
}

// getLogs gets the logs from the given pod
func (c *ClusterController) getLogs(pod corev1.Pod) ([]string, error) {
	req := c.kubeclient.CoreV1().Pods(c.ClusterId).GetLogs(pod.Name, &corev1.PodLogOptions{})
	readCloser, err := req.Stream()
	if err != nil {
		return nil, err
	}

	defer readCloser.Close()

	logs := []string{}
	scanner := bufio.NewScanner(readCloser)
	for scanner.Scan() {
		logs = append(logs, scanner.Text())
	}
	return logs, nil
}

// RemoveSimulator removes a device simulator with the given name
func (c *ClusterController) RemoveSimulator(name string) error {
	log.Infof("Tearing down simulator %s/%s", name, c.ClusterId)
	if err := c.teardownSimulator(name); err != nil {
		return err
	}
	return c.redeployOnosConfig()
}