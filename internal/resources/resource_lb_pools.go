// (C) Copyright 2022 Hewlett Packard Enterprise Development LP

package resources

import (
	"context"

	"github.com/HewlettPackard/hpegl-vmaas-terraform-resources/internal/resources/validations"
	"github.com/HewlettPackard/hpegl-vmaas-terraform-resources/internal/utils"
	"github.com/HewlettPackard/hpegl-vmaas-terraform-resources/pkg/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func LoadBalancerPools() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"lb_id": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Parent lb ID, lb_id can be obtained by using LB datasource/resource.",
				ForceNew:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Network loadbalancer pool name",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Creating the Network loadbalancer pool.",
			},
			"min_active_members": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1,
				Description: "minimum active members for the Network loadbalancer pool",
			},
			"algorithm": {
				Type: schema.TypeString,
				ValidateDiagFunc: validations.StringInSlice([]string{
					"ROUND_ROBIN",
					"WEIGHTED_ROUND_ROBIN",
					"LEAST_CONNECTION",
					"WEIGHTED_LEAST_CONNECTION",
					"IP_HASH",
				}, false),
				Required:     true,
				InputDefault: "ROUND_ROBIN",
				Description:  "Provide the Supported values for pool algorithm",
			},
			"config": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "pool Configuration",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"snat_translation_type": {
							Type:             schema.TypeString,
							ValidateDiagFunc: validations.StringInSlice([]string{"LBSnatAutoMap", "LBSnatDisabled", "LBSnatIpPool"}, false),
							Optional:         true,
							Default:          "LBSnatDisabled",
							Description:      "Network Loadbalancer Supported values are `LBSnatAutoMap`,`LBSnatDisabled`, `LBSnatIpPool`",
						},
						"passive_monitor_path": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "passive_monitor_path for Network loadbalancer pool",
						},
						"active_monitor_paths": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "active_monitor_paths for Network loadbalancer pool",
						},
						"tcp_multiplexing": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "tcp_multiplexing for Network loadbalancer pool",
						},
						"tcp_multiplexing_number": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     6,
							Description: "tcp_multiplexing_number for Network loadbalancer pool",
						},
						"snat_ip_address": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "snat_ip_address for Network loadbalancer pool",
						},
						"member_group": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "member group",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"group": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "name of member group",
									},
									"max_ip_list_size": {
										Type:        schema.TypeInt,
										Optional:    true,
										Description: "max_ip_list_size of member group",
									},
									"ip_revision_filter": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "ipRevisionFilter of member group",
									},
									"port": {
										Type:        schema.TypeInt,
										Optional:    true,
										Description: "port of member group",
									},
								},
							},
						},
					},
				},
			},
			"tags": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "tags Configuration",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"tag": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "tag for Network Load balancer Profile",
						},
						"scope": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "scope for Network Load balancer Profile",
						},
					},
				},
			},
		},
		ReadContext:   loadbalancerPoolReadContext,
		UpdateContext: loadbalancerPoolUpdateContext,
		CreateContext: loadbalancerPoolCreateContext,
		DeleteContext: loadbalancerPoolDeleteContext,
		Description: `loadbalancer Pool resource facilitates creating,
		and deleting NSX-T  Network Load Balancers.`,
	}
}

func loadbalancerPoolUpdateContext(ctx context.Context, rd *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := client.GetClientFromMetaMap(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	data := utils.NewData(rd)
	if err := c.CmpClient.LoadBalancerPool.Update(ctx, data, meta); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func loadbalancerPoolReadContext(ctx context.Context, rd *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := client.GetClientFromMetaMap(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	data := utils.NewData(rd)
	if err := c.CmpClient.LoadBalancerPool.Read(ctx, data, meta); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func loadbalancerPoolCreateContext(ctx context.Context, rd *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := client.GetClientFromMetaMap(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	data := utils.NewData(rd)
	if err := c.CmpClient.LoadBalancerPool.Create(ctx, data, meta); err != nil {
		return diag.FromErr(err)
	}

	return loadbalancerPoolReadContext(ctx, rd, meta)
}

func loadbalancerPoolDeleteContext(ctx context.Context, rd *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, err := client.GetClientFromMetaMap(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	data := utils.NewData(rd)
	if err := c.CmpClient.LoadBalancerPool.Delete(ctx, data, meta); err != nil {
		return diag.FromErr(err)
	}

	return nil
}