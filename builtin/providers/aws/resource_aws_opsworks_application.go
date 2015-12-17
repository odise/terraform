package aws

import (
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/odise/terraform/helper/resource"
)

type opsworksApplicationTypeAttribute struct {
	AttrName  string
	Type      schema.ValueType
	Default   interface{}
	Required  bool
	WriteOnly bool
}

type opsworksApplicationType struct {
	TypeName         string
	DefaultLayerName string
	Attributes       map[string]*opsworksApplicationTypeAttribute
	CustomShortName  bool
}

func resourceAwsOpsworksApplication() *schema.Resource {
	return &schema.Resource{

		Create: resourceAwsOpsworksApplicationCreate,
		Read:   resourceAwsOpsworksApplicationRead,
		Update: resourceAwsOpsworksApplicationUpdate,
		Delete: resourceAwsOpsworksApplicationDelete,
		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"type": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"stack_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"app_source": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},

						"url": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},

						"username": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},

						"password": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},

						"revision": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},

						"ssh_key": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func resourceAwsOpsworksApplicationRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AWSClient).opsworksconn

	req := &opsworks.DescribeAppsInput{
		AppIds: []*string{
			aws.String(d.Id()),
		},
	}

	log.Printf("[DEBUG] Reading OpsWorks app: %s", d.Id())

	resp, err := client.DescribeApps(req)
	if err != nil {
		if awserr, ok := err.(awserr.Error); ok {
			if awserr.Code() == "ResourceNotFoundException" {
				log.Printf("[INFO] App not found: %s", d.Id())
				d.SetId("")
				return nil
			}
		}
		return err
	}

	app := resp.Apps[0]

	d.Set("name", app.Name)
	d.Set("stack_id", app.StackId)
	d.Set("type", app.Type)

	return nil
}

func resourceAwsOpsworksApplicationCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AWSClient).opsworksconn

	// XXX: validate

	req := &opsworks.CreateAppInput{
		Name:    aws.String(d.Get("name").(string)),
		StackId: aws.String(d.Get("stack_id").(string)),
		Type:    aws.String(d.Get("type").(string)),
	}

	var resp *opsworks.CreateAppOutput
	err := resource.Retry(20*time.Minute, func() error {
		var cerr error
		resp, cerr = client.CreateApp(req)
		if cerr != nil {
			log.Printf("[INFO] client error")
			if opserr, ok := cerr.(awserr.Error); ok {
				// XXX: handle errors
				log.Printf("[INFO] OpsWorks error: " + opserr.Code() + "message: " + opserr.Message())
				return cerr
			}
			return resource.RetryError{Err: cerr}
		}
		return nil
	})

	if err != nil {
		return err
	}

	appID := *resp.AppId
	d.SetId(appID)
	d.Set("id", appID)

	return resourceAwsOpsworksApplicationRead(d, meta)
}

func resourceAwsOpsworksApplicationUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AWSClient).opsworksconn
	req := &opsworks.UpdateAppInput{
		Name: aws.String(d.Get("name").(string)),
		Type: aws.String(d.Get("type").(string)),
	}

	log.Printf("[DEBUG] Updating OpsWorks layer: %s", d.Id())

	_, err := client.UpdateApp(req)
	if err != nil {
		return err
	}

	return resourceAwsOpsworksApplicationRead(d, meta)
}

func resourceAwsOpsworksApplicationDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AWSClient).opsworksconn

	req := &opsworks.DeleteAppInput{
		AppId: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting OpsWorks application: %s", d.Id())

	_, err := client.DeleteApp(req)
	return err
}
