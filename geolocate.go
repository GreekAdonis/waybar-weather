package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/mdlayher/wifi"
)

type Address struct {
	MAC      string `json:"macAddress"`
	Age      int64  `json:"age,omitempty"`
	Strength int32  `json:"signalStrength,omitempty"`
}

func geolocate() {
	var ifaceList []*wifi.Interface
	var addresses []Address
	wlan, err := wifi.New()
	if err != nil {
		log.Fatal("failed to initialize wifi: ", err)
	}
	ifaces, err := wlan.Interfaces()
	if err != nil {
		log.Fatal("failed to get interfaces: ", err)
	}
	for _, iface := range ifaces {
		if iface.Type == wifi.InterfaceTypeStation {
			ifaceList = append(ifaceList, iface)
		}
	}

	for _, iface := range ifaceList {
		fmt.Printf("Checking BSS for interface %s\n", iface.Name)
		addresses = append(addresses, Address{MAC: iface.HardwareAddr.String()})

		aps, err := wlan.AccessPoints(iface)
		if err != nil {
			log.Printf("failed to get access points for %q: %s", iface.Name, err)
		}
		for _, ap := range aps {
			if ap.SSID == "" || ap.SSID[0] == '\x00' || strings.HasSuffix(ap.SSID, "_nomap") {
				continue
			}
			addresses = append(addresses, Address{
				MAC:      ap.BSSID.String(),
				Age:      ap.LastSeen.Microseconds(),
				Strength: ap.Signal / 100,
			})
		}
	}

	if err = json.NewEncoder(os.Stdout).Encode(addresses); err != nil {
		log.Fatal("failed to encode addresses: ", err)
	}
}

// curl --json '{"considerIp":true,"wifiAccessPoints":[{"macAddress":"72:30:f6:84:a9:49"},{"macAddress":"ac:15:a2:ac:1b:f7"},{"macAddress":"b4:8a:0a:1b:ac:81"}]}' https://api.positon.xyz/v1/geolocate\?key\=56aba903-ae67-4f26-919b-15288b44bda9 | jq
