package form

type Template interface {
	SaveAsTemplate() bool
	TemplateName() string
	TemplateData() interface{}
}

type BaseTemplate struct {
	SaveAsTpl bool   `json:"saveAsTemplate"`
	TplName   string `json:"templateName"`
}

func (b *BaseTemplate) SaveAsTemplate() bool {
	if b == nil {
		return false
	}
	return b.SaveAsTpl
}

func (b *BaseTemplate) TemplateName() string {
	if b == nil {
		return ""
	}
	return b.TplName
}
