package main

import (
	"testing"
	"time"
	"encoding/json"
	"reflect"
)

func TestAddStop(t *testing.T) {
	tables := []struct {
		route   string
		station string
		time    time.Time
		stop    Stop
	}{
		{"London-Norwich", "A", parseStringToHourMinute("10:00"), Stop{RouteName:"London-Norwich", StationName: "A", Time: parseStringToHourMinute("10:00")}},
		{"London-Norwich", "B", parseStringToHourMinute("9:00"), Stop{RouteName:"London-Norwich", StationName: "B", Time: parseStringToHourMinute("9:00")}},
		{"Norwich-London", "Oxford", parseStringToHourMinute("10:00"), Stop{RouteName:"Norwich-London", StationName: "Oxford", Time: parseStringToHourMinute("10:00")}},
	}

	var timetable = NewTimetable(parseStringToHourMinute("00:00"), parseStringToHourMinute("23:59"))
	for _, table := range tables {
		newStop := timetable.addStop(table.route, table.station, table.time)
		if newStop != table.stop {
			t.Errorf("Error adding %s, %s, %s", table.route, table.station, table.time.Format("15:04"))
		}
	}
}

func TestGetStopsBetween(t *testing.T) {
	tables := []struct {
		stops []Stop
		from  time.Time
		to    time.Time
		count int
	}{
		{[]Stop{{RouteName:"London-Norwich", StationName: "A", Time: parseStringToHourMinute("10:00")},
			{RouteName:"London-Norwich", StationName: "A", Time: parseStringToHourMinute("12:00")}},
			parseStringToHourMinute("13:00"),
			parseStringToHourMinute("15:00"),
			0},
		{[]Stop{{RouteName:"London-Norwich", StationName: "A", Time: parseStringToHourMinute("10:00")},
			{RouteName:"London-Norwich", StationName: "A", Time: parseStringToHourMinute("12:00")}},
			parseStringToHourMinute("10:00"),
			parseStringToHourMinute("11:00"),
			1},
		{[]Stop{{RouteName:"London-Norwich", StationName: "A", Time: parseStringToHourMinute("10:00")},
			{RouteName:"London-Norwich", StationName: "A", Time: parseStringToHourMinute("12:00")}},
			parseStringToHourMinute("10:00"),
			parseStringToHourMinute("12:00"),
			2},
	}

	for n, table := range tables {
		var timetable = NewTimetable(parseStringToHourMinute("00:00"), parseStringToHourMinute("23:59"))
		for _, stop := range table.stops {
			timetable.addStop(stop.RouteName, stop.StationName, stop.Time)
		}
		stopsBetween := getStopsBetween(timetable.StopsPerStation["A"], table.from, table.to)
		if len(stopsBetween) != table.count {
			t.Errorf("Error on dataset entry %d. getStopsBetween count: %d, expected %d", n, len(stopsBetween), table.count)
		}
	}
}

func TestParseParamToTime(t *testing.T) {
	tables := []struct {
		timeParam string
		expected  time.Time
	}{
		{"", parseStringToHourMinute(time.Now().Format("15:04"))}, //it could fail if the minute changes between the call. Very unlikely though...
		{"00:00", parseStringToHourMinute("00:00")},
		{"13:00", parseStringToHourMinute("13:00")},
	}

	for n, table := range tables {
		parsedTime := parseParamToTime(table.timeParam)
		duration := parsedTime.Sub(table.expected)
		if duration.Minutes() != 0 {
			t.Errorf("Error on dataset entry %d. Parsing time param %s, expected %s", n, table.timeParam, table.expected.Format("15:04"))
		}
	}
}

func TestGetSlicedTimetable(t *testing.T) {
	tables := []struct {
		from                   time.Time
		to                     time.Time
		expectedTimetableIndex int
	}{
		{parseStringToHourMinute("00:00"), parseStringToHourMinute("23:59"), 0},
		{parseStringToHourMinute("10:00"), parseStringToHourMinute("10:30"), 1},
		{parseStringToHourMinute("10:00"), parseStringToHourMinute("11:00"), 2},
		{parseStringToHourMinute("8:00"), parseStringToHourMinute("9:00"), 3},
	}

	var timetable = NewTimetable(parseStringToHourMinute("00:00"), parseStringToHourMinute("23:59"))
	timetable.addStop("London-Norwich", "Stevenage", parseStringToHourMinute("10:00"))
	timetable.addStop("London-Norwich", "Baldock", parseStringToHourMinute("11:00"))
	timetable.addStop("London-Cambridge", "Baldock", parseStringToHourMinute("11:00"))
	timetable.addStop("London-Norwich", "Ipswitch", parseStringToHourMinute("12:00"))

	var expected_A = NewTimetable(parseStringToHourMinute("00:00"), parseStringToHourMinute("23:59"))
	expected_A.addStop("London-Norwich", "Stevenage", parseStringToHourMinute("10:00"))
	expected_A.addStop("London-Norwich", "Baldock", parseStringToHourMinute("11:00"))
	expected_A.addStop("London-Cambridge", "Baldock", parseStringToHourMinute("11:00"))
	expected_A.addStop("London-Norwich", "Ipswitch", parseStringToHourMinute("12:00"))

	var expected_B = NewTimetable(parseStringToHourMinute("10:00"), parseStringToHourMinute("10:30"))
	expected_B.addStop("London-Norwich", "Stevenage", parseStringToHourMinute("10:00"))
	expected_B.addStation("Baldock")
	expected_B.addStation("Ipswitch")

	var expected_C = NewTimetable(parseStringToHourMinute("10:00"), parseStringToHourMinute("11:00"))
	expected_C.addStop("London-Norwich", "Stevenage", parseStringToHourMinute("10:00"))
	expected_C.addStop("London-Norwich", "Baldock", parseStringToHourMinute("11:00"))
	expected_C.addStop("London-Cambridge", "Baldock", parseStringToHourMinute("11:00"))
	expected_C.addStation("Ipswitch")

	var expected_D = NewTimetable(parseStringToHourMinute("8:00"), parseStringToHourMinute("9:00"))
	expected_D.addStation("Stevenage")
	expected_D.addStation("Baldock")
	expected_D.addStation("Ipswitch")

	expected := []*Timetable{expected_A, expected_B, expected_C, expected_D}

	for n, table := range tables {
		sliced := timetable.getSlicedTimetable(table.from, table.to)
		if !reflect.DeepEqual(sliced, expected[table.expectedTimetableIndex]) {
			jsonS, _ := json.Marshal(sliced)
			jsonE, _ := json.Marshal(expected[table.expectedTimetableIndex])
			t.Errorf("Error on dataset entry %d. For timetable between %s - %s , expected index %d " +
				"\n Original %s \n Expected %s ", n,
				table.from.Format("15:04"),
				table.to.Format("15:04"),
				table.expectedTimetableIndex,
				jsonS,
				jsonE)
		}
	}
}