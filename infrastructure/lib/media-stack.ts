import { GoFunction } from '@aws-cdk/aws-lambda-go-alpha';
import { Stack, StackProps } from 'aws-cdk-lib';
import { Construct } from 'constructs';
import { Table } from 'aws-cdk-lib/aws-dynamodb';
import { HttpLambdaIntegration } from '@aws-cdk/aws-apigatewayv2-integrations-alpha';
import { AddRoutesOptions, HttpMethod } from '@aws-cdk/aws-apigatewayv2-alpha';
import { LambdaInvoke } from 'aws-cdk-lib/aws-stepfunctions-tasks';
import { DefinitionBody, StateMachine } from 'aws-cdk-lib/aws-stepfunctions';
import { FUNCTIONS_FOLDER } from '../config';
import { HttpStepFunctionsIntegration } from './http-state-machine-integration';

export interface MediaStackProps extends StackProps {
    mediaTable: Table,
    leaderboardTable: Table
}

export class MediaStack extends Stack {
    routeOptions: AddRoutesOptions[];
    stateMachines: StateMachine[];

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
        const backfillPostFunction = new GoFunction(this, 'backfillPostFunction', {
            entry: FUNCTIONS_FOLDER + 'backfill/post'
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
        props.mediaTable.grantReadWriteData(backfillPostFunction);
        props.mediaTable.grantReadWriteData(statusUpdateGetFunction);
        props.mediaTable.grantReadWriteData(statusUpdatePutFunction);

        props.leaderboardTable.grantReadWriteData(backfillPostFunction);
        props.leaderboardTable.grantReadWriteData(statusUpdatePutFunction);

        const backfillPostTask = new LambdaInvoke(this, 'backfillPostInvoke', {
            lambdaFunction: backfillPostFunction,
            outputPath: '$.Payload'
        });
        backfillPostTask.addRetry();
        const backfillPostStateMachine = new StateMachine(this, 'backfillPostStateMachine', {
            definitionBody: DefinitionBody.fromChainable(backfillPostTask)
        });

        const mediaInfoGetIntegration = new HttpLambdaIntegration('mediaInfoGetIntegration', mediaInfoGetFunction);
        const mediaInfoPutIntegration = new HttpLambdaIntegration('mediaInfoPutIntegration', mediaInfoPutFunction);
        const backfillGetIntegration = new HttpLambdaIntegration('backfillGetIntegration', backfillGetFunction);
        const statusUpdateGetIntegration = new HttpLambdaIntegration('statusUpdateGetIntegration', statusUpdateGetFunction);
        const statusUpdatePutIntegration = new HttpLambdaIntegration('statusUpdatePutIntegration', statusUpdatePutFunction);

        const backfillPostIntegration = new HttpStepFunctionsIntegration('backfillPostIntegration', {
            stateMachine: backfillPostStateMachine
        });

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
        const backfillPostRouteOptions: AddRoutesOptions = {
            path: '/backfill/post',
            methods: [HttpMethod.POST],
            integration: backfillPostIntegration
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
            backfillPostRouteOptions,
            statusUpdateGetRouteOptions,
            statusUpdatePutRouteOptions
        ];
    }
}
