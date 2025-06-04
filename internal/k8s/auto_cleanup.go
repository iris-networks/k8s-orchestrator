package k8s

import (
	"context"
	"errors"
	"log"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// ResourceExpirationTime is the duration after which a sandbox will be automatically deleted
	ResourceExpirationTime = 15 * time.Minute
	// DefaultAuthToken is the auth token for external cleanup trigger
	DefaultAuthToken = "k8s-auto-cleanup-token"
)

// CleanupExpiredSandboxesByDuration performs cleanup of sandboxes older than the specified duration
// This function can be triggered via API and requires authentication
func (c *ClientWithTraefik) CleanupExpiredSandboxesByDuration(ctx context.Context, duration time.Duration, authToken string) error {
	// Validate the auth token
	if authToken != DefaultAuthToken {
		return errors.New("unauthorized: invalid auth token")
	}

	log.Printf("Targeting namespace: %s", c.namespace)

	// Define now here so we can use it consistently throughout the function
	now := time.Now()
	
	// Get all deployments in the namespace
	deployments, err := c.clientset.AppsV1().Deployments(c.namespace).List(ctx, metav1.ListOptions{
		// No label selector here to find ALL deployments
	})
	if err != nil {
		return err
	}

	// Print all deployments found in the namespace with detailed info
	log.Printf("Found %d deployments in namespace %s", len(deployments.Items), c.namespace)
	for i, deployment := range deployments.Items {
		createdTime := deployment.CreationTimestamp.Time
		age := now.Sub(createdTime)
		userID := deployment.Labels["user"]
		log.Printf("[%d] Deployment: %s, User: %s, Created: %s, Age: %v",
			i+1, deployment.Name, userID, createdTime.Format(time.RFC3339), age.Round(time.Second))
	}

	log.Printf("Running external cleanup for sandboxes older than %v", duration)
	cleanupCount := 0

	for _, deployment := range deployments.Items {
		// Check if the deployment has been running for more than the specified duration
		creationTime := deployment.CreationTimestamp.Time
		age := now.Sub(creationTime)

		log.Printf("Age of deployment %s: %v (comparing with %v)", deployment.Name, age, duration)
		// Debug log for exact comparison
		log.Printf("Duration comparison: age (%v) >= duration (%v) = %v",
			age.Seconds(), duration.Seconds(), age >= duration)
		if age >= duration {
			// Extract user ID from labels or deployment name
			userID := deployment.Labels["user"]
			log.Printf("Labels for deployment %s: %v", deployment.Name, deployment.Labels)

			// If user label is empty, try to extract from deployment name
			if userID == "" {
				log.Printf("No 'user' label found for deployment %s, trying to extract from name", deployment.Name)
				
				// Try to handle deployment name format: iris-{userId}-deployment
				if strings.HasPrefix(deployment.Name, "iris-") && strings.HasSuffix(deployment.Name, "-deployment") {
					// Extract the middle part between "iris-" and "-deployment"
					namePart := strings.TrimPrefix(deployment.Name, "iris-")
					extractedID := strings.TrimSuffix(namePart, "-deployment")
					userID = extractedID
					log.Printf("Extracted userID '%s' from iris-specific deployment name", userID)
				} else if strings.HasSuffix(deployment.Name, "-deployment") {
					// Try the standard format: {userId}-deployment
					userID = strings.TrimSuffix(deployment.Name, "-deployment")
					log.Printf("Extracted userID '%s' from standard deployment name", userID)
				} else {
					// Last resort: try to split by dash and take the second part
					parts := strings.Split(deployment.Name, "-")
					if len(parts) >= 2 {
						userID = parts[1]
						log.Printf("Extracted userID '%s' using fallback method", userID)
					} else {
						log.Printf("Could not extract userID from deployment name: %s", deployment.Name)
						continue
					}
				}
			}

			if userID == "" {
				log.Printf("Skipping deployment %s: no userID found", deployment.Name)
				continue
			}

			log.Printf("Attempting to delete sandbox for user %s (age: %v)", userID, age.Round(time.Second))
			if err := c.DeleteSandbox(userID); err != nil {
				log.Printf("Error deleting sandbox for user %s: %v", userID, err)
				// Continue with other sandboxes even if this one fails
			} else {
				log.Printf("Successfully deleted sandbox for user %s", userID)
				cleanupCount++
			}
		}
	}

	log.Printf("External cleanup completed: %d sandboxes removed", cleanupCount)
	return nil
}