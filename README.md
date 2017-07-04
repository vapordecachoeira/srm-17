17
==

A timetable service for a train service, that takes routes, stations, and schedules as input, and produces an updated 60-minute time-windowed departure board for each station.


Build 
-----

Build

    go build 17.go
    
Run

    ./17
    
    
Testing
-----

    go test
    
    
Usage / Endpoints
-----

#### Add Sample data

    localhost:8080/addSample

#### Add a stop, example:

    localhost:8080/add?&routeName=London-Cambridge&station=Baldock&time=10:00

#### Display the timetable for the next 60 min from now
 
As text
 
    localhost:8080/
    
As Json
 
    localhost:8080/?format=json 

#### Display the timetable for the next 60 min from a specific time
 
As text
    
    localhost:8080/?from=12:30
    

As Json
    
    localhost:8080/?from=12:30&format=json




