import { IsNotEmpty, IsOptional, IsString } from 'class-validator';
import { ApiProperty } from '@nestjs/swagger';

export class CreateEnvironmentDto {
  @ApiProperty({
    description: 'Username for the environment',
    example: 'alice',
  })
  @IsNotEmpty()
  @IsString()
  username: string;

  @ApiProperty({
    description: 'Storage size for the persistent volume claim',
    example: '10Gi',
    required: false,
  })
  @IsOptional()
  @IsString()
  storageSize?: string;

  @ApiProperty({
    description: 'Storage class for the persistent volume claim',
    example: 'standard',
    required: false,
  })
  @IsOptional()
  @IsString()
  storageClass?: string;
}