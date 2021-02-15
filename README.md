# TcTool

**Dev Environment**
Create an EC2 Env in Cloud9
Open ~/.bashrc and update the following 2 lines
    GOPATH=~/environment/go
    export GOPATH
To increase the EBS size, run the cmd - sh resize.sh 20
Run the following to install aws-sdk
    go get -u github.com/aws/aws-sdk-go/...
To download some examples, run the following cmd
    git clone https://github.com/awsdocs/aws-doc-sdk-examples.git

**Features**
- CIS AWS Foundations Benchmark controls - config tests
https://docs.aws.amazon.com/securityhub/latest/userguide/securityhub-cis-controls.html
https://d1.awsstatic.com/whitepapers/compliance/AWS_CIS_Foundations_Benchmark.pdf

- AWS Foundational Security Best Practices controls - best practices
https://docs.aws.amazon.com/securityhub/latest/userguide/securityhub-standards-fsbp-controls.html

- AWS Inspector run - EC2 vuln tests

- AWS Trust Advisor Results

**Build the tool**
git checkout https://github.com/aseemsethi/tctool
go build src/main.go

** Run the tool**
tctool/src/main.go us-east-1 <Accountid>
jq '.' tctool.log > tctool.logJQ

**Tool Prerequisites**
1) Create an IAM Role in the Source (Dev) Env and attach to it the following policy, in addition to "Admin Access" policy
{
    "Version": "2012-10-17",
    "Statement": {
        "Effect": "Allow",
        "Action": "sts:AssumeRole",
        "Resource": "arn:aws:iam::<targetarn>:role/KVAccess"
    }
}
Attach the above Role to an EC2 machine, from where we will run the TcTool
Also provide SES Email Access Policy to the EC2 machine.

2) In a Target (customer) env, ask them to create a role called KVAccess, and attach the following policies to it
                AWSCloudTrailReadOnlyAccess
                IAMReadOnlyAccess
                IAMAccessAdvisorReadOnly
                AmazonInspectorFullAccess
                AmazonVPCReadOnlyAccess
                AmazonSSMManagedInstanceCore
                AmazonS3ReadOnlyAccess
                AmazonVPCReadOnlyAccess
                AWSConfigUserAccess

Also, create a "Trust Policy" under KVAccess, that grants TcTool from remote Account to assume the KVAccess Role
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.amazonaws.com",
        "AWS": "arn:aws:iam::<tctool account>:root"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}

3) For AWS Inspector run, attach the following Role to every EC2 instance. This
allows access of the SSM agent on EC2 to communicate with EC2 Systems Manager..
Also added to the Policies above, so no additional step is needed.
_AmazonSSMManagedInstanceCore_
_AmazonInspectorFullAccess_

Also, tag all EC2s in the Target Env, where you want inspector to run with tag "inspector" : "true"
Note that all following rules are run - 
Common Vulnerabilities and Exposures-1.1
CIS Operating System Security Configuration Benchmarks-1.0
Network Reachability-1.1
Security Best Practices-1.0

- inspector agent is installed automatically by Sytems Nanager, when it is run
(Manual steps if needed - Need ssm-agent and inspector-agent to be installed in all EC2s.
Run cmd - sudo systemctl status amazon-ssm-agent - to check if ssm-agent is installed and running
Else, install ssm-agent on all EC2s.)


4) Enable AWS Config - needed for SecurityHub Findings
