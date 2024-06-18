package aoscx

import (
	"context"
	"sort"

	"github.com/aruba/aoscxgo"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceL3Interface() *schema.Resource {
	return &schema.Resource{
		Description:   "Resource to configure interface Layer3 attributes on AOS-CX switches.",
		CreateContext: resourceL3InterfaceCreate,
		ReadContext:   resourceL3InterfaceRead,
		UpdateContext: resourceL3InterfaceUpdate,
		DeleteContext: resourceL3InterfaceDelete,

		Schema: map[string]*schema.Schema{
			"interface": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"description": &schema.Schema{
				Type:     schema.TypeString,
				Required: false,
				Optional: true,
			},
			"admin_state": &schema.Schema{
				Type:         schema.TypeString,
				Required:     false,
				Default:      "up",
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"up", "down"}, true),
			},
			"ipv4": &schema.Schema{
				Type:     schema.TypeList,
				Required: false,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
				Default:  nil,
			},
			"ipv6": &schema.Schema{
				Type:     schema.TypeSet,
				Required: false,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
				Default:  nil,
			},
			"vrf": &schema.Schema{
				Type:     schema.TypeString,
				Required: false,
				Default:  "default",
				Optional: true,
			},
		},
	}
}

func resourceL3InterfaceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error

	sw := m.(*aoscxgo.Client)

	tmp_int := aoscxgo.Interface{
		Name:       d.Get("interface").(string),
		AdminState: d.Get("admin_state").(string),
	}

	err = tmp_int.Create(sw)

	if materialized := tmp_int.GetStatus(); !materialized {

		err = tmp_int.Get(sw)

		if err != nil {
			{
				diags = append(diags, diag.Errorf("Error in Creating Interface ", err)...)
				return diags
			}
		}

		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Interface Already Existing",
			Detail:   string(tmp_int.Name),
		})
	}

	tmp_l3_int := aoscxgo.L3Interface{}
	tmp_l3_int.Interface = tmp_int
	tmp_l3_int.Description = d.Get("description").(string)

	// sort supplied ipv4 to match sorted REST response
	tmp_l3_int.Ipv4 = d.Get("ipv4").([]interface{})
	var tmp_splice []string
	for _, ip_addr := range tmp_l3_int.Ipv4 {
		if ip_addr != nil {
			tmp_splice = append(tmp_splice, ip_addr.(string))
		}
		if len(tmp_splice) > 1 {
			sort.Strings(tmp_splice[1:])
		}
	}
	for index := 0; index < len(tmp_splice); index++ {
		tmp_l3_int.Ipv4[index] = tmp_splice[index]
	}

	tmp_set := d.Get("ipv6").(*schema.Set)
	tmp_l3_int.Ipv6 = tmp_set.List()
	tmp_l3_int.Vrf = d.Get("vrf").(string)

	err = tmp_l3_int.Create(sw)

	if materialized := tmp_l3_int.GetStatus(); !materialized {

		get_err := tmp_l3_int.Get(sw)

		if get_err != nil {
			diags = append(diags, diag.Errorf("Error in Creating L3 Interface: %s", err.(*aoscxgo.RequestError).StatusCode)...)
			return diags
		}
	}

	d.SetId(d.Get("interface").(string))
	d.Set("interface", d.Get("interface").(string))

	resourceL3InterfaceRead(ctx, d, m)

	return diags
}

func resourceL3InterfaceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error

	sw := m.(*aoscxgo.Client)

	// Retrieve Interface from sw if existing
	tmp_int := aoscxgo.L3Interface{
		Interface: aoscxgo.Interface{
			Name: d.Get("interface").(string),
		}}
	err = tmp_int.Get(sw)

	if err != nil {
		//Failure in L3Interface retrieval
		d.SetId("")
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Interface Not Found",
			Detail:   "Interface Not Found",
		})
		return diags
	}

	d.Set("interface", tmp_int.Interface.Name)
	d.Set("description", tmp_int.Description)

	d.Set("admin_state", tmp_int.Interface.AdminState)

	d.Set("ipv4", tmp_int.Ipv4)
	d.Set("ipv6", tmp_int.Ipv6)
	d.Set("vrf", tmp_int.Vrf)

	return diags
}

