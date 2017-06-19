package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/fanatic/go-infoblox"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceInfobloxHost() *schema.Resource {
	return &schema.Resource{
		Create: resourceInfobloxHostCreate,
		Read:   resourceInfobloxHostRead,
		Update: resourceInfobloxHostUpdate,
		Delete: resourceInfobloxHostDelete,

		Schema: map[string]*schema.Schema{
			"domain": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"ttl": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "3600",
			},

			"cidr": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: false,
			},

			"address": &schema.Schema{
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
		},
	}
}

func resourceInfobloxHostCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*infoblox.Client)

	var excludedAddresses []string
	if userExcludes := d.Get("exclude"); userExcludes != nil {
		addresses := userExcludes.(*schema.Set).List()
		for _, address := range addresses {
			excludedAddresses = append(excludedAddresses, address.(string))
		}
	}

	cidr := d.Get("cidr").(string)
	if cidr == "" {
		return fmt.Errorf("Invalid cidr block")
	}

	name := d.Get("name").(string)
	if name == "" {
		return fmt.Errorf("Invalid name")
	}

	domain := d.Get("domain").(string)
	if domain == "" {
		return fmt.Errorf("Invalid domain")
	}

	ipAddress, err := findIP(client, cidr, excludedAddresses)

	if err != nil {
		return err
	}

	record := url.Values{}
	record.Set("name", strings.Join([]string{name, domain}, "."))

	if attr, ok := d.GetOk("ttl"); ok {
		record.Set("ttl", attr.(string))
	}
	record.Set("ipv4addr", ipAddress)

	log.Printf("[DEBUG] Infoblox Record create configuration: %#v", record)

	var recID string

	opts := &infoblox.Options{
		ReturnFields: []string{"ttl", "ipv4addr", "name"},
	}
	recID, err = client.RecordA().Create(record, opts, nil)

	if err != nil {
		return fmt.Errorf("Failed to create Infblox Record: %s", err)
	}

	d.SetId(recID)

	log.Printf("[INFO] record ID: %s", d.Id())

	return resourceInfobloxRecordRead(d, meta)
}

func resourceInfobloxHostRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*infoblox.Client)

	rec, err := client.GetRecordA(d.Id(), nil)
	if err != nil {
		return fmt.Errorf("Couldn't find Infoblox A record: %s", err)
	}
	d.Set("address", rec.Ipv4Addr)
	fqdn := strings.Split(rec.Name, ".")
	d.Set("name", fqdn[0])
	d.Set("domain", strings.Join(fqdn[1:], "."))
	d.Set("ttl", rec.Ttl)

	return nil
}

func resourceInfobloxHostUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*infoblox.Client)
	var recID string
	var err, updateErr error

	name := d.Get("name").(string)
	if name == "" {
		return fmt.Errorf("Invalid name")
	}

	domain := d.Get("domain").(string)
	if domain == "" {
		return fmt.Errorf("Invalid domain")
	}

	_, err = client.GetRecordA(d.Id(), nil)

	if err != nil {
		return fmt.Errorf("Couldn't find Infoblox record: %s", err)
	}

	record := url.Values{}
	record.Set("name", strings.Join([]string{name, domain}, "."))
	if attr, ok := d.GetOk("ttl"); ok {
		record.Set("ttl", attr.(string))
	}

	log.Printf("[DEBUG] Infoblox Record update configuration: %#v", record)

	opts := &infoblox.Options{
		ReturnFields: []string{"ttl", "ipv4addr", "name"},
	}
	recID, updateErr = client.RecordAObject(d.Id()).Update(record, opts, nil)

	if updateErr != nil {
		return fmt.Errorf("Failed to update Infblox Record: %s", err)
	}

	d.SetId(recID)

	return resourceInfobloxRecordRead(d, meta)
}

func resourceInfobloxHostDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*infoblox.Client)

	log.Printf("[INFO] Deleting Infoblox Record: %s, %s", d.Get("name").(string), d.Id())

	_, err := client.GetRecordA(d.Id(), nil)
	if err != nil {
		return fmt.Errorf("Couldn't find Infoblox A record: %s", err)
	}

	deleteErr := client.RecordAObject(d.Id()).Delete(nil)
	if deleteErr != nil {
		return fmt.Errorf("Error deleting Infoblox A Record: %s", err)
	}
	return nil
}

func findIP(client *infoblox.Client, cidr string, excludedAddresses []string) (string, error) {
	fieldName := "network"
	query := []infoblox.Condition{
		infoblox.Condition{
			Field: &fieldName,
			Value: cidr,
		},
	}

	results, err := client.Network().Find(query, nil)

	if err != nil {
		log.Printf("[ERROR] Unable to invoke find on cidr: %s, %s", cidr, err)
		return "", err
	}

	if len(results) == 0 {
		return "", fmt.Errorf("Empty response from client.Network().find. Is %s a valid network?", cidr)
	}

	netIP, err := client.NetworkObject(results[0]["_ref"].(string)).NextAvailableIP(1, excludedAddresses)

	if err != nil {
		log.Printf("[ERROR] Unable to allocate NextAvailableIP: %s", err)
		return "", err
	}

	var ipInfo InfobloxIPResponse
	data, err := json.Marshal(netIP)
	if err != nil {
		log.Printf("[Error] Unable to re-serialize response: %s", err)
		return "", err
	}

	err = json.Unmarshal(data, &ipInfo)
	if err != nil {
		log.Printf("[Error] Unable to deserialize response: %s", err)
		return "", err
	}

	if err != nil {
		log.Print("Error: unable to determine IP address from response \n", err)
		return "", err
	}

	if len(ipInfo.IPAddresses) != 1 {
		return "", fmt.Errorf("Error: unable to get an IP address")
	}

	return ipInfo.IPAddresses[0], nil
}
