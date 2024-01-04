package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
	"sync"
	"strconv"
	"encoding/json"
	"database/sql"
	_ "github.com/lib/pq"
)

func getStopsDetails() {
	//Make the request for the stops details
	apiURL := "https://stibmivb.opendatasoft.com/api/explore/v2.1/catalog/datasets/stop-details-production/records?limit=100&offset="
	for i := 0; i < 2600; i += 100 {
		response, err := http.Get(apiURL + strconv.Itoa(i))
		if err != nil {
			fmt.Println("Error making API request:", err)
			return
		}

		// Read the API response
		responseBody, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Println("Error reading API response:", err)
			return
		}

		// Create a directory named "api_responses" if it doesn't exist
		dirName := "stopDetails"
		if _, err := os.Stat(dirName); os.IsNotExist(err) {
			err := os.Mkdir(dirName, os.ModePerm)
			if err != nil {
				fmt.Println("Error creating directory:", err)
				return
			}
		}

		filename := dirName + "/stopsDetails" + fmt.Sprintf("%d", i) + "-" + fmt.Sprintf("%d", i+100) + ".json"

		// Create or open the file for writing in the directory
		file, err := os.Create(filename)
		if err != nil {
			fmt.Println("Error creating file:", err)
			return
		}
		defer file.Close()

		// Write the API response to the file
		_, err = file.Write(responseBody)
		if err != nil {
			fmt.Println("Error writing to file:", err)
			return
		}

		fmt.Printf("API response saved to %s\n", filename)
	}

}

func makeWaitingTimeRequests() {
	var wg sync.WaitGroup

	apiURLFormat  := "https://stibmivb.opendatasoft.com/api/explore/v2.1/catalog/datasets/waiting-time-rt-production/records?limit=100&offset=%d&apikey=abdb9e039bc041977c9c251402ad6d967fef8530dedcacdd2c91385e"
	totalRequests := 33

	// Generate a timestamp when the function is called
	timestamp := time.Now().Format("2006-01-02_15-04-05")

	// Create a directory named "waitingTimes" if it doesn't exist
	dirName := "waitingTimes"
	if _, err := os.Stat(dirName); os.IsNotExist(err) {
		err := os.Mkdir(dirName, os.ModePerm)
		if err != nil {
			fmt.Println("Error creating directory:", err)
			return
		}
	}

	offset := 0

	for i := 0; i < totalRequests; i++ {
		wg.Add(1)
		go func(offset int) {
			defer wg.Done()

			apiURL := fmt.Sprintf(apiURLFormat, offset)

			response, err := http.Get(apiURL)
			if err != nil {
				fmt.Println("Error making API request:", err)
				return
			}
			defer response.Body.Close()

			responseBody, err := ioutil.ReadAll(response.Body)
			if err != nil {
				fmt.Println("Error reading API response:", err)
				return
			}

			filename := fmt.Sprintf("%s/api_response_%s_offset%d.json", dirName, timestamp, offset)

			file, err := os.Create(filename)
			if err != nil {
				fmt.Println("Error creating file:", err)
				return
			}
			defer file.Close()

			_, err = file.Write(responseBody)
			if err != nil {
				fmt.Println("Error writing to file:", err)
				return
			}

			fmt.Printf("API response saved to %s\n", filename)
		}(offset)

		offset += 100
		if offset > 3300 {
			offset = 3300
		}
	}

	// Wait for all API requests to finish
	wg.Wait()
}


type ResponseData struct {
    TotalCount int `json:"total_count"`
    Results    []struct {
        LineID           string `json:"lineid"`
        VehiclePositions string `json:"vehiclepositions"`
    } `json:"results"`
}


