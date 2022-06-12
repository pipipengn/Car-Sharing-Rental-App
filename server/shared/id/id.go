package id

type AccountID string

func (a AccountID) String() string {
	return string(a)
}

type TripID string

func (a TripID) String() string {
	return string(a)
}

type IdentityID string

func (a IdentityID) String() string {
	return string(a)
}

type CarID string

func (a CarID) String() string {
	return string(a)
}

type BlobID string

func (a BlobID) String() string {
	return string(a)
}
