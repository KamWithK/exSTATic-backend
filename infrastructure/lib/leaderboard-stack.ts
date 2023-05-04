import { Stack, StackProps } from 'aws-cdk-lib';
import { Construct } from 'constructs';
import { Table } from 'aws-cdk-lib/aws-dynamodb';
import { AddRoutesOptions, HttpMethod } from '@aws-cdk/aws-apigatewayv2-alpha';
import { GoFunction } from '@aws-cdk/aws-lambda-go-alpha';
import { HttpLambdaIntegration } from '@aws-cdk/aws-apigatewayv2-integrations-alpha';
import { FUNCTIONS_FOLDER } from '../config';

export interface SettingsStackProps extends StackProps {
    leaderboardTable: Table
}

export class LeaderboardStack extends Stack {
    routeOptions: AddRoutesOptions[];

    constructor(scope: Construct, id: string, props: SettingsStackProps) {
        super(scope, id, props);

        const leaderboardFunction = new GoFunction(this, 'leaderboardFunction', {
            entry: FUNCTIONS_FOLDER + 'leaderboard'
        });

        props.leaderboardTable.grantReadWriteData(leaderboardFunction);

        const leaderboardIntegration = new HttpLambdaIntegration('leaderboardIntegration', leaderboardFunction);

        const leaderboardRouteOptions: AddRoutesOptions = {
            path: '/leaderboard',
            methods: [HttpMethod.GET],
            integration: leaderboardIntegration
        };

        this.routeOptions = [
            leaderboardRouteOptions
        ];
    }
}
