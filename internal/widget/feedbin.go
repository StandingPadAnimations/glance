package widget

import (
	"context"
	"html/template"
	"time"

	"github.com/glanceapp/glance/internal/assets"
	"github.com/glanceapp/glance/internal/feed"
)

type Feedbin struct {
	widgetBase          `yaml:",inline"`
	Feeds               feed.FeedBinFeeds `yaml:"-"`
	Limit               int             `yaml:"limit"`
	CollapseAfter	    int		`yaml:"collapse-after"`

	Username	    string `yaml:"username"`
	Password	    string `yaml:"password"`
}

func (widget *Feedbin) Initialize() error {
	widget.withTitle("Feedbin").withCacheDuration(30 * time.Minute)

	if widget.Limit <= 0 {
		widget.Limit = 15
	}

	if widget.CollapseAfter == 0 || widget.CollapseAfter < -1 {
		widget.CollapseAfter = 3
	}

	return nil
}

func (widget *Feedbin) Update(ctx context.Context) {
	feeds, err := feed.FetchFeedbinUnreadFeeds(40, widget.Username, widget.Password)
	if !widget.canContinueUpdateAfterHandlingErr(err) {
		return
	}
	widget.Feeds = feeds
}

func (widget *Feedbin) Render() template.HTML {
	return widget.render(widget, assets.FeedbinPostsTemplate)
}
