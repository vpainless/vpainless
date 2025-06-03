package storage

import (
	"context"
	"net/url"

	"vpainless/internal/hosting/core"

	"github.com/gofrs/uuid/v5"
)

func (s *RepositoryTestSuite) Test_Get_Save_Group() {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	groupID := core.GroupID{UUID: uuid.Must(uuid.NewV4())}
	xrayID1 := core.XrayTemplateID{UUID: uuid.Must(uuid.NewV4())}
	xrayID2 := core.XrayTemplateID{UUID: uuid.Must(uuid.NewV4())}
	sshID := core.SSHKeyID{UUID: uuid.Must(uuid.NewV4())}
	startupScriptID := core.StartUpScriptID{UUID: uuid.Must(uuid.NewV4())}
	group := &core.Group{
		ID:   groupID,
		Name: "my group",
		Host: core.Provider{
			Base:   url.URL{Scheme: "https", Host: "api.vultr.com"},
			Name:   core.Vultr,
			APIKey: "my api key",
		},
		DefaultXrayTemplate: xrayID1,
		XrayTemplates: map[core.XrayTemplateID]core.XrayTemplate{
			xrayID1: {
				ID:   xrayID1,
				Base: "config content",
			},
		},
		DefaultSSHKey: core.SSHKeyPair{
			ID:         sshID,
			Name:       "system ssh key",
			PublicKey:  []byte("public key"),
			PrivateKey: []byte("private key"),
		},
		DefaultStartUpScript: core.StartUpScript{
			ID:      startupScriptID,
			Content: "#!/bin/bash",
		},
	}

	repo := NewRepository(s.db)

	_, err := repo.GetGroup(ctx, group.ID)
	s.Require().ErrorIs(err, core.ErrNotFound, "group should not be there")

	saved, err := repo.SaveGroup(ctx, group)
	s.Require().NoError(err, "should create group successfully")
	s.Require().Equal(group, saved, "saved group should match original")

	actual, err := repo.GetGroup(ctx, group.ID)
	s.Require().NoError(err, "should fetch group successfully")
	s.Require().Equal(group, actual, "saved group should match fetched one")

	group.Name = "another name"
	group.Host = core.Provider{
		Base:   url.URL{Scheme: "https", Host: "api2.vultr.com"},
		Name:   "digital",
		APIKey: "another api key",
	}
	group.DefaultSSHKey.RemoteID = core.SSHKeyID{UUID: uuid.Must(uuid.NewV4())}

	group.XrayTemplates[xrayID2] = core.XrayTemplate{
		ID:   xrayID2,
		Base: "config content 2",
	}
	group.DefaultXrayTemplate = xrayID2

	group.DefaultStartUpScript.RemoteID = core.StartUpScriptID{UUID: uuid.Must(uuid.NewV4())}

	actual, err = repo.SaveGroup(ctx, group)
	s.Require().NoError(err, "should save group successfully")
	s.Require().Equal(group, actual, "saved group should match original")

	newest, err := repo.GetGroup(ctx, group.ID)
	s.Require().NoError(err, "should fetch newest group successfully")
	s.Require().Equal(group, newest, "saved group should match newest")
}
