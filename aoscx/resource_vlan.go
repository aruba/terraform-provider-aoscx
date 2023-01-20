package aoscx

import (
	"context"
	"strconv"

	"github.com/aruba/aoscxgo"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceVlan() *schema.Resource {
	return &schema.Resource{
		Description:   "Resource to configure VLANs on AOS-CX switches.",
		CreateContext: resourceVlanCreate,
		ReadContext:   resourceVlanRead,
		UpdateContext: resourceVlanUpdate,
		DeleteContext: resourceVlanDelete,
		Schema: map[string]*schema.Schema{
			"vlan_id": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"name": &schema.Schema{
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
		},
	}
}

func resourceVlanCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error

	sw := m.(*aoscxgo.Client)
	vlan_id := d.Get("vlan_id").(int)

	tmp_vlan := aoscxgo.Vlan{
		VlanId:      vlan_id,
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		AdminState:  d.Get("admin_state").(string),
	}

	err = tmp_vlan.Create(sw)

	//defer logout(tr, cookie, url)

	if materialized := tmp_vlan.GetStatus(); !materialized {

		err = tmp_vlan.Get(sw)

		if err != nil {
			{
				diags = append(diags, diag.Errorf("Error in Creating VLAN ", err)...)
				return diags
			}
		}

		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "VLAN Already Existing",
			Detail:   string(tmp_vlan.VlanId),
		})

	}

	d.SetId(strconv.Itoa(vlan_id))
	d.Set("vlan_id", vlan_id)

	resourceVlanRead(ctx, d, m)

	return diags
}

func resourceVlanRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error

	sw := m.(*aoscxgo.Client)

	// Retrieve VLAN from sw if existing
	tmp_vlan := aoscxgo.Vlan{
		VlanId: d.Get("vlan_id").(int),
	}

	err = tmp_vlan.Get(sw)

	if err != nil {
		// Failure in VLAN retrieval
		d.SetId("")
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "VLAN Not Found",
			Detail:   "VLAN Not Found",
		})
		return diags
	}

	d.Set("name", tmp_vlan.Name)
	d.Set("description", tmp_vlan.Description)
	d.Set("admin_state", tmp_vlan.AdminState)

	return diags
}

func resourceVlanUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error

	sw := m.(*aoscxgo.Client)

	// Retrieve VLAN from sw if existing
	tmp_vlan := aoscxgo.Vlan{
		VlanId: d.Get("vlan_id").(int),
	}

	err = tmp_vlan.Get(sw)

	if d.HasChange("name") {
		tmp_vlan.Name = d.Get("name").(string)
	}

	if d.HasChange("description") {
		tmp_vlan.Description = d.Get("description").(string)
	}

	if d.HasChange("admin_state") {
		tmp_state := d.Get("admin_state").(string)
		if tmp_state == "" || (tmp_state != "down" && tmp_state != "up") {
			tmp_vlan.AdminState = "down"
			d.Set("admin_state", "down")
		}
		tmp_vlan.AdminState = tmp_state
	}

	err = tmp_vlan.Update(sw)

	if err != nil {
		if err.(*aoscxgo.RequestError).StatusCode == "404 Not Found" {
			diags = append(diags, diag.Errorf("Error Updating VLAN does not exist: %s", err.(*aoscxgo.RequestError).StatusCode)...)
			return diags
		} else if err.(*aoscxgo.RequestError).StatusCode != "204 No Content" {
			diags = append(diags, diag.Errorf("Error in Updating VLAN: %s", err.(*aoscxgo.RequestError).StatusCode)...)
			return diags
		}
	}

	return resourceVlanRead(ctx, d, m)
}

func resourceVlanDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var err error

	sw := m.(*aoscxgo.Client)

	// Retrieve VLAN from sw if existing
	tmp_vlan := aoscxgo.Vlan{
		VlanId: d.Get("vlan_id").(int),
	}

	err = tmp_vlan.Delete(sw)

	if err != nil {
		if err.(*aoscxgo.RequestError).StatusCode == "404 Not Found" {
			diags = append(diags, diag.Errorf("Error Updating VLAN does not exist: %s", err.(*aoscxgo.RequestError).StatusCode)...)
			return diags
		} else if err.(*aoscxgo.RequestError).StatusCode != "204 No Content" {
			diags = append(diags, diag.Errorf("Error in Updating VLAN: %s ", err.(*aoscxgo.RequestError).StatusCode)...)
			return diags
		}
	}

	d.SetId("")
	return nil
}
