package aoscx

import (
	"context"
	"crypto/tls"
	"net/http"

	"github.com/aruba/aoscxgo"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

type Aoscx struct {
	hostname     string
	username     string
	password     string
	rest_version string
	cookie       *http.Cookie
}

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"hostname": {
				Type:        schema.TypeString,
				Required:    true,
				Optional:    false,
				Description: "Hostname/IP address of the AOS-CX switch to connect to",
			},
			"username": {
				Type:        schema.TypeString,
				Optional:    false,
				Required:    true,
				Description: "Username used to authenticate",
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    false,
				Required:    true,
				Description: "Password used to authenticate",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"aoscx_vlan":         resourceVlan(),
			"aoscx_interface":    resourceInterface(),
			"aoscx_l2_interface": resourceL2Interface(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	hostname := d.Get("hostname").(string)
	username := d.Get("username").(string)
	password := d.Get("password").(string)

	var diags diag.Diagnostics

	if (hostname != "") && (username != "") && (password != "") {
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}

		sw, err := aoscxgo.Connect(
			&aoscxgo.Client{
				Hostname:  hostname,
				Username:  username,
				Password:  password,
				Transport: tr,
			},
		)

		if (sw.Cookie == nil) || (err != nil) {
			return nil, diag.FromErr(err)
		}

		return sw, diags
	}

	diags = append(diags, diag.Diagnostic{
		Severity: diag.Error,
		Summary:  "Unable to create AOS-CX client",
		Detail:   "Invalid or no values found for hostname, username, password",
	})

	return nil, diags
}
