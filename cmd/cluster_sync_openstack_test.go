package cmd

import "testing"

func TestClusterSyncOpenStackCommandRegistration(t *testing.T) {
	cluster := NewClusterCmd()
	syncCmd, _, err := cluster.Find([]string{"sync"})
	if err != nil || syncCmd == nil || syncCmd.Name() != "sync" {
		t.Fatalf("cluster sync command not registered: command=%v err=%v", syncCmd, err)
	}
	openstackCmd, _, err := cluster.Find([]string{"sync", "openstack"})
	if err != nil || openstackCmd == nil || openstackCmd.Name() != "openstack" {
		t.Fatalf("cluster sync openstack command not registered: command=%v err=%v", openstackCmd, err)
	}
}
