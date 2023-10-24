/*
Copyright 2017 Heptio Inc.

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

package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cast"
	klog "k8s.io/klog/v2"

	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var resyncInterval = flag.Duration("resync-interval", 3*time.Minute, "K8S resync interval")

var isMetricsEnabled = flag.Bool("metrics-enabled", true, "Enabled metrics.")
var metricsListenAddress = flag.String("metrics-listen-address", ":9090", "The address to listen on for metrics.")

var positionFilePath = flag.String("position-file-path", "", "Path to position path.")

// setup a signal hander to gracefully exit
func sigHandler() <-chan struct{} {
	stop := make(chan struct{})
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c,
			syscall.SIGINT,  // Ctrl+C
			syscall.SIGTERM, // Termination Request
			syscall.SIGSEGV, // FullDerp
			syscall.SIGABRT, // Abnormal termination
			syscall.SIGILL,  // illegal instruction
			syscall.SIGFPE)  // floating point - this is why we can't have nice things
		sig := <-c
		klog.Warningf("Signal (%v) Detected, Shutting Down", sig)
		close(stop)
	}()
	return stop
}

// loadConfig will parse input + config file and return a clientset
func loadConfig() kubernetes.Interface {
	var config *rest.Config
	var err error

	flag.Parse()

	kubeconfig := os.Getenv("KUBECONFIG")
	if len(kubeconfig) > 0 {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		panic(err.Error())
	}

	// creates the clientset from kubeconfig
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	return clientset
}

// main entry point of the program
func main() {
	var wg sync.WaitGroup

	klog.InitFlags(nil)

	err := flag.Set("logtostderr", "true")
	if err != nil {
		panic(err.Error())
	}

	clientset := loadConfig()

	var lastResourceVersionPosition string
	var mostRecentResourceVersion *string

	resourceVersionPositionPath := *positionFilePath
	resourceVersionPositionFunc := func(resourceVersion string) {
		if resourceVersionPositionPath != "" {
			if cast.ToInt(resourceVersion) > cast.ToInt(mostRecentResourceVersion) {
				err := os.WriteFile(resourceVersionPositionPath, []byte(resourceVersion), 0600)
				if err != nil {
					klog.Errorf("failed to write lastResourceVersionPosition")
				} else {
					mostRecentResourceVersion = &resourceVersion
				}
			}
		}
	}

	if resourceVersionPositionPath != "" {
		_, err := os.Stat(resourceVersionPositionPath)
		if !os.IsNotExist(err) {
			resourceVersionBytes, err := os.ReadFile(resourceVersionPositionPath)
			if err != nil {
				klog.Errorf("failed to read resource version bookmark from %s", resourceVersionPositionPath)
			} else {
				lastResourceVersionPosition = string(resourceVersionBytes)
			}
		}
	}

	sharedInformers := informers.NewSharedInformerFactory(clientset, *resyncInterval)
	eventsInformer := sharedInformers.Core().V1().Events()

	// TODO: Support locking for HA https://github.com/kubernetes/kubernetes/pull/42666
	eventRouter := NewEventRouter(clientset, eventsInformer, lastResourceVersionPosition, resourceVersionPositionFunc, *isMetricsEnabled)
	stop := sigHandler()

	// Startup the http listener for Prometheus Metrics endpoint.
	if *isMetricsEnabled {
		go func() {
			klog.Infof("Starting metrics enpoint at %s.", *metricsListenAddress)
			http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintf(w, "OK")
			})
			http.Handle("/metrics", promhttp.Handler())
			server := &http.Server{
				Addr:              *metricsListenAddress,
				ReadHeaderTimeout: 3 * time.Second,
			}
			klog.Warning(server.ListenAndServe())
		}()
	}

	// Startup the EventRouter
	wg.Add(1)
	go func() {
		defer wg.Done()
		eventRouter.Run(stop)
	}()

	// Startup the Informer(s)
	klog.Infof("Starting shared Informer(s)")
	sharedInformers.Start(stop)
	wg.Wait()
	klog.Warningf("Exiting main()")
	os.Exit(1)
}
