package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/fanatic/go-infoblox"
	"github.com/hashicorp/terraform/helper/schema"
)

// InfobloxIPResponse structure for ip serialization
type InfobloxIPResponse struct {
	IPAddresses []string `json:"ips"`
}

func resourceInfobloxIP() *schema.Resource {
	return &schema.Resource{
		Create: resourceInfobloxIPCreate,
		Read:   resourceInfobloxIPRead,
		Update: resourceInfobloxIPUpdate,
		Delete: resourceInfobloxIPDelete,

		Schema: map[string]*schema.Schema{
			"cidr": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},

			"addresses": &schema.Schema{
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
				Required: false,
			},

			"exclude": &schema.Schema{
				Type:     schema.TypeSet,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},

			"ip_count": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1,
			},
		},
	}
}

func resourceInfobloxIPCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*infoblox.Client)

	log.Print("[TRACE] inside resourceInfobloxIPCreate.")

	ntwork := d.Get("cidr")

	log.Printf("[TRACE] CIDR from terraform file: %s", ntwork.(string))

	s := "network"
	q := []infoblox.Condition{
		infoblox.Condition{
			Field: &s,
			Value: ntwork.(string),
		},
	}

	log.Print("[TRACE] invoking client.Network().find")

	out, err := client.Network().Find(q, nil)

	if err != nil {
		log.Printf("[ERROR] Unable to invoke find on cidr: %s, %s", ntwork, err)
		return err
	}

	if len(out) == 0 {
		return fmt.Errorf("Empty response from client.Network().find. Is %s a valid network?", ntwork)
	}

	log.Print("[TRACE] invoking client.NetworkObject().NextAvailableIP")

	var excludedAddresses []string
	if userExcludes := d.Get("exclude"); userExcludes != nil {
		addresses := userExcludes.(*schema.Set).List()
		for _, address := range addresses {
			excludedAddresses = append(excludedAddresses, address.(string))
		}
	}

	log.Printf("[TRACE] Excluding Addresses = %v", excludedAddresses)

	ipCount := d.Get("ip_count").(int)
	ou, err := client.NetworkObject(out[0]["_ref"].(string)).NextAvailableIP(ipCount, excludedAddresses)

	if err != nil {
		log.Printf("[ERROR] Unable to allocate NextAvailableIP: %s", err)
		return err
	}

	log.Print("[TRACE] Walking NextAvailableIP output to get ip")
	log.Printf("[TRACE] NextAvailableIP returned %v", ou)

	var res InfobloxIPResponse
	data, err := json.Marshal(ou)
	if err != nil {
		log.Printf("[Error] Unable to re-serialize response: %s", err)
		return err
	}

	err = json.Unmarshal(data, &res)
	if err != nil {
		log.Printf("[Error] Unable to deserialize response: %s", err)
		return err
	}

	if err != nil {
		log.Print("Error: unable to determine IP address from response \n", err)
		return nil
	}

	log.Printf("[TRACE] returned value in ips structure: %v", res)

	log.Print("[TRACE] Setting ID, locking provisioned IP in terraform")

	d.SetId(strings.Join(res.IPAddresses, " "))

	log.Print("[TRACE] Setting output variable 'ipaddress'")

	if err := d.Set("addresses", res.IPAddresses); err != nil {
		log.Printf("[ERROR] Setting output variable 'addresses', %s", err)
		return err
	}

	log.Print("[TRACE] exiting resourceInfobloxIPCreate.")

	return nil
}

func resourceInfobloxIPRead(d *schema.ResourceData, meta interface{}) error {

	// since the infoblox network object's NextAvailableIP function isn't exactly
	// a resource (you don't really allocate an IP address until you use the record:a or
	// record:host object), we don't actually implement READ, UPDATE, or DELETE

	return nil
}

func resourceInfobloxIPUpdate(d *schema.ResourceData, meta interface{}) error {

	// since the infoblox network object's NextAvailableIP function isn't exactly
	// a resource (you don't really allocate an IP address until you use the record:a or
	// record:host object), we don't actually implement READ, UPDATE, or DELETE

	return nil
}

func resourceInfobloxIPDelete(d *schema.ResourceData, meta interface{}) error {

	// since the infoblox network object's NextAvailableIP function isn't exactly
	// a resource (you don't really allocate an IP address until you use the record:a or
	// record:host object), we don't actually implement READ, UPDATE, or DELETE

	return nil
}
