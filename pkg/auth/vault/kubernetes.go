package vault

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/argoproj-labs/argocd-vault-plugin/pkg/utils"
	"github.com/hashicorp/vault/api"
)

const (
	defaultMountPath   = "auth/kubernetes"
	serviceAccountFile = "/var/run/secrets/kubernetes.io/serviceaccount/token"
)

// K8sAuth TODO
type K8sAuth struct {
	// Optional, will use default path of auth/kubernetes if left blank
	MountPath string

	// Optional, will use default service account if left blank
	TokenPath string

	Role string
}

// NewK8sAuth initializes and returns a K8sAuth Struct
func NewK8sAuth(role, mountPath, tokenPath string) *K8sAuth {
	k8sAuth := &K8sAuth{
		Role:      role,
		MountPath: mountPath,
		TokenPath: tokenPath,
	}

	return k8sAuth
}

// Authenticate authenticates with Vault via K8s and returns a token
func (k *K8sAuth) Authenticate(vaultClient *api.Client) error {
	token, err := k.getJWT()
	if err != nil {
		return err
	}

	payload := map[string]interface{}{
		"role": k.Role,
		"jwt":  token,
	}

	kubeAuthPath := defaultMountPath
	if k.MountPath != "" {
		kubeAuthPath = k.MountPath
	}

	data, err := vaultClient.Logical().Write(fmt.Sprintf("%s/login", kubeAuthPath), payload)
	if err != nil {
		return err
	}

	// If we cannot write the Vault token, we'll just have to login next time. Nothing showstopping.
	err = utils.SetToken(vaultClient, data.Auth.ClientToken)
	if err != nil {
		print(err)
	}

	return nil
}

func (k *K8sAuth) getJWT() (string, error) {
	tokenFilePath := serviceAccountFile
	if k.TokenPath != "" {
		tokenFilePath = k.TokenPath
	}

	f, err := os.Open(tokenFilePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	contentBytes, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(contentBytes)), nil
}
