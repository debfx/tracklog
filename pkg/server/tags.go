package server

import (
	"fmt"
	"net/http"

	"github.com/debfx/tracklog/pkg/utils"
)

type tagsData struct {
	Tags []string
}

type tagData struct {
	TagName  string
	Duration string
	Distance string
	Logs     []*logsDataLog
}

func (s *Server) HandleGetTags(w http.ResponseWriter, r *http.Request) {
	ctx := NewContext(r, w)

	user := ctx.User()
	if user == nil {
		s.redirectToSignIn(w, r)
		return
	}

	tags, err := s.db.UserLogTags(user)
	if err != nil {
		panic(err)
	}

	var data tagsData
	data.Tags = tags

	ctx.SetTitle("Tags")
	ctx.Breadcrumb().Add("Tags", "", true)
	ctx.SetActiveTab("tags")
	ctx.SetData(data)

	s.render(w, r, "tags")
}

func (s *Server) HandleGetTag(w http.ResponseWriter, r *http.Request) {
	ctx := NewContext(r, w)

	user := ctx.User()
	if user == nil {
		s.redirectToSignIn(w, r)
		return
	}

	tag := ctx.Params().ByName("tag")

	logs, err := s.db.UserLogsByTag(user, tag)
	if err != nil {
		panic(err)
	}
	if len(logs) == 0 {
		http.NotFound(w, r)
		return
	}

	var (
		data     tagData
		duration uint
		distance uint
	)

	for _, log := range logs {
		data.Logs = append(data.Logs, &logsDataLog{
			ID:       log.ID,
			Name:     log.Name,
			Start:    log.Start.Format(logTimeFormat),
			Duration: utils.Duration(log.Duration).String(),
			Distance: utils.Distance(log.Distance).String(),
			Tags:     log.Tags,
		})

		distance += log.Distance
		duration += log.Duration
	}

	data.TagName = tag
	data.Duration = utils.Duration(duration).String()
	data.Distance = utils.Distance(distance).String()

	ctx.SetTitle(fmt.Sprintf("Tag %s", tag))
	ctx.Breadcrumb().Add("Tags", "/tags", false)
	ctx.Breadcrumb().Add(tag, "", true)
	ctx.SetActiveTab("tags")
	ctx.SetData(data)

	s.render(w, r, "tag")
}
