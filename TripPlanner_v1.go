package main

import (
    "fmt"
    "time"
    "bytes"
    "errors"
    "strings"
    "io/ioutil"
    "net/http"
    "encoding/json"
    "gopkg.in/mgo.v2"
    "math/rand"
    "github.com/jmoiron/jsonq"
    "gopkg.in/mgo.v2/bson"
    "github.com/julienschmidt/httprouter"
    "strconv"
)

const (
    sandboxURL = "https://sandbox-api.uber.com/v1"
    server_token = "sAVnmYBVi1OzrV1By5dJaKu4NN0boKl3atPXVtTk"
    product_id = "a1111c8c-c720-46c3-8534-2fcdd730040d"
    authorization_token = "Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzY29wZXMiOlsicHJvZmlsZSIsInJlcXVlc3RfcmVjZWlwdCIsInJlcXVlc3QiLCJoaXN0b3J5X2xpdGUiXSwic3ViIjoiM2NiZWVmNWUtOTU3Ni00MWUwLWI3ZGItZjYzMDY3ZDJiNzE3IiwiaXNzIjoidWJlci11czEiLCJqdGkiOiI3ZGNhNTYwMS1jYmE3LTQ3OGItOTIzMS1iNWRmYTA3YzAzOGMiLCJleHAiOjE0NTA0MzUyNzQsImlhdCI6MTQ0Nzg0MzI3NCwidWFjdCI6ImJNc0tXT3U2Q3lma0ZjbGt2MG16V1ZibVVvMlBQcSIsIm5iZiI6MTQ0Nzg0MzE4NCwiYXVkIjoiLVFUQmlyQU8yVkNzWEpPYWFpck05SXdjNVR4dU9QTzYifQ.IQpzaEX2hQri5H1_k-GifU5KJ1Klw7WjnBL-z0cOA96wvndNOmqWtCBg7YV_6NPmn5hyQR2IZvPSNmYxz4zv5EVQ69Nd3-ea-Z7Fu1UyYy4cxMINXLShNG3i7wveTlBMEnr6vb2WCj1CYUIKNqeCgkmYs_E1eTSSpnbm_vCK0dKqPjkprEYOaL3P6pDH2uFa2gWR4esYGqTiSTTO5UfFAlyJfjaBCg1JMW4v9c842InuqoVS3Es4RLB-T6wCcrncJZTtRR0CK6Ghb8UUpCJyAHIWJP49PQSy97SsIBezO3xOhNRbPSkK2fKVnSEnZ2lj3sP1xtmb3zjnnRVMFk2yjg"
)

var tripFinishedForTripIDs map[string]bool

var tripRequestIDAndCordsMapper map[string]map[string]string

type UserData struct {
    Id bson.ObjectId `json:"id" bson:"_id"`
    Name string `json:"name" bson:"name"`
    Address string `json:"address" bson:"address"`
    City string `json:"city" bson:"city"`
    State string `json:"state" bson:"state"`
    Zip string `json:"zip" bson:"zip"`
    Coordinate struct {
        Lat float64 `json:"lat" bson:"lat"`
        Lng float64 `json:"lng" bson:"lng"`
    } `json:"coordinate" bson:"coordinate"`
}

type tripRequest struct {
    StartingLocID string    `json:"starting_from_location_id"`
    LocationIds []string    `json:"location_ids"`
}

type TripResponse struct {
    Id bson.ObjectId `json:"id" bson:"_id"`
    Status string `json:"status" bson:"status"`
    Starting_from_location_id string `json:"starting_from_location_id" bson:"starting_from_location_id"`
    Best_route_location_ids []string `json:"best_route_location_ids" bson:"best_route_location_ids"`
    Total_uber_costs int `json:"total_uber_costs" bson:"total_uber_costs"`
    Total_uber_duration int `json:"total_uber_duration" bson:"total_uber_duration"`
    Total_distance float64 `json:"total_distance" bson:"total_distance"`
}

