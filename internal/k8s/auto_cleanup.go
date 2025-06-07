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
	// DefaultAuthToken is the auth token for external cleanup trigger
	DefaultAuthToken = "k8s-auto-cleanup-token"
)

// ResourceExpirationTime is the duration after which a sandbox will be automatically deleted
// Deprecated: Use Config.SandboxTimeoutDuration instead
var ResourceExpirationTime time.Duration

// StartAutoCleanupService starts a background goroutine that periodically checks for sandboxes
// that have been running for more than ResourceExpirationTime and deletes them
func (c *Client) StartAutoCleanupService(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Println("Auto cleanup service stopped")
				return
			case <-ticker.C:
				if err := c.cleanupExpiredSandboxes(ctx); err != nil {
					log.Printf("Error cleaning up sandboxes: %v", err)
				}
			}
		}
	}()
	timeoutMinutes := int(c.config.SandboxTimeoutDuration.Minutes())
	log.Printf("Auto cleanup service started - sandboxes will be deleted after %d minutes", timeoutMinutes)
}

// cleanupExpiredSandboxes checks for and deletes sandboxes that have been running for too long
func (c *Client) cleanupExpiredSandboxes(ctx context.Context) error {
	// Get all deployments in the namespace
	deployments, err := c.clientset.AppsV1().Deployments(c.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "app=user-sandbox",
	})
	if err != nil {
		return err
	}

	now := time.Now()
	for _, deployment := range deployments.Items {
		// Check if the deployment has been running for more than the configured timeout
		creationTime := deployment.CreationTimestamp.Time
		age := now.Sub(creationTime)

		if age >= c.config.SandboxTimeoutDuration {
			// Extract user ID from labels or deployment name
			userID := deployment.Labels["user"]
			if userID == "" {
				continue
			}

			log.Printf("Deleting sandbox for user %s (age: %v)", userID, age.Round(time.Second))
			if err := c.DeleteSandbox(userID); err != nil {
				log.Printf("Error deleting sandbox for user %s: %v", userID, err)
				// Continue with other sandboxes even if this one fails
			}
		}
	}

	return nil
}


// StartAutoCleanupService starts a background goroutine for auto cleanup with Traefik support
func (c *ClientWithTraefik) StartAutoCleanupService(ctx context.Context) {
	// Reuse the base client's implementation but use our DeleteSandbox
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Println("Auto cleanup service stopped")
				return
			case <-ticker.C:
				if err := c.cleanupExpiredSandboxes(ctx); err != nil {
					log.Printf("Error cleaning up sandboxes: %v", err)
				}
			}
		}
	}()
	timeoutMinutes := int(c.config.SandboxTimeoutDuration.Minutes())
	log.Printf("Auto cleanup service started - sandboxes will be deleted after %d minutes", timeoutMinutes)
}

// cleanupExpiredSandboxes checks for and deletes sandboxes that have been running for too long
func (c *ClientWithTraefik) cleanupExpiredSandboxes(ctx context.Context) error {
	// Get all deployments in the namespace
	deployments, err := c.clientset.AppsV1().Deployments(c.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "app=user-sandbox",
	})
	if err != nil {
		return err
	}

	now := time.Now()
	for _, deployment := range deployments.Items {
		// Check if the deployment has been running for more than the configured timeout
		creationTime := deployment.CreationTimestamp.Time
		age := now.Sub(creationTime)

		if age >= c.config.SandboxTimeoutDuration {
			// Extract user ID from labels or deployment name
			userID := deployment.Labels["user"]
			if userID == "" {
				continue
			}

			log.Printf("Deleting sandbox for user %s (age: %v)", userID, age.Round(time.Second))
			if err := c.DeleteSandbox(userID); err != nil {
				log.Printf("Error deleting sandbox for user %s: %v", userID, err)
				// Continue with other sandboxes even if this one fails
			}
		}
	}

	return nil
}

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
				
				// Try to handle deployment name format: {userId}-deployment
				if strings.HasSuffix(deployment.Name, "-deployment") {
					// Standard format: {userId}-deployment
					userID = strings.TrimSuffix(deployment.Name, "-deployment")
					log.Printf("Extracted userID '%s' from deployment name", userID)
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