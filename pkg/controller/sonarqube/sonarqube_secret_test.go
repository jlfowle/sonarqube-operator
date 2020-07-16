package sonarqube

import (
	"context"
	"fmt"
	"github.com/jlfowle/sonarqube-operator/pkg/api_client"
	sonarsourcev1alpha1 "github.com/jlfowle/sonarqube-operator/pkg/apis/sonarsource/v1alpha1"
	"github.com/jlfowle/sonarqube-operator/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"strings"
	"testing"
)

// TestSonarQubeSecret runs ReconcileSonarQube.ReconcileSecret() against a
// fake client
func TestSonarQubeSecret(t *testing.T) {
	// Set the logger to development mode for verbose logs.
	logf.SetLogger(logf.ZapLogger(true))

	var (
		name           = "sonarqube-operator"
		namespace      = "sonarqube"
		namespacedName = types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		}
	)

	// A SonarQube resource with metadata and spec.
	sonarqube := &sonarsourcev1alpha1.SonarQube{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: sonarsourcev1alpha1.SonarQubeSpec{},
	}
	// Objects to track in the fake client.
	objs := []runtime.Object{
		sonarqube,
	}

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(sonarsourcev1alpha1.SchemeGroupVersion, sonarqube)
	// Create a fake client to mock API calls.
	cl := fake.NewFakeClientWithScheme(s, objs...)
	// Create a ReconcileSonarQube object with the scheme and fake client.
	apiMock := &api_client.APIClientMock{}
	r := &ReconcileSonarQube{client: cl, scheme: s, apiClient: apiMock}

	_, err := r.ReconcileSecret(sonarqube)
	if utils.ReasonForError(err) != utils.ErrorReasonSpecUpdate {
		t.Error("reconcileSecret: spec update error not thrown when secret not set in spec")
	}
	err = r.client.Get(context.TODO(), namespacedName, sonarqube)
	if err != nil {
		t.Fatalf("reconcileSecret: (%v)", err)
	}
	if sonarqube.Spec.Secret == nil {
		t.Error("reconcileSecret: spec not updated with secret name")
	}

	_, err = r.ReconcileSecret(sonarqube)
	if utils.ReasonForError(err) != utils.ErrorReasonResourceCreate {
		t.Error("reconcileSecret: resource created error not thrown when creating secret")
	}
	secret := &corev1.Secret{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: *sonarqube.Spec.Secret, Namespace: sonarqube.Namespace}, secret)
	if err != nil && errors.IsNotFound(err) {
		t.Error("reconcileSecret: secret not created")
	} else if err != nil {
		t.Fatalf("reconcileSecret: (%v)", err)
	}

	secret, err = r.ReconcileSecret(sonarqube)
	if err != nil {
		t.Error("reconcileSecret: returned error even though secret is in expected state")
	}

}

// TestSonarQubeSecret runs ReconcileSonarQube.ReconcileSecret() against a
// fake client with a secret not owned by SonarQube
func TestSonarQubeSecretUnowned(t *testing.T) {
	// Set the logger to development mode for verbose logs.
	logf.SetLogger(logf.ZapLogger(true))

	var (
		name      = "sonarqube-operator"
		namespace = "sonarqube"
	)

	// A SonarQube resource with metadata and spec.
	sonarqube := &sonarsourcev1alpha1.SonarQube{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: sonarsourcev1alpha1.SonarQubeSpec{
			Secret: &[]string{"test"}[0],
		},
	}
	sonarqube2 := &sonarsourcev1alpha1.SonarQube{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s2", name),
			Namespace: namespace,
		},
		Spec: sonarsourcev1alpha1.SonarQubeSpec{
			Secret: &[]string{"test"}[0],
		},
	}
	// Objects to track in the fake client.
	objs := []runtime.Object{
		sonarqube,
	}

	// Register operator types with the runtime scheme.
	s := scheme.Scheme
	s.AddKnownTypes(sonarsourcev1alpha1.SchemeGroupVersion, sonarqube)
	// Create a fake client to mock API calls.
	cl := fake.NewFakeClientWithScheme(s, objs...)
	// Create a ReconcileSonarQube object with the scheme and fake client.
	apiMock := &api_client.APIClientMock{}
	r := &ReconcileSonarQube{client: cl, scheme: s, apiClient: apiMock}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: sonarqube.Namespace,
			Name:      *sonarqube.Spec.Secret,
		},
	}
	err := r.client.Create(context.TODO(), secret)
	if err != nil {
		t.Fatalf("reconcileSecret: (%v)", err)
	}

	_, err = r.ReconcileSecret(sonarqube)
	if utils.ReasonForError(err) != utils.ErrorReasonResourceUpdate {
		t.Error("reconcileSecret: resource update error not returned")
	}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: secret.Name, Namespace: secret.Namespace}, secret)
	if err != nil {
		t.Fatalf("reconcileSecret: (%v)", err)
	}
	if v, ok := secret.GetAnnotations()[sonarsourcev1alpha1.ServerSecretAnnotation]; !ok || v != sonarqube.Name {
		t.Error("reconcileSecret: secret annotation isn't sonarqube name")
	}

	_, err = r.ReconcileSecret(sonarqube2)
	if utils.ReasonForError(err) != utils.ErrorReasonResourceUpdate {
		t.Error("reconcileSecret: resource update error not returned")
	}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: secret.Name, Namespace: secret.Namespace}, secret)
	if err != nil {
		t.Fatalf("reconcileSecret: (%v)", err)
	}
	if v, ok := secret.GetAnnotations()[sonarsourcev1alpha1.ServerSecretAnnotation]; !ok || !strings.Contains(v, sonarqube.Name) {
		t.Error("reconcileSecret: sonarqube2 name not appended to secret annotation")
	}
}
