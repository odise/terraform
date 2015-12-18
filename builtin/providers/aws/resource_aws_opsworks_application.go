package aws

import (
	"log"
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
							Optional: true,
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
			"data_source": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						// AutoSelectOpsworksMysqlInstance, OpsworksMysqlInstance, or RdsDbInstance.
						"type": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"database_name": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
						"arn": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
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
			"enable_ssl": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
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
			"ssl_configuration": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"certificate": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"private_key": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"chain": &schema.Schema{
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
	d.Set("description", app.Description)
	d.Set("domains", app.Domains)
	d.Set("enable_ssl", app.EnableSsl)
	resourceAwsOpsworksSetApplicationAppSource(d, app.AppSource)
	resourceAwsOpsworksSetApplicationEnvironmentVariable(d, app.Environment)
	resourceAwsOpsworksSetApplicationDataSources(d, app.DataSources)
	return nil
}

func resourceAwsOpsworksApplicationCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*AWSClient).opsworksconn

	// XXX: validate

	req := &opsworks.CreateAppInput{
		Name:        aws.String(d.Get("name").(string)),
		StackId:     aws.String(d.Get("stack_id").(string)),
		Type:        aws.String(d.Get("type").(string)),
		Description: aws.String(d.Get("description").(string)),
		Domains:     makeAwsStringList(d.Get("domains").([]interface{})),
		EnableSsl:   aws.Bool(d.Get("enable_ssl").(bool)),
		AppSource:   resourceAwsOpsworksApplicationAppSource(d),
		Environment: resourceAwsOpsworksApplicationEnvironmentVariable(d),
		DataSources: resourceAwsOpsworksApplicationDataSources(d),
	}

	var resp *opsworks.CreateAppOutput
	err := resource.Retry(10*time.Minute, func() error {
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
		AppSource:   resourceAwsOpsworksApplicationAppSource(d),
		Environment: resourceAwsOpsworksApplicationEnvironmentVariable(d),
		DataSources: resourceAwsOpsworksApplicationDataSources(d),
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

func resourceAwsOpsworksApplicationAppSource(d *schema.ResourceData) *opsworks.Source {
	count := d.Get("app_source.#").(int)
	if count == 0 {
		return nil
	}

	return &opsworks.Source{
		Type:     aws.String(d.Get("app_source.0.type").(string)),
		Url:      aws.String(d.Get("app_source.0.url").(string)),
		Username: aws.String(d.Get("app_source.0.username").(string)),
		Password: aws.String(d.Get("app_source.0.password").(string)),
		Revision: aws.String(d.Get("app_source.0.revision").(string)),
		SshKey:   aws.String(d.Get("app_source.0.ssh_key").(string)),
	}
}

func resourceAwsOpsworksSetApplicationAppSource(d *schema.ResourceData, v *opsworks.Source) {
	nv := make([]interface{}, 0, 1)
	if v != nil {
		m := make(map[string]interface{})
		if v.Type != nil {
			m["type"] = *v.Type
		}
		if v.Url != nil {
			m["url"] = *v.Url
		}
		if v.Username != nil {
			m["username"] = *v.Username
		}
		if v.Password != nil {
			m["password"] = *v.Password
		}
		if v.Revision != nil {
			m["revision"] = *v.Revision
		}
		if v.SshKey != nil {
			m["ssh_key"] = *v.SshKey
		}
		nv = append(nv, m)
	}

	err := d.Set("app_source", nv)
	if err != nil {
		// should never happen
		panic(err)
	}
}

func resourceAwsOpsworksApplicationDataSources(d *schema.ResourceData) []*opsworks.DataSource {
	dataSources := d.Get("data_source").(*schema.Set).List()
	result := make([]*opsworks.DataSource, len(dataSources))

	for i := 0; i < len(dataSources); i++ {
		src := dataSources[i].(map[string]interface{})

		result[i] = &opsworks.DataSource{
			Arn:          aws.String(src["arn"].(string)),
			DatabaseName: aws.String(src["database_name"].(string)),
			Type:         aws.String(src["type"].(string)),
		}
	}
	return result
}

func resourceAwsOpsworksSetApplicationDataSources(d *schema.ResourceData, v []*opsworks.DataSource) {
	log.Printf("[DEBUG] data sources: %s %d", v, len(v))
	newValue := make([]*map[string]interface{}, len(v))

	for i := 0; i < len(v); i++ {
		config := v[i]
		data := make(map[string]interface{})
		newValue[i] = &data

		if config.Type != nil {
			data["type"] = *config.Type
		}
		if config.DatabaseName != nil {
			data["database_name"] = *config.DatabaseName
		}
		if config.Arn != nil {
			data["arn"] = *config.Arn
		}
		log.Printf("[DEBUG] v: %s", data)
	}

	log.Printf("[DEBUG] d: %s", d.Get("data_source"))
	d.Set("data_source", newValue)
}
