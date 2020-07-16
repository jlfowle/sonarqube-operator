package v1alpha1

import (
	"github.com/operator-framework/operator-sdk/pkg/status"
)

// Condition Types
const (
	// ConditionInvalid means that there is a misconfiguration that can not be corrected by the operator.
	ConditionInvalid status.ConditionType = "Invalid"
	// ConditionProgressing means that for some reason the state of the resources did not match the expected state.
	// Resources are being updated to meet expected state.
	ConditionProgressing status.ConditionType = "Progressing"
	// ConditionShutdown means that the resources have been shutdown.
	ConditionShutdown status.ConditionType = "Shutdown"
	// ConditionUnavailable means that the application is not available.
	ConditionUnavailable status.ConditionType = "Unavailable"
)

// Condition Reasons
const (
	// ConditionResourcesCreating means that resources are being created
	ConditionResourcesCreating status.ConditionReason = "CreatingResources"
	// ConditionReasourceUpdating means that resources are updating
	ConditionReasourcesUpdating status.ConditionReason = "ResourcesUpdating"
	// ConditionReasourceInvalid means that one or more resources are invalid
	ConditionReasourcesInvalid status.ConditionReason = "ResourcesInvalid"
	// ConditionSpecInvalid means that the current spec would result in an invalid running configuration
	ConditionSpecInvalid status.ConditionReason = "SpecInvalid"
	// ConditionConfigured means that the current spec specified meeting this condition
	ConditionConfigured status.ConditionReason = "Configured"
)

const (
	SecretAnnotation       = "sonarqube.sonarsource.jfowler.github.io/database"
	ServerSecretAnnotation = "sonarqubeserver.sonarsource.jfowler.github.io/database"
)

const (
	KubeAppComponent = "app.kubernetes.io/component"
	KubeAppPartof    = "app.kubernetes.io/part-of"
	KubeAppVersion   = "app.kubernetes.io/version"
	KubeAppInstance  = "app.kubernetes.io/instance"
	KubeAppManagedby = "app.kubernetes.io/managed-by"
	KubeAppName      = "app.kubernetes.io/name"
	TypeLabel        = "sonarsource.jfowler.github.io/SonarQube"
	ServerTypeLabel  = "sonarsource.jfowler.github.io/SonarQubeServer"
)

type ServerType string

const (
	AIO         ServerType = "aio"
	Application ServerType = "application"
	Search      ServerType = "search"
)

const (
	ApplicationWebPort int32 = 9000
	ApplicationPort    int32 = 9003
	ApplicationCEPort  int32 = 9004
	SearchPort         int32 = 9001
)

type DeploymentStatuses map[DeploymentStatus][]string

type DeploymentStatus string

const (
	DeploymentReady       DeploymentStatus = "Ready"
	DeploymentAvailable   DeploymentStatus = "Available"
	DeploymentUpdating    DeploymentStatus = "Updating"
	DeploymentUnavailable DeploymentStatus = "Unavailable"
)