type TripBookingResponse struct {
    Id bson.ObjectId `json:"id" bson:"_id"`
    Status string `json:"status" bson:"status"`
    Starting_from_location_id string `json:"starting_from_location_id" bson:"starting_from_location_id"`
    Next_destination_location_id string `json:"next_destination_location_id" bson:"next_destination_location_id"`
    Best_route_location_ids []string `json:"best_route_location_ids" bson:"best_route_location_ids"`
    Total_uber_costs int `json:"total_uber_costs" bson:"total_uber_costs"`
    Total_uber_duration int `json:"total_uber_duration" bson:"total_uber_duration"`
    Total_distance float64 `json:"total_distance" bson:"total_distance"`
    Uber_wait_time_eta int `json:"uber_wait_time_eta" bson:"uber_wait_time_eta"`
}

type  locationCoordinate struct {
    id string
    latitude float64
    longitude float64
    visited bool
}

type uberInfoFromPoint struct {
    id string
    low_estimate int
    duration int
    distance float64
}

type jsonPostBody struct{
    Start_longitude string `json:"start_longitude"`
    Start_latitude string `json:"start_latitude"`
    Product_id string `json:"product_id"`
}
//To capture the response from the POST call to sandbox to fetch the Request ID
type postResponse struct {
    Status string
    Request_id string
    Eta int
    Surge_multiplier float64
}

type error struct {
    Error_message string `json:"error_message"`
}

func getSession() *mgo.Session {
    //Connect to mongo
    s, err := mgo.Dial("mongodb://kamlendrachauhan:cmpe273@ds045064.mongolab.com:45064/location_service")

    // Check if connection error, is mongo running?
    if err != nil {
        panic(err)
    }
    return s
}
//Get a Location - GET        /locations/{location_id}
func getLocations(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
    location_id :=  p.ByName("location_id")

    if !bson.IsObjectIdHex(location_id) {
        rw.WriteHeader(404)
        return
    }

    original_loc_id := bson.ObjectIdHex(location_id)

    returnObj := UserData{}

    if err := getSession().DB("location_service").C("location").FindId(original_loc_id).One(&returnObj); err != nil {
        rw.WriteHeader(404)
        return
    }

    uj, _ := json.Marshal(returnObj)

    // Write content-type, statuscode, payload
    rw.Header().Set("Content-Type", "application/json")
    rw.WriteHeader(200)
    fmt.Fprintf(rw, "%s", uj)
}

//Get a Trip - GET        /trips/{trip_id}
func getTrips(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
    trip_id :=  p.ByName("trip_id")

    if !bson.IsObjectIdHex(trip_id) {
        rw.WriteHeader(404)
        return
    }

    if _,found := tripRequestIDAndCordsMapper[trip_id]; !found{//if trip has been planned but not yet started
        original_trip_id := bson.ObjectIdHex(trip_id)

        returnObj := TripResponse{}

        if err := getSession().DB("location_service").C("trip").FindId(original_trip_id).One(&returnObj); err != nil {
            rw.WriteHeader(404)
            return
        }

        uj, _ := json.Marshal(returnObj)

        // Write content-type, statuscode, payload
        rw.Header().Set("Content-Type", "application/json")
        rw.WriteHeader(200)
        fmt.Fprintf(rw, "%s", uj)
    }else {//if trip has been started then checking the status
        original_trip_id := bson.ObjectIdHex(trip_id)

        returnObj := TripBookingResponse{}

        if err := getSession().DB("location_service").C("tripbooking").FindId(original_trip_id).One(&returnObj); err != nil {
            rw.WriteHeader(404)
            return
        }

        uj, _ := json.Marshal(returnObj)

        // Write content-type, statuscode, payload
        rw.Header().Set("Content-Type", "application/json")
        rw.WriteHeader(200)
        fmt.Fprintf(rw, "%s", uj)
    }


}

