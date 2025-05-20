import { Injectable, Logger, OnModuleInit } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import * as k8s from '@kubernetes/client-node';

@Injectable()
export class KubernetesService implements OnModuleInit {
  private readonly logger = new Logger(KubernetesService.name);
  private k8sApi: {
    coreV1Api: k8s.CoreV1Api;
    appsV1Api: k8s.AppsV1Api;
    networkingV1Api: k8s.NetworkingV1Api;
  };
  private readonly kubeConfig = new k8s.KubeConfig();

  constructor(private configService: ConfigService) { }

  async onModuleInit() {
    try {
      // Load from default location or from env var
      const kubeConfigPath = this.configService.get<string>('KUBECONFIG');
      if (kubeConfigPath) {
        this.kubeConfig.loadFromFile(kubeConfigPath);
      } else {
        this.kubeConfig.loadFromDefault();
      }

      this.k8sApi = {
        coreV1Api: this.kubeConfig.makeApiClient(k8s.CoreV1Api),
        appsV1Api: this.kubeConfig.makeApiClient(k8s.AppsV1Api),
        networkingV1Api: this.kubeConfig.makeApiClient(k8s.NetworkingV1Api),
      };

      this.logger.log('Kubernetes client initialized successfully');
    } catch (error) {
      this.logger.error(`Failed to initialize Kubernetes client: ${error.message}`);
      throw error;
    }
  }

  // Namespace operations
  async createNamespace(name: string, labels: Record<string, string> = {}): Promise<k8s.V1Namespace> {
    try {
      const namespace = {
        metadata: {
          name,
          labels: {
            ...labels,
            'managed-by': 'k8s-service',
          },
        },
      } as k8s.V1Namespace;

      const response = await this.k8sApi.coreV1Api.createNamespace({ body: namespace });
      this.logger.log(`Created namespace: ${name}`);
      return response;
    } catch (error) {
      if (error.response?.statusCode === 409) {
        this.logger.warn(`Namespace ${name} already exists`);
        return await this.k8sApi.coreV1Api.readNamespace({ name }).then(res => res);
      }
      this.logger.error(`Failed to create namespace ${name}: ${error.message}`);
      throw error;
    }
  }

  async deleteNamespace(name: string): Promise<void> {
    try {
      await this.k8sApi.coreV1Api.deleteNamespace({ name });
      this.logger.log(`Deleted namespace: ${name}`);
    } catch (error) {
      if (error.response?.statusCode === 404) {
        this.logger.warn(`Namespace ${name} not found`);
        return;
      }
      this.logger.error(`Failed to delete namespace ${name}: ${error.message}`);
      throw error;
    }
  }

  // PVC operations
  async createPersistentVolumeClaim(
    namespace: string,
    name: string,
    storageSize: string,
    storageClass?: string,
  ): Promise<k8s.V1PersistentVolumeClaim> {
    try {
      const storageClassName = storageClass || this.configService.get<string>('STORAGE_CLASS', 'standard');
      const pvc = {
        metadata: {
          name,
          namespace,
        },
        spec: {
          accessModes: ['ReadWriteOnce'],
          resources: {
            requests: {
              storage: storageSize,
            },
          },
          storageClassName,
        },
      } as k8s.V1PersistentVolumeClaim;

      const response = await this.k8sApi.coreV1Api.createNamespacedPersistentVolumeClaim({
        namespace,
        body: pvc
      });
      this.logger.log(`Created PVC: ${name} in namespace ${namespace}`);
      return response;
    } catch (error) {
      if (error.response?.statusCode === 409) {
        this.logger.warn(`PVC ${name} already exists in namespace ${namespace}`);
        return await this.k8sApi.coreV1Api.readNamespacedPersistentVolumeClaim({ namespace, name }).then(res => res);
      }
      this.logger.error(`Failed to create PVC ${name} in namespace ${namespace}: ${error.message}`);
      throw error;
    }
  }

  // Deployment operations
  async createDeployment(
    namespace: string,
    name: string,
    image: string,
    volumeMounts: Array<{ name: string; mountPath: string }> = [],
    volumes: Array<{ name: string; persistentVolumeClaim: { claimName: string } }> = [],
    ports: Array<{ containerPort: number }> = [],
    envVars: Array<{ name: string; value: string }> = [],
  ): Promise<k8s.V1Deployment> {
    try {
      const deployment = {
        metadata: {
          name,
          namespace,
          labels: {
            app: name,
            'managed-by': 'k8s-service',
          },
        },
        spec: {
          replicas: 1,
          selector: {
            matchLabels: {
              app: name,
            },
          },
          template: {
            metadata: {
              labels: {
                app: name,
              },
            },
            spec: {
              containers: [
                {
                  name,
                  image,
                  ports,
                  env: envVars,
                  volumeMounts,
                  resources: {
                    limits: {
                      cpu: '1',
                      memory: '1Gi',
                    },
                    requests: {
                      cpu: '100m',
                      memory: '256Mi',
                    },
                  },
                },
              ],
              volumes,
            },
          },
        },
      } as k8s.V1Deployment;

      const response = await this.k8sApi.appsV1Api.createNamespacedDeployment({ namespace, body: deployment });
      this.logger.log(`Created deployment: ${name} in namespace ${namespace}`);
      return response;
    } catch (error) {
      if (error.response?.statusCode === 409) {
        this.logger.warn(`Deployment ${name} already exists in namespace ${namespace}`);
        return await this.k8sApi.appsV1Api.readNamespacedDeployment({ name, namespace }).then(res => res);
      }
      this.logger.error(`Failed to create deployment ${name} in namespace ${namespace}: ${error.message}`);
      throw error;
    }
  }

