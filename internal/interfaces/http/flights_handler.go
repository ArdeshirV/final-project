package http

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/the-go-dragons/final-project/internal/domain"
	"github.com/the-go-dragons/final-project/pkg/mock_api"
)

// Test ArdeshirV code by:
//                 Get all flights: http://localhost:8080/flights
//   Get flights with specified ID: http://localhost:8080/flights?flightno=VN931
// Get flights from A to b in time: http://localhost:8080/flights?city_a=New%20York&city_b=Paris&time=2023-06-14

// TODO: Test Amir Hosein code by:
// http://localhost:8080/flights?...

type Flights []domain.Flight
type SortOption int

const (
	SortByPrice SortOption = iota
	SortByDepartureDatetime
	SortByArrivalDatetime
	SortByDurationDatetime
)

const (

	// sort types
	AscendingSort  = "asc"
	DescendingSort = "desc"

	// params constant
	ParamTime              = "time"
	ParamCityA             = "city_a"
	ParamCityB             = "city_b"
	ParamCommand           = "command"
	ParamFlightNo          = "flightno"
	ParamMinimumCapacity   = "min_capacity"
	ParamDepartureDateTime = "depature_datetime"
	ParamArriveDateTime    = "arrive_datetime"
	ParamAirplane          = "airplane"
	ParamAirline           = "airline"

	// sort params
	ParamSortPrice            = "sort_price"
	ParamSortDuration         = "sort_duration"
	ParamSortArriveDatetime   = "sort_arrive_datetime"
	ParamSortDepatureDatetime = "sort_depature_datetime"
)

func FlightsRoute(e *echo.Echo) {
	e.GET("/flights", flightsHandler)
}

func flightsHandler(ctx echo.Context) error {
	data := make(Flights, 0)

	flightNo := ctx.QueryParam(ParamFlightNo)
	minCapacity := ctx.QueryParam(ParamMinimumCapacity)
	depatureDatetime := ctx.QueryParam(ParamDepartureDateTime)
	arriveDatetime := ctx.QueryParam(ParamArriveDateTime)
	airplane := ctx.QueryParam(ParamAirplane)
	airline := ctx.QueryParam(ParamAirline)
	priceSort := ctx.QueryParam(ParamSortPrice)
	durationSort := ctx.QueryParam(ParamSortDuration)
	arriveDatetimeSort := ctx.QueryParam(ParamSortArriveDatetime)
	depatureDatetimeSort := ctx.QueryParam(ParamSortDepatureDatetime)

	if flightNo != "" {
		fliteredFlight, err := mock_api.GetFlightsByFlightNo(flightNo)
		if err != nil {
			return echoJSON(ctx, http.StatusBadRequest, APIResponse{ Message: fmt.Sprintf("%v", err) })
		}

		return echoJSON(ctx, http.StatusOK, fliteredFlight)
	}

	cityA := ctx.QueryParam(ParamCityA)
	cityB := ctx.QueryParam(ParamCityB)
	timeD := ctx.QueryParam(ParamTime)

	if timeD != "" || cityA != "" || cityB != "" {
		errMsg := ""
		dataIsNotEnough := false

		if timeD == "" {
			dataIsNotEnough = true
			errMsg += "\"time\" is not defined correctly. "
		}

		if cityA == "" {
			dataIsNotEnough = true
			errMsg += "\"city_a\" is not defined correctly. "
		}

		if cityB == "" {
			dataIsNotEnough = true
			errMsg += "\"city_b\" is not defined correctly. "
		}

		if dataIsNotEnough {
			return echoJSON(ctx, http.StatusBadRequest, APIResponse{ Message: errMsg })
		}

		filteredFlights, err := mock_api.GetFlightsFromA2B(timeD, cityA, cityB)

		if err != nil {
			return echoJSON(ctx, http.StatusBadRequest, APIResponse{ Message: fmt.Sprintf("%v", err) })
		}

		data = filteredFlights
	} 

	if len(data) == 0 {
		allFlights, err := mock_api.GetFlights()

		data = allFlights
	
		if err != nil {
			return echoJSON(ctx, http.StatusBadRequest, APIResponse{ Message: fmt.Sprintf("%v", err) })
		}
	}
	
	if minCapacity != "" {
			
		numberMinCapacity, err := strconv.Atoi(minCapacity)

		if err != nil {
			return echoJSON(ctx, http.StatusBadRequest, APIResponse{ Message: fmt.Sprintf("%v", err) })
		}

		data, err = data.FilterFlightsByMinimumCapacity(numberMinCapacity)

		if err != nil {
			return echoJSON(ctx, http.StatusBadRequest, APIResponse{ Message: fmt.Sprintf("%v", err) })
		}
	}

	if airplane != "" {
			
		airplaneId, err := strconv.Atoi(airplane)

		if err != nil {
			return echoJSON(ctx, http.StatusBadRequest, APIResponse{ Message: fmt.Sprintf("%v", err) })
		}

		data, err = data.FilterFlightsByAirplaneId(airplaneId)

		if err != nil {
			return echoJSON(ctx, http.StatusBadRequest, APIResponse{ Message: fmt.Sprintf("%v", err) })
		}
	}

	if airline != "" {
			
		airlineId, err := strconv.Atoi(airline)

		if err != nil {
			return echoJSON(ctx, http.StatusBadRequest, APIResponse{ Message: fmt.Sprintf("%v", err) })
		}

		data, err = data.FilterFlightsByAirlineId(airlineId)

		if err != nil {
			return echoJSON(ctx, http.StatusBadRequest, APIResponse{ Message: fmt.Sprintf("%v", err) })
		}
	}

	if depatureDatetime != "" || arriveDatetime != "" {
		result, err := data.FilterFlightsByDepatureTimeAndArriveTime(depatureDatetime, arriveDatetime)

		if err != nil {
			return echoJSON(ctx, http.StatusBadRequest, APIResponse{ Message: fmt.Sprintf("%v", err) })
		}

		data = result
	}

	fmt.Printf("ParamSortPrice: %v\n", ParamSortPrice)

	data = data.ApplySort(ParamSortPrice, priceSort).ApplySort(ParamSortDuration, durationSort).ApplySort(ParamSortArriveDatetime, arriveDatetimeSort).ApplySort(ParamSortDepatureDatetime, depatureDatetimeSort)

	return echoJSON(ctx, http.StatusOK, data)
}

