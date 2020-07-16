package sonarqube

import (
	"fmt"
	sonarsourcev1alpha1 "github.com/jlfowle/sonarqube-operator/pkg/apis/sonarsource/v1alpha1"
	"github.com/jlfowle/sonarqube-operator/version"
)

func (r *ReconcileSonarQube) Labels(cr *sonarsourcev1alpha1.SonarQube) map[string]string {
	labels := make(map[string]string)

	for k, v := range cr.Labels {
		labels[k] = v
	}

	labels[sonarsourcev1alpha1.ServerTypeLabel] = cr.Name
	labels[sonarsourcev1alpha1.KubeAppName] = "SonarQube"
	labels[sonarsourcev1alpha1.KubeAppInstance] = cr.Name
	labels[sonarsourcev1alpha1.KubeAppVersion] = cr.Status.Revision
	labels[sonarsourcev1alpha1.KubeAppManagedby] = fmt.Sprintf("sonarqube-operator.v%s", version.Version)

	if cr.Spec.Type != nil {
		labels[sonarsourcev1alpha1.KubeAppComponent] = string(*cr.Spec.Type)
	} else {
		labels[sonarsourcev1alpha1.KubeAppComponent] = string(sonarsourcev1alpha1.AIO)
	}

	return labels
}

func (r *ReconcileSonarQube) PodLabels(cr *sonarsourcev1alpha1.SonarQube) map[string]string {
	labels := r.Labels(cr)
	podLabels := make(map[string]string)
	podLabels[sonarsourcev1alpha1.ServerTypeLabel] = labels[sonarsourcev1alpha1.ServerTypeLabel]
	podLabels["Deployment"] = cr.Name

	return labels
}
