package sonarqube

import (
	"fmt"
	sonarsourcev1alpha1 "github.com/jlfowle/sonarqube-operator/pkg/apis/sonarsource/v1alpha1"
	"github.com/jlfowle/sonarqube-operator/pkg/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"strings"
)

const (
	PodGracePeriod       int64  = 3600
	VolumePathData       string = "/opt/sonarqube/data"
	VolumePathLogs       string = "/opt/sonarqube/logs"
	VolumePathTemp       string = "/opt/sonarqube/temp"
	VolumePathExtensions string = "/opt/sonarqube/extensions"
)

// Reconciles Deployment for SonarQube
// Returns: Deployment, Error
// If Error is non-nil, Deployment is not in expected state
// Errors:
//   ErrorReasonResourceCreate: returned when Deployment does not exists
//   ErrorReasonResourceUpdate: returned when Deployment was updated to meet expected state
//   ErrorReasonUnknown: returned when unhandled error from client occurs
func (r *ReconcileSonarQube) ReconcileDeployment(cr *sonarsourcev1alpha1.SonarQube) (*appsv1.Deployment, error) {
	deployment, err := r.findDeployment(cr)
	if err != nil {
		return deployment, err
	}

	err = r.verifyDeployment(cr, deployment)
	if err != nil {
		return deployment, err
	}

	newStatus := cr.DeepCopy()

	newStatus.Status.Deployment = r.getDeploymentStatus([]*appsv1.Deployment{deployment})
	utils.UpdateStatus(r.client, newStatus, cr)

	if utils.GetDeploymentCondition(deployment, appsv1.DeploymentReplicaFailure) == corev1.ConditionTrue {
		return deployment, &utils.Error{
			Reason:  utils.ErrorReasonResourceInvalid,
			Message: "deployment replica failure",
		}
	}

	if deployment.Status.Replicas > 0 && len(newStatus.Status.Deployment[sonarsourcev1alpha1.DeploymentReady]) < 1 {
		return deployment, &utils.Error{
			Reason:  utils.ErrorReasonResourceWaiting,
			Message: "waiting for deployment to be ready",
		}
	}

	if deployment.Status.Replicas > 0 && len(newStatus.Status.Deployment[sonarsourcev1alpha1.DeploymentAvailable]) < 1 && len(newStatus.Status.Deployment[sonarsourcev1alpha1.DeploymentReady]) < 1 {
		return deployment, &utils.Error{
			Reason:  utils.ErrorReasonResourceWaiting,
			Message: "waiting for deployment to be available and not progressing",
		}
	}

	return deployment, nil
}

func (r *ReconcileSonarQube) findDeployment(cr *sonarsourcev1alpha1.SonarQube) (*appsv1.Deployment, error) {
	newDeployment, err := r.newDeployment(cr)
	if err != nil {
		return newDeployment, err
	}

	foundDeployment := &appsv1.Deployment{}

	return foundDeployment, utils.CreateResourceIfNotFound(r.client, newDeployment, foundDeployment)
}

