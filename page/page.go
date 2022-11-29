package page

import (
	"encoding/json"
	"log"
)

type Page struct {
	Title             string
	CSRFToken         string
	ErrorMessages     []string
	ErrorMessagesJSON string
	Data              interface{}
	DataJSON          string
	UIMapData         map[string]interface{}
}

// New ...
func New() *Page {
	return &Page{}
}

// AddError ...
func (p *Page) AddError(msg string) {
	p.ErrorMessages = append(p.ErrorMessages, msg)
	p.ErrorMessagesJSON = p.justJSONMarshal(p.ErrorMessages)
}

// ResetErrors ...
func (p *Page) ResetErrors() {
	p.ErrorMessages = nil
	p.ErrorMessages = make([]string, 0)
	p.ErrorMessagesJSON = p.justJSONMarshal(p.ErrorMessages)
}

// SetData ...
func (p *Page) SetData(v interface{}) {
	p.Data = v
	p.DataJSON = p.justJSONMarshal(p.Data)
}

// JSONify ...
func (p *Page) JSONify() string {
	p.DataJSON = p.justJSONMarshal(p.Data)
	return p.justJSONMarshal(p)
}

func (p *Page) justJSONMarshal(v interface{}) string {
	result, err := json.Marshal(v)
	if err != nil {
		log.Panic(err)
	}
	return string(result)
}
