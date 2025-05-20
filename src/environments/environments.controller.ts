import { Controller, Get, Post, Delete, Param, Body, HttpCode, HttpStatus, Put, Query } from '@nestjs/common';
import { EnvironmentsService } from './environments.service';
import { CreateEnvironmentDto } from './dto/create-environment.dto';
import { Environment } from './models/environment.model';
import { ApiTags, ApiOperation, ApiResponse, ApiParam, ApiQuery, ApiBody } from '@nestjs/swagger';

@ApiTags('environments')
@Controller('environments')
export class EnvironmentsController {
  constructor(private readonly environmentsService: EnvironmentsService) {}

  @Get()
  @ApiOperation({ summary: 'Get all environments', description: 'Returns all environments or filters by username if provided' })
  @ApiQuery({ name: 'username', required: false, description: 'Filter environments by username' })
  @ApiResponse({ status: 200, description: 'List of environments returned successfully', type: [Environment] })
  async findAll(@Query('username') username?: string): Promise<Environment[]> {
    if (username) {
      return this.environmentsService.findByUsername(username);
    }
    return this.environmentsService.findAll();
  }

  @Get(':id')
  @ApiOperation({ summary: 'Get environment by ID', description: 'Returns a single environment by its ID' })
  @ApiParam({ name: 'id', description: 'Environment ID' })
  @ApiResponse({ status: 200, description: 'Environment returned successfully', type: Environment })
  @ApiResponse({ status: 404, description: 'Environment not found' })
  async findOne(@Param('id') id: string): Promise<Environment> {
    return this.environmentsService.findOne(id);
  }

  @Post()
  @ApiOperation({ summary: 'Create new environment', description: 'Creates a new isolated user environment in Kubernetes' })
  @ApiBody({ type: CreateEnvironmentDto, description: 'Environment creation parameters' })
  @ApiResponse({ status: 201, description: 'Environment created successfully', type: Environment })
  @ApiResponse({ status: 400, description: 'Invalid input' })
  async create(@Body() createEnvironmentDto: CreateEnvironmentDto): Promise<Environment> {
    return this.environmentsService.create(createEnvironmentDto);
  }

  @Delete(':id')
  @ApiOperation({ summary: 'Delete environment', description: 'Deletes an environment by its ID' })
  @ApiParam({ name: 'id', description: 'Environment ID' })
  @ApiResponse({ status: 204, description: 'Environment deleted successfully' })
  @ApiResponse({ status: 404, description: 'Environment not found' })
  @HttpCode(HttpStatus.NO_CONTENT)
  async delete(@Param('id') id: string): Promise<void> {
    return this.environmentsService.delete(id);
  }

  @Put(':id/restart')
  @ApiOperation({ summary: 'Restart environment', description: 'Restarts an environment without data loss' })
  @ApiParam({ name: 'id', description: 'Environment ID' })
  @ApiResponse({ status: 200, description: 'Environment restarted successfully', type: Environment })
  @ApiResponse({ status: 404, description: 'Environment not found' })
  async restart(@Param('id') id: string): Promise<Environment> {
    return this.environmentsService.restart(id);
  }
}