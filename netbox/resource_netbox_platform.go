package netbox

import (
	"strconv"

	"github.com/fbreckle/go-netbox/netbox/client"
	"github.com/fbreckle/go-netbox/netbox/client/dcim"
	"github.com/fbreckle/go-netbox/netbox/models"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceNetboxPlatform() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetboxPlatformCreate,
		Read:   resourceNetboxPlatformRead,
		Update: resourceNetboxPlatformUpdate,
		Delete: resourceNetboxPlatformDelete,

		Description: `:meta:subcategory:Data Center Inventory Management (DCIM):From the [official documentation](https://docs.netbox.dev/en/stable/features/devices/#platforms):

> A platform defines the type of software running on a device or virtual machine. This can be helpful to model when it is necessary to distinguish between different versions or feature sets. Note that two devices of the same type may be assigned different platforms: For example, one Juniper MX240 might run Junos 14 while another runs Junos 15.`,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"slug": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(0, 30),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 200),
			},
			"manufacturer_id": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"napalm_driver": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringLenBetween(0, 50),
			},
			"napalm_args": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringIsJSON,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceNetboxPlatformCreate(d *schema.ResourceData, m interface{}) error {
	api := m.(*client.NetBoxAPI)

	name := d.Get("name").(string)
	description := d.Get("description").(string)
	manufacturer_id := d.Get("manufacturer_id").(int64)
	napalm_driver := d.Get("napalm_driver").(string)
	napalm_args := d.Get("napalm_args").(string)

	slugValue, slugOk := d.GetOk("slug")
	var slug string
	// Default slug to generated slug if not given
	if !slugOk {
		slug = getSlug(name)
	} else {
		slug = slugValue.(string)
	}

	params := dcim.NewDcimPlatformsCreateParams().WithData(
		&models.WritablePlatform{
			Name:         &name,
			Slug:         &slug,
			Description:  description,
			Manufacturer: &manufacturer_id,
			NapalmDriver: napalm_driver,
			NapalmArgs:   napalm_args,
			Tags:         []*models.NestedTag{},
		},
	)

	res, err := api.Dcim.DcimPlatformsCreate(params, nil)
	if err != nil {
		//return errors.New(getTextFromError(err))
		return err
	}

	d.SetId(strconv.FormatInt(res.GetPayload().ID, 10))

	return resourceNetboxPlatformRead(d, m)
}

func resourceNetboxPlatformRead(d *schema.ResourceData, m interface{}) error {
	api := m.(*client.NetBoxAPI)
	id, _ := strconv.ParseInt(d.Id(), 10, 64)
	params := dcim.NewDcimPlatformsReadParams().WithID(id)

	res, err := api.Dcim.DcimPlatformsRead(params, nil)

	if err != nil {
		if errresp, ok := err.(*dcim.DcimPlatformsReadDefault); ok {
			errorcode := errresp.Code()
			if errorcode == 404 {
				// If the ID is updated to blank, this tells Terraform the resource no longer exists (maybe it was destroyed out of band). Just like the destroy callback, the Read function should gracefully handle this case. https://www.terraform.io/docs/extend/writing-custom-providers.html
				d.SetId("")
				return nil
			}
		}
		return err
	}

	d.Set("name", res.GetPayload().Name)
	d.Set("slug", res.GetPayload().Slug)
	d.Set("description", res.GetPayload().Description)
	d.Set("manufacturer_id", res.GetPayload().Manufacturer)
	d.Set("napalm_driver", res.GetPayload().NapalmDriver)
	d.Set("napalm_args", res.GetPayload().NapalmArgs)
	return nil
}

func resourceNetboxPlatformUpdate(d *schema.ResourceData, m interface{}) error {
	api := m.(*client.NetBoxAPI)

	id, _ := strconv.ParseInt(d.Id(), 10, 64)
	data := models.WritablePlatform{}

	name := d.Get("name").(string)
	slugValue, slugOk := d.GetOk("slug")
	var slug string
	// Default slug to generated slug if not given
	if !slugOk {
		slug = getSlug(name)
	} else {
		slug = slugValue.(string)
	}

	if d.HasChange("name") {
		data.Name = &name
	}
	if d.HasChange("slug") {
		data.Slug = &slug
	}
	if d.HasChange("description") {
		data.Description = d.Get("description").(string)
	}
	if d.HasChange("manufacturer_id") {
		manufacturer_id := d.Get("manufacturer_id").(int64)
		data.Manufacturer = &manufacturer_id
	}
	if d.HasChange("description") {
		data.NapalmDriver = d.Get("napalm_driver").(string)
	}
	if d.HasChange("description") {
		data.NapalmArgs = d.Get("napalm_args").(string)
	}

	data.Tags = []*models.NestedTag{} // Why???

	params := dcim.NewDcimPlatformsPartialUpdateParams().WithID(id).WithData(&data)

	_, err := api.Dcim.DcimPlatformsPartialUpdate(params, nil)
	if err != nil {
		return err
	}

	return resourceNetboxPlatformRead(d, m)
}

func resourceNetboxPlatformDelete(d *schema.ResourceData, m interface{}) error {
	api := m.(*client.NetBoxAPI)

	id, _ := strconv.ParseInt(d.Id(), 10, 64)
	params := dcim.NewDcimPlatformsDeleteParams().WithID(id)

	_, err := api.Dcim.DcimPlatformsDelete(params, nil)
	if err != nil {
		if errresp, ok := err.(*dcim.DcimPlatformsDeleteDefault); ok {
			if errresp.Code() == 404 {
				d.SetId("")
				return nil
			}
		}
		return err
	}
	return nil
}
