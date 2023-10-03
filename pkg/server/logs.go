package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/debfx/tracklog/pkg/geo"
	"github.com/debfx/tracklog/pkg/heartrate"
	"github.com/debfx/tracklog/pkg/models"
	"github.com/debfx/tracklog/pkg/rdp"
	"github.com/debfx/tracklog/pkg/utils"
)

const logTimeFormat = "2006-01-02 15:04:05"

type logsData struct {
	Years  []logsDataYear
	Groups []*logDataGroup
}

type logsDataYear struct {
	Year   int
	Active bool
}

type logDataGroup struct {
	Title    string
	Duration string
	Distance string
	Logs     []logsDataLog
}

type logsDataLog struct {
	ID       int
	Name     string
	Start    string
	Duration string
	Distance string
	Tags     []string
}

func (s *Server) HandleGetLogs(w http.ResponseWriter, r *http.Request) {
	ctx := NewContext(r, w)

	user := ctx.User()
	if user == nil {
		s.redirectToSignIn(w, r)
		return
	}

	years, err := s.db.UserLogYears(user)
	if err != nil {
		panic(err)
	}

	year := 0
	explicitYear := false
	if s := r.FormValue("year"); s != "" {
		if y, err := strconv.Atoi(s); err == nil {
			year = y
			explicitYear = true
		}
	}
	if year == 0 && len(years) > 0 {
		year = years[len(years)-1]
	}

	logs, err := s.db.UserLogsByYear(user, year)
	if err != nil {
		panic(err)
	}

	var (
		data     logsData
		group    *logDataGroup
		duration uint
		distance uint
		month    time.Month
	)

	for _, log := range logs {
		if month != log.Start.Month() {
			if group != nil {
				group.Duration = utils.Duration(duration).String()
				group.Distance = utils.Distance(distance).String()
			}

			duration = 0
			distance = 0
			month = log.Start.Month()

			group = &logDataGroup{
				Title: fmt.Sprintf("%s %d", month, year),
			}
			data.Groups = append(data.Groups, group)
		}

		group.Logs = append(group.Logs, logsDataLog{
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
	if group != nil {
		group.Duration = utils.Duration(duration).String()
		group.Distance = utils.Distance(distance).String()
	}

	for _, y := range years {
		data.Years = append(data.Years, logsDataYear{
			Year:   y,
			Active: y == year,
		})
	}

	if explicitYear {
		ctx.SetTitle(fmt.Sprintf("Logs %d", year))
		ctx.Breadcrumb().Add("Logs", "/logs", false)
		ctx.Breadcrumb().Add(strconv.Itoa(year), "", true)
	} else {
		ctx.SetTitle("Logs")
		ctx.Breadcrumb().Add("Logs", "", true)
	}

	ctx.SetActiveTab("logs")
	ctx.SetData(data)

	s.render(w, r, "logs")
}

type postLogRequest struct {
	Filename string `json:"filename"`
	GPX      string `json:"gpx"`
}

type postLogResponse struct {
	ID int `json:"id"`
}

func (s *Server) HandlePostLog(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		http.Error(w, "JSON requests only", http.StatusNotAcceptable)
		return
	}

	ctx := NewContext(r, w)
	user := ctx.User()
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req postLogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	name := strings.TrimSuffix(req.Filename, ".gpx")
	log, err := models.NewLog(name, []byte(req.GPX))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := s.db.AddUserLog(user, log); err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	resp := postLogResponse{ID: log.ID}
	json.NewEncoder(w).Encode(resp)
}

type logData struct {
	Log logDataLog
}

type logDataLog struct {
	ID       int              `json:"id"`
	Name     string           `json:"name"`
	Start    string           `json:"start"`
	End      string           `json:"end"`
	Duration string           `json:"duration"`
	Distance string           `json:"distance"`
	Speed    string           `json:"speed"`
	Pace     string           `json:"pace"`
	HR       string           `json:"hr,omitempty"`
	HRZones  logDataHRZones   `json:"hrzones"`
	Tracks   [][]logDataPoint `json:"tracks"`
	Tags     []string         `json:"tags"`
	Ascent   string           `json:"ascent,omitempty"`
	Descent  string           `json:"descent,omitempty"`
}

type logDataHRZones struct {
	Red        float64 `json:"red"`
	Anaerobic  float64 `json:"anaerobic"`
	Aerobic    float64 `json:"aerobic"`
	FatBurning float64 `json:"fatburning"`
	Easy       float64 `json:"easy"`
	None       float64 `json:"none"`
}

type logDataPoint struct {
	Lat               float64 `json:"lat"`
	Lon               float64 `json:"lon"`
	CumulatedDistance float64 `json:"cumulated_distance"`
	Speed             float64 `json:"speed"`
	Ele               float64 `json:"ele"`
	HR                uint    `json:"hr"`
}

func (s *Server) HandleGetLog(w http.ResponseWriter, r *http.Request) {
	ctx := NewContext(r, w)

	user := ctx.User()
	if user == nil {
		s.redirectToSignIn(w, r)
		return
	}

	id, err := strconv.Atoi(ctx.Params().ByName("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	log, err := s.db.UserLogByID(user, id)
	if err != nil {
		panic(err)
	}
	if log == nil {
		http.NotFound(w, r)
		return
	}

	data := logData{
		Log: logDataLog{
			ID:       log.ID,
			Name:     log.Name,
			Start:    log.Start.Format(logTimeFormat),
			End:      log.End.Format(logTimeFormat),
			Duration: utils.Duration(log.Duration).String(),
			Distance: utils.Distance(log.Distance).String(),
			Speed:    utils.Speed(log.Speed()).String(),
			Pace:     utils.Speed(log.Speed()).Pace().String(),
			Tracks:   make([][]logDataPoint, 0, len(log.Tracks)),
			Tags:     log.Tags,
		},
	}
	if data.Log.Tags == nil {
		data.Log.Tags = make([]string, 0, 0)
	}

	var cumDistance, ascent, descent float64
	performReduce := r.FormValue("reduce") != "no"

	for _, track := range log.Tracks {
		points := track.Points

		for i, point := range points {
			// smooth elevation
			smoothElevationNum := 0
			smoothElevationSum := float64(0)
			for j := -5 + 1; j <= 0; j++ {
				if (i+j >= 0) && (points[i+j].Elevation != 0) {
					smoothElevationNum++
					smoothElevationSum += points[i+j].Elevation
				}
			}
			if smoothElevationNum != 0 {
				point.Elevation = smoothElevationSum / float64(smoothElevationNum)
			}
		}

		if performReduce {
			rdpPoints := make([]rdp.Point, 0, len(track.Points))
			for _, point := range track.Points {
				λ, φ := geo.ToRad(point.Longitude), geo.ToRad(point.Latitude)
				x, y := geo.EquirectangularProjection(λ, φ, 0)
				rdpPoints = append(rdpPoints, rdp.Point{
					X:    x,
					Y:    y,
					Data: point,
				})
			}
			const epsilon = 0.0000001 // ≈ 1m
			reducedPoints := rdp.Reduce(rdpPoints, epsilon)
			points = make([]*models.Point, 0, len(reducedPoints))
			for _, rp := range reducedPoints {
				points = append(points, rp.Data.(*models.Point))
			}
		}

		var ps []logDataPoint
		for i, point := range points {
			p := logDataPoint{
				Lat: point.Latitude,
				Lon: point.Longitude,
				Ele: point.Elevation,
				HR:  point.Heartrate,
			}
			if i > 0 {
				lastPoint := points[i-1]
				cumDistance += point.DistanceTo(lastPoint)
				p.CumulatedDistance = cumDistance
				p.Speed = point.SpeedTo(lastPoint)

				dEle := point.Elevation - lastPoint.Elevation
				if dEle >= 0 {
					ascent += dEle
				} else {
					descent += -dEle
				}
			} else if len(points) > 1 {
				nextPoint := points[i+1]
				p.CumulatedDistance = cumDistance
				p.Speed = nextPoint.SpeedTo(point)
			}
			ps = append(ps, p)
		}
		data.Log.Tracks = append(data.Log.Tracks, ps)
	}

	if ascent != 0 || descent != 0 {
		data.Log.Ascent = fmt.Sprintf("%d m", int(ascent))
		data.Log.Descent = fmt.Sprintf("%d m", int(descent))
	}

	hrSummary := heartrate.SummaryForLog(log)
	if len(hrSummary.Rates) > 0 {
		data.Log.HR = strconv.Itoa(hrSummary.Average)

		perc := func(zone heartrate.Zone) float64 {
			return float64(hrSummary.Zones[zone]) / float64(len(hrSummary.Rates)) * 100.0
		}

		data.Log.HRZones.Red = perc(heartrate.ZoneRed)
		data.Log.HRZones.Anaerobic = perc(heartrate.ZoneAnaerobic)
		data.Log.HRZones.Aerobic = perc(heartrate.ZoneAerobic)
		data.Log.HRZones.FatBurning = perc(heartrate.ZoneFatBurning)
		data.Log.HRZones.Easy = perc(heartrate.ZoneEasy)
		data.Log.HRZones.None = perc(heartrate.ZoneNone)
	}

	ctx.SetTitle(log.Name)
	ctx.Breadcrumb().Add("Logs", "/logs", false)
	ctx.Breadcrumb().Add(log.Name, "", true)
	ctx.SetActiveTab("logs")
	ctx.SetData(data)

	s.render(w, r, "log")
}

func (s *Server) HandleDownloadLog(w http.ResponseWriter, r *http.Request) {
	ctx := NewContext(r, w)

	user := ctx.User()
	if user == nil {
		s.redirectToSignIn(w, r)
		return
	}

	id, err := strconv.Atoi(ctx.Params().ByName("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	log, err := s.db.UserLogByID(user, id)
	if err != nil {
		panic(err)
	}
	if log == nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/gpx+json")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%d.gpx", log.ID))
	w.Write([]byte(log.GPX))
}

type patchLogRequest struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

func (s *Server) HandlePatchLog(w http.ResponseWriter, r *http.Request) {
	ctx := NewContext(r, w)

	user := ctx.User()
	if user == nil {
		s.redirectToSignIn(w, r)
		return
	}

	id, err := strconv.Atoi(ctx.Params().ByName("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	log, err := s.db.UserLogByID(user, id)
	if err != nil {
		panic(err)
	}
	if log == nil {
		http.NotFound(w, r)
		return
	}

	var req patchLogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		panic(err)
	}
	log.Name = req.Name
	if req.Tags != nil {
		log.Tags = req.Tags
	}

	if err := s.db.UpdateLog(log); err != nil {
		panic(err)
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) HandleDeleteLog(w http.ResponseWriter, r *http.Request) {
	ctx := NewContext(r, w)

	user := ctx.User()
	if user == nil {
		s.redirectToSignIn(w, r)
		return
	}

	id, err := strconv.Atoi(ctx.Params().ByName("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	log, err := s.db.UserLogByID(user, id)
	if err != nil {
		panic(err)
	}
	if log == nil {
		http.NotFound(w, r)
		return
	}

	if err := s.db.DeleteLog(log); err != nil {
		panic(err)
	}

	http.Redirect(w, r, "/logs", http.StatusSeeOther)
}
