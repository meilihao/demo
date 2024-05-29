package main

import (
	"fmt"
	"log"
	"time"

	ovirtsdk4 "github.com/ovirt/go-ovirt"
)

func main() {
	// Create the connection to the api server
	inputRawURL := "https://192.168.88.152/ovirt-engine/api"
	conn, err := ovirtsdk4.NewConnectionBuilder().
		URL(inputRawURL).
		Username("admin@ovirt@internalsso").
		Password("password").
		Insecure(true).
		Compress(true).
		Timeout(time.Second * 10).
		Build()
	if err != nil {
		log.Fatalf("Make connection failed, reason: %s", err.Error())
	}

	defer conn.Close()

	// Get the reference to the "clusters" service
	clustersService := conn.SystemService().ClustersService()

	// Use the "list" method of the "clusters" service to list all the clusters of the system
	clustersResponse, err := clustersService.List().Send()
	if err != nil {
		fmt.Printf("Failed to get cluster list, reason: %v\n", err)
		return
	}

	if clusters, ok := clustersResponse.Clusters(); ok {
		// Print the datacenter names and identifiers
		fmt.Printf("Cluster: (")
		for _, cluster := range clusters.Slice() {
			if clusterName, ok := cluster.Name(); ok {
				fmt.Printf(" name: %v", clusterName)
			}
			if clusterId, ok := cluster.Id(); ok {
				fmt.Printf(" id: %v", clusterId)
			}
		}
		fmt.Println(")")
	}

	ListVm(conn)
	ListStorage(conn)
}

func ListVm(conn *ovirtsdk4.Connection) {
	vmsService := conn.SystemService().VmsService()

	// Use the "list" method of the "clusters" service to list all the clusters of the system
	vmsResponse, err := vmsService.List().Send()
	if err != nil {
		fmt.Printf("Failed to get vm list, reason: %v\n", err)
		return
	}

	if vms, ok := vmsResponse.Vms(); ok {
		// Print the datacenter names and identifiers
		fmt.Printf("Vm: (")
		for _, vm := range vms.Slice() {
			if vmName, ok := vm.Name(); ok {
				fmt.Printf(" name: %v", vmName)
			}
			if vmId, ok := vm.Id(); ok {
				fmt.Printf(" id: %v", vmId)
			}
		}
		fmt.Println(")")
	}
}

func ListStorage(conn *ovirtsdk4.Connection) {
	sdsService := conn.SystemService().StorageDomainsService()

	// Use the "list" method of the "clusters" service to list all the clusters of the system
	sdsResponse, err := sdsService.List().Send()
	if err != nil {
		fmt.Printf("Failed to get storagedomain list, reason: %v\n", err)
		return
	}

	if sds, ok := sdsResponse.StorageDomains(); ok {
		// Print the datacenter names and identifiers
		fmt.Printf("storagedomain: (")
		for _, sd := range sds.Slice() {
			if sdName, ok := sd.Name(); ok {
				fmt.Printf(" name: %v", sdName)
			}
			if sdId, ok := sd.Id(); ok {
				fmt.Printf(" id: %v", sdId)
			}
		}
		fmt.Println(")")
	}
}
