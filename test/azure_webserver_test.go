package test

import (
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/azure"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

// You normally want to run this under a separate "Testing" subscription
// For lab purposes you will use your assigned subscription under the Cloud Dev/Ops program tenant
var subscriptionID string = "127e7a44-d802-42e4-b654-a434382666ac"

func TestAzureLinuxVMCreation(t *testing.T) {
	// Configure Terraform options
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		// The path to where our Terraform code is located
		TerraformDir: "../",
		// Override the default Terraform variables
		Vars: map[string]interface{}{
			"labelPrefix": "duan0027",
		},
	})

	// Ensure resources are destroyed after the test
	defer func() {
		retry.DoWithRetry(t, "Destroy resources", 5, 60*time.Second, func() (string, error) {
			_, err := terraform.DestroyE(t, terraformOptions)
			if err != nil {
				return "", err
			}
			return "Destroyed successfully", nil
		})
	}()

	// Run `terraform init` and `terraform apply`. Fail the test if there are any errors.
	terraform.InitAndApply(t, terraformOptions)

	// Run `terraform output` to get the values of output variables
	vmName := terraform.Output(t, terraformOptions, "vm_name")
	resourceGroupName := terraform.Output(t, terraformOptions, "resource_group_name")
	nicName := terraform.Output(t, terraformOptions, "nic_name")

	// Test 1: Confirm VM and resource group exist
	assert.Equal(t, "duan0027A05VM", vmName, "VM name should match")
	assert.Equal(t, "duan0027-A05-RG", resourceGroupName, "Resource group name should match")
	assert.True(t, azure.VirtualMachineExists(t, vmName, resourceGroupName, subscriptionID), "VM should exist")

	// Test 2: Confirm NIC exists and is connected to the VM
	nic, err := azure.GetNetworkInterfaceE(nicName, resourceGroupName, subscriptionID)
	assert.NoError(t, err, "Failed to get NIC")
	assert.NotNil(t, nic, "NIC should exist")
	assert.NotNil(t, nic.VirtualMachine, "NIC should be attached to a VM")
	assert.Contains(t, *nic.VirtualMachine.ID, vmName, "NIC should be attached to the correct VM")

	// Test 3: Confirm VM is running Ubuntu 22.04 LTS
	vm := azure.GetVirtualMachine(t, subscriptionID, resourceGroupName, vmName)
	assert.NotNil(t, vm, "VM should exist")
	assert.Equal(t, "Canonical", *vm.StorageProfile.ImageReference.Publisher, "VM should use Canonical image")
	assert.Equal(t, "0001-com-ubuntu-server-jammy", *vm.StorageProfile.ImageReference.Offer, "VM should use Ubuntu Jammy offer")
	assert.Equal(t, "22_04-lts-gen2", *vm.StorageProfile.ImageReference.Sku, "VM should use Ubuntu 22.04 LTS Gen2")
}
