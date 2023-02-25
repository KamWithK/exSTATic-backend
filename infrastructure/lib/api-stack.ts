import { Construct } from 'constructs';
import { Stack, StackProps } from 'aws-cdk-lib';
import { HttpApi } from '@aws-cdk/aws-apigatewayv2-alpha';
import { AddRoutesOptions } from '@aws-cdk/aws-apigatewayv2-alpha';

export interface ApiStackProps extends StackProps {
    routeOptions: AddRoutesOptions[]
}

export class ApiStack extends Stack {
    constructor(scope: Construct, id: string, props: ApiStackProps) {
        super(scope, id, props);
        
        const httpApi = new HttpApi(this, 'httpApi');

        props.routeOptions.forEach((addRouteOption) => httpApi.addRoutes(addRouteOption));
    }
}
