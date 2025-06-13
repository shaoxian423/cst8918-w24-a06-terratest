package test

import (
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/azure"
	"github.com/gruntwork-io/terratest/modules/retry"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
)

// Define the Azure subscription ID, ideally use a dedicated testing subscription
var subscriptionID string = "127e7a44-d802-42e4-b654-a434382666ac"

// TestAzureLinuxVMCreation tests the creation and attributes of an Azure Linux Virtual Machine
func TestAzureLinuxVMCreation(t *testing.T) {
	// Enable parallel testing to improve execution speed
	t.Parallel()

	// Configure Terraform options, specifying the Terraform code path and variables
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: "../", // Directory containing Terraform code
		Vars: map[string]interface{}{
			"labelPrefix": "duan0027", // Set Terraform variable for resource naming
		},
	})

	// Ensure resources are cleaned up after the test to avoid leftover resources
	defer func() {
		// Retry resource destruction up to 5 times with 60-second intervals
		retry.DoWithRetry(t, "Destroy resources", 5, 60*time.Second, func() (string, error) {
			_, err := terraform.DestroyE(t, terraformOptions)
			if err != nil {
				return "", err
			}
			return "Resources destroyed successfully", nil
		})
	}()

	// Run `terraform init` and `terraform apply`, fail the test if errors occur
	terraform.InitAndApply(t, terraformOptions)

	// Retrieve VM name, resource group name, and NIC name from Terraform outputs
	vmName := terraform.OutputRequired(t, terraformOptions, "vm_name")
	resourceGroupName := terraform.OutputRequired(t, terraformOptions, "resource_group_name")
	nicName := terraform.OutputRequired(t, terraformOptions, "nic_name")

	// Test Case 1: Verify VM and resource group existence
	t.Run("Verify VM and Resource Group", func(t *testing.T) {
		// Check if the VM name matches the expected value
		assert.Equal(t, "duan0027A05VM", vmName, "VM name should be duan0027A05VM")
		// Check if the resource group name matches the expected value
		assert.Equal(t, "duan0027-A05-RG", resourceGroupName, "Resource group name should be duan0027-A05-RG")
		// Use the Azure package's VirtualMachineExists function to confirm VM existence
		assert.True(t, azure.VirtualMachineExists(t, vmName, resourceGroupName, subscriptionID), "VM should exist")
	})

	// Test Case 2: Verify network interface (NIC) existence and association with the VM
	t.Run("Verify Network Interface", func(t *testing.T) {
		// Use the Azure package's GetVirtualMachineNics function to get the VM's network interfaces
		nicList := azure.GetVirtualMachineNics(t, vmName, resourceGroupName, subscriptionID)
		// Verify that the expected NIC name is in the list of VM network interfaces
		assert.Contains(t, nicList, nicName, "VM's network interface list should contain %s", nicName)
	})

	// Test Case 3: Verify the VM is running Ubuntu 22.04 LTS
	t.Run("Verify VM Image", func(t *testing.T) {
		// Use the Azure package's GetVirtualMachineImageE function to retrieve VM image details
		vmImage, err := azure.GetVirtualMachineImageE(vmName, resourceGroupName, subscriptionID)
		assert.NoError(t, err, "Should not encounter error when fetching VM image")
		// Verify the image publisher is Canonical
		assert.Equal(t, "Canonical", vmImage.Publisher, "VM image publisher should be Canonical")
		// Verify the image offer is Ubuntu Jammy
		assert.Equal(t, "0001-com-ubuntu-server-jammy", vmImage.Offer, "VM image offer should be 0001-com-ubuntu-server-jammy")
		// Verify the image SKU is Ubuntu 22.04 LTS Gen2
		assert.Equal(t, "22_04-lts-gen2", vmImage.SKU, "VM image SKU should be 22_04-lts-gen2")
	})
}
