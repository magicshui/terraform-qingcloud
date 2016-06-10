package qingcloud

import (
	"fmt"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/magicshui/qingcloud-go/eip"
	"github.com/magicshui/qingcloud-go/instance"
	"github.com/magicshui/qingcloud-go/loadbalancer"
	"github.com/magicshui/qingcloud-go/router"
	"github.com/magicshui/qingcloud-go/volume"
)

func transitionStateRefresh(refreshFunc func() (interface{}, string, error), pending, target []string) (interface{}, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    pending,
		Target:     target,
		Refresh:    refreshFunc,
		Timeout:    10 * time.Minute,
		Delay:      5 * time.Second,
		MinTimeout: 5 * time.Second,
	}
	return stateConf.WaitForState()
}

// LoadbalancerTransitionStateRefresh 刷新 Loadbalancer 的状态，直到状态固定为止
func LoadbalancerTransitionStateRefresh(clt *loadbalancer.LOADBALANCER, id string) (interface{}, error) {
	refreshFunc := func() (interface{}, string, error) {
		params := loadbalancer.DescribeLoadBalancersRequest{}
		params.LoadbalancersN.Add(id)
		params.Verbose.Set(1)

		resp, err := clt.DescribeLoadBalancers(params)
		if err != nil {
			return nil, "", err
		}
		if resp.TotalCount != 1 {
			return nil, "", fmt.Errorf("LB not found: %s", id)
		}
		return resp.LoadbalancerSet[0], resp.LoadbalancerSet[0].TransitionStatus, nil
	}
	pending := []string{"creating", "starting", "stopping", "updating", "suspending", "resuming", "deleting"}
	target := []string{""}

	return transitionStateRefresh(refreshFunc, pending, target)

}

// EipTransitionStateRefresh 等待 EIP 状态稳定下来
func EipTransitionStateRefresh(clt *eip.EIP, id string) (interface{}, error) {
	refreshFunc := func() (interface{}, string, error) {
		params := eip.DescribeEipsRequest{}
		params.EipsN.Add(id)
		params.Verbose.Set(1)

		resp, err := clt.DescribeEips(params)
		if err != nil {
			return nil, "", err
		}
		if resp.TotalCount != 1 {
			return nil, "", fmt.Errorf("Eip not found: %s", id)
		}
		return resp.EipSet[0], resp.EipSet[0].TransitionStatus, nil
	}

	pending := []string{"associating", "dissociating", "suspending", "resuming", "releasing"}
	target := []string{""}
	return transitionStateRefresh(refreshFunc, pending, target)

}

func VolumeTransitionStateRefresh(clt *volume.VOLUME, id string) (interface{}, error) {
	refreshFunc := func() (interface{}, string, error) {
		params := volume.DescribeVolumesRequest{}
		params.VolumesN.Add(id)
		params.Verbose.Set(1)

		resp, err := clt.DescribeVolumes(params)
		if err != nil {
			return nil, "", err
		}
		if resp.TotalCount != 1 {
			return nil, "", fmt.Errorf("Volume not found: %s", id)
		}
		return resp.VolumeSet[0], resp.VolumeSet[0].TransitionStatus, nil
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{"creating", "attaching", "detaching", "suspending", "suspending", "resuming", "deleting", "recovering"}, // creating, attaching, detaching, suspending，resuming，deleting，recovering
		Target:     []string{""},
		Refresh:    refreshFunc,
		Timeout:    10 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 10 * time.Second,
	}
	return stateConf.WaitForState()

}

// RouterTransitionStateRefresh 等待路由器状态稳定下来
func RouterTransitionStateRefresh(clt *router.ROUTER, id string) (interface{}, error) {
	refreshFunc := func() (interface{}, string, error) {
		params := router.DescribeRoutersRequest{}
		params.RoutersN.Add(id)
		params.Verbose.Set(1)
		resp, err := clt.DescribeRouters(params)
		if err != nil {
			return nil, "", err
		}
		if resp.TotalCount != 1 {
			return nil, "", fmt.Errorf("Router not found: %s", id)
		}
		return resp.RouterSet[0], resp.RouterSet[0].TransitionStatus, nil
	}

	pending := []string{"creating", "updating", "suspending", "resuming", "poweroffing", "poweroning", "deleting"}
	target := []string{""}
	return transitionStateRefresh(refreshFunc, pending, target)
}

// InstanceTransitionStateRefresh 等待主机状态稳定下来
func InstanceTransitionStateRefresh(clt *instance.INSTANCE, id string) (interface{}, error) {
	refreshFunc := func() (interface{}, string, error) {
		params := instance.DescribeInstanceRequest{}
		params.InstancesN.Add(id)
		params.Verbose.Set(1)
		resp, err := clt.DescribeInstances(params)
		if err != nil {
			return nil, "", err
		}
		if resp.TotalCount != 1 {
			return nil, "", fmt.Errorf("Instance not found: %s", id)
		}
		return resp.InstanceSet[0], resp.InstanceSet[0].TransitionStatus, nil
	}

	pending := []string{"creating", "updating", "suspending", "resuming", "poweroffing", "poweroning", "deleting"}
	target := []string{""}
	return transitionStateRefresh(refreshFunc, pending, target)
}