func resourceL3InterfaceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var err error

	sw := m.(*aoscxgo.Client)

	// Retrieve Interface from sw if existing
	tmp_l3_int := aoscxgo.L3Interface{
		Interface: aoscxgo.Interface{
			Name: d.Get("interface").(string),
		}}

	//tmp_vlan.GetStatus() will return if existing
	err = tmp_l3_int.Get(sw)

	if err != nil {
		{
			diags = append(diags, diag.Errorf("Error in Retrieving Interface ", err)...)
			return diags
		}
	}

	// Flag to determine if put should be used instead of patch
	use_put := false

	if d.HasChange("description") {
		tmp_l3_int.Description = d.Get("description").(string)
	}

	if d.HasChange("ipv4") {
		tmp_l3_int.Ipv4 = d.Get("ipv4").([]interface{})
		// sort supplied ipv4 to match sorted REST response
		tmp_l3_int.Ipv4 = d.Get("ipv4").([]interface{})
		var tmp_splice []string
		for _, ip_addr := range tmp_l3_int.Ipv4 {
			if ip_addr != nil {
				tmp_splice = append(tmp_splice, ip_addr.(string))
			}
			if len(tmp_splice) > 1 {
				sort.Strings(tmp_splice[1:])
			}
		}
		for index := 0; index < len(tmp_splice); index++ {
			tmp_l3_int.Ipv4[index] = tmp_splice[index]
		}

		use_put = true
	}

	if d.HasChange("ipv6") {
		tmp_set := d.Get("ipv6").(*schema.Set)
		tmp_l3_int.Ipv6 = tmp_set.List()
	}

	if d.HasChange("vrf") {
		tmp_l3_int.Vrf = d.Get("vrf").(string)
	}

	if d.HasChange("admin_state") {
		tmp_state := d.Get("admin_state").(string)
		if tmp_state == "" || (tmp_state != "down" && tmp_state != "up") {
			tmp_l3_int.Interface.AdminState = "down"
			d.Set("admin_state", "down")
			// Should enforce value can only be up or down
		}
		tmp_l3_int.Interface.AdminState = tmp_state
	}

	err = tmp_l3_int.Update(sw, use_put)

	if err != nil {
		if err.(*aoscxgo.RequestError).StatusCode == "404 Not Found" {
			diags = append(diags, diag.Errorf("Error Updating Interface does not exist: %s", err.(*aoscxgo.RequestError).StatusCode)...)
			return diags
		} else if err.(*aoscxgo.RequestError).StatusCode != "204 No Content" {
			diags = append(diags, diag.Errorf("Error in Updating Interface: %s", err.(*aoscxgo.RequestError).StatusCode)...)
			return diags
		}
	}

	return resourceL3InterfaceRead(ctx, d, m)
}

func resourceL3InterfaceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var err error

	sw := m.(*aoscxgo.Client)

	// Delete Interface Obj
	tmp_int := aoscxgo.Interface{
		Name: d.Get("interface").(string),
	}

	err = tmp_int.Delete(sw)

	if err != nil {
		if err.(*aoscxgo.RequestError).StatusCode == "404 Not Found" {
			diags = append(diags, diag.Errorf("Error Deleting Interface does not exist: %s", err.(*aoscxgo.RequestError).StatusCode)...)
			return diags
		} else if err.(*aoscxgo.RequestError).StatusCode != "204 No Content" {
			diags = append(diags, diag.Errorf("Error in Deleting Interface: %s ", err.(*aoscxgo.RequestError).StatusCode)...)
			return diags
		}
	}

	d.SetId("")
	return nil
}
