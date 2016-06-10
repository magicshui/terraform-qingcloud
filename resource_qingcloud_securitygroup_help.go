package qingcloud

import (
	"github.com/magicshui/qingcloud-go/securitygroup"
)

func applySecurityGroupUpdates(meta interface{}, id string) error {
	clt := meta.(*QingCloudClient).securitygroup
	params := securitygroup.ApplySecurityGroupRequest{}
	params.SecurityGroup.Set(id)

	if _, err := clt.ApplySecurityGroup(params); err != nil {
		return err
	}

	_, err := SecurityGroupTransitionStateRefresh(clt, id)
	return err
}
