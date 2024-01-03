package setup

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/canonical/k8s/pkg/k8s/certutils"
	"github.com/canonical/k8s/pkg/k8s/utils"
)

// InitFolders creates the necessary folders for service arguments and certificates.
func InitFolders() error {
	argsDir := filepath.Join(utils.SNAP_DATA, "args")
	err := os.MkdirAll(argsDir, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create arguments directory: %w", err)
	}

	err = os.MkdirAll(certutils.KubePkiPath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to create pki directory: %w", err)
	}

	return nil
}