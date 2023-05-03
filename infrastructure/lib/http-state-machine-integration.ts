import { HttpIntegrationSubtype, HttpIntegrationType, HttpRouteIntegration, HttpRouteIntegrationBindOptions, HttpRouteIntegrationConfig, ParameterMapping, PayloadFormatVersion } from "@aws-cdk/aws-apigatewayv2-alpha";
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
        
        return {
            type: HttpIntegrationType.AWS_PROXY,
            subtype: HttpIntegrationSubtype.STEPFUNCTIONS_START_EXECUTION,
            credentials: {credentialsArn: httpApiRole.roleArn},
            payloadFormatVersion: PayloadFormatVersion.VERSION_1_0,
            parameterMapping: new ParameterMapping()
                .custom('StateMachineArn', this.props.stateMachine.stateMachineArn)
                .custom('Input', '$request.body.input')
        };
    }
}
