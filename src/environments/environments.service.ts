import { Injectable, Logger, NotFoundException } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { KubernetesService } from '../kubernetes/kubernetes.service';
import { Environment, EnvironmentStatus } from './models/environment.model';
import { CreateEnvironmentDto } from './dto/create-environment.dto';
import { v4 as uuidv4 } from 'uuid';

@Injectable()
export class EnvironmentsService {
  private readonly logger = new Logger(EnvironmentsService.name);
  private environments: Map<string, Environment> = new Map();

  constructor(
    private readonly k8sService: KubernetesService,
    private readonly configService: ConfigService,
  ) { }

  private getVncImage(): string {
    return this.configService.get<string>('VNC_IMAGE', 'novnc/noVNC:latest');
  }

  private getStorageSize(): string {
    return this.configService.get<string>('STORAGE_SIZE', '10Gi');
  }

  private getBaseDomain(): string {
    return this.configService.get<string>('BASE_DOMAIN', 'local.dev');
  }

  private getNextNodePort(): number {
    const startPort = this.configService.get<number>('NODE_PORT_RANGE_START', 30000);
    const maxPort = 32767; // Kubernetes NodePort range upper bound

    // Find the highest used port
    let highestPort = startPort - 1;
    for (const env of this.environments.values()) {
      // This is simplified - in a real system you'd track this differently
      const port = parseInt(env.id, 10) + startPort;
      if (port > highestPort) {
        highestPort = port;
      }
    }

    const nextPort = highestPort + 1;
    if (nextPort > maxPort) {
      throw new Error('No available NodePorts');
    }

    return nextPort;
  }

  async findAll(): Promise<Environment[]> {
    return Array.from(this.environments.values());
  }

  async findOne(id: string): Promise<Environment> {
    const environment = this.environments.get(id);
    if (!environment) {
      throw new NotFoundException(`Environment with ID ${id} not found`);
    }
    return environment;
  }

  async findByUsername(username: string): Promise<Environment[]> {
    return Array.from(this.environments.values()).filter(env => env.username === username);
  }

  async create(createDto: CreateEnvironmentDto): Promise<Environment> {
    const id = uuidv4();
    const { username } = createDto;
    const namespace = `user-${username}`;
    const subdomain = `${username}.${this.getBaseDomain()}`;

    // Create new environment record
    const environment: Environment = {
      id,
      username,
      namespace,
      subdomain,
      status: EnvironmentStatus.CREATING,
      createdAt: new Date(),
      updatedAt: new Date(),
    };

    this.environments.set(id, environment);
    this.logger.log(`Created environment record for user ${username} with ID ${id}`);

    try {
      // Start the async provisioning process
      this.provisionEnvironment(id, createDto).catch(error => {
        this.logger.error(`Failed to provision environment ${id}: ${error.message}`);
        if (this.environments.has(id)) {
          const failedEnv = this.environments.get(id);
          if (failedEnv) {
            failedEnv.status = EnvironmentStatus.ERROR;
            failedEnv.updatedAt = new Date();
            this.environments.set(id, failedEnv);
          }
        }
      });

      return environment;
    } catch (error) {
      this.logger.error(`Failed to create environment for user ${username}: ${error.message}`);
      this.environments.delete(id);
      throw error;
    }
  }

