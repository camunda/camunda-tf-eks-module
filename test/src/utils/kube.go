package utils

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"time"

	batchv1 "k8s.io/api/batch/v1"
)

func WaitForJobCompletion(clientset *kubernetes.Clientset, namespace, jobName string, timeout time.Duration, listOptions metav1.ListOptions) error {
	// Create a context
	ctx := context.Background()

	// Watch for job events
	watch, err := clientset.BatchV1().Jobs(namespace).Watch(ctx, listOptions)
	if err != nil {
		return fmt.Errorf("Error creating watcher: %v", err)
	}
	defer watch.Stop()

	// Channel to receive events
	eventChan := watch.ResultChan()

	// Loop to wait for job completion
	for {
		select {
		case event := <-eventChan:
			job, ok := event.Object.(*batchv1.Job)
			if !ok {
				return fmt.Errorf("Unexpected event received: %v", event)
			}
			if job.Name == jobName {
				// Check if the job has completed
				if job.Status.CompletionTime != nil {
					// The job has completed, check if it succeeded
					if job.Status.Succeeded == 1 {
						return nil // Job completed successfully
					} else {
						return fmt.Errorf("Job completed with errors")
					}
				}
			}
		case <-time.After(timeout):
			return fmt.Errorf("Timeout: job did not complete after %v minutes", timeout)
		}
	}
}
