import { Construct } from 'constructs';
import { aws_cognito, Stack, StackProps } from 'aws-cdk-lib';
import { AccountRecovery, StringAttribute, UserPool, UserPoolClient } from 'aws-cdk-lib/aws-cognito';

export interface AuthenticationStackProps extends StackProps {
    environmentType: string
}

export class AuthenticationStack extends Stack {
    userPool: aws_cognito.UserPool;

    constructor(scope: Construct, id: string, props: AuthenticationStackProps) {
        super(scope, id, props);
        
        this.userPool = new UserPool(this, 'userPool', {
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
            deletionProtection: props.environmentType === "prod"
        });
        this.userPool.addClient('userPoolClient', {generateSecret: true});
    }
}
