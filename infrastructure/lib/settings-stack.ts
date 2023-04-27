import { GoFunction } from '@aws-cdk/aws-lambda-go-alpha';
import { Stack, StackProps } from 'aws-cdk-lib';
import { Construct } from 'constructs';
import { Table } from 'aws-cdk-lib/aws-dynamodb';
import { HttpLambdaIntegration } from '@aws-cdk/aws-apigatewayv2-integrations-alpha';
import { AddRoutesOptions, HttpMethod } from '@aws-cdk/aws-apigatewayv2-alpha';
import { FUNCTIONS_FOLDER } from '../config';

export interface SettingsStackProps extends StackProps {
    settingsTable: Table
}

export class SettingsStack extends Stack {
    routeOptions: AddRoutesOptions[];

    constructor(scope: Construct, id: string, props: SettingsStackProps) {
        super(scope, id, props);
        
        const settingsGetFunction = new GoFunction(this, 'settingsGetFunction', {
            entry: FUNCTIONS_FOLDER + 'settings/get'
        });

        const settingsPutFunction = new GoFunction(this, 'settingsPutFunction', {
            entry: FUNCTIONS_FOLDER + 'settings/put'
        });

        props.settingsTable.grantReadWriteData(settingsGetFunction);
        props.settingsTable.grantReadWriteData(settingsPutFunction);

        const settingsGetIntegration = new HttpLambdaIntegration('settingsGetIntegration', settingsGetFunction);
        const settingsPutIntegration = new HttpLambdaIntegration('settingsPutIntegration', settingsPutFunction);

        const settingsGetRouteOptions = {
            path: '/settings/get',
            methods: [HttpMethod.GET],
            integration: settingsGetIntegration
        };

        const settingsPutRouteOptions = {
            path: '/settings/put',
            methods: [HttpMethod.PUT],
            integration: settingsPutIntegration
        };

        this.routeOptions = [settingsGetRouteOptions, settingsPutRouteOptions];
    }
}
