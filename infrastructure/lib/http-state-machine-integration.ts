import { HttpIntegrationType, HttpRouteIntegration, HttpRouteIntegrationBindOptions, HttpRouteIntegrationConfig, PayloadFormatVersion } from "@aws-cdk/aws-apigatewayv2-alpha";
import { CfnIntegration } from "aws-cdk-lib/aws-apigatewayv2";
import { Role, ServicePrincipal } from "aws-cdk-lib/aws-iam";
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
        const apiRole = new Role(options.scope, 'httpApiRole', {
            assumedBy: new ServicePrincipal('apigateway.amazonaws.com'),
        });
        this.props.stateMachine.grantStartExecution(apiRole);
        
        // Create the AWS_PROXY integration with Step Functions
        const httpStepFunctionIntegration = new CfnIntegration(options.scope, 'httpStepFunctionIntegration', {
            apiId: options.route.httpApi.httpApiId,
            integrationType: HttpIntegrationType.AWS_PROXY,
            integrationSubtype: 'StepFunctions-StartExecution',
            payloadFormatVersion: PayloadFormatVersion.VERSION_1_0.version,
            credentialsArn: apiRole.roleArn,
            requestParameters: {
                StateMachineArn: this.props.stateMachine.stateMachineArn,
                Input: '$request.body.input',
            },
            timeoutInMillis: this.props.timeoutInMillis,
        });
        
        return {
            type: HttpIntegrationType.AWS_PROXY,
            uri: `integrations/${httpStepFunctionIntegration.ref}`,
            payloadFormatVersion: PayloadFormatVersion.VERSION_1_0
        };
    }
}
