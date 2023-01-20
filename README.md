# Terraform Provider for AOS-CX

The Terraform Provider for AOS-CX provides a set of configuration management modules and resources specifically designed to manage/configure AOS-CX switches using REST API.


## Requirements

-	[Terraform](https://www.terraform.io/downloads.html) >= 0.13.x
-	[Go](https://golang.org/doc/install) >= 1.18
-  Install AOS-CX Terraform Provider from Terraform Registry  
    -  Terraform 0.13 added support for automatically downloading providers from the terraform registry. Add the following to your terraform project

        ```
        terraform {
          required_providers {
            aoscx = {
              version = "=> 1.0.0"
              source  = "aruba/aoscx"
            }
          }
        }
        ```


## Using the AOS-CX provider

To use the AOS-CX Terraform provider you'll need to define the switch connection details inside a provider block with the following variables:
- `hostname`: IP address of the switch
- `username`: Username used to login to the switch using REST API
- `password`: Password used to login to the switch using REST API
    - see Terraform's documentation on how to [Protect Sensitive Input Variables](https://developer.hashicorp.com/terraform/tutorials/configuration-language/sensitive-variables)
```
provider "aoscx" {
  hostname = "10.6.7.16"
  username = "admin"
  password = "admin"
}
```

Once the provider is defined then you'll define the resources you want Terraform manage on your CX switch. To see all supported resources and their required/optional values see the [/docs](https://github.com/aruba/terraform-provider-aoscx/tree/master/docs) directory.  

Here's an example:  
```
resource "aoscx_vlan" "vlan42" {
        vlan_id = 42
        name = "terraform vlan"
}

resource "aoscx_interface" "int_1_1_14" {
        name = "1/1/14"
        admin_state = "down"
        description = "terraform_uplink"
}

resource "aoscx_l2_interface" "int_1_1_15" {
        interface = "1/1/15"
        admin_state = "up"
        description = "terraform_downlink"
        vlan_mode = "access"
        vlan_tag = 42
}
resource "aoscx_l2_interface" "int_1_1_16" {
        interface = "1/1/16"
        admin_state = "down"
        vlan_mode = "trunk"
        vlan_ids = [20, 42]
        native_vlan_tag = true
}
```
