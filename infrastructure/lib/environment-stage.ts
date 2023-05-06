import { Construct } from 'constructs';
import { Stage, StageProps } from 'aws-cdk-lib';
import { DataStack } from './data-stack';
import { SettingsStack } from './settings-stack';
import { MediaStack } from './media-stack';
import { LeaderboardStack } from './leaderboard-stack';
import { ApiStack } from './api-stack';
import { AuthenticationStack } from './authentication-stack';

export interface EnvironmentStageProps extends StageProps {
    environmentType: string;
}

export class EnvironmentStage extends Stage {
    constructor(scope: Construct, id: string, props: EnvironmentStageProps) {
        super(scope, id, props);

        const dataStack = new DataStack(this, 'dataStack', {
            environmentType: props.environmentType
        });


        const settingsStack = new SettingsStack(this, 'settingsStack', {
            settingsTable: dataStack.settingsTable
        });
        const mediaStack = new MediaStack(this, 'mediaStack', {
            mediaTable: dataStack.mediaTable,
            leaderboardTable: dataStack.leaderboardTable
        });
        const leaderboardStack = new LeaderboardStack(this, 'leaderboardStack', {
            leaderboardTable: dataStack.leaderboardTable
        });

        if (props.environmentType !== 'local') {
            const authenticationStack = new AuthenticationStack(this, 'authenticationStack', {
                environmentType: props.environmentType
            });
            const apiStack = new ApiStack(this, 'apiStack', {
                userPoolAuthoriser: authenticationStack?.userPoolAuthoriser,
                routeOptions: [
                    ...settingsStack.routeOptions,
                    ...mediaStack.routeOptions,
                    ...leaderboardStack.routeOptions
                ]
            });
        }
    }
}
