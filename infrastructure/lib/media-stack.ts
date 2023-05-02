import { GoFunction } from '@aws-cdk/aws-lambda-go-alpha';
import { Stack, StackProps } from 'aws-cdk-lib';
import { Construct } from 'constructs';
import { Table } from 'aws-cdk-lib/aws-dynamodb';
import { HttpLambdaIntegration } from '@aws-cdk/aws-apigatewayv2-integrations-alpha';
import { AddRoutesOptions, HttpApi, HttpMethod } from '@aws-cdk/aws-apigatewayv2-alpha';
import { LambdaInvoke } from 'aws-cdk-lib/aws-stepfunctions-tasks';
import { StateMachine } from 'aws-cdk-lib/aws-stepfunctions';
import { FUNCTIONS_FOLDER } from '../config';
import { HttpStepFunctionsIntegration } from './http-state-machine-integration';

export interface MediaStackProps extends StackProps {
    mediaTable: Table,
    leaderboardTable: Table
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

        props.leaderboardTable.grantReadWriteData(backfillPutFunction);
        props.leaderboardTable.grantReadWriteData(statusUpdatePutFunction);

        const backfillPutTask = new LambdaInvoke(this, 'backfillPutInvoke', {
            lambdaFunction: backfillPutFunction,
            outputPath: '$.Payload'
        });
        backfillPutTask.addRetry();
        const backfillPutStateMachine = new StateMachine(this, 'backfillPutStateMachine', {
            definition: backfillPutTask
        });

        const mediaInfoGetIntegration = new HttpLambdaIntegration('mediaInfoGetIntegration', mediaInfoGetFunction);
        const mediaInfoPutIntegration = new HttpLambdaIntegration('mediaInfoPutIntegration', mediaInfoPutFunction);
        const backfillGetIntegration = new HttpLambdaIntegration('backfillGetIntegration', backfillGetFunction);
        const statusUpdateGetIntegration = new HttpLambdaIntegration('statusUpdateGetIntegration', statusUpdateGetFunction);
        const statusUpdatePutIntegration = new HttpLambdaIntegration('statusUpdatePutIntegration', statusUpdatePutFunction);

        const backfillPutIntegration = new HttpStepFunctionsIntegration('backfillPutIntegration', {
            stateMachine: backfillPutStateMachine
        })

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
