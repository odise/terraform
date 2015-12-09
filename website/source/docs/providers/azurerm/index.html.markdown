---
layout: "azurerm"
page_title: "Provider: Azure Resource Manager"
sidebar_current: "docs-azurerm-index"
description: |-
  The Azure Resource Manager provider is used to interact with the many resources supported by Azure, via the ARM API. This supercedes the Azure provider, which interacts with Azure using the Service Management API. The provider needs to be configured with a credentials file, or credentials needed to generate OAuth tokens for the ARM API.
---

# Azure Resource Manager Provider

The Azure Resource Manager provider is used to interact with the many resources
supported by Azure, via the ARM API. This supercedes the Azure provider, which
interacts with Azure using the Service Management API. The provider needs to be
configured with a credentials file, or credentials needed to generate OAuth
tokens for the ARM API.

Use the navigation to the left to read about the available resources.

## Example Usage

```
# Configure the Azure Resource Manager Provider
provider "azurerm" {
    arm_config_file = "${file("~/.azure/credentials.json")}"
}

# Create a resource group
resource "azurerm_resource_group" "production" {
    name = "production"
    location = "West US"
}

# Create a virtual network in the web_servers resource group
resource "azurerm_virtual_network" "network" {
    name = "productionNetwork"
    address_space = ["10.0.0.0/16"]
    location = "West US"
    resource_group_name = "${azurerm_resource_group.production.name}"

    subnet {
        name = "subnet1"
        address_prefix = "10.0.1.0/24"
    }

    subnet {
        name = "subnet2"
        address_prefix = "10.0.2.0/24"
    }
    
    subnet {
        name = "subnet3"
        address_prefix = "10.0.3.0/24"
    }
}

```

## Argument Reference

The following arguments are supported:

* `arm_config_file` - (Optional) The contents of a JASON file containing the
  credentials necessary to generate OAuth tokens for use with the ARM API.
  The requirements are [documented here][armconfig]. If this file is provided
  none of the other authentication options are required.

* `subscription_id` - (Optional) The subscription ID to use. It can also
  be sourced from the `ARM_SUBSCRIPTION_ID` environment variable.

* `client_id` - (Optional) The client ID to use. It can also be sourced from
  the `ARM_CLIENT_ID` environment variable.

* `client_secret` - (Optional) The client secret to use. It can also be sourced from
  the `ARM_CLIENT_SECRET` environment variable.

* `tenant_id` - (Optional) The tenant ID to use. It can also be sourced from the
  `ARM_TENANT_ID` environment variable.

## Testing:

The following environment variables must be set for the running of the
acceptance test suite:

* A valid combination of the above which are required for authentication.

* `AZURE_STORAGE` - The name of a storage account to be used in tests which
  require a storage backend. The storage account needs to be located in
  the Western US Azure region.

[armconfig]: https://msdn.microsoft.com/en-us/library/azure/dn790557.aspx#bk_portal
