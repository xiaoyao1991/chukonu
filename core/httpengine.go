package core

import "net/http"

type HttpEngine struct {
	err error
	http.Client
}

func (e *HttpEngine) Run(requestProvider RequestProvider) error {
	for r := range requestProvider {
		if err := e.RunRequest(r); err != nil {
			return err
		}
	}
	return nil
}

func (e *HttpEngine) RunRequest(request ChukonuRequest) error {
	return nil
}

func (e *HttpEngine) Err() error {
	return e.err
}