  async deleteDeployment(namespace: string, name: string): Promise<void> {
    try {
      await this.k8sApi.appsV1Api.deleteNamespacedDeployment({ name, namespace });
      this.logger.log(`Deleted deployment: ${name} from namespace ${namespace}`);
    } catch (error) {
      if (error.response?.statusCode === 404) {
        this.logger.warn(`Deployment ${name} not found in namespace ${namespace}`);
        return;
      }
      this.logger.error(`Failed to delete deployment ${name} from namespace ${namespace}: ${error.message}`);
      throw error;
    }
  }

  // Service operations
  async createService(
    namespace: string,
    name: string,
    ports: Array<{ port: number; targetPort: number; protocol?: string }>,
    nodePort?: number,
  ): Promise<k8s.V1Service> {
    try {
      const serviceType = nodePort ? 'NodePort' : 'ClusterIP';
      const servicePorts = ports.map(port => ({
        port: port.port,
        targetPort: port.targetPort,
        protocol: port.protocol || 'TCP',
        nodePort: nodePort,
      }));

      const service = {
        metadata: {
          name,
          namespace,
          labels: {
            app: name,
            'managed-by': 'k8s-service',
          },
        },
        spec: {
          selector: {
            app: name,
          },
          ports: servicePorts,
          type: serviceType,
        },
      } as k8s.V1Service;

      const response = await this.k8sApi.coreV1Api.createNamespacedService({ namespace, body: service });
      this.logger.log(`Created service: ${name} in namespace ${namespace}`);
      return response;
    } catch (error) {
      if (error.response?.statusCode === 409) {
        this.logger.warn(`Service ${name} already exists in namespace ${namespace}`);
        return await this.k8sApi.coreV1Api.readNamespacedService({ name, namespace }).then(res => res);
      }
      this.logger.error(`Failed to create service ${name} in namespace ${namespace}: ${error.message}`);
      throw error;
    }
  }

  async deleteService(namespace: string, name: string): Promise<void> {
    try {
      await this.k8sApi.coreV1Api.deleteNamespacedService({ name, namespace });
      this.logger.log(`Deleted service: ${name} from namespace ${namespace}`);
    } catch (error) {
      if (error.response?.statusCode === 404) {
        this.logger.warn(`Service ${name} not found in namespace ${namespace}`);
        return;
      }
      this.logger.error(`Failed to delete service ${name} from namespace ${namespace}: ${error.message}`);
      throw error;
    }
  }

  // Ingress operations
  async createIngress(
    namespace: string,
    name: string,
    hostname: string,
    serviceName: string,
    servicePort: number,
  ): Promise<k8s.V1Ingress> {
    try {
      const ingress = {
        metadata: {
          name,
          namespace,
          labels: {
            app: name,
            'managed-by': 'k8s-service',
          },
          annotations: {
            'kubernetes.io/ingress.class': 'nginx',
            'cert-manager.io/cluster-issuer': 'selfsigned-issuer',
          },
        },
        spec: {
          tls: [
            {
              hosts: [hostname],
              secretName: `${name}-tls`,
            },
          ],
          rules: [
            {
              host: hostname,
              http: {
                paths: [
                  {
                    path: '/',
                    pathType: 'Prefix',
                    backend: {
                      service: {
                        name: serviceName,
                        port: {
                          number: servicePort,
                        },
                      },
                    },
                  },
                ],
              },
            },
          ],
        },
      } as k8s.V1Ingress;

      const response = await this.k8sApi.networkingV1Api.createNamespacedIngress({ namespace, body: ingress });
      this.logger.log(`Created ingress: ${name} in namespace ${namespace} for host ${hostname}`);
      return response;
    } catch (error) {
      if (error.response?.statusCode === 409) {
        this.logger.warn(`Ingress ${name} already exists in namespace ${namespace}`);
        return await this.k8sApi.networkingV1Api.readNamespacedIngress({ name, namespace }).then(res => res);
      }
      this.logger.error(`Failed to create ingress ${name} in namespace ${namespace}: ${error.message}`);
      throw error;
    }
  }

  async deleteIngress(namespace: string, name: string): Promise<void> {
    try {
      await this.k8sApi.networkingV1Api.deleteNamespacedIngress({ name, namespace });
      this.logger.log(`Deleted ingress: ${name} from namespace ${namespace}`);
    } catch (error) {
      if (error.response?.statusCode === 404) {
        this.logger.warn(`Ingress ${name} not found in namespace ${namespace}`);
        return;
      }
      this.logger.error(`Failed to delete ingress ${name} from namespace ${namespace}: ${error.message}`);
      throw error;
    }
  }
}