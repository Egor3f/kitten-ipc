package common

type Parser struct {
	Files []string
}

func (p *Parser) AddFile(path string) {
	p.Files = append(p.Files, path)
}
