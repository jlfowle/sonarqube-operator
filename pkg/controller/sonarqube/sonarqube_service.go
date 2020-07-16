package sonarqube

import (
	sonarsourcev1alpha1 "github.com/jlfowle/sonarqube-operator/pkg/apis/sonarsource/v1alpha1"
	"github.com/jlfowle/sonarqube-operator/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// Reconciles Service for SonarQube
// Returns: Service, Error
// If Error is non-nil, Service is not in expected state
// Errors:
//   ErrorReasonResourceCreate: returned when Service does not exists
//   ErrorReasonResourceUpdate: returned when Service was updated to meet expected state
//   ErrorReasonUnknown: returned when unhandled error from client occurs
func (r *ReconcileSonarQube) ReconcileService(cr *sonarsourcev1alpha1.SonarQube) (*corev1.Service, error) {
	service, err := r.findService(cr)
	if err != nil {
		return service, err
	}

	newStatus := cr.DeepCopy()

	newStatus.Status.Service = service.Name

	utils.UpdateStatus(r.client, newStatus, cr)

	newService, err := r.newService(cr)
	if err != nil {
		return service, err
	}
	if err := utils.VerifyService(r.client, service, newService); err != nil {
		return service, err
	}

	return service, nil
}

func (r *ReconcileSonarQube) findService(cr *sonarsourcev1alpha1.SonarQube) (*corev1.Service, error) {
	newService, err := r.newService(cr)
	if err != nil {
		return newService, err
	}

	foundService := &corev1.Service{}

	return foundService, utils.CreateResourceIfNotFound(r.client, newService, foundService)
}

func (r *ReconcileSonarQube) newService(cr *sonarsourcev1alpha1.SonarQube) (*corev1.Service, error) {
	labels := r.Labels(cr)

	var nodeType sonarsourcev1alpha1.ServerType
	if cr.Spec.Type == nil {
		nodeType = sonarsourcev1alpha1.AIO
	} else {
		nodeType = *cr.Spec.Type
	}

	dep := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: cr.Namespace,
			Name:      cr.Name,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Type:     corev1.ServiceTypeClusterIP,
			Ports:    utils.ServicePorts(nodeType),
		},
	}

	if err := controllerutil.SetControllerReference(cr, dep, r.scheme); err != nil {
		return dep, err
	}

	return dep, nil
}
