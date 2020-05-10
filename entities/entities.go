package entities

import (
	"encoding/csv"
	"io"
	"log"
	"math"
	"math/rand"
	"os"
	"strconv"
	"time"
)

// Wheather enum
type Weather int
const (
	HEAVY = iota
	LIGHT
	MODERATE
)

type Station struct {
	Name string
	Lat  float64
	Lon  float64
}

type DroneRoute struct {
	Id int
	Seq int
	Lat  float64
	Lon  float64
	DateTime time.Time
	Time time.Time
}

type DroneCommand struct {
	Route DroneRoute
	CountdownTime int
}

type DroneCommandResult struct {
	DroneId int
	HasReport bool
	Station string
	CurrentTraffic	 string
}

// Returns an int >= min, < max
func randomInt(min, max int) int {
	rand.Seed(time.Now().UnixNano())
	return min + rand.Intn(max-min)
}

// Returns a random Traffic string
func GetRandomTrafficCondition() string {

	switch rndWeather := Weather(randomInt(0, 3)); rndWeather {
	case HEAVY:
		return "HEAVY"
	case LIGHT:
		return "LIGHT"
	case MODERATE:
		return "MODERATE"
	default:
		return "... UPS something was wrong ..."
	}
}

func GetStationsFromRepo(stations *[]Station, repoPath string) {
	// Open the file
	csvFile, err := os.Open(repoPath)
	if err != nil {
		log.Fatalln("Couldn't open the csv file", err)
	}

	// Parse the file
	r := csv.NewReader(csvFile)

	// Iterate through the records
	for {
		// Read each record from csv
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		lat, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			log.Fatal(err)
		}
		lon, err := strconv.ParseFloat(record[2], 64)
		if err != nil {
			log.Fatal(err)
		}

		*stations = append(*stations, Station {
			record[0],
			lat,
			lon,
		})
	}
}

func FindStationByGPS(stations []Station, lat float64, lon float64) Station {
	var station Station

	if stations == nil {
		log.Fatalln("Couldn't search inside a nil station")
	}

	for _, n := range stations {
		if lat == n.Lat && lon == n.Lon {
			return n
		}
	}

	return station
}

func ContainsStation(stations []Station, name string) bool {
	for _, n := range stations {
		if name == n.Name {
			return true
		}
	}

	return false;
}

func GetNextDroneRoute(droneRoutes []DroneRoute, seq int) DroneRoute {
	if droneRoutes == nil {
		log.Fatalln("Cannot be null")
	}

	if (seq > len(droneRoutes)) {
		log.Fatalln("drone route index overflow")
	}

	return droneRoutes[seq]
}

func GetDroneRoute(droneRoute *[]DroneRoute, routePath string) {
	// Open the file
	csvFile, err := os.Open(routePath)
	if err != nil {
		log.Fatalln("Couldn't open the csv file", err)
	}

	// Parse the file
	r := csv.NewReader(csvFile)

	seq := 1

	// Iterate through the records
	for {

		// Read each record from csv
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		id, err := strconv.Atoi(record[0])
		if err != nil {
			log.Fatal(err)
		}

		lat, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			log.Fatal(err)
		}
		lon, err := strconv.ParseFloat(record[2], 64)
		if err != nil {
			log.Fatal(err)
		}

		// time parse format layout
		layout := "2006-01-02 03:04:05"
		dateTime, err := time.Parse(layout, record[3])
		if err != nil {
			log.Fatal(err)
		}

		nextTime, err := time.Parse("03:04:05", dateTime.Format("15:04:05"))
		if err != nil {
			log.Fatal(err)
		}

		// append all the drone route entries
		*droneRoute = append(*droneRoute, DroneRoute {
			id,
			seq,
			lat,
			lon,
			dateTime,
			nextTime,
		})

		seq = seq + 1
	}
}

func GetNearStation(stations []Station, droneLat float64, droneLong float64) Station {
	for _, c  := range stations {
		if ( StationDistance(droneLat, droneLong, c.Lon, c.Lat, "K") <= 350) {
			return c
		}
	}

	return Station{}
}

func StationDistance(lat1 float64, lng1 float64, lat2 float64, lng2 float64, unit ...string) float64 {
	const PI float64 = 3.141592653589793

	radlat1 := float64(PI * lat1 / 180)
	radlat2 := float64(PI * lat2 / 180)

	theta := float64(lng1 - lng2)
	radtheta := float64(PI * theta / 180)

	dist := math.Sin(radlat1) * math.Sin(radlat2) + math.Cos(radlat1) * math.Cos(radlat2) * math.Cos(radtheta)

	if dist > 1 {
		dist = 1
	}

	dist = math.Acos(dist)
	dist = dist * 180 / PI
	dist = dist * 60 * 1.1515

	if len(unit) > 0 {
		if unit[0] == "K" {
			dist = (dist * 1.609344) / 1000
		} else if unit[0] == "N" {
			dist = dist * 0.8684
		}
	}

	return dist
}
