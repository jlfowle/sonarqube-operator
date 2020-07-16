package sonarqube

import (
	"context"
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
	"testing"
)

// TestSonarQubeServiceAccount runs ReconcileSonarQube.ReconcileAppServiceAccount() against a
// fake client
func TestSonarQubeServiceAccount(t *testing.T) {
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

	_, err := r.ReconcileServiceAccount(sonarqube)
	if utils.ReasonForError(err) != utils.ErrorReasonResourceCreate {
		t.Error("reconcileServiceAccount: resource created error not thrown when creating ServiceAccount")
	}
	ServiceAccount := &corev1.ServiceAccount{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: sonarqube.Name, Namespace: sonarqube.Namespace}, ServiceAccount)
	if err != nil && errors.IsNotFound(err) {
		t.Error("reconcileServiceAccount: ServiceAccount not created")
	} else if err != nil {
		t.Fatalf("reconcileServiceAccount: (%v)", err)
	}

	ServiceAccount, err = r.ReconcileServiceAccount(sonarqube)
	if err != nil {
		t.Error("reconcileServiceAccount: returned error even though ServiceAccount is in expected state")
	}
}
