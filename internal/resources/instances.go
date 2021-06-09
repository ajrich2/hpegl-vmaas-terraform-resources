// (C) Copyright 2021 Hewlett Packard Enterprise Development LP

package resources

import (
	"context"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hpe-hcss/vmaas-terraform-resources/internal/utils"
	"github.com/hpe-hcss/vmaas-terraform-resources/pkg/client"
)

const (
	instanceAvailableTimeout = 60 * time.Minute
	instanceReadTimeout      = 2 * time.Minute
	instanceDeleteTimeout    = 60 * time.Minute
	instanceRetryTimeout     = 10 * time.Minute
	instanceRetryDelay       = 60 * time.Second
	instanceRetryMinTimeout  = 30 * time.Second
)

func Instances() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the instance to be provisioned.",
			},
			"cloud_id": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: f(generalDDesc, "cloud"),
			},
			"group_id": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: f(generalDDesc, "group"),
			},
			"plan_id": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: f(generalDDesc, "plan"),
			},
			"layout_id": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: f(generalDDesc, "layout"),
			},
			"instance_code": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Unique code used to identify the instance type.",
			},
			"network": {
				Type:        schema.TypeList,
				Required:    true,
				Description: "Details of the network to which the instance should belong.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: f(generalDDesc, "network"),
						},
					},
				},
			},
			"volume": {
				Type:     schema.TypeList,
				Required: true,
				Description: `A list of volumes to be created inside a provisioned instance.
				It can have a root volume and other secondary volumes.`,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Unique name for the volume.",
						},
						"size": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Size of the volume in GB.",
						},
						"datastore_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: f(generalDDesc, "datastore"),
						},
						"root": {
							Type:        schema.TypeBool,
							Default:     true,
							Optional:    true,
							Description: "If true then the given volume as considered as root volume.",
						},
						"id": {
							Computed:    true,
							Type:        schema.TypeInt,
							Description: "ID for the volume",
						},
						"persist_volume_on_update": {
							Optional: true,
							Type:     schema.TypeBool,
						},
					},
				},
			},
			"labels": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "A string used for labelling instances.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"tags": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "A list of key and value pairs used to tag instances of similar type.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"hostname": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Hostname for the instance",
			},
			"config": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "Configuration details for the instance to be provisioned.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"resource_pool_id": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: f(generalDDesc, "resource pool"),
						},
						"template": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Unique ID for the template",
						},
						"no_agent": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							Description: "If true agent will not be installed on the instance.",
						},
						"vm_folder": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Folder name where will be stored.",
						},
						"create_user": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "If true new user will be created",
						},
						"asset_tag": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Asset tag",
						},
					},
				},
			},
			"scale": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1,
				Description: "Number of nodes within an instance.",
			},
			"evars": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Environment Variables to be added to the provisioned instance.",
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
				Description: `Status of the instance .It can be one among these:
				 Provisioning/Failed/Unknown/Running.`,
			},
			"clone": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "If Clone is provided, this instance will created from cloning an existing instance",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"source_instance_id": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Instance ID of the source.",
						},
					},
				},
			},
			"power_schedule_id": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Scheduled power operations",
			},
		},
		SchemaVersion:  0,
		StateUpgraders: nil,
		CreateContext:  instanceCreateContext,
		ReadContext:    instanceReadContext,
		UpdateContext:  instanceUpdateContext,
		DeleteContext:  instanceDeleteContext,
		CustomizeDiff:  nil,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(instanceAvailableTimeout),
			Update: schema.DefaultTimeout(instanceAvailableTimeout),
			Delete: schema.DefaultTimeout(instanceDeleteTimeout),
			Read:   schema.DefaultTimeout(instanceReadTimeout),
		},
		Description: `Instance resource facilitates creating,
		updating and deleting virtual machines.
		For creating an instance, provide a unique name and all the Mandatory(Required) parameters.
		It is recommend to use the Vmware type for provisioning.`,
	}
}

func instanceCreateContext(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := client.GetClientFromMetaMap(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	data := utils.NewData(d)
	if err := c.CmpClient.Instance.Create(ctx, data); err != nil {
		return diag.FromErr(err)
	}

	// Wait for the status to be running
	createStateConf := resource.StateChangeConf{
		Delay:      instanceRetryDelay,
		Pending:    []string{"provisioning"},
		Target:     []string{"running"},
		Timeout:    instanceRetryTimeout,
		MinTimeout: instanceRetryMinTimeout,
		Refresh: func() (result interface{}, state string, err error) {
			if err := c.CmpClient.Instance.Read(ctx, data); err != nil {
				return nil, "", err
			}

			return d.Get("name"), data.GetString("status"), nil
		},
	}
	_, err = createStateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func instanceReadContext(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Print("[INFO] this a log")
	c, err := client.GetClientFromMetaMap(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	data := utils.NewData(d)
	err = c.CmpClient.Instance.Read(ctx, data)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func instanceDeleteContext(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := client.GetClientFromMetaMap(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	data := utils.NewData(d)
	if err := c.CmpClient.Instance.Delete(ctx, data); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func instanceUpdateContext(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}