module example

go 1.16

replace (
	go.etcd.io/etcd/api/v3 => /home/chen/test/etcd-master/api
	go.etcd.io/etcd/client/v3 => /home/chen/test/etcd-master/client/v3
	go.etcd.io/etcd/pkg/v3 => /home/chen/test/etcd-master/pkg
)

require (
	github.com/golang/protobuf v1.4.3 // indirect
	github.com/google/uuid v1.2.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	go.etcd.io/etcd/client/v3 v3.0.0-00010101000000-000000000000
	golang.org/x/net v0.0.0-20200625001655-4c5254603344 // indirect
	golang.org/x/sys v0.0.0-20201214210602-f9fddec55a1e // indirect
	golang.org/x/text v0.3.2 // indirect
	golang.org/x/tools v0.0.0-20200103221440-774c71fcf114 // indirect
	gopkg.in/yaml.v2 v2.3.0 // indirect
)
