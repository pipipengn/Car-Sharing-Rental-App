package profile

import (
	"context"
	rentalpb "coolcar/rental/api/gen/v1"
	"coolcar/shared/id"
	"encoding/base64"
	"fmt"
	"google.golang.org/protobuf/proto"
)

type Fetcher interface {
	GetProfile(c context.Context, request *rentalpb.GetProfileRequest) (*rentalpb.Profile, error)
}

type Manager struct {
	Fetcher Fetcher
}

func (m *Manager) Verify(c context.Context, aid id.AccountID) (id.IdentityID, error) {
	profile, err := m.Fetcher.GetProfile(c, &rentalpb.GetProfileRequest{})
	if err != nil {
		return "", fmt.Errorf("cannot get profile: %v", err)
	}

	if profile.IdentityStatus != rentalpb.IdentityStatus_VERIFIED {
		return "", fmt.Errorf("invalid identity status: %v", err)
	}

	b, err := proto.Marshal(profile.Identity)
	if err != nil {
		return "", fmt.Errorf("cannot marshal identity: %v", err)
	}
	return id.IdentityID(base64.StdEncoding.EncodeToString(b)), nil
}
