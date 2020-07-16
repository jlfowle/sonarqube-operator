package sonarqube

import (
	"context"
	"github.com/jlfowle/sonarqube-operator/pkg/api_client"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"testing"

	sonarsourcev1alpha1 "github.com/jlfowle/sonarqube-operator/pkg/apis/sonarsource/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

const (
	ReconcileErrorFormat string = "reconcile: (%v)"
)

// TestSonarQubeController runs ReconcileSonarQube.Reconcile() against a
// fake client that tracks a SonarQube object.
func TestSonarQubeController(t *testing.T) {
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
	apiMock.InfoOutput = &api_client.Status{
		Version: api_client.SystemVersion{
			Major: 8,
			Minor: 3,
		},
	}

	// Mock request to simulate Reconcile() being called on an event for a
	// watched resource .
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}

	res, err := r.Reconcile(req)
	if err != nil {
		t.Fatalf(ReconcileErrorFormat, err)
	}
	// Check the result of reconciliation to make sure it has the desired state.
	if !res.Requeue {
		t.Error("reconcile did not requeue")
	}

	res, err = r.Reconcile(req)
	if err != nil {
		t.Fatalf(ReconcileErrorFormat, err)
	}
	// Check the result of reconciliation to make sure it has the desired state.
	if !res.Requeue {
		t.Error("reconcile did not requeue")
	}
	err = r.client.Get(context.TODO(), req.NamespacedName, sonarqube)
	if err != nil {
		t.Fatalf(ReconcileErrorFormat, err)
	}
	if !sonarqube.Status.Conditions.IsTrueFor(sonarsourcev1alpha1.ConditionProgressing) {
		t.Errorf("condition progressing not set")
	}
	secret := &corev1.Secret{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: *sonarqube.Spec.Secret, Namespace: sonarqube.Namespace}, secret)
	if err != nil && errors.IsNotFound(err) {
		t.Error("reconcile: secret not created")
	} else if err != nil {
		t.Fatalf(ReconcileErrorFormat, err)
	}

	res, err = r.Reconcile(req)
	if err != nil {
		t.Fatalf(ReconcileErrorFormat, err)
	}
	// Check the result of reconciliation to make sure it has the desired state.
	if !res.Requeue {
		t.Error("reconcile did not requeue")
	}
	err = r.client.Get(context.TODO(), req.NamespacedName, sonarqube)
	if err != nil {
		t.Fatalf(ReconcileErrorFormat, err)
	}
	if !sonarqube.Status.Conditions.IsTrueFor(sonarsourcev1alpha1.ConditionProgressing) {
		t.Errorf("condition progressing not set")
	}
	serviceAccount := &corev1.ServiceAccount{}
	err = r.client.Get(context.TODO(), req.NamespacedName, serviceAccount)
	if err != nil && errors.IsNotFound(err) {
		t.Error("reconcile: service account not created")
	} else if err != nil {
		t.Fatalf(ReconcileErrorFormat, err)
	}

	res, err = r.Reconcile(req)
	if err != nil {
		t.Fatalf(ReconcileErrorFormat, err)
	}
	// Check the result of reconciliation to make sure it has the desired state.
	if !res.Requeue {
		t.Error("reconcile did not requeue")
	}
	err = r.client.Get(context.TODO(), req.NamespacedName, sonarqube)
	if err != nil {
		t.Fatalf(ReconcileErrorFormat, err)
	}
	if !sonarqube.Status.Conditions.IsTrueFor(sonarsourcev1alpha1.ConditionProgressing) {
		t.Errorf("condition progressing not set")
	}
	service := &corev1.Service{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: sonarqube.Name, Namespace: sonarqube.Namespace}, service)
	if err != nil && errors.IsNotFound(err) {
		t.Error("reconcile: service not created")
	} else if err != nil {
		t.Fatalf(ReconcileErrorFormat, err)
	}

	res, err = r.Reconcile(req)
	if err != nil {
		t.Fatalf(ReconcileErrorFormat, err)
	}
	// Check the result of reconciliation to make sure it has the desired state.
	if !res.Requeue {
		t.Error("reconcile did not requeue")
	}
	err = r.client.Get(context.TODO(), req.NamespacedName, sonarqube)
	if err != nil {
		t.Fatalf(ReconcileErrorFormat, err)
	}
	if !sonarqube.Status.Conditions.IsTrueFor(sonarsourcev1alpha1.ConditionProgressing) {
		t.Errorf("condition progressing not set")
	}
	dataPVC := &corev1.PersistentVolumeClaim{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: sonarqube.Name, Namespace: namespace}, dataPVC)
	if err != nil && errors.IsNotFound(err) {
		t.Error("reconcile: data pvc not created")
	} else if err != nil {
		t.Fatalf(ReconcileErrorFormat, err)
	}

	res, err = r.Reconcile(req)
	if err != nil {
		t.Fatalf(ReconcileErrorFormat, err)
	}
	// Check the result of reconciliation to make sure it has the desired state.
	if !res.Requeue {
		t.Error("reconcile did not requeue")
	}
	err = r.client.Get(context.TODO(), req.NamespacedName, sonarqube)
	if err != nil {
		t.Fatalf(ReconcileErrorFormat, err)
	}
	if !sonarqube.Status.Conditions.IsTrueFor(sonarsourcev1alpha1.ConditionProgressing) {
		t.Errorf("condition progressing not set")
	}
	deployment := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: sonarqube.Name, Namespace: namespace}, deployment)
	if err != nil && errors.IsNotFound(err) {
		t.Error("reconcile: stateful set not created")
	} else if err != nil {
		t.Fatalf(ReconcileErrorFormat, err)
	}

	deployment.Status.Conditions = append(deployment.Status.Conditions, appsv1.DeploymentCondition{
		Type:   appsv1.DeploymentAvailable,
		Status: corev1.ConditionTrue,
	})
	err = r.client.Status().Update(context.TODO(), deployment)
	if err != nil {
		t.Fatalf("reconcileDeployment: (%v)", err)
	}

	res, err = r.Reconcile(req)
	if err != nil {
		t.Fatalf(ReconcileErrorFormat, err)
	}
	// Check the result of reconciliation to make sure it has the desired state.
	if !res.Requeue {
		t.Error("reconcile did not requeue to set version")
	}
	err = r.client.Get(context.TODO(), req.NamespacedName, sonarqube)
	if err != nil {
		t.Fatalf(ReconcileErrorFormat, err)
	}
	if sonarqube.Spec.Version == nil {
		t.Error("sonarqube version not set")
	}

	apiMock.UpgradesOutput = &api_client.Upgrades{
		Upgrades:            []api_client.Upgrade{},
		UpdateCenterRefresh: "",
	}

	res, err = r.Reconcile(req)
	if err != nil {
		t.Fatalf(ReconcileErrorFormat, err)
	}
	// Check the result of reconciliation to make sure it has the desired state.
	if res.Requeue {
		t.Error("reconcile requeued even though everything should be good")
	}
}
