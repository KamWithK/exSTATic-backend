import { GoFunction } from '@aws-cdk/aws-lambda-go-alpha';
import { Stack, StackProps } from 'aws-cdk-lib';
import { Construct } from 'constructs';
import { Table } from 'aws-cdk-lib/aws-dynamodb';
import { HttpLambdaIntegration } from '@aws-cdk/aws-apigatewayv2-integrations-alpha';
import { AddRoutesOptions, HttpMethod } from '@aws-cdk/aws-apigatewayv2-alpha';
import { FUNCTIONS_FOLDER } from '../config';

export interface MediaStackProps extends StackProps {
    mediaTable: Table
}

export class MediaStack extends Stack {
    routeOptions: AddRoutesOptions[];

    constructor(scope: Construct, id: string, props: MediaStackProps) {
        super(scope, id, props);
        
        const mediaInfoFunction = new GoFunction(this, 'mediaInfoFunction', {
            entry: FUNCTIONS_FOLDER + 'media_info'
        });
        
        const backfillFunction = new GoFunction(this, 'backfillFunction', {
            entry: FUNCTIONS_FOLDER + 'backfill'
        });
        
        const statusUpdateFunction = new GoFunction(this, 'statusUpdateFunction', {
            entry: FUNCTIONS_FOLDER + 'status_update'
        });
        
        props.mediaTable.grantReadWriteData(mediaInfoFunction);
        props.mediaTable.grantReadWriteData(backfillFunction);
        props.mediaTable.grantReadWriteData(statusUpdateFunction);

        const mediaInfoIntegration = new HttpLambdaIntegration('mediaInfoIntegration', mediaInfoFunction);
        const backfillIntegration = new HttpLambdaIntegration('backfillIntegration', backfillFunction);
        const statusUpdateIntegration = new HttpLambdaIntegration('statusUpdateIntegration', statusUpdateFunction);

        const mediaInfoRouteOptions: AddRoutesOptions = {
            path: '/mediaInfo',
            methods: [HttpMethod.ANY],
            integration: mediaInfoIntegration
        };
        const backfillRouteOptions: AddRoutesOptions = {
            path: '/backfill',
            methods: [HttpMethod.ANY],
            integration: backfillIntegration
        };
        const statusUpdateRouteOptions: AddRoutesOptions = {
            path: '/statusUpdate',
            methods: [HttpMethod.ANY],
            integration: statusUpdateIntegration
        };

        this.routeOptions = [
            mediaInfoRouteOptions,
            backfillRouteOptions,
            statusUpdateRouteOptions
        ];
    }
}
