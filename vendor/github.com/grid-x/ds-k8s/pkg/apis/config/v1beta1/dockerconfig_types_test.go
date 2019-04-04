/*
 * Author: Joel Hermanns <j.hermanns@gridx.ai>
 */

package v1beta1

import (
	"testing"

	"github.com/onsi/gomega"
	"golang.org/x/net/context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	appsv1beta1 "github.com/grid-x/ds-k8s/pkg/apis/apps/v1beta1"
)

func TestStorageDockerConfig(t *testing.T) {
	key := types.NamespacedName{
		Name:      "foo",
		Namespace: "default",
	}
	created := &DockerConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: "default",
		},
		Spec: DockerConfigSpec{
			Registry: "https://630781358184.dkr.ecr.eu-central-1.amazonaws.com",
			Selector: appsv1beta1.Selector{
				MatchByLabels: map[string]string{
					"foo": "bar",
				},
			},
			Credentials: DockerConfigCredentails{
				AWS: &AWSCredentialProvider{
					AccessKey:       "AKIAJYFSS3HBFTQQFGIK",
					SecretAccessKey: "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",
					Region:          "eu-central-1",
				},
			},
		}}
	g := gomega.NewGomegaWithT(t)

	// Test Create
	fetched := &DockerConfig{}
	g.Expect(c.Create(context.TODO(), created)).NotTo(gomega.HaveOccurred())

	g.Expect(c.Get(context.TODO(), key, fetched)).NotTo(gomega.HaveOccurred())
	g.Expect(fetched).To(gomega.Equal(created))

	// Test Updating the Labels
	updated := fetched.DeepCopy()
	updated.Labels = map[string]string{"hello": "world"}
	g.Expect(c.Update(context.TODO(), updated)).NotTo(gomega.HaveOccurred())

	g.Expect(c.Get(context.TODO(), key, fetched)).NotTo(gomega.HaveOccurred())
	g.Expect(fetched).To(gomega.Equal(updated))

	// Test Delete
	g.Expect(c.Delete(context.TODO(), fetched)).NotTo(gomega.HaveOccurred())
	g.Expect(c.Get(context.TODO(), key, fetched)).To(gomega.HaveOccurred())
}
