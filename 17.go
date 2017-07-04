package main

import (
	"fmt"
	"net/http"
	"log"
	"time"
	"bytes"
	"encoding/json"
	"sort"
	"strings"
)

type Stop struct {
	StationName string
	Time        time.Time
	RouteName   string
}

type StopByTime []Stop

func (s StopByTime) Less(i, j int) bool {
	return s[i].Time.Before(s[j].Time)
}
func (s StopByTime) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s StopByTime) Len() int {
	return len(s)
}

type Timetable struct {
	From            time.Time        `json:"from"`
	To              time.Time        `json:"to"`
	StopsPerStation map[string]StopByTime `json:"stopsPerStation"`
}

func NewTimetable(from time.Time, to time.Time) *Timetable {
	return &Timetable{StopsPerStation: make(map[string]StopByTime), From: from, To: to}
}

func (self *Timetable) addStop(route string, station string, time time.Time) Stop {
	stops := self.StopsPerStation[station]
	if stops == nil {
		self.addStation(station)
	}
	newStop := Stop{RouteName: route, StationName: station, Time: time}
	self.StopsPerStation[station] = append(self.StopsPerStation[station],
		newStop)
	return newStop
}

func (self *Timetable) addStation(station string) string {
	self.StopsPerStation[station] = StopByTime{}
	return station
}

func (self *Timetable) getSlicedTimetable(from time.Time, to time.Time) *Timetable {
	var slicedTimetable = NewTimetable(from, to)
	for station, stops := range self.StopsPerStation {
		stopsToSort := getStopsBetween(stops, from, to)
		sort.Sort(stopsToSort)
		slicedTimetable.StopsPerStation[station] = stopsToSort
	}
	return slicedTimetable
}

func (self *Timetable) getJsonTimetable(from time.Time, to time.Time) string {
	result, err := json.Marshal(self.getSlicedTimetable(from, to))
	if err != nil {
		panic(err)
	}
	return string(result)
}

func (self *Timetable) getTextTimetable(from time.Time, to time.Time) string {
	var buffer bytes.Buffer
	slicedTimetable := self.getSlicedTimetable(from, to)
	buffer.WriteString(strings.Join([]string{"\nDEPARTURES FROM", from.Format("15:04"), "to", to.Format("15:04"), "\n"}, " "))
	for station, stops := range slicedTimetable.StopsPerStation {
		buffer.WriteString("\nSTATION: " + station + "\n")
		if len(stops) > 0 {
			for _, stop := range stops {
				buffer.WriteString(stop.Time.Format("15:04") + " - " + stop.RouteName + "\n")
			}
		} else {
			buffer.WriteString("-- No departures\n")
		}
		buffer.WriteString("---------------------\n")
	}
	return buffer.String()
}

func getStopsBetween(stops []Stop, from time.Time, to time.Time) StopByTime {
	result := StopByTime{}
	fromLimit := from.Add(time.Minute * -1)
	toLimit := to.Add(time.Minute * 1)
	for _, stop := range stops {
		if stop.Time.After(fromLimit) && stop.Time.Before(toLimit) {
			result = append(result, stop)
		}
	}
	return result
}

func loadSample(table *Timetable) *Timetable {
	table.addStop("London-Norwich", "Stevenage", parseStringToHourMinute("10:00"))
	table.addStop("London-Norwich", "Baldock", parseStringToHourMinute("11:00"))
	table.addStop("London-Norwich", "Ipswitch", parseStringToHourMinute("12:00"))
	table.addStop("London-Norwich", "XPTO", parseStringToHourMinute("13:00"))
	table.addStop("Norwich-London", "Baldock", parseStringToHourMinute("9:00"))
	table.addStop("Norwich-London", "Oxford", parseStringToHourMinute("10:00"))
	table.addStop("Norwich-London", "Ipswitch", parseStringToHourMinute("11:00"))
	table.addStop("Norwich-London", "XPTO", parseStringToHourMinute("11:00"))
	return table
}

func parseStringToHourMinute(str string) time.Time {
	hourMinute := "15:04"
	t, err := time.Parse(hourMinute, str)
	if err != nil {
		log.Print(err)
		panic(err)
	}
	return t
}

func parseParamToTime(timeString string) time.Time {
	if timeString != "" {
		return parseStringToHourMinute(timeString)
	} else {
		return parseStringToHourMinute(time.Now().Format("15:04"))
	}
}

func addSampleHandler(w http.ResponseWriter, r *http.Request) {
	timetable = loadSample(timetable)
	timetableHandler(w, r)
}

func addStopHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		defer func() {
			if r := recover(); r != nil {
				// TODO improve the error handling and message
				fmt.Fprintf(w, "Something went wrong when adding the stop. Error message: ", r)
			}
		}()
		routeName := r.URL.Query().Get("routeName")
		station := r.URL.Query().Get("station")
		time := parseStringToHourMinute((r.URL.Query().Get("time")))
		timetable.addStop(routeName, station, time)
	}
	timetableHandler(w, r)
}

func timetableHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		from := parseParamToTime(r.URL.Query().Get("from"))
		to := from.Add(time.Minute * 60)
		format := r.URL.Query().Get("format")
		if strings.EqualFold(format, "JSON") {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintln(w, timetable.getJsonTimetable(from, to))
		} else {
			fmt.Fprintln(w, timetable.getTextTimetable(from, to))
		}
	}
}

var timetable = NewTimetable(parseStringToHourMinute("00:00"), parseStringToHourMinute("23:59"))
func main() {
	http.HandleFunc("/add", addStopHandler)
	http.HandleFunc("/addSample", addSampleHandler)
	http.HandleFunc("/", timetableHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}