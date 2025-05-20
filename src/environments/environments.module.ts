import { Module } from '@nestjs/common';
import { EnvironmentsService } from './environments.service';
import { EnvironmentsController } from './environments.controller';
import { KubernetesModule } from '../kubernetes/kubernetes.module';

@Module({
  imports: [KubernetesModule],
  providers: [EnvironmentsService],
  controllers: [EnvironmentsController]
})
export class EnvironmentsModule {}