package common

import (
	"fmt"

	"efprojects.com/kitten-ipc/kitcom/api"
)

type Parser struct {
	Files []string
}

func (p *Parser) AddFile(path string) {
	p.Files = append(p.Files, path)
}

func (p *Parser) MapFiles(parseFile func(path string) ([]api.Endpoint, error)) (*api.Api, error) {
	var apis api.Api

	for _, f := range p.Files {
		endpoints, err := parseFile(f)
		if err != nil {
			return nil, fmt.Errorf("parse file: %w", err)
		}
		apis.Endpoints = append(apis.Endpoints, endpoints...)
	}

	if len(apis.Endpoints) == 0 {
		return nil, fmt.Errorf("no endpoints found")
	}

	return &apis, nil
}