func (r *ReconcileSonarQube) newDeployment(cr *sonarsourcev1alpha1.SonarQube) (*appsv1.Deployment, error) {
	labels := r.Labels(cr)
	podLabels := r.PodLabels(cr)

	serviceAccount, secret, pvc, service, err := r.getDeploymentDeps(cr)
	if err != nil {
		return nil, err
	}

	sqImage := utils.GetImage(cr.Spec.Edition, cr.Spec.Version)

	var replicas *int32
	if cr.Spec.Shutdown == nil || *cr.Spec.Shutdown == false {
		replicas = &[]int32{1}[0]
	} else {
		replicas = &[]int32{0}[0]
	}

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: cr.Namespace,
			Name:      cr.Name,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RecreateDeploymentStrategyType,
			},
			Replicas: replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: podLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      cr.Name,
					Namespace: cr.Namespace,
					Labels:    podLabels,
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "temp",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
						{
							Name: "conf",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: secret.Name,
									Optional:   &[]bool{true}[0],
								},
							},
						},
						{
							Name: "storage",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: pvc.Name,
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:  "sonarqube",
							Image: sqImage,
							Env: []corev1.EnvVar{
								{
									Name:  "SONARR_WEB_PORT",
									Value: fmt.Sprintf("%v", sonarsourcev1alpha1.ApplicationWebPort),
								},
								{
									Name:  "SONARR_PATH_DATA",
									Value: VolumePathData,
								},
								{
									Name:  "SONARR_PATH_LOGS",
									Value: VolumePathLogs,
								},
								{
									Name:  "SONARR_PATH_TEMP",
									Value: VolumePathTemp,
								},
								{
									Name:  "SONARR_PATH_EXTENSIONS",
									Value: VolumePathExtensions,
								},
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "storage",
									MountPath: VolumePathData,
									SubPath:   "data",
								},
								{
									Name:      "storage",
									MountPath: VolumePathLogs,
									SubPath:   "logs",
								},
								{
									Name:      "temp",
									MountPath: VolumePathTemp,
								},
								{
									Name:      "storage",
									MountPath: VolumePathExtensions,
									SubPath:   "extensions",
								},
								{
									Name:      "conf",
									MountPath: "/opt/sonarqube/conf/",
								},
							},
							LivenessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									TCPSocket: &corev1.TCPSocketAction{
										Port: intstr.IntOrString{
											Type:   intstr.Int,
											IntVal: sonarsourcev1alpha1.ApplicationWebPort,
											StrVal: "",
										},
									},
								},
								InitialDelaySeconds: 60,
								TimeoutSeconds:      1,
								PeriodSeconds:       10,
								SuccessThreshold:    1,
								FailureThreshold:    3,
							},
							ReadinessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/api/system/status",
										Port: intstr.IntOrString{
											Type:   intstr.Int,
											IntVal: sonarsourcev1alpha1.ApplicationWebPort,
											StrVal: "",
										},
										Scheme: corev1.URISchemeHTTP,
									},
								},
								InitialDelaySeconds: 0,
								TimeoutSeconds:      1,
								PeriodSeconds:       10,
								SuccessThreshold:    1,
								FailureThreshold:    3,
							},
							ImagePullPolicy: corev1.PullAlways,
						},
					},
					RestartPolicy:                 corev1.RestartPolicyAlways,
					TerminationGracePeriodSeconds: &[]int64{PodGracePeriod}[0],
					DNSPolicy:                     corev1.DNSClusterFirst,
					ServiceAccountName:            serviceAccount.Name,
					Affinity: &corev1.Affinity{
						NodeAffinity:    cr.Spec.NodeConfig.NodeAffinity,
						PodAffinity:     cr.Spec.NodeConfig.PodAffinity,
						PodAntiAffinity: cr.Spec.NodeConfig.PodAntiAffinity,
					},
				},
			},
		},
	}

	if cr.Spec.NodeConfig.Resources != nil {
		dep.Spec.Template.Spec.Containers[0].Resources = *cr.Spec.NodeConfig.Resources
	}

	if cr.Spec.NodeConfig.NodeSelector != nil {
		dep.Spec.Template.Spec.NodeSelector = *cr.Spec.NodeConfig.NodeSelector
	}

	if cr.Spec.NodeConfig.PriorityClass != nil {
		dep.Spec.Template.Spec.PriorityClassName = *cr.Spec.NodeConfig.PriorityClass
	}

	var nodeType sonarsourcev1alpha1.ServerType
	if cr.Spec.Type == nil {
		nodeType = sonarsourcev1alpha1.AIO
	} else {
		nodeType = *cr.Spec.Type
	}

	switch nodeType {
	case sonarsourcev1alpha1.AIO:
		dep.Spec.Template.Spec.Containers[0].Ports = []corev1.ContainerPort{
			{
				Name:          "web",
				ContainerPort: sonarsourcev1alpha1.ApplicationWebPort,
				Protocol:      corev1.ProtocolTCP,
			},
		}
	case sonarsourcev1alpha1.Application:
		dep.Spec.Template.Spec.Containers[0].Ports = []corev1.ContainerPort{
			{
				Name:          "web",
				ContainerPort: sonarsourcev1alpha1.ApplicationWebPort,
				Protocol:      corev1.ProtocolTCP,
			},
			{
				Name:          "node",
				ContainerPort: sonarsourcev1alpha1.ApplicationPort,
				Protocol:      corev1.ProtocolTCP,
			},
			{
				Name:          "ce",
				ContainerPort: sonarsourcev1alpha1.ApplicationCEPort,
				Protocol:      corev1.ProtocolTCP,
			},
		}
		hosts := cr.Spec.Hosts
		if !utils.ContainsString(hosts, service.Spec.ClusterIP) {
			hosts = append(hosts, service.Spec.ClusterIP)
		}
		searchHosts := cr.Spec.SearchHosts
		if !utils.ContainsString(searchHosts, service.Spec.ClusterIP) {
			searchHosts = append(searchHosts, service.Spec.ClusterIP)
		}

		clusteredEnv := []corev1.EnvVar{
			{
				Name:  "SONAR_CLUSTER_ENABLED",
				Value: "true",
			},
			{
				Name:  "SONAR_CLUSTER_NODE_TYPE",
				Value: string(nodeType),
			},
			{
				Name:  "SONAR_CLUSTER_SEARCH_HOSTS",
				Value: strings.Join(searchHosts, ","),
			},
			{
				Name:  "SONAR_CLUSTER_HOSTS",
				Value: strings.Join(hosts, ","),
			},
			{
				Name: "SONAR_CLUSTER_NODE_HOST",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						APIVersion: "v1",
						FieldPath:  "status.podIP",
					},
				},
			},
			{
				Name:  "SONAR_CLUSTER_NODE_NAME",
				Value: dep.Name,
			},
		}
		dep.Spec.Template.Spec.Containers[0].Env = append(dep.Spec.Template.Spec.Containers[0].Env, clusteredEnv...)
	case sonarsourcev1alpha1.Search:
		dep.Spec.Template.Spec.Containers[0].Ports = []corev1.ContainerPort{
			{
				Name:          "search",
				ContainerPort: sonarsourcev1alpha1.SearchPort,
				Protocol:      corev1.ProtocolTCP,
			},
		}
		searchHosts := cr.Spec.SearchHosts
		if !utils.ContainsString(cr.Spec.SearchHosts, service.Spec.ClusterIP) {
			searchHosts = append(searchHosts, service.Spec.ClusterIP)
		}

		clusteredEnv := []corev1.EnvVar{
			{
				Name:  "SONAR_CLUSTER_ENABLED",
				Value: "true",
			},
			{
				Name:  "SONAR_CLUSTER_NODE_TYPE",
				Value: string(nodeType),
			},
			{
				Name: "SONAR_CLUSTER_NODE_HOST",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						APIVersion: "v1",
						FieldPath:  "status.podIP",
					},
				},
			},
			{
				Name:  "SONAR_CLUSTER_NODE_NAME",
				Value: dep.Name,
			},
			{
				Name:  "SONAR_CLUSTER_SEARCH_HOSTS",
				Value: strings.Join(searchHosts, ","),
			},
			{
				Name: "SONAR_SEARCH_HOST",
				ValueFrom: &corev1.EnvVarSource{
					FieldRef: &corev1.ObjectFieldSelector{
						APIVersion: "v1",
						FieldPath:  "status.podIP",
					},
				},
			},
		}
		dep.Spec.Template.Spec.Containers[0].Env = append(dep.Spec.Template.Spec.Containers[0].Env, clusteredEnv...)
		dep.Spec.Template.Spec.Containers[0].LivenessProbe.Handler.TCPSocket = &corev1.TCPSocketAction{
			Port: intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: sonarsourcev1alpha1.SearchPort,
				StrVal: "",
			},
		}
		dep.Spec.Template.Spec.Containers[0].ReadinessProbe.Handler.HTTPGet = nil
		dep.Spec.Template.Spec.Containers[0].ReadinessProbe.Handler.TCPSocket = &corev1.TCPSocketAction{
			Port: intstr.IntOrString{
				Type:   intstr.Int,
				IntVal: sonarsourcev1alpha1.SearchPort,
				StrVal: "",
			},
		}
	}

	if err := controllerutil.SetControllerReference(cr, dep, r.scheme); err != nil {
		return dep, err
	}

	return dep, nil
}

