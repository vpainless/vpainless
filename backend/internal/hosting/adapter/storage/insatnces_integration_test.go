package storage

import (
	"context"
	"net"
	"time"

	"vpainless/internal/hosting/core"
	"vpainless/internal/pkg/authz"

	"github.com/gofrs/uuid/v5"
)

func (s *RepositoryTestSuite) Test_Get_Save_Instance() {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	now := time.Date(1984, 11, 5, 4, 32, 15, 0, time.UTC)
	userID := core.UserID{UUID: uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000")}
	instanceID := core.InstanceID{UUID: uuid.Must(uuid.NewV4())}

	instance := &core.Instance{
		ID:         instanceID,
		RemoteID:   instanceID,
		Owner:      userID,
		CreatedAt:  now,
		Status:     core.StatusInitializing,
		Config:     core.XrayConfig{},
		PrivateKey: []byte("my private key"),
	}

	repo := NewRepository(s.db)
	repo.now = func() time.Time {
		return now
	}
	nilPartial := authz.Clause{}
	_, err := repo.GetInstance(ctx, instanceID, nilPartial)
	s.Require().ErrorIs(err, core.ErrNotFound, "should not find the instance")

	actual, err := repo.SaveInstance(ctx, instance)
	s.Require().NoError(err, "should save instance without any error")
	s.Require().Equal(instance, actual, "saved instance should match the original one")

	actual, err = repo.GetInstance(ctx, instanceID, authz.Clause{
		Condition: "user_id = ?",
		Values:    []any{userID},
	})
	s.Require().NoError(err, "should get instance without any error")
	s.Require().Equal(instance, actual, "fetched instance should match the original one")

	instance.IP = net.ParseIP("192.168.0.1")
	instance.Status = core.StatusOK
	instance.Config.ConnectionString = "vless://my-vpn"
	actual, err = repo.SaveInstance(ctx, instance)
	s.Require().NoError(err, "should save modified instance without any error")
	s.Require().Equal(instance, actual, "saved instance should match the modified one")
}

func (s *RepositoryTestSuite) Test_Find_Save_Instance() {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	now := time.Date(1984, 11, 5, 4, 32, 15, 0, time.UTC)
	userID := core.UserID{UUID: uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000")}
	instanceID := core.InstanceID{UUID: uuid.Must(uuid.NewV4())}

	instance := &core.Instance{
		ID:        instanceID,
		RemoteID:  instanceID,
		Owner:     userID,
		IP:        net.ParseIP("192.168.0.1"),
		CreatedAt: now,
		Status:    core.StatusInitializing,
		Config: core.XrayConfig{
			ConnectionString: "xray://my-vpn",
		},
		PrivateKey: []byte("my private key"),
	}

	repo := NewRepository(s.db)
	repo.now = func() time.Time {
		return now
	}
	_, err := repo.FindInstance(ctx, userID)
	s.Require().ErrorIs(err, core.ErrNotFound, "should not find the instance")

	actual, err := repo.SaveInstance(ctx, instance)
	s.Require().NoError(err, "should save instance without any error")
	s.Require().Equal(instance, actual, "saved instance should match the original one")

	actual, err = repo.FindInstance(ctx, userID)
	s.Require().NoError(err, "should get instance without any error")
	s.Require().Equal(instance, actual, "fetched instance should match the original one")
}

func fakeInstance(id core.InstanceID, owner core.UserID, created time.Time) *core.Instance {
	return &core.Instance{
		ID:        id,
		RemoteID:  id,
		Owner:     owner,
		IP:        net.ParseIP("192.168.0.1"),
		CreatedAt: created,
		Status:    core.StatusInitializing,
		Config: core.XrayConfig{
			ConnectionString: "xray://my-vpn",
		},
		PrivateKey: []byte("my private key"),
	}
}

func (s *RepositoryTestSuite) Test_List_Instances() {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	now := time.Date(1984, 11, 5, 4, 32, 15, 0, time.UTC)
	adminID := core.UserID{UUID: uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000")}
	firstClient := core.UserID{UUID: uuid.FromStringOrNil("22000000-0000-0000-0000-000000000000")}
	secondClient := core.UserID{UUID: uuid.FromStringOrNil("33000000-0000-0000-0000-000000000000")}
	instanceIDs := []core.InstanceID{
		{UUID: uuid.Must(uuid.NewV4())},
		{UUID: uuid.Must(uuid.NewV4())},
		{UUID: uuid.Must(uuid.NewV4())},
	}

	instances := []*core.Instance{
		fakeInstance(instanceIDs[0], adminID, now),
		fakeInstance(instanceIDs[1], firstClient, now),
		fakeInstance(instanceIDs[2], secondClient, now),
	}

	repo := NewRepository(s.db)
	repo.now = func() time.Time {
		return now
	}

	for _, instance := range instances {
		_, err := repo.SaveInstance(ctx, instance)
		s.Require().NoError(err, "should save instance without any error")
	}

	actual, err := repo.ListInstances(ctx, authz.Clause{})
	s.Require().NoError(err, "should list instance without any error")
	s.Require().ElementsMatch(actual, instances, "should list all instances")

	actual, err = repo.ListInstances(ctx, authz.Clause{
		Condition: "group_id = ?",
		Values:    []any{uuid.FromStringOrNil("00000000-0000-0000-0000-111111111111")},
	})
	s.Require().NoError(err, "should list instance without any error")
	s.Require().ElementsMatch(actual, instances[:2], "should list all group instancesl")

	actual, err = repo.ListInstances(ctx, authz.Clause{
		Condition: "user_id = ?",
		Values:    []any{firstClient},
	})
	s.Require().NoError(err, "should list instance without any error")
	s.Require().ElementsMatch(actual, instances[1:2], "should list first client instance")
}

func (s *RepositoryTestSuite) Test_Delete_Instance() {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	now := time.Date(1984, 11, 5, 4, 32, 15, 0, time.UTC)
	userID := core.UserID{UUID: uuid.FromStringOrNil("11000000-0000-0000-0000-000000000000")}
	instanceID := core.InstanceID{UUID: uuid.Must(uuid.NewV4())}

	instance := &core.Instance{
		ID:         instanceID,
		RemoteID:   instanceID,
		Owner:      userID,
		CreatedAt:  now,
		Status:     core.StatusInitializing,
		Config:     core.XrayConfig{},
		PrivateKey: []byte("my private key"),
	}

	repo := NewRepository(s.db)
	repo.now = func() time.Time {
		return now
	}
	nilPartial := authz.Clause{}
	_, err := repo.GetInstance(ctx, instanceID, nilPartial)
	s.Require().ErrorIs(err, core.ErrNotFound, "should not find the instance")

	actual, err := repo.SaveInstance(ctx, instance)
	s.Require().NoError(err, "should save instance without any error")
	s.Require().Equal(instance, actual, "saved instance should match the original one")

	err = repo.DeleteInstance(ctx, instanceID, nilPartial)
	s.Require().NoError(err, "should delete instance without any error")

	_, err = repo.SaveInstance(ctx, instance)
	s.Require().NoError(err, "save instance after deletion should have no effect")

	_, err = repo.GetInstance(ctx, instanceID, authz.Clause{
		Condition: "user_id = ?",
		Values:    []any{userID},
	})
	s.Require().ErrorIs(err, core.ErrNotFound, "should not get instance after delettion")

	_, err = repo.FindInstance(ctx, userID)
	s.Require().ErrorIs(err, core.ErrNotFound, "should not find instance after delettion")
}

func (s *RepositoryTestSuite) Test_Delete_Instance_Partials() {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	now := time.Date(1984, 11, 5, 4, 32, 15, 0, time.UTC)
	userID := core.UserID{UUID: uuid.FromStringOrNil("22000000-0000-0000-0000-000000000000")}
	groupID := core.GroupID{UUID: uuid.FromStringOrNil("00000000-0000-0000-0000-111111111111")}
	instanceID := core.InstanceID{UUID: uuid.Must(uuid.NewV4())}

	instance := &core.Instance{
		ID:         instanceID,
		RemoteID:   instanceID,
		Owner:      userID,
		CreatedAt:  now,
		Status:     core.StatusInitializing,
		Config:     core.XrayConfig{},
		PrivateKey: []byte("my private key"),
	}

	repo := NewRepository(s.db)
	repo.now = func() time.Time {
		return now
	}
	nilPartial := authz.Clause{}
	actual, err := repo.SaveInstance(ctx, instance)
	s.Require().NoError(err, "should save instance without any error")
	s.Require().Equal(instance, actual, "saved instance should match the original one")

	err = repo.DeleteInstance(ctx, instanceID, nilPartial)
	s.Require().NoError(err, "should delete instance without any error")

	instanceID = core.InstanceID{UUID: uuid.Must(uuid.NewV4())}
	instance.ID = instanceID
	actual, err = repo.SaveInstance(ctx, instance)
	s.Require().NoError(err, "should save instance without any error")
	s.Require().Equal(instance, actual, "saved instance should match the original one")

	err = repo.DeleteInstance(ctx, instanceID, authz.Clause{
		Condition: "user_id = ?",
		Values:    []any{uuid.Must(uuid.NewV4())},
	})
	s.Require().ErrorIs(err, core.ErrNotFound, "should not find the instance for deletion")

	err = repo.DeleteInstance(ctx, instanceID, authz.Clause{
		Condition: "user_id = ?",
		Values:    []any{userID},
	})
	s.Require().NoError(err, "should be able to delete the user with correct partial")

	instanceID = core.InstanceID{UUID: uuid.Must(uuid.NewV4())}
	instance.ID = instanceID
	actual, err = repo.SaveInstance(ctx, instance)
	s.Require().NoError(err, "should save instance without any error")
	s.Require().Equal(instance, actual, "saved instance should match the original one")

	err = repo.DeleteInstance(ctx, instanceID, authz.Clause{
		Condition: "user_id in (select id from users where group_id = ?)",
		Values:    []any{uuid.Must(uuid.NewV4())},
	})
	s.Require().ErrorIs(err, core.ErrNotFound, "should not find the instance for deletion")

	err = repo.DeleteInstance(ctx, instanceID, authz.Clause{
		Condition: "user_id in (select id from users where group_id = ?)",
		Values:    []any{groupID},
	})
	s.Require().NoError(err, "should be able to delete the instance with group partial")
}