func getLocationCoordinatesFromDB(locationID string) (coordinates locationCoordinate){
    returnObj := UserData{}
    if !bson.IsObjectIdHex(locationID) {
        fmt.Println("Location ID does not found")
        return
    }

    original_loc_id := bson.ObjectIdHex(locationID)

    if err := getSession().DB("location_service").C("location").FindId(original_loc_id).One(&returnObj); err != nil {
        fmt.Print(err)
        panic(err)
    }

    locCords := locationCoordinate{
        id:locationID,
        latitude: returnObj.Coordinate.Lat,
        longitude: returnObj.Coordinate.Lng,
        visited:false,
    }
    return locCords
}
//Plan a trip using UBER - POST        /trips
func createTrip(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
    var request tripRequest
    var otherCoordinates []locationCoordinate
    var bestRoute []uberInfoFromPoint
    err := json.NewDecoder(req.Body).Decode(&request)
    if err != nil {
        panic(err)
    }
    //checking if destination Ids are given or not
    if len(request.LocationIds) == 0{
        jsonResponse := &error{
            "Destination Location Ids are missing.",
        }
        resp,_ := json.Marshal(jsonResponse)
        rw.Header().Set("Content-Type", "application/json")
        rw.WriteHeader(400)
        fmt.Fprintf(rw, "%s", resp)
        return
    }
    //Fetch the coordinates of the starting location
    startCoordinates := getLocationCoordinatesFromDB(request.StartingLocID)
    tempStartCoordinates := startCoordinates
    //fmt.Println(startCoordinates)

    //Fetch the coordinates of the rest of the locations
    for i := 0; i < len(request.LocationIds); i++  {
        otherCoordinates = append(otherCoordinates, getLocationCoordinatesFromDB(request.LocationIds[i]))
    }
    loopCount := len(otherCoordinates)
    for loopCount > 0{
        var uberModelInfo []uberInfoFromPoint

        for i := 0; i< loopCount; i ++ {

            url := fmt.Sprintf("/estimates/price?start_latitude=%f&start_longitude=%f&end_latitude=%f&end_longitude=%f&", startCoordinates.latitude, startCoordinates.longitude, otherCoordinates[i].latitude, otherCoordinates[i].longitude)
            servertoken := "server_token="+server_token
            url = sandboxURL + url + servertoken
            fmt.Println(url)

            //Calling UBER API to fetch the Price estimation based on UBER model
            response, err := http.Get(url)

            if response.StatusCode == 422 {
                fmt.Println("Distance Exceeded ")
                continue
            }
            if err != nil {
                return
            }
            defer response.Body.Close()

            resp := make(map[string]interface{})
            body, _ := ioutil.ReadAll(response.Body)
            err = json.Unmarshal(body, &resp)
            if err != nil {
                return
            }
            var estimated_prices []int
            count := resp["prices"].([]interface{})
            jq := jsonq.NewQuery(resp)
            // prices, err := jq.Array("prices")
            for k, _ := range count {
                le, _ := jq.Int("prices", fmt.Sprintf("%d", k), "low_estimate")
                du, _ := jq.Int("prices", fmt.Sprintf("%d", k), "duration")
                dis, _ := jq.Float("prices", fmt.Sprintf("%d", k), "distance")

                estimated_prices = append(estimated_prices, le)
                modelInfo := uberInfoFromPoint{
                    id:otherCoordinates[i].id,
                    low_estimate:le,
                    duration:du,
                    distance:dis,
                }
                uberModelInfo = append(uberModelInfo, modelInfo)
            }
           // fmt.Println(uberModelInfo)
            minimum_estimated_price := estimated_prices[0]
            var minimum_model_index int
            for j := 0; j< len(estimated_prices); j++ {
                if minimum_estimated_price > estimated_prices[j] {
                    minimum_estimated_price = estimated_prices[j]
                    minimum_model_index = j
                }
            }

            //fmt.Println(minimum_estimated_price)
            fmt.Println(uberModelInfo[minimum_model_index])
        }
        var minimumDestination uberInfoFromPoint
        minimumDestination = uberModelInfo[0]
        for i := 0; i<len(uberModelInfo); i++ {
            if minimumDestination.low_estimate > uberModelInfo[i].low_estimate && minimumDestination.duration > uberModelInfo[i].duration {
                minimumDestination = uberModelInfo[i]
            }
        }
        bestRoute = append(bestRoute,minimumDestination)
        //fmt.Println(minimumDestination)
        for loop:=0;loop < loopCount; loop++ {
            if minimumDestination.id == otherCoordinates[loop].id{
                //Deleting minimal element from the array
                fmt.Println("IDs matched"+minimumDestination.id)
                otherCoordinates = append(otherCoordinates[:loop],otherCoordinates[loop+1:]...)
                break
            }
        }
        startCoordinates = getLocationCoordinatesFromDB(minimumDestination.id)
        loopCount--
    }
    //fmt.Println("Best Route",bestRoute)
    var totalCost int
    var totalDuration int
    var totalDistance float64
    var locationIds []string
    for i:=0; i<len(bestRoute); i++{
        totalCost = totalCost + bestRoute[i].low_estimate
        totalDuration = totalDuration + bestRoute[i].duration
        totalDistance = totalDistance + bestRoute[i].distance
        locationIds = append(locationIds, bestRoute[i].id)
    }
    lastPointCords := getLocationCoordinatesFromDB(bestRoute[len(bestRoute)-1].id)
    //Add the cost of return journey

    url := fmt.Sprintf("/estimates/price?start_latitude=%f&start_longitude=%f&end_latitude=%f&end_longitude=%f&", lastPointCords.latitude, lastPointCords.longitude, tempStartCoordinates.latitude, tempStartCoordinates.longitude)
    servertoken := "server_token="+server_token
    url = sandboxURL + url + servertoken
    fmt.Println(url)

    //Calling UBER API to fetch the Price estimation based on UBER model
    response, err := http.Get(url)

    if err != nil {
        return
    }
    defer response.Body.Close()

    resp := make(map[string]interface{})
    body, _ := ioutil.ReadAll(response.Body)
    err = json.Unmarshal(body, &resp)
    if err != nil {
        return
    }
  //  count := resp["prices"].([]interface{})
    jq := jsonq.NewQuery(resp)

    le, _ := jq.Int("prices", fmt.Sprintf("%d", 0), "low_estimate")
    du, _ := jq.Int("prices", fmt.Sprintf("%d", 0), "duration")
    dis, _ := jq.Float("prices", fmt.Sprintf("%d", 0), "distance")

    totalCost = totalCost +le
    totalDuration = totalDuration + du
    totalDistance = totalDistance + dis
    //Code ends here for return journey

    tripResponseObject := TripResponse{
        Id: bson.NewObjectId(),
        Status : "planning",
        Starting_from_location_id:tempStartCoordinates.id,
        Best_route_location_ids : locationIds,
        Total_uber_costs: totalCost,
        Total_uber_duration: totalDuration,
        Total_distance:totalDistance,
    }

    //fmt.Println(tripResponseObject)
    //Persisting Data
    getSession().DB("location_service").C("trip").Insert(tripResponseObject)


    // Marshal provided interface into JSON structure
    marshalledResponseObj, _ := json.Marshal(tripResponseObject)

    // Write content-type, status code, payload
    rw.Header().Set("Content-Type", "application/json")
    rw.WriteHeader(201)
    fmt.Fprintf(rw, "%s", marshalledResponseObj)
}

