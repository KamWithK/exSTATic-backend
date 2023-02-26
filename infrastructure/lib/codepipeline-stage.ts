import { Construct } from 'constructs';
import { Stage, StackProps } from 'aws-cdk-lib';
import { DataStack } from './data-stack';
import { SettingsStack } from './settings-stack';
import { MediaStack } from './media-stack';
import { LeaderboardStack } from './leaderboard-stack';
import { ApiStack } from './api-stack';
import { AuthenticationStack } from './authentication-stack';

export interface CodePipelineStageProps extends StackProps {
    environmentType: string;
}

export class CodePipelineStage extends Stage {
    constructor(scope: Construct, id: string, props: CodePipelineStageProps) {
        super(scope, id, props);

        const dataStack = new DataStack(this, 'dataStack', {
            environmentType: props.environmentType
        });
        const authenticationStack = new AuthenticationStack(this, 'authenticationStack', {
            environmentType: props.environmentType
        });


        const settingsStack = new SettingsStack(this, 'settingsStack', {
            settingsTable: dataStack.settingsTable
        });
        const mediaStack = new MediaStack(this, 'mediaStack', {
            mediaTable: dataStack.mediaTable
        });
        const leaderboardStack = new LeaderboardStack(this, 'leaderboardStack', {
            leaderboardTable: dataStack.leaderboardTable
        });

        const apiStack = new ApiStack(this, 'apiStack', {
            userPool: authenticationStack.userPool,
            routeOptions: [
                ...settingsStack.routeOptions,
                ...mediaStack.routeOptions,
                ...leaderboardStack.routeOptions
            ]
        });
    }
}
