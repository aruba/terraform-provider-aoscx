package aoscx

import (
	"context"
	"hash/fnv"
	"strconv"

	"github.com/aruba/aoscxgo"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceFullConfig() *schema.Resource {
	return &schema.Resource{
		Description:   "Resource to manage full running-config on AOS-CX switches.",
		CreateContext: resourceFullConfigCreate,
		ReadContext:   resourceFullConfigRead,
		UpdateContext: resourceFullConfigUpdate,
		DeleteContext: resourceFullConfigDelete,
		Schema: map[string]*schema.Schema{
			"filename": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				Optional: false,
				ForceNew: true,
			},
			"config": &schema.Schema{
				Type:     schema.TypeString,
				Required: false,
				Optional: true,
				Computed: true,
			},
			"diff": &schema.Schema{
				Type:     schema.TypeString,
				Required: false,
				Optional: true,
			},
		},
	}
}

func resourceFullConfigCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error

	sw := m.(*aoscxgo.Client)
	filename := d.Get("filename").(string)

	config_obj := aoscxgo.FullConfig{
		FileName: filename,
	}

	_, err = config_obj.Create(sw)

	if err != nil {
		diags = append(diags, diag.Errorf("Error in Creating FullConfig\n%v", err)...)
		return diags
	}

	h := fnv.New32()
	h.Write([]byte(config_obj.Config))
	d.SetId(strconv.Itoa(int(h.Sum32())))

	d.Set("filename", filename)
	d.Set("diff", "")
	d.Set("config", config_obj.Config)

	resourceFullConfigRead(ctx, d, m)

	return diags
}

func resourceFullConfigRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error

	sw := m.(*aoscxgo.Client)

	current_config := aoscxgo.FullConfig{}

	filename := d.Get("filename").(string)

	current_config.FileName = filename

	local_config_str, err := current_config.ReadConfigFile(filename)

	err = current_config.Get(sw)

	if err != nil {
		// Failure in Config retrieval
		d.SetId("")
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "Config Not Found",
			Detail:   "Config Not Found",
		})
		return diags
	}

	d.Set("config", current_config.Config)
	d.Set("diff", current_config.CompareConfig(local_config_str))
	d.Set("filename", d.Get("filename").(string))

	return diags
}

func resourceFullConfigUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	var err error

	sw := m.(*aoscxgo.Client)

	current_config := aoscxgo.FullConfig{}

	filename := d.Get("filename").(string)

	current_config.FileName = filename

	err = current_config.Get(sw)

	if err != nil {
		diags = append(diags, diag.Errorf("Error in Retrieving FullConfig %v", err)...)
		return diags
	}

	push_config := false

	if d.HasChange("diff") {
		push_config = true
	}

	if d.HasChange("filename") {
		current_config.FileName = filename
		push_config = true
	}

	if d.HasChange("config") {
		local_config_str, _ := current_config.ReadConfigFile(filename)
		d.Set("diff", current_config.CompareConfig(local_config_str))
		push_config = true
	}

	if push_config {
		config_obj := aoscxgo.FullConfig{
			FileName: filename,
		}

		_, err := config_obj.Create(sw)

		if err != nil {
			diags = append(diags, diag.Errorf("Error in Updating FullConfig %v", err)...)
			return diags
		}
	}

	return resourceFullConfigRead(ctx, d, m)
}

func resourceFullConfigDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {

	d.SetId("")
	d.Set("config", "")
	d.Set("diff", "")
	d.Set("filename", "")
	return nil
}
