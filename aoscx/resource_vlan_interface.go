package aoscx

import (
	"context"
	"fmt"
	"sort"

	"github.com/aruba/aoscxgo"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceVlanInterface() *schema.Resource {
	return &schema.Resource{
		Description:   "Resource to configure Vlan interface attributes on AOS-CX switches.",
		CreateContext: resourceVlanInterfaceCreate,
		ReadContext:   resourceVlanInterfaceRead,
		UpdateContext: resourceVlanInterfaceUpdate,
		DeleteContext: resourceVlanInterfaceDelete,

		Schema: map[string]*schema.Schema{
			"vlan_id": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
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

func resourceVlanInterfaceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error

	sw := m.(*aoscxgo.Client)
	// Create Vlan Object, if Vlan doesn't exist error will occur
	tmp_vlan := aoscxgo.Vlan{
		VlanId: d.Get("vlan_id").(int),
	}

	tmp_vlan_int := aoscxgo.VlanInterface{
		Vlan: tmp_vlan,
	}
	tmp_vlan_int.Vlan.AdminState = d.Get("admin_state").(string)
	tmp_vlan_int.Description = d.Get("description").(string)
	// sort supplied ipv4 to match sorted REST response
	tmp_vlan_int.Ipv4 = d.Get("ipv4").([]interface{})
	var tmp_splice []string
	for _, ip_addr := range tmp_vlan_int.Ipv4 {
		if ip_addr != nil {
			tmp_splice = append(tmp_splice, ip_addr.(string))
		}
		if len(tmp_splice) > 1 {
			sort.Strings(tmp_splice[1:])
		}
	}
	for index := 0; index < len(tmp_splice); index++ {
		tmp_vlan_int.Ipv4[index] = tmp_splice[index]
	}
	tmp_set := d.Get("ipv6").(*schema.Set)
	tmp_vlan_int.Ipv6 = tmp_set.List()
	tmp_vlan_int.Vrf = d.Get("vrf").(string)

	err = tmp_vlan_int.Create(sw)

	if materialized := tmp_vlan_int.GetStatus(); !materialized {

		err = tmp_vlan_int.Get(sw)

		if err != nil {
			diags = append(diags, diag.Errorf("Error in Creating Vlan Interface: %s", err.(*aoscxgo.RequestError).StatusCode)...)
			return diags
		}
	}
	str_vlanint_id := fmt.Sprintf("vlanint_%v", tmp_vlan_int.Vlan.VlanId)
	d.SetId(str_vlanint_id)
	d.Set("vlan_id", tmp_vlan_int.Vlan.VlanId)

	resourceVlanInterfaceRead(ctx, d, m)

	return diags
}

func resourceVlanInterfaceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error

	sw := m.(*aoscxgo.Client)

	// Retrieve VLAN from sw if existing
	tmp_vlan := aoscxgo.Vlan{
		VlanId: d.Get("vlan_id").(int),
	}

	tmp_vlan_int := aoscxgo.VlanInterface{
		Vlan: tmp_vlan,
	}

	err = tmp_vlan_int.Get(sw)

	if err != nil {
		//Failure in VlanInterface retrieval
		d.SetId("")
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "VlanInterface Not Found",
			Detail:   "VlanInterface Not Found",
		})
		return diags
	}

	d.Set("vlan_id", tmp_vlan_int.Vlan.VlanId)
	d.Set("description", tmp_vlan_int.Description)

	d.Set("admin_state", tmp_vlan_int.Vlan.AdminState)

	d.Set("ipv4", tmp_vlan_int.Ipv4)
	d.Set("ipv6", tmp_vlan_int.Ipv6)
	d.Set("vrf", tmp_vlan_int.Vrf)

	return diags
}

func resourceVlanInterfaceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var err error

	sw := m.(*aoscxgo.Client)
	// Retrieve VLAN from sw if existing
	tmp_vlan := aoscxgo.Vlan{
		VlanId: d.Get("vlan_id").(int),
	}

	err = tmp_vlan.Get(sw)

	if err != nil {
		diags = append(diags, diag.Errorf("VLAN missing - Error in Updating VlanInterface ", err)...)
		return diags
	}
	tmp_vlan_int := aoscxgo.VlanInterface{
		Vlan: tmp_vlan,
	}
	err = tmp_vlan_int.Get(sw)

	if err != nil {
		diags = append(diags, diag.Errorf("VLANInterface missing - Error in Updating VlanInterface ", err)...)
		return diags
	}

	// Flag to determine if put should be used instead of patch
	use_put := false

	if d.HasChange("description") {
		tmp_vlan_int.Description = d.Get("description").(string)
	}

	if d.HasChange("ipv4") {
		// sort supplied ipv4 to match sorted REST response
		tmp_vlan_int.Ipv4 = d.Get("ipv4").([]interface{})
		var tmp_splice []string
		for _, ip_addr := range tmp_vlan_int.Ipv4 {
			if ip_addr != nil {
				tmp_splice = append(tmp_splice, ip_addr.(string))
			}
			if len(tmp_splice) > 1 {
				sort.Strings(tmp_splice[1:])
			}
		}
		for _, ip_addr := range tmp_vlan_int.Ipv4 {
			if ip_addr != nil {
				tmp_splice = append(tmp_splice, ip_addr.(string))
			}
			if len(tmp_splice) > 1 {
				sort.Strings(tmp_splice[1:])
			}
		}
		for index := 0; index < len(tmp_splice); index++ {
			if index < len(tmp_vlan_int.Ipv4) {
				tmp_vlan_int.Ipv4[index] = tmp_splice[index]
			}
		}

		use_put = true
	}

	if d.HasChange("ipv6") {
		tmp_set := d.Get("ipv6").(*schema.Set)
		tmp_vlan_int.Ipv6 = tmp_set.List()
	}

	if d.HasChange("vrf") {
		tmp_vlan_int.Vrf = d.Get("vrf").(string)
	}

	if d.HasChange("admin_state") {
		tmp_state := d.Get("admin_state").(string)
		if tmp_state == "" || (tmp_state != "down" && tmp_state != "up") {
			tmp_vlan_int.Vlan.AdminState = "down"
			d.Set("admin_state", "down")
			// Should enforce value can only be up or down
		}
		tmp_vlan_int.Vlan.AdminState = tmp_state
	}

	err = tmp_vlan_int.Update(sw, use_put)

	if err != nil {
		if err.(*aoscxgo.RequestError).StatusCode == "404 Not Found" {
			diags = append(diags, diag.Errorf("Error Updating Interface does not exist: %s", err.(*aoscxgo.RequestError).StatusCode)...)
			return diags
		} else if err.(*aoscxgo.RequestError).StatusCode != "204 No Content" {
			diags = append(diags, diag.Errorf("Error in Updating Interface: %s", err.(*aoscxgo.RequestError).StatusCode)...)
			return diags
		}
	}

	return resourceVlanInterfaceRead(ctx, d, m)
}

func resourceVlanInterfaceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var err error

	sw := m.(*aoscxgo.Client)

	// Delete Interface Obj
	tmp_vlan := aoscxgo.Vlan{
		VlanId: d.Get("vlan_id").(int),
	}

	tmp_vlan_int := aoscxgo.VlanInterface{
		Vlan: tmp_vlan,
	}

	err = tmp_vlan_int.Delete(sw)

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
