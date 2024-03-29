import { Construct } from 'constructs';
import { Stack, StackProps } from 'aws-cdk-lib';
import { AccountRecovery, OAuthScope, UserPool, UserPoolClientIdentityProvider } from 'aws-cdk-lib/aws-cognito';
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
            signInAliases: { email: true, username: true },
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
        const userPoolClient = userPool.addClient('userPoolClient', {
            supportedIdentityProviders: [UserPoolClientIdentityProvider.COGNITO],
            authFlows: {
                userPassword: true,
                userSrp: true
            },
            oAuth: {
                flows: {
                    authorizationCodeGrant: true
                },
                scopes: [ OAuthScope.OPENID, OAuthScope.PROFILE, OAuthScope.EMAIL ],
                callbackUrls: [ `http${props.environmentType === 'prod' ? 's://exstatic.io' : '://localhost:5173'}/callback` ],
                logoutUrls: [ `http${props.environmentType === 'prod' ? 's://exstatic.io' : '://localhost:5173'}` ]
            }
        });
        userPool.addDomain('userPoolDomain', {
            cognitoDomain: {
                domainPrefix: `exstatic-${props.environmentType}`
            }
        });

        this.userPoolAuthoriser = new HttpUserPoolAuthorizer('userPoolAuthoriser', userPool, {
            userPoolClients: [userPoolClient]
        });
    }
}
