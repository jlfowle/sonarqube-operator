package sonarqube

import (
	"context"
	"github.com/jlfowle/sonarqube-operator/pkg/api_client"
	sonarsourcev1alpha1 "github.com/jlfowle/sonarqube-operator/pkg/apis/sonarsource/v1alpha1"
	"github.com/jlfowle/sonarqube-operator/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
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

// TestSonarQubeDeployment runs ReconcileSonarQube.ReconcileDeployment() against a
// fake client
func TestSonarQubeDeployment(t *testing.T) {
	// Set the logger to development mode for verbose logs.
	logf.SetLogger(logf.ZapLogger(true))

	var (
		name      = "sonarqube-operator"
		namespace = "sonarqube"
	)

	// A SonarQube resource with metadata and spec.
	sonarqubeList := []*sonarsourcev1alpha1.SonarQube{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: sonarsourcev1alpha1.SonarQubeSpec{},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: sonarsourcev1alpha1.SonarQubeSpec{
				Type: &[]sonarsourcev1alpha1.ServerType{sonarsourcev1alpha1.AIO}[0],
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: sonarsourcev1alpha1.SonarQubeSpec{
				Type: &[]sonarsourcev1alpha1.ServerType{sonarsourcev1alpha1.Application}[0],
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: sonarsourcev1alpha1.SonarQubeSpec{
				Type: &[]sonarsourcev1alpha1.ServerType{sonarsourcev1alpha1.Search}[0],
			},
		},
	}

	for _, sonarqube := range sonarqubeList {
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

		// Take care of dependencies, if there is an unkown error here there is not much to do
		for {
			_, err := r.ReconcileServiceAccount(sonarqube)
			if err != nil && utils.ReasonForError(err) == utils.ErrorReasonUnknown {
				t.Fatalf("reconcileServiceAccount: (%v)", err)
			} else if err == nil {
				break
			}
		}

		for {
			_, err := r.ReconcileSecret(sonarqube)
			if err != nil && utils.ReasonForError(err) == utils.ErrorReasonUnknown {
				t.Fatalf("reconcileServiceAccount: (%v)", err)
			} else if err == nil {
				break
			}
		}

		for {
			_, err := r.ReconcilePVC(sonarqube)
			if err != nil && utils.ReasonForError(err) == utils.ErrorReasonUnknown {
				t.Fatalf("reconcileServiceAccount: (%v)", err)
			} else if err == nil {
				break
			}
		}

		for {
			_, err := r.ReconcileService(sonarqube)
			if err != nil && utils.ReasonForError(err) == utils.ErrorReasonUnknown {
				t.Fatalf("reconcileServiceAccount: (%v)", err)
			} else if err == nil {
				break
			}
		}

		_, err := r.ReconcileDeployment(sonarqube)
		if utils.ReasonForError(err) != utils.ErrorReasonResourceCreate {
			t.Error("reconcileDeployment: resource created error not thrown when creating Deployment")
		}
		deployment := &appsv1.Deployment{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: sonarqube.Name, Namespace: sonarqube.Namespace}, deployment)
		if err != nil && errors.IsNotFound(err) {
			t.Error("reconcileDeployment: Deployment not created")
		} else if err != nil {
			t.Fatalf("reconcileDeployment: (%v)", err)
		}

		deployment.Status.Conditions = append(deployment.Status.Conditions, appsv1.DeploymentCondition{
			Type:   appsv1.DeploymentAvailable,
			Status: corev1.ConditionTrue,
		})
		err = r.client.Status().Update(context.TODO(), deployment)
		if err != nil {
			t.Fatalf("reconcileDeployment: (%v)", err)
		}

		deployment, err = r.ReconcileDeployment(sonarqube)
		if err != nil {
			t.Error("reconcileDeployment: returned error even though Deployment is in expected state")
		}
	}
}
