package aoscx

import (
	"context"

	"github.com/aruba/aoscxgo"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceL2Interface() *schema.Resource {
	return &schema.Resource{
		Description:   "Resource to configure interface Layer2 attributes on AOS-CX switches.",
		CreateContext: resourceL2InterfaceCreate,
		ReadContext:   resourceL2InterfaceRead,
		UpdateContext: resourceL2InterfaceUpdate,
		DeleteContext: resourceL2InterfaceDelete,

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
			"vlan_mode": &schema.Schema{
				Type:         schema.TypeString,
				Required:     false,
				Default:      "access",
				Optional:     true,
				ValidateFunc: validation.StringInSlice([]string{"access", "trunk"}, true),
			},
			"vlan_tag": &schema.Schema{
				Type:     schema.TypeInt,
				Required: false,
				Default:  1,
				Optional: true,
			},
			"vlan_ids": &schema.Schema{
				Type:     schema.TypeSet,
				Required: false,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
				Optional: true,
			},
			"trunk_allowed_all": &schema.Schema{
				Type:     schema.TypeBool,
				Required: false,
				Default:  false,
				Optional: true,
			},
			"native_vlan_tag": &schema.Schema{
				Type:     schema.TypeBool,
				Required: false,
				Default:  false,
				Optional: true,
			},
		},
	}
}

func resourceL2InterfaceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

		vlan_mode := d.Get("vlan_mode").(string)

		tmp_l2_int := aoscxgo.L2Interface{}
		if vlan_mode == "access" {
			tmp_l2_int.Interface = tmp_int
			tmp_l2_int.Description = d.Get("description").(string)
			tmp_l2_int.VlanMode = d.Get("vlan_mode").(string)
			tmp_l2_int.VlanTag = d.Get("vlan_tag").(int)

		} else if vlan_mode == "trunk" {
			tmp_l2_int.Interface = tmp_int
			tmp_l2_int.Description = d.Get("description").(string)
			tmp_l2_int.VlanMode = d.Get("vlan_mode").(string)
			tmp_l2_int.NativeVlanTag = d.Get("native_vlan_tag").(bool)
			tmp_l2_int.VlanTag = d.Get("vlan_tag").(int)
			tmp_l2_int.TrunkAllowedAll = d.Get("trunk_allowed_all").(bool)
			tmp_set := d.Get("vlan_ids").(*schema.Set)
			tmp_l2_int.VlanIds = tmp_set.List()
		}

		err = tmp_l2_int.Create(sw)

		if err != nil {
			if err.(*aoscxgo.RequestError).StatusCode == "404 Not Found" {
				diags = append(diags, diag.Errorf("Error Updating Interface does not exist: %s", err.(*aoscxgo.RequestError).StatusCode)...)
				return diags
			} else if err.(*aoscxgo.RequestError).StatusCode != "204 No Content" {
				diags = append(diags, diag.Errorf("Error in Updating Interface: %s", err.(*aoscxgo.RequestError).StatusCode)...)
				return diags
			}
		}

	}

	d.SetId(d.Get("interface").(string))
	d.Set("interface", d.Get("interface").(string))

	resourceL2InterfaceRead(ctx, d, m)

	return diags
}

func resourceL2InterfaceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error

	sw := m.(*aoscxgo.Client)

	// Retrieve Interface from sw if existing
	tmp_int := aoscxgo.L2Interface{
		Interface: aoscxgo.Interface{
			Name: d.Get("interface").(string),
		}}

	err = tmp_int.Get(sw)

	if err != nil {
		//Failure in Interface retrieval
		d.SetId("")
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Interface Not Found",
			Detail:   "Interface Not Found",
		})
		return diags
	}

	if tmp_int.VlanMode == "access" || tmp_int.VlanMode == "" {
		d.Set("vlan_mode", "access")
		d.Set("vlan_tag", tmp_int.VlanTag)

	} else if tmp_int.VlanMode == "trunk" || tmp_int.VlanMode == "native-tagged" || tmp_int.VlanMode == "native-untagged" {
		d.Set("vlan_mode", "trunk")
		d.Set("native_vlan_tag", tmp_int.NativeVlanTag)
		d.Set("trunk_allowed_all", tmp_int.TrunkAllowedAll)
		d.Set("vlan_ids", tmp_int.VlanIds)
		if tmp_int.VlanTag == 0 {
			d.Set("vlan_tag", 1)
		} else {
			d.Set("vlan_tag", tmp_int.VlanTag)
		}

	}

	d.Set("interface", tmp_int.Interface.Name)
	d.Set("description", tmp_int.Description)

	d.Set("admin_state", tmp_int.Interface.AdminState)

	return diags
}

