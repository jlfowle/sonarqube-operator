package sonarqube

import (
	"fmt"
	sonarsourcev1alpha1 "github.com/jlfowle/sonarqube-operator/pkg/apis/sonarsource/v1alpha1"
	"github.com/jlfowle/sonarqube-operator/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"strings"
)

// Reconciles Secret for SonarQube
// Returns: Secret, Error
// If Error is non-nil, Service is not in expected state
// Errors:
//   ErrorReasonSpecUpdate: returned when spec does not have secret name
//   ErrorReasonResourceCreate: returned when secret does not exists
//   ErrorReasonResourceUpdate: returned when secret was updated to meet expected state
//   ErrorReasonUnknown: returned when unhandled error from client occurs
func (r *ReconcileSonarQube) ReconcileSecret(cr *sonarsourcev1alpha1.SonarQube) (*corev1.Secret, error) {
	foundSecret, err := r.findSecret(cr)
	if err != nil {
		return foundSecret, err
	}

	if !utils.IsOwner(cr, foundSecret) {
		annotations := foundSecret.GetAnnotations()
		if val, ok := annotations[sonarsourcev1alpha1.ServerSecretAnnotation]; ok && !strings.Contains(val, cr.Name) {
			annotations[sonarsourcev1alpha1.ServerSecretAnnotation] = fmt.Sprintf("%s,%s", val, cr.Name)
			foundSecret.SetAnnotations(annotations)
			return foundSecret, utils.UpdateResource(r.client, foundSecret, utils.ErrorReasonResourceUpdate, "updated secret annotation")
		} else if !ok {
			if annotations == nil {
				annotations = make(map[string]string)
			}
			annotations[sonarsourcev1alpha1.ServerSecretAnnotation] = cr.Name
			foundSecret.SetAnnotations(annotations)
			return foundSecret, utils.UpdateResource(r.client, foundSecret, utils.ErrorReasonResourceUpdate, "updated secret annotation")
		}
	}

	err = r.verifySecret(cr, foundSecret)
	if err != nil {
		return foundSecret, nil
	}

	return foundSecret, nil
}

func (r *ReconcileSonarQube) findSecret(cr *sonarsourcev1alpha1.SonarQube) (*corev1.Secret, error) {
	newSecret, err := r.newSecret(cr)
	if err != nil {
		return newSecret, err
	}

	foundSecret := &corev1.Secret{}

	return foundSecret, utils.CreateResourceIfNotFound(r.client, newSecret, foundSecret)
}

func (r *ReconcileSonarQube) newSecret(cr *sonarsourcev1alpha1.SonarQube) (*corev1.Secret, error) {
	labels := r.Labels(cr)

	dep := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		StringData: map[string]string{
			"sonar.properties": "",
			"wrapper.conf":     "",
		},
		Type: corev1.SecretTypeOpaque,
	}

	if cr.Spec.Secret == nil {
		cr.Spec.Secret = &[]string{fmt.Sprintf("%s-config", cr.Name)}[0]
		return dep, utils.UpdateResource(r.client, cr, utils.ErrorReasonSpecUpdate, "updated secret")
	}

	dep.Name = *cr.Spec.Secret

	if err := controllerutil.SetControllerReference(cr, dep, r.scheme); err != nil {
		return dep, err
	}

	return dep, nil
}

func (r *ReconcileSonarQube) verifySecret(cr *sonarsourcev1alpha1.SonarQube, _ *corev1.Secret) error {
	/*var sonarProperties *properties.Properties
	var sonarPropertiesExists bool
	if v, ok := s.Data["sonar.properties"]; ok {
		sonarPropertiesExists = ok
		sonarProperties, _ = properties.Load(v, properties.UTF8)
	}*/

	var nodeType sonarsourcev1alpha1.ServerType
	if cr.Spec.Type == nil {
		nodeType = sonarsourcev1alpha1.AIO
	} else {
		nodeType = *cr.Spec.Type
	}

	switch nodeType {
	case sonarsourcev1alpha1.Application, sonarsourcev1alpha1.Search:
	}

	return nil
}
