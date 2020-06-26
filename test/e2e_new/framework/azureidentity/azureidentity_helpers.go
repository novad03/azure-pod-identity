// +build e2e_new

package azureidentity

import (
	"context"
	"fmt"
	"time"

	aadpodv1 "github.com/Azure/aad-pod-identity/pkg/apis/aadpodidentity/v1"
	"github.com/Azure/aad-pod-identity/test/e2e_new/framework"
	"github.com/Azure/aad-pod-identity/test/e2e_new/framework/azure"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	createTimeout = 10 * time.Second
	createPolling = 1 * time.Second

	updateTimeout = 10 * time.Second
	updatePolling = 1 * time.Second

	deleteTimeout = 10 * time.Second
	deletePolling = 1 * time.Second

	resourceIDTemplate = "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.ManagedIdentity/userAssignedIdentities/%s"
)

// CreateInput is the input for Create.
type CreateInput struct {
	Creator      framework.Creator
	Config       *framework.Config
	AzureClient  azure.Client
	Name         string
	Namespace    string
	IdentityName string
	IdentityType aadpodv1.IdentityType
}

// Create creates an AzureIdentity resource.
func Create(input CreateInput) *aadpodv1.AzureIdentity {
	Expect(input.Creator).NotTo(BeNil(), "input.Creator is required for AzureIdentity.Create")
	Expect(input.Config).NotTo(BeNil(), "input.Config is required for AzureIdentity.Create")
	Expect(input.AzureClient).NotTo(BeNil(), "input.AzureClient is required for AzureIdentity.Create")
	Expect(input.Name).NotTo(BeEmpty(), "input.Name is required for AzureIdentity.Create")
	Expect(input.Namespace).NotTo(BeEmpty(), "input.Namespace is required for AzureIdentity.Create")
	Expect(input.IdentityName).NotTo(BeNil(), "input.IdentityName is required for AzureIdentity.Create")
	Expect(input.IdentityType).NotTo(BeNil(), "input.IdentityType is required for AzureIdentity.Create")

	By(fmt.Sprintf("Creating AzureIdentity \"%s\"", input.Name))

	identityClientID := input.AzureClient.GetIdentityClientID(input.IdentityName)
	azureIdentity := &aadpodv1.AzureIdentity{
		ObjectMeta: metav1.ObjectMeta{
			Name:      input.Name,
			Namespace: input.Namespace,
		},
		// TODO: account for SP type
		Spec: aadpodv1.AzureIdentitySpec{
			Type:       input.IdentityType,
			ResourceID: fmt.Sprintf(resourceIDTemplate, input.Config.SubscriptionID, input.Config.IdentityResourceGroup, input.IdentityName),
			ClientID:   identityClientID,
		},
	}

	Eventually(func() error {
		return input.Creator.Create(context.TODO(), azureIdentity)
	}, createTimeout, createPolling).Should(Succeed())

	return azureIdentity
}

// UpdateInput is the input for Update.
type UpdateInput struct {
	Updater             framework.Updater
	Config              *framework.Config
	AzureClient         azure.Client
	AzureIdentity       *aadpodv1.AzureIdentity
	UpdatedIdentityName string
}

// Update updates an AzureIdentity resource.
func Update(input UpdateInput) *aadpodv1.AzureIdentity {
	Expect(input.Updater).NotTo(BeNil(), "input.Updater is required for AzureIdentity.Update")
	Expect(input.Config).NotTo(BeNil(), "input.Config is required for AzureIdentity.Update")
	Expect(input.AzureClient).NotTo(BeNil(), "input.AzureClient is required for AzureIdentity.Update")
	Expect(input.AzureIdentity).NotTo(BeNil(), "input.AzureIdentity is required for AzureIdentity.Update")
	Expect(input.UpdatedIdentityName).NotTo(BeEmpty(), "input.UpdatedIdentityName is required for AzureIdentity.Update")

	By(fmt.Sprintf("Updating AzureIdentity \"%s\" to use \"%s\"", input.AzureIdentity.Name, input.UpdatedIdentityName))

	identityClientID := input.AzureClient.GetIdentityClientID(input.UpdatedIdentityName)
	Expect(identityClientID).NotTo(BeEmpty(), "identityClientID is required for AzureIdentity.Update")

	input.AzureIdentity.Spec.ClientID = identityClientID
	input.AzureIdentity.Spec.ResourceID = fmt.Sprintf(resourceIDTemplate, input.Config.SubscriptionID, input.Config.IdentityResourceGroup, input.UpdatedIdentityName)

	Eventually(func() error {
		return input.Updater.Update(context.TODO(), input.AzureIdentity)
	}, updateTimeout, updatePolling).Should(Succeed())

	return input.AzureIdentity
}

// DeleteInput is the input for Delete.
type DeleteInput struct {
	Deleter       framework.Deleter
	AzureIdentity *aadpodv1.AzureIdentity
}

// Delete deletes an AzureIdentity resource.
func Delete(input DeleteInput) {
	Expect(input.Deleter).NotTo(BeNil(), "input.Deleter is required for AzureIdentity.Delete")
	Expect(input.AzureIdentity).NotTo(BeNil(), "input.AzureIdentity is required for AzureIdentity.Delete")

	By(fmt.Sprintf("Deleting AzureIdentity \"%s\"", input.AzureIdentity.Name))

	Eventually(func() error {
		return input.Deleter.Delete(context.TODO(), input.AzureIdentity)
	}, deleteTimeout, deletePolling).Should(Succeed())
}
