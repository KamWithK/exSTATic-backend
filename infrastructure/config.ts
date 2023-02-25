import { Environment } from 'aws-cdk-lib';

export const FUNCTIONS_FOLDER = '../functions/'

export const PIPELINE_CONFIG = {
    'repo_string': 'KamWithK/exSTATic-backend',
    'branch': 'master'
}

const SYDNEY_REGION = 'ap-southeast-2'

export const DEV_ENV_ENVIRONMENT: Environment = {
    'account': '868004641356',
    'region': SYDNEY_REGION
}
export const TEST_ENV_ENVIRONMENT: Environment = {
    'account': '305354055033',
    'region': SYDNEY_REGION
}
export const PROD_ENV_ENVIRONMENT: Environment = {
    'account': '136138178459',
    'region': SYDNEY_REGION
}