func processRtPositionAPIResponse(mapOfDirections map[string]string) (map[string]string) {
	brusselsLocation, err := time.LoadLocation("Europe/Brussels")

    // Make a GET request to an external API
    response, err := http.Get("https://stibmivb.opendatasoft.com/api/explore/v2.1/catalog/datasets/vehicle-position-rt-production/records?limit=80&apikey=abdb9e039bc041977c9c251402ad6d967fef8530dedcacdd2c91385e")
    if err != nil {
        fmt.Println("Error making API request:", err)
        
    }
    defer response.Body.Close()

    // Read the API response
    responseBody, err := ioutil.ReadAll(response.Body)
    if err != nil {
        fmt.Println("Error reading API response:", err)
        
    }

    // Parse the JSON response
    var data ResponseData
    if err := json.Unmarshal(responseBody, &data); err != nil {
        fmt.Println("Error unmarshaling JSON:", err)
        
    }

    // Establish a database connection (adjust the connection string)
    db, err := sql.Open("postgres", "user=postgres password=postgres dbname=brussels sslmode=disable")
    if err != nil {
        fmt.Println("Error connecting to the database:", err)
        
    }
    defer db.Close()

    // Iterate through the results and process the data
    for _, result := range data.Results {
        lineID := result.LineID

        // Parse the "vehiclepositions" field, which contains an array of position data
        var positions []struct {
            DirectionID      string `json:"directionId"`
            DistanceFromPoint int    `json:"distanceFromPoint"`
            PointID          string `json:"pointId"`
        }

        if err := json.Unmarshal([]byte(result.VehiclePositions), &positions); err != nil {
            fmt.Println("Error unmarshaling vehicle positions:", err)
            continue
        }
		
		count := 1
        // Iterate through vehicle positions for this line
        for _, position := range positions {
            directionID := position.DirectionID
            distanceFromPoint := position.DistanceFromPoint
            pointID := position.PointID
			

            // Execute your SQL queries here using the database connection (db)
			
			if directionID != pointID {

				// Query 0: check if the vehicle is in deviation
				query0 := `SELECT EXISTS (
								SELECT 1
								FROM stops_by_lines
								WHERE lineid = $2
								AND stops[array_upper(stops, 1)] = $3
								AND $1 = ANY(stops)
							) AS result;
				`
				
				var is_not_deviated bool
				err := db.QueryRow(query0, pointID, lineID, directionID).Scan(&is_not_deviated)
				if err != nil {
					if err == sql.ErrNoRows {
						fmt.Println("Query 0 returned no result for LineID: %s and pointID: %s for directionID: %s", lineID,pointID,directionID)
					} else if err != nil {
						fmt.Println("Error executing query 0:", err)
					}
					continue
				}

				if is_not_deviated == false{

					tripid := lineID + "_" + strconv.Itoa(count)
					count++

					

					//The bus is in a deviation we simply update the trip  with the coodinate of the stop
					// and mark the trip as deviated in the table.

					// Query 2: Get points of the last visited stop
					query0_2 := `WITH points AS (
						SELECT ST_SetSRID(ST_MakePoint(p1.stop_lon,p1.stop_lat),4326) as point1	
						FROM stops as p1
						WHERE p1.stop_id LIKE '%' || $1 || '%'
					)
					SELECT point1 FROM points`

					var point1 sql.NullString
					err = db.QueryRow(query0_2, pointID).Scan(&point1)
					if err != nil {
						if err == sql.ErrNoRows {
							fmt.Println("Query 0_2 returned no result for LineID: %s and pointID: %s for directionID: %s", lineID,pointID,directionID)
						} else if err != nil {
							fmt.Println("Error executing query 0_2:", err)
						}
						continue
					}

					

					currentDate := time.Now().In(brusselsLocation)
					year := currentDate.Year()
					month := int(currentDate.Month())
					day := currentDate.Day()
					hour := currentDate.Hour()
					minute := currentDate.Minute()
					second := currentDate.Second()
					formattedTime := fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d", year, month, day, hour, minute, second)
					
					date := strconv.Itoa(year) + "-" + strconv.Itoa(month) + "-" + strconv.Itoa(day)

					tgeompoint := point1.String + "@" + formattedTime
					fmt.Println(tgeompoint)

					if val, exist := mapOfDirections[tripid] ; exist {
						if val != directionID {
							//query to update current in the table
							queryUpdateCurrent := `UPDATE stib_trips
													SET current = false
													WHERE day = $1 AND lineid = $2 AND tripid = $3 AND directionid = $4 AND current = true`
													
							_, err := db.Exec(queryUpdateCurrent, date, lineID, tripid, directionID)

							if err != nil {
								if err == sql.ErrNoRows {
									fmt.Println("Query update current returned no result for LineID: %s and pointID: %s for directionID: %s", lineID,pointID,directionID)
								} else if err != nil {
									fmt.Println("Error executing query update current:", err)
									fmt.Println("Query update current returned no result for LineID: %s and pointID: %s for directionID: %s", lineID,pointID,directionID)
									fmt.Println("===========================================")
								}
								continue
							}
							mapOfDirections[tripid] = directionID
						}
					}else {
						mapOfDirections[tripid] = directionID
					}

					query0_3 := `SELECT insert_or_update_stib_trip($1, $2, $3, $4, $5, $6, $7, $8, $9::tgeompoint);`
					
									
					_,err = db.Exec(query0_3, date, lineID, tripid, directionID,formattedTime,formattedTime,true,true, tgeompoint)
					if err != nil {
						if err == sql.ErrNoRows {
							fmt.Println("Query 0_3 returned no result for LineID: %s and pointID: %s for directionID: %s", lineID,pointID,directionID)
						} else if err != nil {
							fmt.Println("Error executing query 0_3:", err)
							fmt.Println("Query 0_3 returned no result for LineID: %s and pointID: %s for directionID: %s", lineID,pointID,directionID)
							fmt.Println("===========================================")
						}
						continue
					} else {
						fmt.Println("Trip  added or updated sucessfully from a devation", lineID,pointID,directionID)
					}

				} else {

					tripid := lineID + "_" + strconv.Itoa(count)
					count++
					
					// Query 1: Get the next stop ID
					query1 := `WITH get_real_terminus AS (
						SELECT stopid as stop  from  stopdetails where stopid LIKE '%' || $1 || '%' 
						)
						SELECT stop FROM get_real_terminus`

					var terminusID sql.NullString
					err := db.QueryRow(query1, directionID).Scan(&terminusID)
					if err != nil {
						if err == sql.ErrNoRows {
							fmt.Println("Query 1 returned no result for LineID: %s and pointID: %s for directionID: %s", lineID,pointID,directionID)
						} else if err != nil {
							fmt.Println("Error executing query 1:", err)
						}
						continue
					}



					// Query 2: Get points of the last visited stop
					query2 := `WITH points AS (
						SELECT ST_SetSRID(ST_MakePoint(p1.stop_lon,p1.stop_lat),4326) as point1	
						FROM stops as p1
						WHERE p1.stop_id LIKE '%' || $1 || '%'
					)
					SELECT point1 FROM points`

					var point1 sql.NullString
					err = db.QueryRow(query2, pointID).Scan(&point1)
					if err != nil {
						if err == sql.ErrNoRows {
							fmt.Println("Query 2 returned no result for LineID: %s and pointID: %s for directionID: %s", lineID,pointID,directionID)
						} else if err != nil {
							fmt.Println("Error executing query 2:", err)
						}
						continue
					}
				
					// Query 3: Calculate the fractions of the trajectory
					query3 := `WITH fractions AS (
						SELECT
							ST_LineLocatePoint(trajectory.trajectory, $1) AS start_fraction
						FROM
							terminus_shapes AS trajectory 
						WHERE
							trajectory.route_short_name = $2 AND trajectory.stopid = $3
					)
					SELECT start_fraction FROM fractions`

					var startFraction sql.NullFloat64
					err = db.QueryRow(query3, point1, lineID, terminusID).Scan(&startFraction)
					if err != nil {
						if err == sql.ErrNoRows {
							fmt.Println("Query 3 returned no result for LineID: %s and PointID: %s for DirectionID: %s", lineID,pointID,directionID)
						} else if err != nil {
							fmt.Println("Error executing query 3:", err)
						}
						continue
					}

					// Query 4: Extract the sub-trajectory
					query4 := `WITH sub_trajectory AS (
						SELECT
							ST_SetSRID(ST_LineSubstring(trajectory.trajectory, $1,1),4326) AS sub_trajectory
						FROM
							terminus_shapes AS trajectory
						WHERE 
							trajectory.route_short_name = $2 AND trajectory.stopid = $3
						limit 1
					)
					SELECT ST_LineInterpolatePoint(st.sub_trajectory, $4/ST_Length(st.sub_trajectory::geography))
					FROM sub_trajectory  as st`

					var interpolatedPoint sql.NullString
					err = db.QueryRow(query4, startFraction, lineID, terminusID,distanceFromPoint).Scan(&interpolatedPoint)
					if err != nil {
						if err == sql.ErrNoRows {
							fmt.Println("Query 4 returned no result for LineID: %s and pointID: %s for directionID: %s", lineID,pointID,directionID)
						} else if err != nil {
							fmt.Println("Error executing query 4:", err)
							fmt.Println("Query 4 returned no result for LineID: %s and pointID: %s for directionID: %s", lineID,pointID,directionID)
							fmt.Println("===========================================")
						}
						continue
					}
					
					currentDate := time.Now().In(brusselsLocation)
					year := currentDate.Year()
					month := int(currentDate.Month())
					day := currentDate.Day()
					hour := currentDate.Hour()
					minute := currentDate.Minute()
					second := currentDate.Second()
					formattedTime := fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d", year, month, day, hour, minute, second)
					
					date := strconv.Itoa(year) + "-" + strconv.Itoa(month) + "-" + strconv.Itoa(day)

					tgeompoint := interpolatedPoint.String + "@" + formattedTime
					fmt.Println(tgeompoint)


					if val, exist := mapOfDirections[tripid] ; exist {
						if val != directionID {
							//queery  to update current in the table
							queryUpdateCurrent := `UPDATE stib_trips
													SET current = false
													WHERE day = $1 AND lineid = $2 AND tripid = $3 AND directionid = $4 AND current = true`
													
							_, err := db.Exec(queryUpdateCurrent, date, lineID, tripid, directionID)

							if err != nil {
								if err == sql.ErrNoRows {
									fmt.Println("Query update current returned no result for LineID: %s and pointID: %s for directionID: %s", lineID,pointID,directionID)
								} else if err != nil {
									fmt.Println("Error executing query update current:", err)
									fmt.Println("Query update current returned no result for LineID: %s and pointID: %s for directionID: %s", lineID,pointID,directionID)
									fmt.Println("===========================================")
								}
								continue
							}
							mapOfDirections[tripid] = directionID
						}
					}else {
						mapOfDirections[tripid] = directionID
					}

					query5 := `SELECT insert_or_update_stib_trip($1, $2, $3, $4, $5, $6, $7, $8, $9::tgeompoint);`
					
									
					_,err = db.Exec(query5, date, lineID, tripid, directionID, formattedTime,formattedTime, false, true, tgeompoint)
					if err != nil {
						if err == sql.ErrNoRows {
							fmt.Println("Query 5 returned no result for LineID: %s and pointID: %s for directionID: %s", lineID,pointID,directionID)
						} else if err != nil {
							fmt.Println("Error executing query 5:", err)
							fmt.Println("Query 5 returned no result for LineID: %s and pointID: %s for directionID: %s", lineID,pointID,directionID)
							fmt.Println("===========================================")
						}
						continue
					} else {
						fmt.Println("Trip  added or updated sucessfully in the usual  way", lineID,tripid,pointID,directionID)
					}

				}

			} else {

				fmt.Println("The ID of stop is the same as for the terminus no need to  interpolate")
				// Query 6: Get coordinated of the terminus
				query6 := `WITH points AS (
					SELECT ST_SetSRID(ST_MakePoint(p1.stop_lon,p1.stop_lat),4326) as point1	
					FROM stops as p1
					WHERE p1.stop_id LIKE '%' || $1 || '%'
				)
				SELECT point1 FROM points`

				var point1 sql.NullString
				err = db.QueryRow(query6, pointID).Scan(&point1)
				if err != nil {
					if err == sql.ErrNoRows {
						fmt.Println("Query 6 returned no result for LineID: %s and pointID: %s for directionID: %s", lineID,pointID,directionID)
					} else if err != nil {
						fmt.Println("Error executing query 6:", err)
						fmt.Println("Query 6 returned no result for LineID: %s and pointID: %s for directionID: %s", lineID,pointID,directionID)
						fmt.Println("===========================================")
					}
					continue
				}
				
				// Process the results from query 6
				fmt.Println("Query6")
				fmt.Printf("LineID: %s, DirectionID: %s, Interpolated Point: %s\n", lineID, directionID, point1.String)

				tripid := lineID + "_" +strconv.Itoa(count)
				count++
				
				currentDate := time.Now().In(brusselsLocation)
				year := currentDate.Year()
				month := int(currentDate.Month())
				day := currentDate.Day()
				hour := currentDate.Hour()
				minute := currentDate.Minute()
				second := currentDate.Second()
				formattedTime := fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d", year, month, day, hour, minute, second)
				
				date := strconv.Itoa(year) + "-" + strconv.Itoa(month) + "-" + strconv.Itoa(day)

				tgeompoint := point1.String + "@" + formattedTime
				fmt.Println(tgeompoint)
				
				//fmt.Println(mapOfDirections)
				
				if val, exist := mapOfDirections[tripid] ; exist {
					if val != directionID {
						//queery  to update current in the table
						queryUpdateCurrent := `UPDATE stib_trips
												SET current = false
												WHERE day = $1 AND lineid = $2 AND tripid = $3 AND directionid = $4 AND current = true`
												
						_, err := db.Exec(queryUpdateCurrent, date, lineID, tripid, directionID)

						if err != nil {
							if err == sql.ErrNoRows {
								fmt.Println("Query update current returned no result for LineID: %s and pointID: %s for directionID: %s", lineID,pointID,directionID)
							} else if err != nil {
								fmt.Println("Error executing query update current:", err)
								fmt.Println("Query update current returned no result for LineID: %s and pointID: %s for directionID: %s", lineID,pointID,directionID)
								fmt.Println("===========================================")
							}
							continue
						}
						mapOfDirections[tripid] = directionID
					}
				}else {
					mapOfDirections[tripid] = directionID
				}

				query7 := `SELECT insert_or_update_stib_trip($1, $2, $3, $4, $5, $6, $7, $8, $9::tgeompoint);`
				
				_,err = db.Exec(query7, date, lineID, tripid, directionID, formattedTime,formattedTime, false, true, tgeompoint)
				if err != nil {
					if err == sql.ErrNoRows {
						fmt.Println("Query 7 returned no result for LineID: %s and pointID: %s for directionID: %s", lineID,pointID,directionID)
					} else if err != nil {
						fmt.Println("Error executing query 7:", err)
						fmt.Println("Query 7 returned no result for LineID: %s and pointID: %s for directionID: %s", lineID,pointID,directionID)
						fmt.Println("===========================================")
					}
					continue
				} else {
					fmt.Println("Trip  added or updated sucessfully when  point id is the terminus")
				}
			}
		}
    }

	return mapOfDirections
}


func main() {
	// Start a timer that triggers the API request every 20 seconds
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()
	mapOfDirections := make(map[string]string)

	// Initial API request
	//getStopsDetails() //run this only if it is the first  time that the code is executed
	mapOfDirections = processRtPositionAPIResponse(mapOfDirections)
	//makeWaitingTimeRequests()

	// Set up a goroutine to perform the API request on a timer
	go func() {

		for {
			select {
			case <-ticker.C:
				mapOfDirections = processRtPositionAPIResponse(mapOfDirections)
				//makeWaitingTimeRequests()
			}
		}
	}()

	// Start the server on port 8080
	fmt.Println("Server is running on :8080...")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, Go Server!")
	})
	http.ListenAndServe(":8080", nil)
}