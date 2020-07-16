package api_client

type Status struct {
	ID      string        `json:"id"`
	Version SystemVersion `json:"version"`
	Status  SystemStatus  `json:"status"`
}

type SystemStatus string

const (
	SystemUp                 SystemStatus = "UP"
	SystemDown               SystemStatus = "DOWN"
	SystemStarting           SystemStatus = "STARTING"
	SystemRestarting         SystemStatus = "RESTARTING"
	SystemDBMigrationNeeded  SystemStatus = "DB_MIGRATION_NEEDED"
	SystemDBMigrationRunning SystemStatus = "DB_MIGRATION_RUNNING"
)

type SonarQube struct {
	Version string
}
