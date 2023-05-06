#!/usr/bin/env node

import { config } from 'dotenv';
import 'source-map-support/register';
import * as cdk from 'aws-cdk-lib';
import { env } from 'process';
import { CodePipelineStack } from '../lib/codepipeline-stack';
import { DEV_ENV_ENVIRONMENT } from '../config';
import { EnvironmentStage } from '../lib/environment-stage';

config();

const app = new cdk.App();

if (env.PIPELINE?.toLowerCase() === 'false') {
    const devStage = new EnvironmentStage(app, 'localStage', {
        environmentType: 'local'
    });
} else {
    new CodePipelineStack(app, 'codePipelineStack', {
        env: DEV_ENV_ENVIRONMENT
    });
}

app.synth();