  private async provisionEnvironment(id: string, createDto: CreateEnvironmentDto): Promise<void> {
    const environment = this.environments.get(id);
    if (!environment) {
      throw new NotFoundException(`Environment with ID ${id} not found`);
    }

    const { username } = createDto;
    const { namespace, subdomain } = environment;
    const deploymentName = `vnc-${username}`;
    const serviceName = `vnc-service-${username}`;
    const ingressName = `vnc-ingress-${username}`;
    const pvcName = `user-data-${username}`;

    try {
      // 1. Create namespace
      this.logger.log(`Creating namespace ${namespace}`);
      await this.k8sService.createNamespace(namespace, { username });

      // 2. Create PVC for persistent storage
      this.logger.log(`Creating PVC ${pvcName} in namespace ${namespace}`);
      const storageSize = createDto.storageSize || this.getStorageSize();
      await this.k8sService.createPersistentVolumeClaim(
        namespace,
        pvcName,
        storageSize,
        createDto.storageClass,
      );

      // 3. Create VNC deployment
      this.logger.log(`Creating VNC deployment ${deploymentName} in namespace ${namespace}`);
      await this.k8sService.createDeployment(
        namespace,
        deploymentName,
        this.getVncImage(),
        [{ name: 'user-data', mountPath: '/home/user/data' }],
        [{ name: 'user-data', persistentVolumeClaim: { claimName: pvcName } }],
        [{ containerPort: 8080 }],
        [{ name: 'USERNAME', value: username }],
      );

      // 4. Create service
      this.logger.log(`Creating service ${serviceName} in namespace ${namespace}`);
      const nodePort = this.getNextNodePort();
      await this.k8sService.createService(
        namespace,
        serviceName,
        [{ port: 80, targetPort: 8080 }],
        nodePort,
      );

      // 5. Create ingress for subdomain routing
      this.logger.log(`Creating ingress ${ingressName} in namespace ${namespace} for host ${subdomain}`);
      await this.k8sService.createIngress(
        namespace,
        ingressName,
        subdomain,
        serviceName,
        80,
      );

      // 6. Update environment status
      environment.status = EnvironmentStatus.RUNNING;
      environment.updatedAt = new Date();
      this.environments.set(id, environment);
      this.logger.log(`Environment ${id} for user ${username} is now running`);
    } catch (error) {
      this.logger.error(`Error provisioning environment ${id}: ${error.message}`);
      environment.status = EnvironmentStatus.ERROR;
      environment.updatedAt = new Date();
      this.environments.set(id, environment);
      throw error;
    }
  }

  async delete(id: string): Promise<void> {
    const environment = this.environments.get(id);
    if (!environment) {
      throw new NotFoundException(`Environment with ID ${id} not found`);
    }

    // Update status to indicate deletion is in progress
    environment.status = EnvironmentStatus.DELETING;
    environment.updatedAt = new Date();
    this.environments.set(id, environment);

    try {
      const { namespace, username } = environment;
      const ingressName = `vnc-ingress-${username}`;
      const serviceName = `vnc-service-${username}`;
      const deploymentName = `vnc-${username}`;

      // Delete resources in reverse order of creation
      this.logger.log(`Deleting ingress ${ingressName} in namespace ${namespace}`);
      await this.k8sService.deleteIngress(namespace, ingressName);

      this.logger.log(`Deleting service ${serviceName} in namespace ${namespace}`);
      await this.k8sService.deleteService(namespace, serviceName);

      this.logger.log(`Deleting deployment ${deploymentName} in namespace ${namespace}`);
      await this.k8sService.deleteDeployment(namespace, deploymentName);

      // Note: We don't delete the PVC to ensure data persistence
      // We also don't delete the namespace in case there are other resources

      // Remove environment from our records
      this.environments.delete(id);
      this.logger.log(`Environment ${id} for user ${username} has been deleted`);
    } catch (error) {
      this.logger.error(`Error deleting environment ${id}: ${error.message}`);
      environment.status = EnvironmentStatus.ERROR;
      environment.updatedAt = new Date();
      this.environments.set(id, environment);
      throw error;
    }
  }

  async restart(id: string): Promise<Environment> {
    const environment = this.environments.get(id);
    if (!environment) {
      throw new NotFoundException(`Environment with ID ${id} not found`);
    }

    // Update status to indicate restart is in progress
    environment.status = EnvironmentStatus.RESTARTING;
    environment.updatedAt = new Date();
    this.environments.set(id, environment);

    try {
      const { namespace, username } = environment;
      const deploymentName = `vnc-${username}`;

      // Restart only the deployment (this will recreate the pods)
      this.logger.log(`Restarting deployment ${deploymentName} in namespace ${namespace}`);
      await this.k8sService.deleteDeployment(namespace, deploymentName);

      // Recreate the deployment
      await this.k8sService.createDeployment(
        namespace,
        deploymentName,
        this.getVncImage(),
        [{ name: 'user-data', mountPath: '/home/user/data' }],
        [{ name: 'user-data', persistentVolumeClaim: { claimName: `user-data-${username}` } }],
        [{ containerPort: 8080 }],
        [{ name: 'USERNAME', value: username }],
      );

      // Update status
      environment.status = EnvironmentStatus.RUNNING;
      environment.updatedAt = new Date();
      this.environments.set(id, environment);

      this.logger.log(`Environment ${id} for user ${username} has been restarted`);
      return environment;
    } catch (error) {
      this.logger.error(`Error restarting environment ${id}: ${error.message}`);
      environment.status = EnvironmentStatus.ERROR;
      environment.updatedAt = new Date();
      this.environments.set(id, environment);
      throw error;
    }
  }
}