func (flights Flights) FilterFlightsByMinimumCapacity(minimumCapacity int) (Flights, error) {

	filteredFlights := make(Flights, 0)

	for _, flight := range flights {
		if flight.RemainingCapacity >= minimumCapacity {
			filteredFlights = append(filteredFlights, flight)
		}
	}

	return filteredFlights, nil
}

func (flights Flights) FilterFlightsByDepatureTimeAndArriveTime(depatureDatetime string, arriveDateTime string) (Flights, error) {
	filteredFlights := make(Flights, 0)

	var parsedDepatureDatetime time.Time
	var parsedArriveDatetime time.Time
	var err error

	if depatureDatetime != "" {
		parsedDepatureDatetime, err = time.Parse("2006-01-02T15:04:05Z", depatureDatetime)
		if err != nil {
			return nil, err
		}
	}

	if arriveDateTime != "" {
		parsedArriveDatetime, err = time.Parse("2006-01-02T15:04:05Z", arriveDateTime)
		if err != nil {
			return nil, err
		}
	}

	for _, flight := range flights {
		if (depatureDatetime == "" || flight.DepartureTime == parsedDepatureDatetime) &&
			(arriveDateTime == "" || flight.ArrivalTime == parsedArriveDatetime) {
			filteredFlights = append(filteredFlights, flight)
		}
	}

	return filteredFlights, nil
}

func (flights Flights) FilterFlightsByAirplaneId(airplaneId int) (Flights, error) {

	filteredFlights := make(Flights, 0)

	for _, flight := range flights {
		fmt.Printf("flight.AirplaneID: %v\n", flight.AirplaneID)
		if int(flight.AirplaneID) == airplaneId {
			filteredFlights = append(filteredFlights, flight)
		}
	}

	return filteredFlights, nil
}

func (flights Flights) FilterFlightsByAirlineId(airlineId int) (Flights, error) {

	filteredFlights := make(Flights, 0)

	for _, flight := range flights {
		if int(flight.Airplane.AirlineID) == airlineId {
			filteredFlights = append(filteredFlights, flight)
		}
	}

	return filteredFlights, nil
}

func (flights Flights) SortBy(sortOption SortOption, ascending bool) Flights {
	switch sortOption {
		case SortByPrice:{
				sort.Slice(flights, func(i, j int) bool {
					if ascending {
						return flights[i].Price < flights[j].Price
					} else {
						return flights[i].Price > flights[j].Price
					}
				})
				break
			}
		case SortByDepartureDatetime: {
			sort.Slice(flights, func(i, j int) bool {
				if ascending {
					return flights[i].DepartureTime.Before(flights[j].DepartureTime)
				} else {
					return flights[i].DepartureTime.After(flights[j].DepartureTime)
				}
			})
			break
		}
		case SortByArrivalDatetime: {
			sort.Slice(flights, func(i, j int) bool {
				if ascending {
					return flights[i].ArrivalTime.Before(flights[j].ArrivalTime)
				} else {
					return flights[i].ArrivalTime.After(flights[j].ArrivalTime)
				}
			})
			break
		}
		case SortByDurationDatetime: {
			sort.Slice(flights, func(i, j int) bool {
				durationI := flights[i].ArrivalTime.Sub(flights[i].DepartureTime)
				durationJ := flights[j].ArrivalTime.Sub(flights[j].DepartureTime)
				if ascending {
					return durationI < durationJ
				} else {
					return durationI > durationJ
				}
			})
			break
		}
	}

	return flights
}

func (flights Flights) ApplySort(sortName string, sortValue string) Flights {
	// newFlights := make(Flights, 0)
	var newFlights Flights = flights
	isSortAscending := sortValue == "asc"

	fmt.Printf("sortName: %v %v\n", sortName, sortValue !="")

	if sortName == ParamSortPrice && sortValue != "" {
		newFlights = flights.SortBy(SortByPrice, isSortAscending)
	}

	if sortName == ParamSortDuration && sortValue != "" {
		newFlights = flights.SortBy(SortByDurationDatetime, isSortAscending)
	}

	if sortName == ParamSortArriveDatetime && sortValue != "" {
		newFlights = flights.SortBy(SortByArrivalDatetime, isSortAscending)
	}

	if sortName == ParamSortDepatureDatetime && sortValue != "" {
		newFlights = flights.SortBy(SortByDepartureDatetime, isSortAscending)
	}

	return newFlights
}
