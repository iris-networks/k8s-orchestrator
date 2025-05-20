import { Module } from '@nestjs/common';
import { AppController } from './app.controller';
import { AppService } from './app.service';
import { KubernetesModule } from './kubernetes/kubernetes.module';
import { EnvironmentsModule } from './environments/environments.module';
import { ConfigModule } from './config/config.module';

@Module({
  imports: [KubernetesModule, EnvironmentsModule, ConfigModule],
  controllers: [AppController],
  providers: [AppService],
})
export class AppModule {}
