Build a system that creates isolated Kubernetes environments for users with persistent storage and dynamic subdomains. We are using docker desktop, use an image that provides vnc access on a virtual desktop.
Requirements:

Create a go orchestration service that:

Creates isolated user environments in Kubernetes (each user has his own container, and their own volume mounts)
Provisions persistent storage via PVCs for each user
Generates and configures subdomain access for each environment
Supports destroying and recreating containers without data loss


Each user environment should include:

A VNC container (noVNC) 
Persistent storage for user data
Network isolation from other users
Auto-generated subdomain


Local testing setup:

Docker desktop cluster
DNS configuration


Example Implementation Flow:
Api is called to provision container for a user
System creates namespace, PVC, deployment, service, and ingress
User accesses their environment at username.local.dev
Data persists across container restarts/rebuilds

Build the system to be extensible for future production deployment with actual DNS and Let's Encrypt certificates.
Add swagger documentation on apis. 

Think of it, as though you had to build gitpod, but not for programming but for users to be able to manage their virtual desktops. To reduce cost we would turn off the computer but keep the volume to mount later if the user comes back online.

First we will test this locally, then use it to manage containers on a kubernetes cluster on google / aws.

This is an internal service which can be called from external services. but user management is supposed to happen on the external service. When called always provide results by calling kubectl. 

For this project, i want to expose two ports 5901 and 6901, use this image for testing: accetto/ubuntu-vnc-xfce-firefox-g3. Now these will be default, but the user should be able to change them and also the volume mounts from the api itself. 