import { Construct } from 'constructs';

import { AttributeType, BillingMode, ProjectionType, Table, TableClass, TableEncryption } from 'aws-cdk-lib/aws-dynamodb';
import { RemovalPolicy, Stack, StackProps } from 'aws-cdk-lib';

export interface DataStackProps extends StackProps {
    environmentType: string
}

export class DataStack extends Stack {
    settingsTable: Table;
    mediaTable: Table;
    leaderboardTable: Table;

    constructor(scope: Construct, id: string, props: DataStackProps) {
        super(scope, id, props);
        
        this.settingsTable = new Table(this, 'settingsTable', {
            tableName: 'settings',
            
            partitionKey: {
                name: 'username',
                type: AttributeType.STRING
            },
            sortKey: {
                name: 'media_type',
                type: AttributeType.STRING
            },
            
            billingMode: BillingMode.PAY_PER_REQUEST,
            tableClass: TableClass.STANDARD,
            encryption: TableEncryption.DEFAULT,
            removalPolicy: RemovalPolicy.RETAIN,
            pointInTimeRecovery: props.environmentType === "prod"
        });
        
        this.mediaTable = new Table(this, 'mediaTable', {
            tableName: 'media',
            
            partitionKey: {
                name: 'pk',
                type: AttributeType.STRING
            },
            sortKey: {
                name: 'sk',
                type: AttributeType.STRING
            },
            
            billingMode: BillingMode.PAY_PER_REQUEST,
            tableClass: TableClass.STANDARD,
            encryption: TableEncryption.DEFAULT,
            removalPolicy: RemovalPolicy.RETAIN,
            pointInTimeRecovery: props.environmentType === "prod"
        });
        
        this.mediaTable.addLocalSecondaryIndex({
            indexName: 'lastUpdatedIndex',
            sortKey: {
                name: 'last_update',
                type: AttributeType.NUMBER
            },
            projectionType: ProjectionType.ALL
        });
        
        this.leaderboardTable = new Table(this, 'leaderboardTable', {
            tableName: 'leaderboard',
            
            partitionKey: {
                name: 'section',
                type: AttributeType.STRING
            },
            sortKey: {
                name: 'username',
                type: AttributeType.STRING
            },
            
            billingMode: BillingMode.PAY_PER_REQUEST,
            tableClass: TableClass.STANDARD,
            encryption: TableEncryption.DEFAULT,
            removalPolicy: RemovalPolicy.RETAIN,
            pointInTimeRecovery: props.environmentType === "prod"
        });
        
        this.leaderboardTable.addLocalSecondaryIndex({
            indexName: 'timeReadIndex',
            sortKey: {
                name: 'time_read',
                type: AttributeType.NUMBER
            },
            projectionType: ProjectionType.ALL
        });
        this.leaderboardTable.addLocalSecondaryIndex({
            indexName: 'charsReadIndex',
            sortKey: {
                name: 'chars_read',
                type: AttributeType.NUMBER
            },
            projectionType: ProjectionType.ALL
        });
    }
}
