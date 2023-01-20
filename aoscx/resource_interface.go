package aoscx

import (
	"context"

	"github.com/aruba/aoscxgo"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceInterface() *schema.Resource {
	return &schema.Resource{
		Description:   "Resource to configure interfaces physical attributes on AOS-CX switches.",
		CreateContext: resourceInterfaceCreate,
		ReadContext:   resourceInterfaceRead,
		UpdateContext: resourceInterfaceUpdate,
		DeleteContext: resourceInterfaceDelete,

		Schema: map[string]*schema.Schema{
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

func resourceInterfaceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var err error

	sw := m.(*aoscxgo.Client)

	tmp_int := aoscxgo.Interface{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		AdminState:  d.Get("admin_state").(string),
	}

	err = tmp_int.Create(sw)

	//defer logout(tr, cookie, url)

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

		err = tmp_int.Update(sw)

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

	d.SetId(d.Get("name").(string))
	d.Set("name", d.Get("name").(string))

	resourceInterfaceRead(ctx, d, m)

	return diags
}

func resourceInterfaceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var err error

	sw := m.(*aoscxgo.Client)

	// Retrieve Interface from sw if existing
	tmp_int := aoscxgo.Interface{
		Name: d.Get("name").(string),
	}
	//tmp_vlan.GetStatus() will return if existing
	err = tmp_int.Get(sw)

	if err != nil {
		//Failure in VLAN retrieval
		d.SetId("")
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Interface Not Found",
			Detail:   "Interface Not Found",
		})
		return diags
	}

	d.Set("name", tmp_int.Name)
	d.Set("description", tmp_int.Description)
	d.Set("admin_state", tmp_int.AdminState)

	return diags
}

func resourceInterfaceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var err error

	sw := m.(*aoscxgo.Client)

	// Retrieve Interface from sw if existing
	tmp_int := aoscxgo.Interface{
		Name: d.Get("name").(string),
	}
	//tmp_vlan.GetStatus() will return if existing
	err = tmp_int.Get(sw)

	if d.HasChange("description") {
		tmp_int.Description = d.Get("description").(string)
	}

	if d.HasChange("admin_state") {
		tmp_state := d.Get("admin_state").(string)
		if tmp_state == "" || (tmp_state != "down" && tmp_state != "up") {
			tmp_int.AdminState = "down"
			d.Set("admin_state", "down")
			// Should enforce value can only be up or down
		}
		tmp_int.AdminState = tmp_state
	}

	err = tmp_int.Update(sw)

	if err != nil {
		if err.(*aoscxgo.RequestError).StatusCode == "404 Not Found" {
			diags = append(diags, diag.Errorf("Error Updating Interface does not exist: %s", err.(*aoscxgo.RequestError).StatusCode)...)
			return diags
		} else if err.(*aoscxgo.RequestError).StatusCode != "204 No Content" {
			diags = append(diags, diag.Errorf("Error in Updating Interface: %s", err.(*aoscxgo.RequestError).StatusCode)...)
			return diags
		}
	}

	return resourceInterfaceRead(ctx, d, m)
}

func resourceInterfaceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics
	var err error

	sw := m.(*aoscxgo.Client)

	// Create Interface Obj
	tmp_int := aoscxgo.Interface{
		Name: d.Get("name").(string),
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