//Create New Location - POST        /locations
func saveLocations(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
    var u UserData
    URL := "http://maps.google.com/maps/api/geocode/json?address="
    //Populate the data in local object
    json.NewDecoder(req.Body).Decode(&u)

    //Randomly generated unique ID
    // u.Id = randomString(10)
    u.Id = bson.NewObjectId()

    URL = URL +u.Address+ " " + u.City + " " + u.State + " " + u.Zip+"&sensor=false"
    URL = strings.Replace(URL, " ", "+", -1)
    fmt.Println("URL "+ URL)

    //calling google map API
    response, err := http.Get(URL)
    if err != nil {
        return
    }
    defer response.Body.Close()

    resp := make(map[string]interface{})
    body, _ := ioutil.ReadAll(response.Body)
    err = json.Unmarshal(body, &resp)
    if err != nil {
        return
    }

    jq := jsonq.NewQuery(resp)
    status, err := jq.String("status")
    //fmt.Println(status)
    if err != nil {
        return
    }
    if status != "OK" {
        err = errors.New(status)
        return
    }

    latitude, err := jq.Float("results" ,"0","geometry", "location", "lat")
    if err != nil {
        fmt.Println(err)
        return
    }
    longitude, err := jq.Float("results", "0","geometry", "location", "lng")
    if err != nil {
        fmt.Println(err)
        return
    }

    u.Coordinate.Lat = latitude
    u.Coordinate.Lng = longitude

    //Persisting Data
    getSession().DB("location_service").C("location").Insert(u)


    // Marshal provided interface into JSON structure
    uj, _ := json.Marshal(u)

    // Write content-type, status code, payload
    rw.Header().Set("Content-Type", "application/json")
    rw.WriteHeader(201)
    fmt.Fprintf(rw, "%s", uj)

}

