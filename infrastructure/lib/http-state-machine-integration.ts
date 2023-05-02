import { HttpIntegrationType, HttpRouteIntegration, HttpRouteIntegrationBindOptions, HttpRouteIntegrationConfig, PayloadFormatVersion } from "@aws-cdk/aws-apigatewayv2-alpha";
import { CfnOutput, Resource } from "aws-cdk-lib";
import { CfnIntegration } from "aws-cdk-lib/aws-apigatewayv2";
import { Effect, PolicyDocument, PolicyStatement, Role, ServicePrincipal } from "aws-cdk-lib/aws-iam";
import { StateMachine } from "aws-cdk-lib/aws-stepfunctions";
import { Construct } from "constructs";

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
        new CfnOutput(options.scope, 'httpApiID', {
            value: options.route.httpApi.apiId,
            description: 'HttpApi ID',
        });

        new CfnOutput(options.scope, 'stateMachineARN', {
            value: this.props.stateMachine.stateMachineArn,
            description: 'State Machine ARN',
        });
        
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
            apiId: options.route.httpApi.httpApiId,
            integrationType: HttpIntegrationType.AWS_PROXY,
            integrationSubtype: 'StepFunctions-StartExecution',
            payloadFormatVersion: PayloadFormatVersion.VERSION_1_0.version,
            credentialsArn: httpApiRole.roleArn,
            requestParameters: {
                StateMachineArn: this.props.stateMachine.stateMachineArn,
                Input: '$request.body.input',
            },
            timeoutInMillis: this.props.timeoutInMillis,
        });

        new CfnOutput(options.scope, 'httpStepFunctionIntegrationRef', {
            value: httpStepFunctionIntegration.ref,
            description: 'HTTP Step Function Integration Ref',
        });
        
        const roleResource = httpApiRole.node.defaultChild as Construct;
        if (roleResource instanceof Resource) {
            httpStepFunctionIntegration.node.addDependency(roleResource);
        }
        
        return {
            type: HttpIntegrationType.AWS_PROXY,
            uri: this.props.stateMachine.stateMachineArn,
            payloadFormatVersion: PayloadFormatVersion.VERSION_1_0
        };
    }
}