func resourceL2InterfaceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error

	sw := m.(*aoscxgo.Client)

	// Retrieve Interface from sw if existing
	tmp_l2_int := aoscxgo.L2Interface{
		Interface: aoscxgo.Interface{
			Name: d.Get("interface").(string),
		}}

	err = tmp_l2_int.Get(sw)

	// Flag to determine if put should be used instead of patch
	use_put := false

	if d.HasChange("description") {
		tmp_l2_int.Description = d.Get("description").(string)
	}
	if d.HasChange("vlan_mode") {
		tmp_vlan_mode := d.Get("vlan_mode").(string)
		use_put = true

		if tmp_vlan_mode == "access" {
			tmp_l2_int.VlanMode = tmp_vlan_mode
			tmp_l2_int.VlanTag = d.Get("vlan_tag").(int)

		} else if tmp_vlan_mode == "trunk" {
			// need to figure out how to retrieve value from terraform config file
			// include logic that will only
			// I need to change VlanMode attribute to include logic to handle translation be trunk = ["native-untagged", "native-tagged"]
			tmp_l2_int.VlanMode = tmp_vlan_mode
			tmp_l2_int.VlanTag = d.Get("vlan_tag").(int)
			tmp_l2_int.NativeVlanTag = d.Get("native_vlan_tag").(bool)
			tmp_l2_int.TrunkAllowedAll = d.Get("trunk_allowed_all").(bool)
			tmp_set := d.Get("vlan_ids").(*schema.Set)
			tmp_l2_int.VlanIds = tmp_set.List()
		}
	}
	if d.HasChange("vlan_tag") {
		tmp_l2_int.VlanTag = d.Get("vlan_tag").(int)
	}

	if d.HasChange("vlan_ids") {
		tmp_set := d.Get("vlan_ids").(*schema.Set)
		tmp_l2_int.VlanIds = tmp_set.List()
	}

	if d.HasChange("trunk_allowed_all") {
		tmp_l2_int.TrunkAllowedAll = d.Get("trunk_allowed_all").(bool)
	}

	if d.HasChange("native_vlan_tag") {
		tmp_l2_int.NativeVlanTag = d.Get("native_vlan_tag").(bool)
	}

	if d.HasChange("admin_state") {
		tmp_state := d.Get("admin_state").(string)
		if tmp_state == "" || (tmp_state != "down" && tmp_state != "up") {
			tmp_l2_int.Interface.AdminState = "down"
			d.Set("admin_state", "down")
		}
		tmp_l2_int.Interface.AdminState = tmp_state
	}

	err = tmp_l2_int.Update(sw, use_put)

	if err != nil {
		if err.(*aoscxgo.RequestError).StatusCode == "404 Not Found" {
			diags = append(diags, diag.Errorf("Error Updating Interface does not exist: %s", err.(*aoscxgo.RequestError).StatusCode)...)
			return diags
		} else if err.(*aoscxgo.RequestError).StatusCode != "204 No Content" {
			diags = append(diags, diag.Errorf("Error in Updating Interface: %s", err.(*aoscxgo.RequestError).StatusCode)...)
			return diags
		}
	}

	return resourceL2InterfaceRead(ctx, d, m)
}

func resourceL2InterfaceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error

	sw := m.(*aoscxgo.Client)

	// Create Interface Obj
	tmp_int := aoscxgo.Interface{
		Name: d.Get("interface").(string),
	}

	err = tmp_int.Delete(sw)

	if err != nil {
		if err.(*aoscxgo.RequestError).StatusCode == "404 Not Found" {
			diags = append(diags, diag.Errorf("Error Updating Interface does not exist: %s", err.(*aoscxgo.RequestError).StatusCode)...)
			return diags
		} else if err.(*aoscxgo.RequestError).StatusCode != "204 No Content" {
			diags = append(diags, diag.Errorf("Error in Updating Interface: %s ", err.(*aoscxgo.RequestError).StatusCode)...)
			return diags
		}
	}

	d.SetId("")
	return nil
}