//Request Uber - PUT    /trips/{trip_id}/request
func requestUber(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
    var estimatedTimeOfArrival int
    trip_id :=  p.ByName("trip_id")
    tripBookingResp := TripBookingResponse{}
    if !bson.IsObjectIdHex(trip_id) {
        rw.WriteHeader(404)
        return
    }

    original_trip_id := bson.ObjectIdHex(trip_id)

    returnObj := TripResponse{}
    if tripFinishedForTripIDs[trip_id] {
        if err := getSession().DB("location_service").C("tripbooking").FindId(original_trip_id).One(&tripBookingResp); err != nil {
            rw.WriteHeader(404)
            return
        }
        if(tripBookingResp.Status == "finished"){
            jsonResponse := &error{
                "This trip has already been completed. Please start a new trip.",
            }
            resp,_ := json.Marshal(jsonResponse)
            rw.Header().Set("Content-Type", "application/json")
            rw.WriteHeader(400)
            fmt.Fprintf(rw, "%s", resp)
            return
        }
    }
    if err := getSession().DB("location_service").C("trip").FindId(original_trip_id).One(&returnObj); err != nil {
        rw.WriteHeader(404)
        return
    }
    fmt.Println("Trip ID ::::::",trip_id)
    //Best Route IDs
    bestRouteIDs := returnObj.Best_route_location_ids
    startingLocID := returnObj.Starting_from_location_id
    //fmt.Println("startingLocID",startingLocID)

    //Forming the response
    tripBookingResp.Id = original_trip_id
    tripBookingResp.Best_route_location_ids = bestRouteIDs
    tripBookingResp.Total_uber_costs = returnObj.Total_uber_costs
    tripBookingResp.Total_uber_duration = returnObj.Total_uber_duration
    tripBookingResp.Total_distance = returnObj.Total_distance

    if requestIDandCordsMap,found := tripRequestIDAndCordsMapper[trip_id]; found{

        //fmt.Println(requestIDandCordsMap)
        var startLocID string
        var endLocID string
        for key := range requestIDandCordsMap {
            startLocID = requestIDandCordsMap[key]
            //fmt.Println(startLocID)
            delete(requestIDandCordsMap, key)
        }
        //fmt.Println("Empty Map :",requestIDandCordsMap)

        for i:=0;i<len(bestRouteIDs);i++{
            //fmt.Println("best route ids ",bestRouteIDs[i])
            if (bestRouteIDs[i] == startLocID) && (i != len(bestRouteIDs)-1){
                endLocID = bestRouteIDs[i+1]
                tripBookingResp.Status = "requesting"
            }else if (bestRouteIDs[i] == startLocID) && (i == len(bestRouteIDs)-1){
                endLocID = startingLocID
                tripBookingResp.Status = "requesting"
            }
        }
        if startingLocID == startLocID {
            fmt.Println("Both starts are same journey finished")
            tripFinishedForTripIDs[trip_id] = true
            tripBookingResp.Status = "finished"
            endLocID = ""
        }
        fmt.Println("startLocID",startLocID)
        fmt.Println("endLocID",endLocID)

        requestID, eta := doPostForUberBooking(startLocID, endLocID,trip_id)
        fmt.Println("Request ID : ",requestID)
        estimatedTimeOfArrival = eta

        tripBookingResp.Starting_from_location_id = startLocID
        tripBookingResp.Next_destination_location_id = endLocID
        tripBookingResp.Uber_wait_time_eta = estimatedTimeOfArrival

    }else {
      requestID, eta := doPostForUberBooking(startingLocID, bestRouteIDs[0],trip_id)
        fmt.Println("Request ID : ",requestID)
        estimatedTimeOfArrival = eta

        tripBookingResp.Status = "requesting"
        tripBookingResp.Starting_from_location_id = startingLocID
        tripBookingResp.Next_destination_location_id = bestRouteIDs[0]
        tripBookingResp.Uber_wait_time_eta = estimatedTimeOfArrival
    }

    // Marshal provided interface into JSON structure
    uj, _ := json.Marshal(tripBookingResp)

    if err := getSession().DB("location_service").C("tripbooking").RemoveId(original_trip_id); err != nil {
    }
    //Persisting trip Response in DB to make sure that after the trip is finished, the same trip id cannot be used for consecutive bookings
    getSession().DB("location_service").C("tripbooking").Insert(tripBookingResp)

    // Write content-type, status code, payload
    rw.Header().Set("Content-Type", "application/json")
    rw.WriteHeader(200)
    fmt.Fprintf(rw, "%s", uj)

}
//Internal Function to request the uber
func doPostForUberBooking(startPoint string, endPoint string, trip_id string) (request_id string, eta int){
    if endPoint == ""{
        return
    }
    startCoordinates := getLocationCoordinatesFromDB(startPoint)
    destinationCoordiates := getLocationCoordinatesFromDB(endPoint)

    //Request Uber for start to next destination
    url := fmt.Sprintf("/requests?start_latitude=%f&start_longitude=%f&end_latitude=%f&end_longitude=%f&", startCoordinates.latitude, startCoordinates.longitude, destinationCoordiates.latitude, destinationCoordiates.longitude)
    productId := "product_id="+product_id
    url = sandboxURL + url + productId
    fmt.Println(url)
    //fmt.Println(startCoordinates.latitude)
    //fmt.Println(startCoordinates.longitude)
    //POST request to above url to fetch the request_id
    json_post_body := &jsonPostBody{
        strconv.FormatFloat(startCoordinates.longitude,'f',7,64),
        strconv.FormatFloat(startCoordinates.latitude,'f',7,64),
        product_id,
    }
    uj, err := json.Marshal(json_post_body)
    if err != nil {
        panic (err)
    }

    //fmt.Println(string(uj))
    var jsonByteStr = []byte(string(uj))
    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonByteStr))

    req.Header.Set("Authorization", authorization_token)
    req.Header.Set("Content-Type", "application/json")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    //fmt.Println("response Status:", resp.Status)
    body, _ := ioutil.ReadAll(resp.Body)
    //fmt.Println("response Body:", string(body))
    var postResponseData postResponse
    err = json.Unmarshal(body, &postResponseData)
    if err != nil {
        panic(err)
    }
    fmt.Println("response request id :", postResponseData.Request_id)

    //Put Request ID, Trip ID and next Destination in the Map to be accessed next time
    tempMap := make(map[string]string)
    tempMap[postResponseData.Request_id] = endPoint
    tripRequestIDAndCordsMapper[trip_id] = tempMap
    //fmt.Println("tripRequestIDAndCordsMapper :", tripRequestIDAndCordsMapper)

    return postResponseData.Request_id, postResponseData.Eta
}
//Create New Location - POST        /locations
func updateLocations(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
    var u UserData
    location_id :=  p.ByName("location_id")

    URL := "http://maps.google.com/maps/api/geocode/json?address="

    //Populate the data in local object
    json.NewDecoder(req.Body).Decode(&u)

    URL = URL +u.Address+ " " + u.City + " " + u.State + " " + u.Zip+"&sensor=false"
    URL = strings.Replace(URL, " ", "+", -1)
    fmt.Println("URL "+ URL)

    //calling google map API
    response, err := http.Get(URL)
    if err != nil {
        return
    }
    defer response.Body.Close()

    resp := make(map[string]interface{})
    body, _ := ioutil.ReadAll(response.Body)
    err = json.Unmarshal(body, &resp)
    if err != nil {
        return
    }

    jq := jsonq.NewQuery(resp)
    status, err := jq.String("status")
    //fmt.Println(status)
    if err != nil {
        return
    }
    if status != "OK" {
        err = errors.New(status)
        return
    }

    latitude, err := jq.Float("results" ,"0","geometry", "location", "lat")
    if err != nil {
        fmt.Println(err)
        return
    }
    longitude, err := jq.Float("results", "0","geometry", "location", "lng")
    if err != nil {
        fmt.Println(err)
        return
    }

    u.Coordinate.Lat = latitude
    u.Coordinate.Lng = longitude

    original_loc_id := bson.ObjectIdHex(location_id)
    var data = UserData{
        Address: u.Address,
        City: u.City,
        State: u.State,
        Zip: u.Zip,
    }
    //updateData := bson.M{ "$set": data}
    fmt.Println(data)
    //Persisting Data
    getSession().DB("location_service").C("location").Update(bson.M{"_id":original_loc_id }, bson.M{"$set": bson.M{ "address": u.Address,
        "city": u.City, "state": u.State,"zip": u.Zip, "coordinate.lat":u.Coordinate.Lat, "coordinate.lng":u.Coordinate.Lng}})

    returnObj := UserData{}

    //fetch the response data
    if err := getSession().DB("location_service").C("location").FindId(original_loc_id).One(&returnObj); err != nil {
        rw.WriteHeader(404)
        return
    }
    // Marshal provided interface into JSON structure
    uj, _ := json.Marshal(returnObj)

    // Write content-type, status code, payload
    rw.Header().Set("Content-Type", "application/json")
    rw.WriteHeader(201)
    fmt.Fprintf(rw, "%s", uj)

}

