import { GoFunction } from '@aws-cdk/aws-lambda-go-alpha';
import { Stack, StackProps } from 'aws-cdk-lib';
import { Construct } from 'constructs';
import { Table } from 'aws-cdk-lib/aws-dynamodb';
import { HttpLambdaIntegration } from '@aws-cdk/aws-apigatewayv2-integrations-alpha';
import { AddRoutesOptions, HttpMethod } from '@aws-cdk/aws-apigatewayv2-alpha';
import { FUNCTIONS_FOLDER } from '../config';
import { Runtime } from 'aws-cdk-lib/aws-lambda';

export interface SettingsStackProps extends StackProps {
    settingsTable: Table
}

export class SettingsStack extends Stack {
    routeOptions: AddRoutesOptions[];

    constructor(scope: Construct, id: string, props: SettingsStackProps) {
        super(scope, id, props);
        
        const settingsFunction = new GoFunction(this, 'settingsFunction', {
            entry: FUNCTIONS_FOLDER + 'settings'
        });

        props.settingsTable.grantReadWriteData(settingsFunction);

        const settingsIntegration = new HttpLambdaIntegration('settingsIntegration', settingsFunction);

        const settingsRouteOptions = {
            path: '/settings',
            methods: [HttpMethod.ANY],
            integration: settingsIntegration
        };

        this.routeOptions = [settingsRouteOptions];
    }
}
