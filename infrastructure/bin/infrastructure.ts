#!/usr/bin/env node
import 'source-map-support/register';
import * as cdk from 'aws-cdk-lib';
import { CodePipelineStack } from '../lib/codepipeline-stack';
import { DEV_ENV_ENVIRONMENT } from '../config';

const app = new cdk.App();
new CodePipelineStack(app, 'codePipelineStack', {
    env: DEV_ENV_ENVIRONMENT
});
