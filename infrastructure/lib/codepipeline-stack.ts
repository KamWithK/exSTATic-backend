import { Construct } from 'constructs';
import { Stack, StackProps } from 'aws-cdk-lib';
import { CodePipeline, CodePipelineSource, ManualApprovalStep, ShellStep } from 'aws-cdk-lib/pipelines';

import { PIPELINE_CONFIG, DEV_ENV_ENVIRONMENT, TEST_ENV_ENVIRONMENT, PROD_ENV_ENVIRONMENT } from '../config';
import { CodePipelineStage } from './codepipeline-stage';

export class CodePipelineStack extends Stack {
    constructor(scope: Construct, id: string, props: StackProps) {
        super(scope, id, props);

        const synth = new ShellStep('synth', {
            input: CodePipelineSource.gitHub(PIPELINE_CONFIG.repo_string, PIPELINE_CONFIG.branch),
            commands: ['cd infrastructure', 'npm ci', 'npm run build', 'npx cdk synth']
        });
        const pipeline = new CodePipeline(this, 'pipeline', {
            pipelineName: 'pipeline',
            synth: synth,
            crossAccountKeys: true
        });

        const devStage = new CodePipelineStage(this, 'devStage', {
            env: DEV_ENV_ENVIRONMENT,
            environmentType: 'dev'
        });
        const testStage = new CodePipelineStage(this, 'testStage', {
            env: TEST_ENV_ENVIRONMENT,
            environmentType: 'test'
        });
        const prodStage = new CodePipelineStage(this, 'prodStage', {
            env: PROD_ENV_ENVIRONMENT,
            environmentType: 'prod'
        });

        pipeline.addStage(devStage);
        pipeline.addStage(testStage, {
            pre: [
                new ManualApprovalStep('testStageApproval')
            ]
        });
        pipeline.addStage(prodStage, {
            pre: [
                new ManualApprovalStep('prodStageApproval')
            ]
        });
    }
}
