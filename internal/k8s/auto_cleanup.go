package k8s

import (
	"context"
	"log"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// ResourceExpirationTime is the duration after which a sandbox will be automatically deleted
	ResourceExpirationTime = 15 * time.Minute
)

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
	log.Println("Auto cleanup service started - sandboxes will be deleted after 15 minutes")
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
		// Check if the deployment has been running for more than ResourceExpirationTime
		creationTime := deployment.CreationTimestamp.Time
		age := now.Sub(creationTime)

		if age >= ResourceExpirationTime {
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
	log.Println("Auto cleanup service started - sandboxes will be deleted after 15 minutes")
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
		// Check if the deployment has been running for more than ResourceExpirationTime
		creationTime := deployment.CreationTimestamp.Time
		age := now.Sub(creationTime)

		if age >= ResourceExpirationTime {
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