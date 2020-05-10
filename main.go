package main

import (
	"bufio"
	"derivco/entities"
	"fmt"
	"log"
	"os"
	"time"
)

const (
	STATIONS_REPO = "assets/tube.csv"
	DRONE_6043_ROUTE_REPO = "assets/6043.csv"
	DRONE_5937_ROUTE_REPO = "assets/5937.csv"
	DISPATCHER_DURATION_SECONDS = 1380000
	START_TIME = "07:47:50"
	STOP_TIME = "08:10:00"
	TIME_LAYOUT = "03:04:05"
	DRON_A_ID = 6043
	DRON_B_ID = 5937
)

var (
	currentStations []entities.Station
	drone_6043_route []entities.DroneRoute
	drone_5937_route []entities.DroneRoute
)

func getAllDroneRoutes() {
	entities.GetDroneRoute(&drone_6043_route, DRONE_6043_ROUTE_REPO)
	entities.GetDroneRoute(&drone_5937_route, DRONE_5937_ROUTE_REPO)
}

func droneControl(id int, commandReceived <-chan entities.DroneCommand, result chan<- entities.DroneCommandResult ) {
	command := <- commandReceived
	if (command.Route == entities.DroneRoute{} ) {
		log.Fatalln("Received command is wrong")
	}

	if (command.Route.Id != id) {
		return
	}

	fmt.Printf("[Drone %d][command received]: %d seconds to reach the target\n", command.Route.Id, command.CountdownTime)

	if (command.CountdownTime > 0) {
		timer1 := time.NewTimer(time.Duration(command.CountdownTime) * time.Second)
		<-timer1.C
	}

	// station := entities.FindStationByGPS(currentStations, command.Route.Lat, command.Route.Lon)
	station := entities.GetNearStation(currentStations, command.Route.Lat, command.Route.Lon)
	var commandResult = entities.DroneCommandResult{
		DroneId: id,
	}

	if (station == entities.Station{}) {
		fmt.Printf("[Drone %d][info] arrives at lat: %g and lon: %g but any station was found. \n", command.Route.Id, command.Route.Lat, command.Route.Lon)
		commandResult.HasReport = false
	} else {
		commandResult.HasReport = true
		commandResult.Station = station.Name
		commandResult.CurrentTraffic = entities.GetRandomTrafficCondition()
	}

	result <- commandResult
}

func startDispatcher() {

	// channels
	var commands = make(chan entities.DroneCommand)
	var results = make(chan entities.DroneCommandResult)

	drn6043_route_seq := 0
	drn5937_route_seq := 0

	startTime, err := time.Parse(TIME_LAYOUT, START_TIME)
	if err != nil {
		log.Fatal(err)
	}

	stopTime, err := time.Parse(TIME_LAYOUT, STOP_TIME)
	if err != nil {
		log.Fatal(err)
	}

	// Start drone 6043
	go droneControl(DRON_A_ID, commands, results)
	drn6043_next_route := entities.GetNextDroneRoute(drone_6043_route, drn6043_route_seq)
	var newCommand = entities.DroneCommand{
		Route:         drn6043_next_route,
		CountdownTime: int(drn6043_next_route.Time.Sub(startTime).Seconds()),
	}
	commands <- newCommand

	// Start drone 5937
	go droneControl(DRON_B_ID, commands, results)
	drn5937_next_route := entities.GetNextDroneRoute(drone_5937_route, drn5937_route_seq)
	newCommand = entities.DroneCommand{
		Route:         drn5937_next_route,
		CountdownTime: int(drn5937_next_route.Time.Sub(startTime).Seconds()),
	}
	commands <- newCommand

	// Process all drones commands
	ticker := time.NewTicker(1000 * time.Millisecond)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			case _ = <-ticker.C:

				select {
					case commandResult := <-results:

					if (commandResult.HasReport) {
						fmt.Printf("[Drone %d][Time %s] Traffic at %s is %s\n", commandResult.DroneId, startTime.Format(TIME_LAYOUT), commandResult.Station, commandResult.CurrentTraffic)
					}

					if (commandResult.DroneId == DRON_A_ID) {
						drn6043_route_seq = drn6043_route_seq + 1
						drn6043_next_route = entities.GetNextDroneRoute(drone_6043_route, drn6043_route_seq)

						seconds := int(drn6043_next_route.Time.Sub(startTime).Seconds())
						if (seconds < 0) {
							seconds = 0
						}
						newCommand = entities.DroneCommand{
							Route:         drn6043_next_route,
							CountdownTime: seconds,
						}
						go droneControl(DRON_A_ID, commands, results)
						commands <- newCommand
					} else if (commandResult.DroneId == DRON_B_ID) {
						drn5937_route_seq = drn5937_route_seq + 1
						drn5937_next_route = entities.GetNextDroneRoute(drone_5937_route, drn5937_route_seq)

						seconds := int(drn6043_next_route.Time.Sub(startTime).Seconds())
						if (seconds < 0) {
							seconds = 0
						}

						newCommand = entities.DroneCommand{
							Route:         drn5937_next_route,
							CountdownTime: seconds,
						}

						go droneControl(DRON_B_ID, commands, results)
						commands <- newCommand
					}
					default:
				}

				startTime = startTime.Add(time.Second * 1)
				// fmt.Println("[Dispatcher][info] Clock ", startTime.Format(TIME_LAYOUT))
			}

			if (startTime.After(stopTime)) {
				fmt.Println("SHUTDOWN")
				ticker.Stop()
				done <- true
			}
		}
	}()

	close(commands)
	close(results)
}

func main() {
	fmt.Println("[App] Derivco test start")
	fmt.Println("[Info] Press any key to exit...")
	entities.GetStationsFromRepo(&currentStations, STATIONS_REPO)

	getAllDroneRoutes()

	startDispatcher()

	reader := bufio.NewReader(os.Stdin)
	_, _ = reader.ReadString('\n')
}