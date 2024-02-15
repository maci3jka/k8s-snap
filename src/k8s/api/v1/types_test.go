package v1

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestBootstrapConfigFromMap(t *testing.T) {
	g := NewWithT(t)
	// Create a new BootstrapConfig with default values
	bc := &BootstrapConfig{
		Components:  []string{"dns", "network", "storage"},
		ClusterCIDR: "10.1.0.0/16",
	}

	// Convert the BootstrapConfig to a map
	m, err := bc.ToMap()
	g.Expect(err).To(BeNil())

	// Unmarshal the YAML string from the map into a new BootstrapConfig instance
	bcyaml, err := BootstrapConfigFromMap(m)

	// Check for errors
	g.Expect(err).To(BeNil())
	// Compare the unmarshaled BootstrapConfig with the original one
	g.Expect(bcyaml).To(Equal(bc)) // Note the *bc here to compare values, not pointers

}