func (r *ReconcileSonarQube) getDeploymentDeps(cr *sonarsourcev1alpha1.SonarQube) (*corev1.ServiceAccount, *corev1.Secret, *corev1.PersistentVolumeClaim, *corev1.Service, error) {

	serviceAccount, err := r.ReconcileServiceAccount(cr)
	if err != nil {
		return serviceAccount, nil, nil, nil, err
	}

	secret, err := r.ReconcileSecret(cr)
	if err != nil {
		return serviceAccount, secret, nil, nil, err
	}

	pvc, err := r.ReconcilePVC(cr)
	if err != nil {
		return serviceAccount, secret, pvc, nil, err
	}

	service, err := r.ReconcileService(cr)
	if err != nil {
		return serviceAccount, secret, pvc, service, err
	}

	return serviceAccount, secret, pvc, service, nil
}

func (r *ReconcileSonarQube) verifyDeployment(cr *sonarsourcev1alpha1.SonarQube, deployment *appsv1.Deployment) error {
	newDeployment, err := r.newDeployment(cr)
	if err != nil {
		return err
	}

	if !reflect.DeepEqual(*deployment.Spec.Replicas, *newDeployment.Spec.Replicas) {
		deployment.Spec.Replicas = newDeployment.Spec.Replicas
		return utils.UpdateResource(r.client, deployment, utils.ErrorReasonResourceUpdate, "updated deployment replicas")
	}

	if !r.envEqual(newDeployment.Spec.Template.Spec.Containers[0].Env, deployment.Spec.Template.Spec.Containers[0].Env) {
		deployment.Spec.Template.Spec.Containers[0].Env = newDeployment.Spec.Template.Spec.Containers[0].Env
		return utils.UpdateResource(r.client, deployment, utils.ErrorReasonResourceUpdate, "updated deployment env")
	}

	if !reflect.DeepEqual(deployment.Spec.Template.Spec.Containers[0].ReadinessProbe, newDeployment.Spec.Template.Spec.Containers[0].ReadinessProbe) {
		deployment.Spec.Template.Spec.Containers[0].ReadinessProbe = newDeployment.Spec.Template.Spec.Containers[0].ReadinessProbe
		return utils.UpdateResource(r.client, deployment, utils.ErrorReasonResourceUpdate, "updated deployment readiness probe")
	}

	if !reflect.DeepEqual(deployment.Spec.Template.Spec.Containers[0].LivenessProbe, newDeployment.Spec.Template.Spec.Containers[0].LivenessProbe) {
		deployment.Spec.Template.Spec.Containers[0].LivenessProbe = newDeployment.Spec.Template.Spec.Containers[0].LivenessProbe
		return utils.UpdateResource(r.client, deployment, utils.ErrorReasonResourceUpdate, "updated deployment liveness probe")
	}

	if !reflect.DeepEqual(deployment.Labels, newDeployment.Labels) {
		deployment.Labels = newDeployment.Labels
		return utils.UpdateResource(r.client, deployment, utils.ErrorReasonResourceUpdate, "updated deployment labels")
	}

	return nil
}

