package main

import (
	"log"

	"github.com/fanatic/go-infoblox"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

//Provider returns a terraform.ResourceProvider.
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"username": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("INFOBLOX_USERNAME", nil),
				Description: "Infoblox Username",
			},
			"password": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("INFOBLOX_PASSWORD", nil),
				Description: "Infoblox User Password",
			},
			"host": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("INFOBLOX_HOST", nil),
				Description: "Infoblox Base Url(defaults to testing)",
			},
			"sslverify": &schema.Schema{
				Type:        schema.TypeBool,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("INFOBLOX_SSLVERIFY", true),
				Description: "Enable ssl",
			},
			"usecookies": &schema.Schema{
				Type:        schema.TypeBool,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("INFOBLOX_USECOOKIES", false),
				Description: "Use cookies",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"infoblox_record": resourceInfobloxRecord(),
			"infoblox_ip":     resourceInfobloxIP(),
		},

		ConfigureFunc: provideConfigure,
	}
}

func provideConfigure(d *schema.ResourceData) (interface{}, error) {
	username := d.Get("username").(string)
	password := d.Get("password").(string)
	host := d.Get("host").(string)
	sslVerify := d.Get("sslverify").(bool)
	useCookies := d.Get("usecookies").(bool)

	client := infoblox.NewClient(host, username, password, sslVerify, useCookies)
	log.Printf("[INFO] Infoblox Client configured for user: %s", username)
	return client, nil
}
