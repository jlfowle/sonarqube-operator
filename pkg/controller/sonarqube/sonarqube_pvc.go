package sonarqube

import (
	sonarsourcev1alpha1 "github.com/jlfowle/sonarqube-operator/pkg/apis/sonarsource/v1alpha1"
	"github.com/jlfowle/sonarqube-operator/pkg/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// Reconciles map[Volume]*PersistentVolumeClaim for SonarQube
// Returns: map[Volume]*PersistentVolumeClaim, Error
// If Error is non-nil, map[Volume]*PersistentVolumeClaim is not in expected state
// Errors:
//   ErrorReasonResourceCreate: returned when any PersistentVolumeClaim does not exists
//   ErrorReasonResourceUpdate: returned when any PersistentVolumeClaim was updated to meet expected state
//   ErrorReasonUnknown: returned when unhandled error from client occurs
func (r *ReconcileSonarQube) ReconcilePVC(cr *sonarsourcev1alpha1.SonarQube) (*corev1.PersistentVolumeClaim, error) {
	pvc, err := r.findPVC(cr)
	if err != nil {
		return pvc, err
	}

	newStatus := cr.DeepCopy()

	utils.UpdateStatus(r.client, newStatus, cr)
	return pvc, nil
}

func (r *ReconcileSonarQube) findPVC(cr *sonarsourcev1alpha1.SonarQube) (*corev1.PersistentVolumeClaim, error) {
	newPVC, err := r.newPVC(cr)
	if err != nil {
		return newPVC, err
	}

	foundPVC := &corev1.PersistentVolumeClaim{}

	return foundPVC, utils.CreateResourceIfNotFound(r.client, newPVC, foundPVC)
}

func (r *ReconcileSonarQube) newPVC(cr *sonarsourcev1alpha1.SonarQube) (*corev1.PersistentVolumeClaim, error) {
	labels := r.Labels(cr)

	dep := &corev1.PersistentVolumeClaim{
		ObjectMeta: v1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{},
			},
			VolumeMode: &[]corev1.PersistentVolumeMode{corev1.PersistentVolumeFilesystem}[0],
		},
	}

	var storageSize string
	if cr.Spec.NodeConfig.StorageSize == nil {
		storageSize = DefaultVolumeSize
	} else {
		storageSize = *cr.Spec.NodeConfig.StorageSize
	}

	if cr.Spec.NodeConfig.StorageClass != nil {
		dep.Spec.StorageClassName = cr.Spec.NodeConfig.StorageClass
	}

	if size, err := resource.ParseQuantity(storageSize); err != nil {
		return nil, err
	} else {
		dep.Spec.Resources.Requests[corev1.ResourceStorage] = size
	}

	if err := controllerutil.SetControllerReference(cr, dep, r.scheme); err != nil {
		return dep, err
	}

	return dep, nil
}

type Volume string

const (
	DefaultVolumeSize = "1Gi"
)
