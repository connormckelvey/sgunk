package project

import "context"

type PageMetaValue struct {
	Title    *string `yaml:"title"`
	Property *string `yaml:"property"`
	Content  *string `yaml:"content"`
	Name     *string `yaml:"name"`
}

type PageLinksValue struct {
	Rel  *string `yaml:"rel"`
	Href *string `yaml:"href"`
	Type *string `yaml:"type"`
	Page *string `yaml:"page"`
	As   *string `yaml:"as"`
}

type BuildContext struct {
	ctx context.Context
}

type buildContextKey string

const (
	buildContextPageMetaKey  buildContextKey = "page.meta"
	buildContextPageLinksKey buildContextKey = "page.links"
)

func (pc BuildContext) WithValue(key any, value any) BuildContext {
	return BuildContext{
		ctx: context.WithValue(pc.ctx, key, value),
	}
}

func (pc BuildContext) WithPageMeta(meta []PageMetaValue) BuildContext {
	return BuildContext{
		ctx: context.WithValue(pc.ctx, buildContextPageMetaKey, meta),
	}
}

func (pc BuildContext) GetPageMeta() ([]PageMetaValue, bool) {
	v := pc.ctx.Value(buildContextPageMetaKey)
	if v == nil {
		return nil, false
	}
	vv, ok := v.([]PageMetaValue)
	return vv, ok
}

func (pc BuildContext) WithPageLinks(meta []PageLinksValue) BuildContext {
	return BuildContext{
		ctx: context.WithValue(pc.ctx, buildContextPageLinksKey, meta),
	}
}

func (pc BuildContext) GetPageLinks() ([]PageLinksValue, bool) {
	v := pc.ctx.Value(buildContextPageLinksKey)
	if v == nil {
		return nil, false
	}
	vv, ok := v.([]PageLinksValue)
	return vv, ok
}