//Delete a Location - DELETE /locations/{location_id}
func removeLocations(rw http.ResponseWriter, req *http.Request, p httprouter.Params) {
    location_id :=  p.ByName("location_id")

    if !bson.IsObjectIdHex(location_id) {
        rw.WriteHeader(404)
        return
    }

    original_loc_id := bson.ObjectIdHex(location_id)

    // Remove user
    if err := getSession().DB("location_service").C("location").RemoveId(original_loc_id); err != nil {
        rw.WriteHeader(404)
        return
    }

    rw.WriteHeader(200)
}

func main() {
    mux := httprouter.New()

    //Initialize request and coordinate mapper
    tripRequestIDAndCordsMapper = make(map[string]map[string]string)
    tripFinishedForTripIDs = make(map[string]bool)
    mux.GET("/locations/:location_id", getLocations)
    mux.GET("/trips/:trip_id", getTrips)
    mux.POST("/locations", saveLocations)
    mux.POST("/trips",createTrip)
    mux.PUT("/locations/:location_id", updateLocations)
    mux.PUT("/trips/:trip_id/request", requestUber)
    mux.DELETE("/locations/:location_id", removeLocations)
    rand.Seed( time.Now().UTC().UnixNano())

    server := http.Server{
        Addr:        "0.0.0.0:8880",
        Handler: mux,
    }
    server.ListenAndServe()
}