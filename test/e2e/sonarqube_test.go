package e2e

import (
	goctx "context"
	"fmt"
	"strings"
	"testing"

	"github.com/jlfowle/sonarqube-operator/pkg/apis"
	operator "github.com/jlfowle/sonarqube-operator/pkg/apis/sonarsource/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"

	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSonarQube(t *testing.T) {
	sonarqubeserverList := &operator.SonarQubeList{}
	err := framework.AddToFrameworkScheme(apis.AddToScheme, sonarqubeserverList)
	if err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}
	// run subtests
	t.Run("sonarqubeserver-group", func(t *testing.T) {
		t.Run("server1", SonarQube)
	})
}

func sonarqubeserverDeployTest(t *testing.T, f *framework.Framework, ctx *framework.Context) error {
	namespace, err := ctx.GetWatchNamespace()
	name := strings.Split(t.Name(), "/")[2]
	if err != nil {
		return fmt.Errorf("could not get namespace: %v", err)
	}
	// create sonarqubeserver custom resource
	sonarQubeServer := &operator.SonarQube{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: operator.SonarQubeSpec{
			Shutdown: &[]bool{true}[0],
		},
	}
	// use TestCtx's create helper to create the object and add a cleanup function for the new object
	err = f.Client.Create(goctx.TODO(), sonarQubeServer, &framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		return err
	}

	// wait for sonarqubeserver to reach 0 replica
	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, name, 0, retryInterval, timeout)
	if err != nil {
		return err
	}

	err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, sonarQubeServer)
	if err != nil {
		return err
	}

	sonarQubeServer.Spec.Shutdown = &[]bool{false}[0]
	err = f.Client.Update(goctx.TODO(), sonarQubeServer)
	if err != nil {
		return err
	}

	// wait for sonarqubeserver to reach 1 replica
	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, name, 1, retryInterval, timeout)
	if err != nil {
		if err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, sonarQubeServer); err != nil {
			return err
		}
		t.Logf("Deployment wait timeout (CR=>%+v)\n", sonarQubeServer.Status)
		deployment := &appsv1.Deployment{}
		if err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, deployment); err != nil {
			return err
		}
		t.Logf("Deployment wait timeout (deployment=>%s)\n", deployment.String())
		replicaSetList := &appsv1.ReplicaSetList{}
		listOpts := []client.ListOption{
			client.InNamespace(namespace),
		}
		if err := f.Client.List(goctx.TODO(), replicaSetList, listOpts...); err != nil {
			return err
		}
		t.Logf("Deployment wait timeout (ReplicaSet=>%s)\n", replicaSetList.String())
		return err
	}

	return wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		err = f.Client.Get(goctx.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, sonarQubeServer)
		if err != nil {
			return false, err
		}

		if sonarQubeServer.Status.Conditions.IsFalseFor(operator.ConditionProgressing) {
			return true, nil
		}
		t.Logf("Waiting for full availability of %s sonarqube server (Progressing=>%s)\n", name,
			sonarQubeServer.Status.Conditions.GetCondition(operator.ConditionProgressing).Message)
		return false, nil
	})
}

func SonarQube(t *testing.T) {
	t.Parallel()
	ctx := framework.NewContext(t)
	defer ctx.Cleanup()
	err := ctx.InitializeClusterResources(&framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		t.Fatalf("failed to initialize cluster resources: %v", err)
	}
	t.Log("Initialized cluster resources")
	namespace, err := ctx.GetWatchNamespace()
	if err != nil {
		t.Fatal(err)
	}
	// get global framework variables
	f := framework.Global
	// wait for sonarqubeserver-operator to be ready
	err = e2eutil.WaitForOperatorDeployment(t, f.KubeClient, namespace, "sonarqube-operator", 1, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	}

	if err = sonarqubeserverDeployTest(t, f, ctx); err != nil {
		t.Fatal(err)
	}
}
