{
  build(
    snsTopicArn,
    repoS3Region,
    repoS3Bucket,
    repoS3Prefix='',
    logS3Buckets=[],
  ):: {

    local toBucketResource(bucket) = [
      'arn:aws:s3:::' + bucket,
      'arn:aws:s3:::' + bucket + '/*',
    ],

    local sqsPolicy = {
      Version: '2012-10-17',
      Id: 'BadFilterSQSPolicy',
      Statement: [{
        Sid: 'BadFilterSQSPolicy001',
        Effect: 'Allow',
        Principal: '*',
        Action: 'sqs:SendMessage',
        Resource: '${CreatedEventQueue.Arn}',
        Condition: {
          ArnEquals: { 'aws:SourceArn': snsTopicArn },
        },
      }],
    },

    local lambdaRoleArn = { 'Fn::GetAtt': 'LambdaRole.Arn' },
    local LambdaRoleTemplate = {
      LambdaRole: {
        Type: 'AWS::IAM::Role',
        Properties: {
          AssumeRolePolicyDocument: {
            Version: '2012-10-17',
            Statement: [
              {
                Effect: 'Allow',
                Principal: { Service: ['lambda.amazonaws.com'] },
                Action: ['sts:AssumeRole'],
              },
            ],
          },
          Path: '/',
          ManagedPolicyArns: [
            'arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole',
          ],
          Policies: [
            {
              PolicyName: 'S3Access',
              PolicyDocument: {
                Version: '2012-10-17',
                Statement: [
                  {
                    Effect: 'Allow',
                    Action: [
                      's3:GetObject',
                      's3:PutObject',
                      's3:ListBucket',
                    ],
                    Resource: [
                      'arn:aws:s3:::' + repoS3Bucket,
                      'arn:aws:s3:::' + repoS3Bucket + '/' + repoS3Prefix + '*',
                    ],
                  },
                ],
              },
            },
            {
              PolicyName: 'S3Readable',
              PolicyDocument: {
                Version: '2012-10-17',
                Statement: [
                  {
                    Effect: 'Allow',
                    Action: [
                      's3:GetObject',
                      's3:GetObjectVersion',
                      's3:ListBucket',
                    ],
                    Resource: std.join([], std.map(toBucketResource, logS3Buckets)),
                  },
                ],
              },
            },

            {
              PolicyName: 'SQSAccess',
              PolicyDocument: {
                Version: '2012-10-17',
                Statement: [
                  {
                    Effect: 'Allow',
                    Action: [
                      'sqs:SendMessage',
                      'sqs:ReceiveMessage',
                      'sqs:DeleteMessage',
                      'sqs:GetQueueAttributes',
                    ],
                    Resource: [
                      { 'Fn::GetAtt': 'CreatedEventQueue.Arn' },
                    ],
                  },
                ],
              },
            },
          ],
        },
      },
    },

    // --- main template ---------------------------------------------
    AWSTemplateFormatVersion: '2010-09-09',
    Transform: 'AWS::Serverless-2016-10-31',

    Resources: {
      CreatedEventQueue: {
        Type: 'AWS::SQS::Queue',
        Properties: {
          VisibilityTimeout: 600,
        },
      },
      CreatedEventQueuePolicy: {
        Type: 'AWS::SQS::QueuePolicy',
        Properties: {
          PolicyDocument: { 'Fn::Sub': std.toString(sqsPolicy) },
          Queues: [{ Ref: 'CreatedEventQueue' }],
        },
      },
      CreatedEventQueueSubscription: {
        Type: 'AWS::SNS::Subscription',
        Properties: {
          Endpoint: { 'Fn::GetAtt': 'CreatedEventQueue.Arn' },
          Protocol: 'sqs',
          TopicArn: snsTopicArn,
        },
      },

      updater: {
        Type: 'AWS::Serverless::Function',
        Properties: {
          CodeUri: 'build',
          Handler: 'updater',
          Runtime: 'go1.x',
          Timeout: 60,
          MemorySize: 1024,
          Role: lambdaRoleArn,
          Environment: {
            Variables: {
              REPO_S3_REGION: repoS3Region,
              REPO_S3_BUCKET: repoS3Bucket,
              REPO_S3_PREFIX: repoS3Prefix,
              LOG_LEVEL: 'DEBUG',
            },
          },
          Events: {
            Every1hour: {
              Type: 'Schedule',
              Properties: { Schedule: 'rate(1 hour)' },
            },
          },
        },
      },

      matcher: {
        Type: 'AWS::Serverless::Function',
        Properties: {
          CodeUri: 'build',
          Handler: 'matcher',
          Runtime: 'go1.x',
          Timeout: 60,
          MemorySize: 1024,
          Role: lambdaRoleArn,
          Environment: {
            Variables: {
              REPO_S3_REGION: repoS3Region,
              REPO_S3_BUCKET: repoS3Bucket,
              REPO_S3_PREFIX: repoS3Prefix,
              LOG_LEVEL: 'DEBUG',
            },
          },
          Events: {
            CreatedEventQueue: {
              Type: 'SQS',
              Properties: {
                Queue: { 'Fn::GetAtt': 'CreatedEventQueue.Arn' },
                BatchSize: 10,
              },
            },
          },
        },
      },
    } + LambdaRoleTemplate,
  },
}
