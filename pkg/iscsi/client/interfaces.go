package client

type IscsiClient interface {
	Login(targets []string) error
	Logout(targets []string) error
}
