package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

// These tests assume the existence of predefined Opsworks IAM roles named `aws-opsworks-ec2-role`
// and `aws-opsworks-service-role`.

func TestAccAWSOpsworksApplication(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsOpsworksApplicationDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAwsOpsworksApplicationCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"aws_opsworks_application.tf-acc", "name", "tf-ops-acc-application",
					),
					resource.TestCheckResourceAttr(
						"aws_opsworks_application.tf-acc", "type", "other",
					),
				),
			},
			/*
				resource.TestStep{
					Config: testAccAwsOpsworksApplicationUpdate,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(
							"aws_opsworks_application.tf-acc", "name", "tf-ops-acc-application",
						),
						resource.TestCheckResourceAttr(
							"aws_opsworks_application.tf-acc", "type", "static",
						),
					),
				},
			*/
		},
	})
}

func testAccCheckAwsOpsworksApplicationDestroy(s *terraform.State) error {
	if len(s.RootModule().Resources) > 0 {
		return fmt.Errorf("Expected all resources to be gone, but found: %#v", s.RootModule().Resources)
	}

	return nil
}

var testAccAwsOpsworksApplicationIam = `
# service role
#####################

resource "aws_iam_role" "service-role" {
    name = "tf-acc-opsworks_service_role"
    assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "opsworks.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "service-role-policy" {
    name = "tf-acc-service-role-policy"
    role = "${aws_iam_role.service-role.id}"
    policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Resource":"*",
      "Action": [
        "cloudwatch:GetMetricStatistics",
        "ec2:*",
        "opsworks:*",
        "elasticloadbalancing:*",
        "iam:PassRole",
        "rds:*"
      ],
      "Effect": "Allow"
    }
  ]
}
EOF
}


# instance profile
#####################

resource "aws_iam_instance_profile" "opsworks-instance-profile" {
    name = "xxx-tf-acc-opsworks-instance-profile"
    roles = ["${aws_iam_role.opsworks-instance-role.name}"]
}

resource "aws_iam_role" "opsworks-instance-role" {
    name = "tf-acc-opsworks-instance-role"
    path = "/"
    assume_role_policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": "sts:AssumeRole",
            "Principal": {
              "Service": "ec2.amazonaws.com"
            },
            "Effect": "Allow",
            "Sid": ""
        }
    ]
}
EOF
}

resource "aws_iam_role_policy" "opsworks-instance-policy" {
    name = "tf-acc-opsworks-instance-policy"
    role = "${aws_iam_role.opsworks-instance-role.id}"
    policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "ec2:Describe*"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
}
`

var testAccAwsOpsworksApplicationBasics = `
resource "aws_opsworks_stack" "tf-acc" {
  depends_on = [ "aws_iam_role.service-role", "aws_iam_role_policy.service-role-policy", "aws_iam_instance_profile.opsworks-instance-profile" ]
  name = "tf-opsworks-acc"
  region = "eu-west-1"
  service_role_arn = "${aws_iam_role.service-role.arn}"
  default_instance_profile_arn = "${aws_iam_instance_profile.opsworks-instance-profile.arn}"
  default_availability_zone = "eu-west-1a"
  use_opsworks_security_groups = true
}
`

var testAccAwsOpsworksApplicationCreate = testAccAwsOpsworksApplicationIam + testAccAwsOpsworksApplicationBasics + `
resource "aws_opsworks_application" "tf-acc" {
  stack_id = "${aws_opsworks_stack.tf-acc.id}"
  name = "tf-ops-acc-application"
  type = "other"
}
`
var testAccAwsOpsworksApplicationUpdate = testAccAwsOpsworksApplicationIam + testAccAwsOpsworksApplicationBasics + `
resource "aws_opsworks_application" "tf-acc" {
  stack_id = "${aws_opsworks_stack.tf-acc.id}"
  name = "tf-ops-acc-application"
  type = "static"
}
`
