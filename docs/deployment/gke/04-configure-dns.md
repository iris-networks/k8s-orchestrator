# Configuring DNS with Namecheap for tryiris.dev

After deploying your GKE cluster and ingress controller, you'll need to configure the DNS records in Namecheap:

## Get the Ingress Controller's External IP

1. Get the Traefik Ingress Controller's external IP address:

```bash
# Get the Traefik IP
INGRESS_IP=$(kubectl get service traefik -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
echo $INGRESS_IP  # Make note of this IP for Namecheap configuration
```

## Configure DNS Records

2. In Namecheap's dashboard:
   - Go to Dashboard → Domain List → Manage → Advanced DNS
   - Add the following A records:

   | Type | Host | Value | TTL |
   |------|------|-------|-----|
   | A Record | api.pods | {INGRESS_IP} | Automatic |
   | A Record | *.pods | {INGRESS_IP} | Automatic |

3. Wait for DNS propagation (can take up to 24-48 hours, but often much faster)

## Verify DNS Configuration

You can verify that the DNS configuration is working by using the `dig` command:

```bash
# Check the API domain
dig api.pods.tryiris.dev

# Check a wildcard domain (example user)
dig user1.pods.tryiris.dev
```

You should see the INGRESS_IP in the response. If not, wait longer for DNS propagation or check your Namecheap configuration.

## Next Step

Once your DNS is configured, proceed to [Deploy with Helm](05-deploy-with-helm.md).