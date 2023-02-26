import { Construct } from 'constructs';
import { Stack, StackProps } from 'aws-cdk-lib';
import { AccountRecovery, UserPool } from 'aws-cdk-lib/aws-cognito';
import { HttpUserPoolAuthorizer } from '@aws-cdk/aws-apigatewayv2-authorizers-alpha';

export interface AuthenticationStackProps extends StackProps {
    environmentType: string
}

export class AuthenticationStack extends Stack {
    userPoolAuthoriser: HttpUserPoolAuthorizer;

    constructor(scope: Construct, id: string, props: AuthenticationStackProps) {
        super(scope, id, props);
        
        const userPool = new UserPool(this, 'userPool', {
            signInCaseSensitive: false,
            signInAliases: { email: true, username: true, preferredUsername: true },
            accountRecovery: AccountRecovery.PHONE_WITHOUT_MFA_AND_EMAIL,
            passwordPolicy: {
                minLength: 16,
                requireLowercase: true,
                requireUppercase: true,
                requireDigits: true,
                requireSymbols: true
            },
            selfSignUpEnabled: true,
            autoVerify: { email: true },
            deletionProtection: props.environmentType === 'prod'
        });
        userPool.addClient('userPoolClient', {generateSecret: true});

        this.userPoolAuthoriser = new HttpUserPoolAuthorizer('userPoolAuthoriser', userPool);
    }
}
