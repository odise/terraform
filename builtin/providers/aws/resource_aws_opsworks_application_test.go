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
				),
			},
			/*
				resource.TestStep{
					Config: testAccAwsOpsworksCustomLayerConfigUpdate,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(
							"aws_opsworks_custom_layer.tf-acc", "name", "tf-ops-acc-application",
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

var testAccAwsOpsworksApplicationCreate = testAccAwsOpsworksStackConfigNoVpcCreate + `
resource "aws_opsworks_application" "tf-acc" {
  stack_id = "${aws_opsworks_stack.tf-acc.id}"
  name = "tf-ops-acc-application"
  type = "other"
}
`
