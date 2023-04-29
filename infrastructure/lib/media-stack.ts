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
        const backfillGetFunction = new GoFunction(this, 'backfillGetFunction', {
            entry: FUNCTIONS_FOLDER + 'backfill/get'
        });
        const backfillPutFunction = new GoFunction(this, 'backfillPutFunction', {
            entry: FUNCTIONS_FOLDER + 'backfill/put'
        });
        const statusUpdateGetFunction = new GoFunction(this, 'statusUpdateGetFunction', {
            entry: FUNCTIONS_FOLDER + 'status_update/get'
        });
        const statusUpdatePutFunction = new GoFunction(this, 'statusUpdatePutFunction', {
            entry: FUNCTIONS_FOLDER + 'status_update/put'
        });

        props.mediaTable.grantReadWriteData(mediaInfoGetFunction);
        props.mediaTable.grantReadWriteData(mediaInfoPutFunction);
        props.mediaTable.grantReadWriteData(backfillGetFunction);
        props.mediaTable.grantReadWriteData(backfillPutFunction);
        props.mediaTable.grantReadWriteData(statusUpdateGetFunction);
        props.mediaTable.grantReadWriteData(statusUpdatePutFunction);

        const mediaInfoGetIntegration = new HttpLambdaIntegration('mediaInfoGetIntegration', mediaInfoGetFunction);
        const mediaInfoPutIntegration = new HttpLambdaIntegration('mediaInfoPutIntegration', mediaInfoPutFunction);
        const backfillGetIntegration = new HttpLambdaIntegration('backfillGetIntegration', backfillGetFunction);
        const backfillPutIntegration = new HttpLambdaIntegration('backfillPutIntegration', backfillPutFunction);
        const statusUpdateGetIntegration = new HttpLambdaIntegration('statusUpdateGetIntegration', statusUpdateGetFunction);
        const statusUpdatePutIntegration = new HttpLambdaIntegration('statusUpdatePutIntegration', statusUpdatePutFunction);

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
        const backfillGetRouteOptions: AddRoutesOptions = {
            path: '/backfill/get',
            methods: [HttpMethod.GET],
            integration: backfillGetIntegration
        };
        const backfillPutRouteOptions: AddRoutesOptions = {
            path: '/backfill/put',
            methods: [HttpMethod.PUT],
            integration: backfillPutIntegration
        };
        const statusUpdateGetRouteOptions: AddRoutesOptions = {
            path: '/statusUpdate/get',
            methods: [HttpMethod.GET],
            integration: statusUpdateGetIntegration
        };
        const statusUpdatePutRouteOptions: AddRoutesOptions = {
            path: '/statusUpdate/put',
            methods: [HttpMethod.PUT],
            integration: statusUpdatePutIntegration
        };

        this.routeOptions = [
            mediaInfoGetRouteOptions,
            mediaInfoPutRouteOptions,
            backfillGetRouteOptions,
            backfillPutRouteOptions,
            statusUpdateGetRouteOptions,
            statusUpdatePutRouteOptions
        ];
    }
}
