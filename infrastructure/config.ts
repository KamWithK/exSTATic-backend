import { Environment } from 'aws-cdk-lib';
import { ConnectionSourceOptions } from 'aws-cdk-lib/pipelines'

export const FUNCTIONS_FOLDER = '../functions/';
export const INFRASTRUCTURE_FOLDER = 'infrastructure';


const SYDNEY_REGION = 'ap-southeast-2';

export const DEV_ENV_ENVIRONMENT: Environment = {
    account: '868004641356',
    region: SYDNEY_REGION
};
export const TEST_ENV_ENVIRONMENT: Environment = {
    account: '305354055033',
    region: SYDNEY_REGION
};
export const PROD_ENV_ENVIRONMENT: Environment = {
    account: '136138178459',
    region: SYDNEY_REGION
};


const GITHUB_CONNECTION: ConnectionSourceOptions = {
    connectionArn: `arn:aws:codestar-connections:${DEV_ENV_ENVIRONMENT.region}:${DEV_ENV_ENVIRONMENT.account}:connection/304019ff-fd81-4f11-8fff-52f2fa04df10`
};

export const PIPELINE_CONFIG = {
    'repo_string': 'KamWithK/exSTATic-backend',
    'branch': 'master',
    'connection': GITHUB_CONNECTION
};
