package sonarqube

import (
	"context"
	"fmt"

	"github.com/jlfowle/sonarqube-operator/pkg/api_client"
	sonarsourcev1alpha1 "github.com/jlfowle/sonarqube-operator/pkg/apis/sonarsource/v1alpha1"
	"github.com/jlfowle/sonarqube-operator/pkg/utils"
)

func (r *ReconcileSonarQube) ReconcileServer(cr *sonarsourcev1alpha1.SonarQube) error {
	service, err := r.ReconcileService(cr)
	if err != nil {
		return err
	}

	var url string
	if cr.Spec.ExternalURL != nil {
		url = *cr.Spec.ExternalURL
	} else {
		url = fmt.Sprintf("http://%s:%v", service.Spec.ClusterIP, service.Spec.Ports[0].Port)
	}
	apiClient := r.apiClient.New(url)

	/*err = apiClient.Ping()
	if err != nil {
		return &utils.Error{
			Reason:  utils.ErrorReasonServerWaiting,
			Message: fmt.Sprintf("waiting for api to respond (%s)", err.Error()),
		}
	}*/

	status, err := r.verifyServerStatus(cr, apiClient)

	err = r.verifyServerVersion(cr, status)
	if err != nil {
		return err
	}

	err = r.verifyUpgrades(cr, apiClient)
	if err != nil {
		return err
	}

	return nil
}

func (r *ReconcileSonarQube) verifyServerStatus(_ *sonarsourcev1alpha1.SonarQube, apiClient api_client.APIReader) (*api_client.Status, error) {
	status, err := apiClient.Status()
	if err != nil {
		return status, err
	}

	switch status.Status {
	case api_client.SystemDown:
		return status, &utils.Error{
			Reason:  utils.ErrorReasonServerDown,
			Message: fmt.Sprintf("sonarqube server status %s", status.Status),
		}
	case api_client.SystemStarting, api_client.SystemRestarting, api_client.SystemDBMigrationRunning, api_client.SystemDBMigrationNeeded:
		return status, &utils.Error{
			Reason:  utils.ErrorReasonServerWaiting,
			Message: fmt.Sprintf("sonarqube server status %s", status.Status),
		}
	case api_client.SystemUp:
		return status, nil
	default:
		return status, &utils.Error{
			Reason:  utils.ErrorReasonServerWaiting,
			Message: fmt.Sprintf("waiting for server status to report"),
		}
	}
}

func (r *ReconcileSonarQube) verifyServerVersion(cr *sonarsourcev1alpha1.SonarQube, status *api_client.Status) error {

	mmVersion := status.Version.MajorMinorPatch()
	testVerions := &api_client.SystemVersion{
		Major: 0,
		Minor: 0,
		Patch: 0,
		Build: "",
	}

	if testVerions.MajorMinorPatch() == status.Version.MajorMinorPatch() {
		return &utils.Error{
			Reason:  utils.ErrorReasonServerWaiting,
			Message: fmt.Sprintf("waiting for server to report non-zero version"),
		}
	}

	if cr.Spec.Version == nil {
		cr.Spec.Version = &mmVersion
		if cr.Spec.Edition == nil {
			cr.Spec.Edition = &[]string{"community"}[0]
		}
		err := r.client.Update(context.TODO(), cr)
		if err != nil {
			return err
		}
		return &utils.Error{
			Reason:  utils.ErrorReasonSpecUpdate,
			Message: "set version",
		}
	}

	version, _ := status.Version.MarshalJSON()
	newStatus := cr.DeepCopy()
	newStatus.Status.ObservedVersion = string(version)
	utils.UpdateStatus(r.client, newStatus, cr)

	return nil
}

func (r *ReconcileSonarQube) verifyUpgrades(cr *sonarsourcev1alpha1.SonarQube, apiClient api_client.APIReader) error {
	upgrades, err := apiClient.Upgrades()
	if err != nil {
		return err
	} else if upgrades == nil {
		return fmt.Errorf("nil returned for upgrades")
	}

	newStatus := cr.DeepCopy()

	newStatus.Status.Upgrades = sonarsourcev1alpha1.Upgrades{
		Compatible:   []string{},
		Incompatible: []string{},
	}

	for _, v := range upgrades.Upgrades {
		if len(v.Plugins.Incompatible) > 0 {
			newStatus.Status.Upgrades.Incompatible = append(newStatus.Status.Upgrades.Incompatible, v.Version.MajorMinorPatch())
		} else {
			newStatus.Status.Upgrades.Compatible = append(newStatus.Status.Upgrades.Compatible, v.Version.MajorMinorPatch())
		}
	}

	utils.UpdateStatus(r.client, newStatus, cr)

	return nil
}
