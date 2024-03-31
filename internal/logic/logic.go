package logic

import (
	"bytes"
	"fmt"
	"sync/atomic"
	"text/template"
)

type LogicApp struct {
	countReq int32
	tpl      *template.Template
}

func NewLogic() *LogicApp {
	// init template
	const tmpl = "Hello, {{.Name}}! Your number is {{.CountReqNumber}}."
	t, _ := template.New("greeting").Parse(tmpl)

	return &LogicApp{
		tpl: t,
	}
}

func (l *LogicApp) GenerateText(name string) (string, error) {

	val := atomic.LoadInt32(&l.countReq)

	var buf bytes.Buffer

	data := struct {
		Name           string
		CountReqNumber int32
	}{
		Name:           name,
		CountReqNumber: val,
	}

	err := l.tpl.Execute(&buf, data)
	if err != nil {
		return "", fmt.Errorf("cannot execute tmpl: %w", err)
	}
	atomic.AddInt32(&l.countReq, 1)

	return buf.String(), nil
}
