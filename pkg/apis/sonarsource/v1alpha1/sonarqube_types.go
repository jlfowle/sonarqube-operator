package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/operator-framework/operator-sdk/pkg/status"
	corev1 "k8s.io/api/core/v1"
)

// SonarQubeSpec defines the desired state of SonarQube
type SonarQubeSpec struct {
	// Shutdown SonarQube server
	// +optional
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Shutdown"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:booleanSwitch"
	Shutdown *bool `json:"shutdown,omitempty"`

	// if empty operator will start latest version of selected edition then lock the version
	// +optional
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Version"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:text,urn:alm:descriptor:com.tectonic.ui:advanced"
	Version *string `json:"version,omitempty"`

	// community, developer, or enterprise (default is community)
	// +optional
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Edition"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:text,urn:alm:descriptor:com.tectonic.ui:advanced"
	// +kubebuilder:validation:Enum=community;developer;enterprise
	Edition *string `json:"edition,omitempty"`

	// Automatically apply minor version updates
	// +optional
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Minor"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:checkbox,urn:alm:descriptor:com.tectonic.ui:advanced,urn:alm:descriptor:com.tectonic.ui:fieldGroup:updates"
	UpdatesMinor *bool `json:"updatesMinor,omitempty"`

	// Automatically apply major version updates
	// +optional
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Major"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:checkbox,urn:alm:descriptor:com.tectonic.ui:advanced,urn:alm:descriptor:com.tectonic.ui:fieldGroup:updates"
	UpdatesMajor *bool `json:"updatesMajor,omitempty"`

	// Secret with sonar configuration files (sonar.properties, wrapper.properties).
	// Don't add cluster properties to configuration files as this could cause unexpected results
	// +optional
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Secret"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:io.kubernetes:Secret"
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors.x-descriptors="urn:alm:descriptor:io.kubernetes:Secret"
	Secret *string `json:"secret,omitempty"`

	// Sonar Node Type application or search when clustering is enabled otherwise aio (all-in-one)
	// +optional
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Server Type"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:select:aio,urn:alm:descriptor:com.tectonic.ui:select:application,urn:alm:descriptor:com.tectonic.ui:select:search"
	// +kubebuilder:validation:Enum=aio;application;search
	Type *ServerType `json:"type,omitempty"`

	// SonarQube application hosts list
	// +optional
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Hosts"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:text,urn:alm:descriptor:com.tectonic.ui:arrayFieldGroup:hosts,urn:alm:descriptor:com.tectonic.ui:advanced"
	Hosts []string `json:"hosts,omitempty"`

	// SonarQube search hosts list
	// +optional
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Search Hosts"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:text,urn:alm:descriptor:com.tectonic.ui:arrayFieldGroup:searchHosts,urn:alm:descriptor:com.tectonic.ui:advanced"
	SearchHosts []string `json:"searchHosts,omitempty"`

	// Service Account
	// +optional
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Service Account"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:text"
	ServiceAccount *string `json:"serviceAccount,omitempty"`

	// External base URL
	// +optional
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="External URL"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:text"
	ExternalURL *string `json:"externalURL,omitempty"`

	// Node Configuration
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=false
	NodeConfig NodeConfig `json:"nodeConfig,omitempty"`
}

type NodeConfig struct {
	// Node selector
	// +optional
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Node Selector"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:selector:Node,urn:alm:descriptor:com.tectonic.ui:advanced"
	NodeSelector *map[string]string `json:"nodeSelector,omitempty"`

	// Node Affinity
	// +optional
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Node Affinity"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:nodeAffinity,urn:alm:descriptor:com.tectonic.ui:advanced"
	NodeAffinity *corev1.NodeAffinity `json:"nodeAffinity,omitempty"`

	// Pod Affinity
	// +optional
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Pod Affinity"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:podAffinity,urn:alm:descriptor:com.tectonic.ui:advanced"
	PodAffinity *corev1.PodAffinity `json:"podAffinity,omitempty"`

	// Pod AntiAffinity
	// +optional
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Pod AntiAffinity"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:podAntiAffinity,urn:alm:descriptor:com.tectonic.ui:advanced"
	PodAntiAffinity *corev1.PodAntiAffinity `json:"podAntiAffinity,omitempty"`

	// Priority Class Name
	// +optional
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Priority Class"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:text,urn:alm:descriptor:com.tectonic.ui:advanced"
	PriorityClass *string `json:"priorityClass,omitempty"`

	// Resource requirements
	// +optional
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Resources"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:resourceRequirements,urn:alm:descriptor:com.tectonic.ui:advanced"
	Resources *corev1.ResourceRequirements `json:"resources,omitempty"`

	// Storage class
	// +optional
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Storage Class"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:io.kubernetes:StorageClass"
	StorageClass *string `json:"storageClass,omitempty"`

	// Size of Storage (ex 1Gi)
	// +optional
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.displayName="Storage Size"
	// +operator-sdk:gen-csv:customresourcedefinitions.specDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:text"
	StorageSize *string `json:"storageSize,omitempty"`
}

// SonarQubeStatus defines the observed state of SonarQube
type SonarQubeStatus struct {
	// Conditions represent the latest available observations of an object's state
	Conditions status.Conditions `json:"conditions,omitempty"`

	// Kubernetes service that can be used to expose SonarQube
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors.displayName="Service"
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors.x-descriptors="urn:alm:descriptor:io.kubernetes:Service"
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors=true
	Service string `json:"service,omitempty"`

	// Status of pods
	// +optional
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors.displayName="Pod Statuses"
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors=true
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:podStatuses"
	Deployment DeploymentStatuses `json:"deployment,omitempty"`

	// Hash of latest spec & controller version for revision tracking
	Revision string `json:"revision,omitempty"`

	// Current observed version of SonarQube
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors.displayName="Observed Version"
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors.x-descriptors="urn:alm:descriptor:com.tectonic.ui:text"
	// +operator-sdk:gen-csv:customresourcedefinitions.statusDescriptors=false
	ObservedVersion string `json:"observedVersion,omitempty"`

	Upgrades Upgrades `json:"upgrades,omitempty"`
}

type Upgrades struct {
	Compatible   []string `json:"compatible,omitempty"`
	Incompatible []string `json:"incompatible,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SonarQube is the Schema for the sonarqubes API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=sonarqubes,scope=Namespaced
// +operator-sdk:gen-csv:customresourcedefinitions.displayName="SonarQube Server"
// +operator-sdk:gen-csv:customresourcedefinitions.resources="Service,v1,\"\""
// +operator-sdk:gen-csv:customresourcedefinitions.resources="Secret,v1,\"\""
// +operator-sdk:gen-csv:customresourcedefinitions.resources="Deployment,v1,\"\""
// +operator-sdk:gen-csv:customresourcedefinitions.resources="PersistentVolumeClaim,v1,\"\""
type SonarQube struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SonarQubeSpec   `json:"spec,omitempty"`
	Status SonarQubeStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SonarQubeList contains a list of SonarQube
type SonarQubeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SonarQube `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SonarQube{}, &SonarQubeList{})
}
