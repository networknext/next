# --------------------------------------------------------------------------

terraform {
  required_providers {
    hivelocity = {
      version  = "0.5.0"
      source   = "hivelocity/hivelocity"
    }
  }
}

// Find a plan with 16GB of memory in Tampa.
data "hivelocity_product" "tampa_product" {
  first = true

  filter {
    name   = "product_memory"
    values = ["16GB"]
  }

  filter {
    name   = "data_center"
    values = ["TPA1"]
  }

  filter {
    name   = "stock"
    values = ["limited", "available"]
  }
}

// Create a SSH Key
resource "hivelocity_ssh_key" "my_ssh_key" {
  name       = "This is my Terraform SSH Key"
  public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDYYGqXPIJ1yQ7pMTSmdB+LrjvnPh2jSJewvinxYwuk/CBvbsUgFxOTBtjzfU0kHIlo5QyIhJ7tjh8PhT8LdJPGB86ItwTg3Lmt1q5UFxbHmZ0kPmmzaDI/9aakOal3P93D14HDCzBnkTHfC8/JZ5JpDxp86XM+gWQ9sFMkLx83ZOwONNu3E8PowTEpsp0jMx2B2aFeSM+T4bLkJJQtA5Cp6lgRAc5AklXaqmpdAil/fIL/+gvRf8kJIccAe/2oVj4flaMK7mgZ39qhzYYUjTEokEYvJf17QbdtFTxxtIQ+hTxzKzwT6p8cMu6DNQLfq6oxzbuBVGTvKD79MR5vjx+RNpaIru8wzIspHTez9eGDzdR0316GWDcxmMwVIDMM+3pjopDJV6DILfs6hVlAuH11yCX8YwwGHYpsdzLLd00FEEaGLGVRDr/hvduZ1caQIvdNln6Gr7k6W51U1VTC3NRq49yoxYSsXxn30PfTe2IxFaZyhQXHunCLaMCF+TrAOc0= someone@somewhere.com"
}

// Provision your devices with CentOS 7.
resource "hivelocity_bare_metal_device" "webserver" {
  product_id        = data.hivelocity_product.tampa_product.product_id
  os_name           = "CentOS 7.x"
  location_name     = data.hivelocity_product.tampa_product.data_center
  hostname          = "webserver.terraform.test"
  tags              = ["hello", "world"]
  script            = file("${path.module}/cloud_init_example.yaml")
  public_ssh_key_id = hivelocity_ssh_key.my_ssh_key.ssh_key_id
}

resource "hivelocity_bare_metal_device" "database" {
  product_id        = data.hivelocity_product.tampa_product.product_id
  os_name           = "CentOS 7.x"
  location_name     = data.hivelocity_product.tampa_product.data_center
  hostname          = "database.terraform.test"
  tags              = ["hello", "world"]
  script            = file("${path.module}/cloud_init_example.yaml")
  public_ssh_key_id = hivelocity_ssh_key.my_ssh_key.ssh_key_id
}

// Create a VLAN connecting servers
resource "hivelocity_vlan" "private_vlan" {
  device_ids    = [
      hivelocity_bare_metal_device.webserver.device_id,
      hivelocity_bare_metal_device.database.device_id,
  ]
}

# --------------------------------------------------------------------------
