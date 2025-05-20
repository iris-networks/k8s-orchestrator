import { ApiProperty } from '@nestjs/swagger';

export enum EnvironmentStatus {
  CREATING = 'creating',
  RUNNING = 'running',
  ERROR = 'error',
  DELETING = 'deleting',
  RESTARTING = 'restarting',
}

export class Environment {
  @ApiProperty({
    description: 'Unique identifier for the environment',
    example: 'a1b2c3d4-e5f6-7890-abcd-ef1234567890',
  })
  id: string;

  @ApiProperty({
    description: 'Username associated with this environment',
    example: 'alice',
  })
  username: string;

  @ApiProperty({
    description: 'Kubernetes namespace for this environment',
    example: 'user-alice',
  })
  namespace: string;

  @ApiProperty({
    description: 'Subdomain for accessing the environment',
    example: 'alice.local.dev',
  })
  subdomain: string;

  @ApiProperty({
    description: 'Current status of the environment',
    enum: EnvironmentStatus,
    example: EnvironmentStatus.RUNNING,
  })
  status: EnvironmentStatus;

  @ApiProperty({
    description: 'Creation timestamp',
    example: '2023-09-01T12:00:00.000Z',
  })
  createdAt: Date;

  @ApiProperty({
    description: 'Last update timestamp',
    example: '2023-09-01T12:05:30.000Z',
  })
  updatedAt: Date;
}