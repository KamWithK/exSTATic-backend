import { HttpIntegrationType, HttpRouteIntegration, HttpRouteIntegrationBindOptions, HttpRouteIntegrationConfig, PayloadFormatVersion } from "@aws-cdk/aws-apigatewayv2-alpha";
import { Aws, CfnOutput } from "aws-cdk-lib";
import { IntegrationType } from "aws-cdk-lib/aws-apigateway";
import { CfnIntegration } from "aws-cdk-lib/aws-apigatewayv2";
import { Effect, PolicyDocument, PolicyStatement, Role, ServicePrincipal } from "aws-cdk-lib/aws-iam";
import { StateMachine } from "aws-cdk-lib/aws-stepfunctions";

interface HttpStepFunctionsIntegrationProps {
    stateMachine: StateMachine;
    timeoutInMillis?: number;
}

export class HttpStepFunctionsIntegration extends HttpRouteIntegration {
    private readonly props: HttpStepFunctionsIntegrationProps;
    
    constructor(id: string, props: HttpStepFunctionsIntegrationProps) {
        super(id);
        
        this.props = props;
    }

    public bind(options: HttpRouteIntegrationBindOptions): HttpRouteIntegrationConfig {
        // Create the IAM role for API Gateway
        const httpApiRole = new Role(options.scope, 'httpApiRole', {
            assumedBy: new ServicePrincipal('apigateway.amazonaws.com'),
            inlinePolicies: {
                AllowSFNExec: new PolicyDocument({
                    statements: [new PolicyStatement({
                        actions: ['states:StartExecution', 'states:StartSyncExecution'],
                        effect: Effect.ALLOW,
                        resources: [this.props.stateMachine.stateMachineArn]
                    })]
                })
            }
        });
        this.props.stateMachine.grantStartExecution(httpApiRole);
        
        new CfnOutput(options.scope, 'httpApiRoleARN', {
            value: httpApiRole.roleArn,
            description: 'API Role ARN',
        });

        // Create the AWS_PROXY integration with Step Functions
        const httpStepFunctionIntegration = new CfnIntegration(options.scope, 'httpStepFunctionIntegration', {
            apiId: options.route.httpApi.apiId,
            integrationType: IntegrationType.AWS,
            // integrationSubtype: 'StepFunctions-StartExecution',
            payloadFormatVersion: PayloadFormatVersion.VERSION_1_0.version,
            credentialsArn: httpApiRole.roleArn,
            requestParameters: {
                StateMachineArn: this.props.stateMachine.stateMachineArn,
                Input: '$request.body.input',
            },
            timeoutInMillis: this.props.timeoutInMillis,
        });

        return {
            type: HttpIntegrationType.AWS_PROXY,
            uri: `arn:aws:apigateway:${Aws.REGION}:states:action/StartExecution`,
            credentials: {credentialsArn: httpApiRole.roleArn},
            payloadFormatVersion: PayloadFormatVersion.VERSION_1_0
        };
    }
}
