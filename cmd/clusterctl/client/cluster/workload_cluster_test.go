/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cluster

import (
	"testing"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/internal/test"
	"sigs.k8s.io/cluster-api/util/secret"
)

func Test_WorkloadCluster_GetKubeconfig(t *testing.T) {
	var (
		validKubeConfig = `
clusters:
- cluster:
    certificate-authority-data: stuff
    server: https://test-cluster-api:6443
  name: test1
contexts:
- context:
    cluster: test1
    user: test1-admin
  name: test1-admin@test1
current-context: test1-admin@test1
kind: Config
preferences: {}
users:
- name: test1-admin
  user:
    client-certificate-data: stuff-cert-data
    client-key-data: stuff-key-data
`

		validSecret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test1-kubeconfig",
				Namespace: "test",
				Labels:    map[string]string{clusterv1.ClusterLabelName: "test1"},
			},
			Data: map[string][]byte{
				secret.KubeconfigDataName: []byte(validKubeConfig),
			},
		}
	)

	tests := []struct {
		name      string
		expectErr bool
		proxy     Proxy
		want      string
	}{
		{
			name:      "return secret data",
			expectErr: false,
			proxy:     test.NewFakeProxy().WithObjs(validSecret),
			want:      string(validSecret.Data[secret.KubeconfigDataName]),
		},
		{
			name:      "return error if cannot find secret",
			expectErr: true,
			proxy:     test.NewFakeProxy(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			wc := newWorkloadCluster(tt.proxy)
			data, err := wc.GetKubeconfig("test1", "test")

			if tt.expectErr {
				g.Expect(err).To(HaveOccurred())
				return
			}

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(data).To(Equal(tt.want))
		})
	}
}

func Test_WorkloadCluster_GetUserKubeconfig(t *testing.T) {
	var (
		validUserKubeconfig = `
clusters:
- cluster:
    certificate-authority-data: randomstring
    server: https://test-cluster-api.us-east-1.eks.amazonaws.com
  name: test1
contexts:
- context:
    cluster: test1
    user: test1-admin-user
  name: test1-admin@test1
current-context: test1-admin@test1
kind: Config
preferences: {}
users
- name: test1-admin-user
  user:
    client-certificate-data: stuff-cert-data
    client-key-data: stuff-key-data
`
		validUserSecret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test1-user-kubeconfig",
				Namespace: "test",
				Labels:    map[string]string{clusterv1.ClusterLabelName: "test1-user"},
			},
			Data: map[string][]byte{
				secret.KubeconfigDataName: []byte(validUserKubeconfig),
			},
			Type: clusterv1.ClusterSecretType,
		}
	)

	tests := []struct {
		name           string
		expectErr      bool
		proxy          Proxy
		userKubeConfig bool
		want           string
	}{
		{
			name:      "return error if cannot find secret",
			expectErr: true,
			proxy:     test.NewFakeProxy(),
		},
		{
			name:           "return user secret data ",
			expectErr:      false,
			proxy:          test.NewFakeProxy().WithObjs(validUserSecret),
			userKubeConfig: true,
			want:           string(validUserSecret.Data[secret.KubeconfigDataName]),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			wcluster := newWorkloadCluster(tt.proxy)
			data, err := wcluster.GetUserKubeconfig("test1", "test")

			if tt.expectErr {
				g.Expect(err).To(HaveOccurred())
				return
			}

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(data).To(Equal(tt.want))
		})
	}
}
