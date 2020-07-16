package api_client

type APIClientMock struct {
	PingError      error
	InfoOutput     *Status
	InfoError      error
	UpgradesOutput *Upgrades
	UpgradesError  error
}

func (r *APIClientMock) New(string) APIReader {
	return r
}

func (r *APIClientMock) Ping() error {
	return r.PingError
}

func (r *APIClientMock) Status() (*Status, error) {
	return r.InfoOutput, r.InfoError
}

func (r *APIClientMock) Upgrades() (*Upgrades, error) {
	return r.UpgradesOutput, r.UpgradesError
}
