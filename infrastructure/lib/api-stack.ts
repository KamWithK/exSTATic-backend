import { Construct } from 'constructs';
import { Duration, Stack, StackProps } from 'aws-cdk-lib';
import { HttpApi, HttpIntegrationSubtype, HttpIntegrationType, HttpRouteKey, PayloadFormatVersion } from '@aws-cdk/aws-apigatewayv2-alpha';
import { AddRoutesOptions } from '@aws-cdk/aws-apigatewayv2-alpha';
import { HttpUserPoolAuthorizer } from '@aws-cdk/aws-apigatewayv2-authorizers-alpha';
import { StateMachine } from 'aws-cdk-lib/aws-stepfunctions';
import { CfnIntegration, CfnRoute } from 'aws-cdk-lib/aws-apigatewayv2';
import { Effect, PolicyDocument, PolicyStatement, Role, ServicePrincipal } from 'aws-cdk-lib/aws-iam';

export interface ApiStackProps extends StackProps {
    userPoolAuthoriser: HttpUserPoolAuthorizer,
    routeOptions: AddRoutesOptions[],
    stateMachines: StateMachine[]
}

function createStateMachineRoute(scope: Construct, httpApi: HttpApi, stateMachine: StateMachine, timeoutInMillis?: number) {
    // Create IAM role for API Gateway
    const httpApiRole = new Role(scope, 'httpApiRole', {
        assumedBy: new ServicePrincipal('apigateway.amazonaws.com'),
        inlinePolicies: {
            AllowSFNExec: new PolicyDocument({
                statements: [new PolicyStatement({
                    actions: ['states:StartExecution', 'states:StartSyncExecution'],
                    effect: Effect.ALLOW,
                    resources: [stateMachine.stateMachineArn]
                })]
            })
        }
    });
    stateMachine.grantStartExecution(httpApiRole);

    // Create the AWS_PROXY integration with Step Functions
    const httpStepFunctionIntegration = new CfnIntegration(scope, 'httpStepFunctionIntegration', {
        apiId: httpApi.httpApiId,
        integrationType: HttpIntegrationType.AWS_PROXY,
        integrationSubtype: HttpIntegrationSubtype.STEPFUNCTIONS_START_EXECUTION,
        payloadFormatVersion: PayloadFormatVersion.VERSION_1_0.version,
        credentialsArn: httpApiRole.roleArn,
        requestParameters: {
            StateMachineArn: stateMachine.stateMachineArn,
            Input: '$request.body.input',
        },
        timeoutInMillis: timeoutInMillis,
    });

    new CfnRoute(scope, 'httpStepFunctionRoute', {
        apiId: httpApi.httpApiId,
        routeKey: HttpRouteKey.DEFAULT.key,
        target: `integrations/${httpStepFunctionIntegration.ref}`
    });
}

export class ApiStack extends Stack {
    constructor(scope: Construct, id: string, props: ApiStackProps) {
        super(scope, id, props);

        const httpApi = new HttpApi(this, 'httpApi', {
            defaultAuthorizer: props.userPoolAuthoriser
        });

        props.routeOptions.forEach((addRouteOption) => httpApi.addRoutes(addRouteOption));
        props.stateMachines.forEach((stateMachine) => createStateMachineRoute(this, httpApi, stateMachine, Duration.minutes(1).toMilliseconds()))
    }
}
