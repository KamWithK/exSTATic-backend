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
        
        const mediaInfoGetFunction = new GoFunction(this, 'mediaInfoGetFunction', {
            entry: FUNCTIONS_FOLDER + 'media_info/get'
        });
        const mediaInfoPutFunction = new GoFunction(this, 'mediaInfoPutFunction', {
            entry: FUNCTIONS_FOLDER + 'media_info/put'
        });
        
        const backfillFunction = new GoFunction(this, 'backfillFunction', {
            entry: FUNCTIONS_FOLDER + 'backfill'
        });
        
        const statusUpdateFunction = new GoFunction(this, 'statusUpdateFunction', {
            entry: FUNCTIONS_FOLDER + 'status_update'
        });
        
        props.mediaTable.grantReadWriteData(mediaInfoGetFunction);
        props.mediaTable.grantReadWriteData(mediaInfoPutFunction);
        props.mediaTable.grantReadWriteData(backfillFunction);
        props.mediaTable.grantReadWriteData(statusUpdateFunction);

        const mediaInfoGetIntegration = new HttpLambdaIntegration('mediaInfoIntegration', mediaInfoGetFunction);
        const mediaInfoPutIntegration = new HttpLambdaIntegration('mediaInfoIntegration', mediaInfoPutFunction);
        const backfillIntegration = new HttpLambdaIntegration('backfillIntegration', backfillFunction);
        const statusUpdateIntegration = new HttpLambdaIntegration('statusUpdateIntegration', statusUpdateFunction);

        const mediaInfoGetRouteOptions: AddRoutesOptions = {
            path: '/mediaInfo/get',
            methods: [HttpMethod.GET],
            integration: mediaInfoGetIntegration
        };
        const mediaInfoPutRouteOptions: AddRoutesOptions = {
            path: '/mediaInfo/put',
            methods: [HttpMethod.PUT],
            integration: mediaInfoPutIntegration
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
            mediaInfoGetRouteOptions,
            mediaInfoPutRouteOptions,
            backfillRouteOptions,
            statusUpdateRouteOptions
        ];
    }
}
