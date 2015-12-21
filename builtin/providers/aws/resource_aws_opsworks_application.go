package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
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
			// aws-flow-ruby | java | rails | php | nodejs | static | other
			"type": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"stack_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"app_source_type": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},

			"app_source_url": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},

			"app_source_username": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},

			"app_source_password": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},

			"app_source_revision": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},

			"app_source_ssh_key": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			// AutoSelectOpsworksMysqlInstance, OpsworksMysqlInstance, or RdsDbInstance.
			// anything beside auto select will lead into failure in case the instance doen't existence
			// XXX: validation?
			"data_source_type": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"data_source_database_name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"data_source_arn": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"domains": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"environment": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				//Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"value": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"secure": &schema.Schema{
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
					},
				},
			},
			"enable_ssl": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"ssl_certificate": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"ssl_private_key": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"ssl_chain": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceAwsOpsworksApplicationValidate(d *schema.ResourceData) error {
	sslKey := d.Get("ssl_private_key").(string)
	sslCert := d.Get("ssl_certificate").(string)
	if (len(sslKey) > 0 || len(sslCert) > 0) && (len(sslKey)+len(sslCert) < 2) {
		return fmt.Errorf("ssl_private_key and ssl_certivicate must be set")
	}

	return nil
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
	d.Set("description", app.Description)
	d.Set("domains", unwrapAwsStringList(app.Domains))
	d.Set("enable_ssl", app.EnableSsl)
	d.Set("ssl_private_key", app.SslConfiguration.PrivateKey)
	d.Set("ssl_certificate", app.SslConfiguration.Certificate)
	d.Set("ssl_chain", app.SslConfiguration.Chain)
	d.Set("app_source_type", app.AppSource.Type)
	d.Set("app_source_url", app.AppSource.Url)
	d.Set("app_source_username", app.AppSource.Username)
	d.Set("app_source_password", app.AppSource.Password)
	d.Set("app_source_revision", app.AppSource.Revision)
	d.Set("app_source_ssh_key", app.AppSource.SshKey)
	resourceAwsOpsworksSetApplicationDataSources(d, app.DataSources)
	resourceAwsOpsworksSetApplicationEnvironmentVariable(d, app.Environment)
	return nil
}

func resourceAwsOpsworksApplicationCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AWSClient).opsworksconn

	err := resourceAwsOpsworksApplicationValidate(d)
	if err != nil {
		return err
	}

	req := &opsworks.CreateAppInput{
		Name:        aws.String(d.Get("name").(string)),
		StackId:     aws.String(d.Get("stack_id").(string)),
		Type:        aws.String(d.Get("type").(string)),
		Description: aws.String(d.Get("description").(string)),
		Domains:     makeAwsStringList(d.Get("domains").([]interface{})),
		EnableSsl:   aws.Bool(d.Get("enable_ssl").(bool)),
		SslConfiguration: &opsworks.SslConfiguration{
			Certificate: aws.String(strings.TrimRight(d.Get("ssl_certificate").(string), "\n")),
			PrivateKey:  aws.String(strings.TrimRight(d.Get("ssl_private_key").(string), "\n")),
			Chain:       aws.String(strings.TrimRight(d.Get("ssl_chain").(string), "\n")),
		},
		AppSource: &opsworks.Source{
			Type:     aws.String(d.Get("app_source_type").(string)),
			Url:      aws.String(d.Get("app_source_url").(string)),
			Username: aws.String(d.Get("app_source_username").(string)),
			Password: aws.String(d.Get("app_source_password").(string)),
			Revision: aws.String(d.Get("app_source_revision").(string)),
			SshKey:   aws.String(d.Get("app_source_ssh_key").(string)),
		},
		DataSources: resourceAwsOpsworksApplicationDataSources(d),
		Environment: resourceAwsOpsworksApplicationEnvironmentVariable(d),
	}

	var resp *opsworks.CreateAppOutput
	err = resource.Retry(10*time.Minute, func() error {
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
		AppId:       aws.String(d.Id()),
		Name:        aws.String(d.Get("name").(string)),
		Type:        aws.String(d.Get("type").(string)),
		Description: aws.String(d.Get("description").(string)),
		Domains:     makeAwsStringList(d.Get("domains").([]interface{})),
		EnableSsl:   aws.Bool(d.Get("enable_ssl").(bool)),
		SslConfiguration: &opsworks.SslConfiguration{
			Certificate: aws.String(strings.TrimRight(d.Get("ssl_certificate").(string), "\n")),
			PrivateKey:  aws.String(strings.TrimRight(d.Get("ssl_private_key").(string), "\n")), // Required
			Chain:       aws.String(strings.TrimRight(d.Get("ssl_chain").(string), "\n")),
		},
		AppSource: &opsworks.Source{
			Type:     aws.String(d.Get("app_source_type").(string)),
			Url:      aws.String(d.Get("app_source_url").(string)),
			Username: aws.String(d.Get("app_source_username").(string)),
			Password: aws.String(d.Get("app_source_password").(string)),
			Revision: aws.String(d.Get("app_source_revision").(string)),
			SshKey:   aws.String(d.Get("app_source_ssh_key").(string)),
		},
		DataSources: resourceAwsOpsworksApplicationDataSources(d),
		Environment: resourceAwsOpsworksApplicationEnvironmentVariable(d),
	}

	log.Printf("[DEBUG] Updating OpsWorks layer: %s", d.Id())

	var resp *opsworks.UpdateAppOutput
	err := resource.Retry(10*time.Minute, func() error {
		var cerr error
		resp, cerr = client.UpdateApp(req)
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

func resourceAwsOpsworksSetApplicationEnvironmentVariable(d *schema.ResourceData, v []*opsworks.EnvironmentVariable) {
	log.Printf("[DEBUG] envs: %s %d", v, len(v))
	if len(v) == 0 {
		d.Set("environment", nil)
		return
	}
	newValue := make([]*map[string]interface{}, len(v))

	for i := 0; i < len(v); i++ {
		config := v[i]
		data := make(map[string]interface{})
		newValue[i] = &data

		if config.Key != nil {
			data["key"] = *config.Key
		}
		if config.Value != nil {
			data["value"] = *config.Value
		}
		if config.Secure != nil {

			if bool(*config.Secure) {
				data["secure"] = &opsworksTrueString
			} else {
				data["secure"] = &opsworksFalseString
			}
		}
		log.Printf("[DEBUG] v: %s", data)
	}

	d.Set("environment", newValue)
}

func resourceAwsOpsworksApplicationEnvironmentVariable(d *schema.ResourceData) []*opsworks.EnvironmentVariable {
	environmentVariables := d.Get("environment").(*schema.Set).List()
	result := make([]*opsworks.EnvironmentVariable, len(environmentVariables))

	for i := 0; i < len(environmentVariables); i++ {
		env := environmentVariables[i].(map[string]interface{})

		result[i] = &opsworks.EnvironmentVariable{
			Key:    aws.String(env["key"].(string)),
			Value:  aws.String(env["value"].(string)),
			Secure: aws.Bool(env["secure"].(bool)),
		}
	}
	return result
}

func resourceAwsOpsworksApplicationDataSources(d *schema.ResourceData) []*opsworks.DataSource {
	arn := d.Get("data_source_arn").(string)
	databaseName := d.Get("data_source_database_name").(string)
	databaseType := d.Get("data_source_type").(string)

	result := make([]*opsworks.DataSource, 1)

	if len(arn) > 0 || len(databaseName) > 0 || len(databaseType) > 0 {
		result[0] = &opsworks.DataSource{
			Arn:          aws.String(arn),
			DatabaseName: aws.String(databaseName),
			Type:         aws.String(databaseType),
		}
	}
	return result
}

func resourceAwsOpsworksSetApplicationDataSources(d *schema.ResourceData, v []*opsworks.DataSource) {
	d.Set("data_source_arn", nil)
	d.Set("data_source_database_name)", nil)
	d.Set("data_source_type)", nil)

	if len(v) == 0 {
		return
	}

	d.Set("data_source_arn", v[0].Arn)
	d.Set("data_source_database_name)", v[0].DatabaseName)
	d.Set("data_source_type)", v[0].Type)
}
