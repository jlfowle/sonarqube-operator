package api_client

type Upgrades struct {
	Upgrades            []Upgrade `json:"upgrades,omitempty"`
	UpdateCenterRefresh string    `json:"updateCenterRefresh,omitempty"`
}

type Upgrade struct {
	Version     SystemVersion `json:"version,omitempty"`
	Description string        `json:"description,omitempty"`
	ReleaseDate string        `json:"releaseDate,omitempty"`
	Plugins     Plugins       `json:"plugins,omitempty"`
}

type Plugins struct {
	RequireUpdate []Plugin `json:"requireUpdate,omitempty"`
	Incompatible  []Plugin `json:"incompatible,omitempty"`
}

type Plugin struct {
	Key     string `json:"key,omitempty"`
	Name    string `json:"name,omitempty"`
	Version string `json:"version,omitempty"`
}
