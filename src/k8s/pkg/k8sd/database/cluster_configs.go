package database

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"

	"github.com/canonical/microcluster/cluster"
	"gopkg.in/yaml.v2"
)

var (
	clusterConfigsStmts = map[string]int{
		"insert-v1alpha1": mustPrepareStatement("cluster-configs", "insert-v1alpha1.sql"),
		"select-v1alpha1": mustPrepareStatement("cluster-configs", "select-v1alpha1.sql"),
	}
)

type ClusterConfigAPIServer struct {
	SecurePort          int    `yaml:"secure-port,omitempty"`
	AuthorizationMode   string `yaml:"authorization-mode,omitempty"`
	ServiceAccountKey   string `yaml:"service-account-key,omitempty"`
	Datastore           string `yaml:"datastore,omitempty"`
	DatastoreURL        string `yaml:"datastore-url,omitempty"`
	DatastoreCA         string `yaml:"datastore-ca,omitempty"`
	DatastoreClientCert string `yaml:"datastore-client-crt,omitempty"`
	DatastoreClientKey  string `yaml:"datastore-client-key,omitempty"`
}

type ClusterConfigKubelet struct {
	CloudProvider string `yaml:"cloud-provider,omitempty"`
	ClusterDNS    string `yaml:"cluster-dns,omitempty"`
	ClusterDomain string `yaml:"cluster-domain,omitempty"`
}

type ClusterConfigCertificates struct {
	CACert                 string `yaml:"ca-crt,omitempty"`
	CAKey                  string `yaml:"ca-key,omitempty"`
	APIServerToKubeletCert string `yaml:"apiserver-to-kubelet-crt,omitempty"`
	APIServerToKubeletKey  string `yaml:"apiserver-to-kubelet-key,omitempty"`
	K8sDqliteCert          string `yaml:"k8s-dqlite-crt,omitempty"`
	K8sDqliteKey           string `yaml:"k8s-dqlite-key,omitempty"`
	FrontProxyCACert       string `yaml:"front-proxy-ca-crt,omitempty"`
	FrontProxyCAKey        string `yaml:"front-proxy-ca-key,omitempty"`
}

type ClusterConfigCluster struct {
	CIDR string `yaml:"cidr,omitempty"`
}

// ClusterConfig is the control plane configuration format of the k8s cluster.
// ClusterConfig should attempt to use structured fields wherever possible.
type ClusterConfig struct {
	Cluster      ClusterConfigCluster      `yaml:"cluster"`
	Certificates ClusterConfigCertificates `yaml:"certificates"`
	Kubelet      ClusterConfigKubelet      `yaml:"kubelet"`
	APIServer    ClusterConfigAPIServer    `yaml:"apiserver"`
}

// SetClusterConfig updates the cluster configuration with any non-empty values that are set.
// SetClusterConfig will attempt to merge the existing and new configs, and return an error if any protected fields have changed.
func SetClusterConfig(ctx context.Context, tx *sql.Tx, new ClusterConfig) error {
	old, err := GetClusterConfig(ctx, tx)
	if err != nil {
		return fmt.Errorf("failed to fetch existing cluster config: %w", err)
	}
	config, err := mergeConfig(old, new)
	if err != nil {
		return fmt.Errorf("failed to update cluster config: %w", err)
	}

	b, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to encode cluster config: %w", err)
	}
	insertTxStmt, err := cluster.Stmt(tx, clusterConfigsStmts["insert-v1alpha1"])
	if err != nil {
		return fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	if _, err := insertTxStmt.ExecContext(ctx, string(b)); err != nil {
		return fmt.Errorf("failed to insert v1alpha1 config: %w", err)
	}
	return nil
}

// GetClusterConfig retrieves the cluster configuration from the database.
func GetClusterConfig(ctx context.Context, tx *sql.Tx) (ClusterConfig, error) {
	txStmt, err := cluster.Stmt(tx, clusterConfigsStmts["select-v1alpha1"])
	if err != nil {
		return ClusterConfig{}, fmt.Errorf("failed to prepare statement: %w", err)
	}

	var s string
	if err := txStmt.QueryRowContext(ctx).Scan(&s); err != nil {
		if err == sql.ErrNoRows {
			return ClusterConfig{}, nil
		}
		return ClusterConfig{}, fmt.Errorf("failed to retrieve v1alpha1 config: %w", err)
	}

	var clusterConfig ClusterConfig
	if err := yaml.Unmarshal([]byte(s), &clusterConfig); err != nil {
		return ClusterConfig{}, fmt.Errorf("failed to parse v1alpha1 config: %w", err)
	}

	return clusterConfig, nil
}