func (r *ReconcileSonarQube) envEqual(c, p []corev1.EnvVar) bool {
	equal := true
	for _, c := range c {
		if !equal {
			break
		}
		var found bool
		for _, p := range p {
			if c.Name == p.Name {
				found = true
				if !reflect.DeepEqual(c.ValueFrom, p.ValueFrom) || c.Value != p.Value {
					equal = false
					break
				}
				break
			}
		}
		if !found {
			equal = false
			break
		}
	}
	for _, p := range p {
		if !equal {
			break
		}
		var found bool
		for _, c := range c {
			if c.Name == p.Name {
				found = true
				if !reflect.DeepEqual(c.ValueFrom, p.ValueFrom) || c.Value != p.Value {
					equal = false
					break
				}
				break
			}
		}
		if !found {
			equal = false
			break
		}
	}
	return equal
}

func (r *ReconcileSonarQube) getDeploymentStatus(deployments []*appsv1.Deployment) sonarsourcev1alpha1.DeploymentStatuses {
	status := sonarsourcev1alpha1.DeploymentStatuses{
		sonarsourcev1alpha1.DeploymentAvailable:   []string{},
		sonarsourcev1alpha1.DeploymentUpdating:    []string{},
		sonarsourcev1alpha1.DeploymentUnavailable: []string{},
		sonarsourcev1alpha1.DeploymentReady:       []string{},
	}

	for _, dep := range deployments {
		if *dep.Spec.Replicas == 0 {
			break
		}
		if dep.Status.Replicas > dep.Status.UpdatedReplicas {
			status[sonarsourcev1alpha1.DeploymentUpdating] = append(status[sonarsourcev1alpha1.DeploymentUpdating], dep.Name)
			break
		}
		if dep.Status.Replicas == dep.Status.ReadyReplicas {
			status[sonarsourcev1alpha1.DeploymentReady] = append(status[sonarsourcev1alpha1.DeploymentReady], dep.Name)
			break
		}
		if dep.Status.Replicas == dep.Status.AvailableReplicas {
			status[sonarsourcev1alpha1.DeploymentAvailable] = append(status[sonarsourcev1alpha1.DeploymentAvailable], dep.Name)
			break
		}
		if dep.Status.Replicas == dep.Status.UnavailableReplicas {
			status[sonarsourcev1alpha1.DeploymentUnavailable] = append(status[sonarsourcev1alpha1.DeploymentUnavailable], dep.Name)
			break
		}
	}

	return status
}
