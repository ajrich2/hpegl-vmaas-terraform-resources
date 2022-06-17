// (C) Copyright 2022 Hewlett Packard Enterprise Development LP

package cmp

import (
	"context"
	"fmt"

	"github.com/HewlettPackard/hpegl-vmaas-cmp-go-sdk/pkg/client"
	"github.com/HewlettPackard/hpegl-vmaas-cmp-go-sdk/pkg/models"
	"github.com/HewlettPackard/hpegl-vmaas-terraform-resources/internal/utils"
	"github.com/tshihad/tftags"
)

type loadBalancerPool struct {
	lbClient *client.LoadBalancerAPIService
}

func newLoadBalancerPool(loadBalancerClient *client.LoadBalancerAPIService) *loadBalancerPool {
	return &loadBalancerPool{
		lbClient: loadBalancerClient,
	}
}

func (lb *loadBalancerPool) Read(ctx context.Context, d *utils.Data, meta interface{}) error {
	var lbPoolResp models.GetSpecificLBPoolResp
	if err := tftags.Get(d, &lbPoolResp); err != nil {
		return err
	}

	lbDetails, err := lb.lbClient.GetLoadBalancers(ctx)
	if err != nil {
		return err
	}

	getlbPoolResp, err := lb.lbClient.GetSpecificLBPool(ctx, lbDetails.GetNetworkLoadBalancerResp[0].ID, lbPoolResp.ID)
	if err != nil {
		return err
	}

	return tftags.Set(d, getlbPoolResp.GetSpecificLBPoolResp)
}

func (lb *loadBalancerPool) Create(ctx context.Context, d *utils.Data, meta interface{}) error {
	createReq := models.CreateLBPool{}
	if err := tftags.Get(d, &createReq.CreateLBPoolReq); err != nil {
		return err
	}

	lbDetails, err := lb.lbClient.GetLoadBalancers(ctx)
	if err != nil {
		return err
	}

	lbPoolResp, err := lb.lbClient.CreateLBPool(ctx, createReq, lbDetails.GetNetworkLoadBalancerResp[0].ID)
	if err != nil {
		return err
	}
	if !lbPoolResp.Success {
		return fmt.Errorf(successErr, "creating loadBalancer Pool")
	}

	// wait until created
	retry := &utils.CustomRetry{
		RetryDelay:   1,
		InitialDelay: 1,
	}
	_, err = retry.Retry(ctx, meta, func(ctx context.Context) (interface{}, error) {
		return lb.lbClient.GetSpecificLBPool(ctx, lbDetails.GetNetworkLoadBalancerResp[0].ID, lbPoolResp.LBPoolResp.ID)
	})
	if err != nil {
		return err
	}

	return tftags.Set(d, createReq.CreateLBPoolReq)
}

func (lb *loadBalancerPool) Update(ctx context.Context, d *utils.Data, meta interface{}) error {
	return nil
}

func (lb *loadBalancerPool) Delete(ctx context.Context, d *utils.Data, meta interface{}) error {
	lbPoolID := d.GetID()
	lbDetails, err := lb.lbClient.GetLoadBalancers(ctx)
	if err != nil {
		return err
	}

	_, err = lb.lbClient.DeleteLBPool(ctx, lbDetails.GetNetworkLoadBalancerResp[0].ID, lbPoolID)
	if err != nil {
		return err
	}

	return nil
}
