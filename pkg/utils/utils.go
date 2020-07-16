package utils

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	sonarsourcev1alpha1 "github.com/jlfowle/sonarqube-operator/pkg/apis/sonarsource/v1alpha1"
	"github.com/magiconair/properties"
	"github.com/operator-framework/operator-sdk/pkg/status"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const (
	DefaultImage = "sonarqube"
)

var log = logf.Log.WithName("controller_sonarqube")

func IsOwner(owner, child metav1.Object) bool {
	ownerUID := owner.GetUID()
	for _, v := range child.GetOwnerReferences() {
		if v.UID == ownerUID {
			return true
		}
	}
	return false
}

func ContainsString(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func GetProperties(s *corev1.Secret, f string) (*properties.Properties, error) {
	if v, ok := s.Data[f]; ok {
		return properties.Load(v, properties.UTF8)
	} else {
		return nil, &Error{
			Reason:  ErrorReasonSpecInvalid,
			Message: fmt.Sprintf("%s doesn't exist in secret %s", f, s.Name),
		}
	}
}

func GetDeploymentCondition(deployment *appsv1.Deployment, condition appsv1.DeploymentConditionType) corev1.ConditionStatus {
	for _, v := range deployment.Status.Conditions {
		if v.Type == condition {
			return v.Status
		}
	}
	return corev1.ConditionUnknown
}

func UpdateResource(client client.Writer, object runtime.Object, reason ErrorType, message string) error {
	err := client.Update(context.TODO(), object)
	if err != nil {
		return err
	}
	return &Error{
		Reason:  reason,
		Message: message,
	}
}

func CreateResourceIfNotFound(client client.Client, object, output runtime.Object) error {
	metaObject := object.(metav1.Object)
	err := client.Get(context.TODO(), types.NamespacedName{Name: metaObject.GetName(), Namespace: metaObject.GetNamespace()}, output)
	if err != nil && errors.IsNotFound(err) {
		err := client.Create(context.TODO(), object)
		if err != nil {
			return err
		}
		return &Error{
			Reason:  ErrorReasonResourceCreate,
			Message: fmt.Sprintf("created %s %s", object.GetObjectKind().GroupVersionKind().Kind, metaObject.GetName()),
		}
	} else if err != nil {
		return err
	}

	return nil
}

func ClearConditions(conditions status.Conditions) status.Conditions {
	var cList []status.ConditionType
	for _, c := range conditions {
		// Filter out excluded condition types
		for _, e := range []status.ConditionType{sonarsourcev1alpha1.ConditionUnavailable} {
			if e == c.Type {
				break
			}
		}
		cList = append(cList, c.Type)
	}

	for _, c := range cList {
		conditions.SetCondition(status.Condition{
			Type:   c,
			Status: corev1.ConditionFalse,
		})
	}

	return conditions
}

func ParseErrorForReconcileResult(client client.Client, object interface{}, err error) (reconcile.Result, error) {
	objectRuntime := object.(runtime.Object)
	objectMeta := object.(metav1.Object)
	reqLogger := log.WithValues("SonarQube.Namespace", objectMeta.GetNamespace(), "SonarQube.Name", objectMeta.GetName())
	newStatus := objectRuntime.DeepCopyObject()
	var statusConditions *status.Conditions
	switch t := newStatus.(type) {
	case *sonarsourcev1alpha1.SonarQube:
		statusConditions = &t.Status.Conditions
	}

	if statusConditions == nil {
		statusConditions = &status.Conditions{}
	}

	if err != nil && ReasonForError(err) != ErrorReasonUnknown {
		sqErr := err.(*Error)
		switch sqErr.Type() {
		case ErrorReasonSpecUpdate, ErrorReasonResourceCreate, ErrorReasonResourceUpdate, ErrorReasonResourceWaiting, ErrorReasonServerWaiting:
			*statusConditions = ClearConditions(*statusConditions)
			var reason status.ConditionReason
			switch sqErr.Type() {
			case ErrorReasonSpecUpdate:
				reason = sonarsourcev1alpha1.ConditionConfigured
			case ErrorReasonResourceCreate:
				reason = sonarsourcev1alpha1.ConditionResourcesCreating
			case ErrorReasonResourceUpdate, ErrorReasonResourceWaiting:
				reason = sonarsourcev1alpha1.ConditionReasourcesUpdating
			}
			statusConditions.SetCondition(status.Condition{
				Type:    sonarsourcev1alpha1.ConditionProgressing,
				Status:  corev1.ConditionTrue,
				Reason:  reason,
				Message: sqErr.Error(),
			})
			UpdateStatus(client, newStatus, object)
			reqLogger.Info(sqErr.Error())
			switch sqErr.Type() {
			case ErrorReasonServerWaiting, ErrorReasonResourceWaiting:
				return reconcile.Result{RequeueAfter: 5 * time.Second}, nil
			default:
				return reconcile.Result{Requeue: true}, nil
			}
		case ErrorReasonSpecInvalid, ErrorReasonResourceInvalid:
			*statusConditions = ClearConditions(*statusConditions)
			var reason status.ConditionReason
			switch sqErr.Type() {
			case ErrorReasonSpecInvalid:
				reason = sonarsourcev1alpha1.ConditionSpecInvalid
			case ErrorReasonResourceInvalid:
				reason = sonarsourcev1alpha1.ConditionReasourcesInvalid
			}
			statusConditions.SetCondition(status.Condition{
				Type:    sonarsourcev1alpha1.ConditionInvalid,
				Status:  corev1.ConditionTrue,
				Reason:  reason,
				Message: sqErr.Error(),
			})
			UpdateStatus(client, newStatus, object)
			reqLogger.Info(sqErr.Error())
			return reconcile.Result{}, nil
		case ErrorReasonResourceShutdown:
			*statusConditions = ClearConditions(*statusConditions)
			statusConditions.SetCondition(status.Condition{
				Type:    sonarsourcev1alpha1.ConditionShutdown,
				Status:  corev1.ConditionTrue,
				Reason:  sonarsourcev1alpha1.ConditionConfigured,
				Message: sqErr.Error(),
			})
			UpdateStatus(client, newStatus, object)
			reqLogger.Info(sqErr.Error())
			return reconcile.Result{}, nil
		default:
			reqLogger.Error(sqErr, "unhandled sonarqube error")
			return reconcile.Result{}, sqErr
		}
	}
	return reconcile.Result{}, err
}

type Status interface {
	DeepCopy()
}

func UpdateStatus(client client.Client, newObject interface{}, object interface{}) {
	objectRuntime := object.(runtime.Object)
	objectMetav1 := object.(metav1.Object)

	err := client.Get(context.TODO(), types.NamespacedName{Name: objectMetav1.GetName(), Namespace: objectMetav1.GetNamespace()}, objectRuntime)
	if err != nil {
		log.Error(err, "failed to get updated object")
	}

	var requiresUpdate bool
	switch t := object.(type) {
	case *sonarsourcev1alpha1.SonarQube:
		newSonarQube := newObject.(*sonarsourcev1alpha1.SonarQube)
		if !reflect.DeepEqual(newSonarQube.Status, t.Status) {
			t.Status = *newSonarQube.Status.DeepCopy()
			requiresUpdate = true
		}
	}
	reqLogger := log.WithValues("SonarQube.Namespace", objectMetav1.GetNamespace(), "SonarQube.Name", objectMetav1.GetName())

	if requiresUpdate {
		err := client.Status().Update(context.TODO(), objectRuntime)
		if err != nil {
			reqLogger.Error(err, "failed to update status")
		}
		err = client.Get(context.TODO(), types.NamespacedName{Name: objectMetav1.GetName(), Namespace: objectMetav1.GetNamespace()}, objectRuntime)
		if err != nil {
			reqLogger.Error(err, "failed to get updated object")
		}
	}
}

type SecretMapper struct {
	Annotation string
}

func (r *SecretMapper) Map(o handler.MapObject) []reconcile.Request {
	var output []reconcile.Request
	for k, v := range o.Meta.GetAnnotations() {
		if k == r.Annotation {
			for _, e := range strings.Split(v, ",") {
				output = append(output, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Namespace: o.Meta.GetNamespace(),
						Name:      e,
					},
				})
			}
		}
	}
	return output
}

func GenVersion(spec interface{}, secret []byte) (string, error) {
	toBeHashed, err := json.Marshal(spec)
	if err != nil {
		return "", err
	}
	toBeHashed = append(toBeHashed, secret...)

	h := sha1.New()

	h.Write(toBeHashed)

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func GetImage(edition, version *string) string {
	var sqImage, sqEdition string

	if edition != nil {
		sqEdition = *edition
	} else {
		sqEdition = "community"
	}

	if version != nil {
		sqImage = fmt.Sprintf("%s:%s-%s", DefaultImage, *version, sqEdition)
	} else {
		sqImage = fmt.Sprintf("%s:%s", DefaultImage, sqEdition)
	}

	return sqImage
}
