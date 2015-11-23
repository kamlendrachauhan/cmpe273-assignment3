# CMPE 273 Assignment 3

##Trip Planner

The trip planner is a feature that will take a set of locations from the database and will then check against UBER’s price estimates API to suggest the best possible route in terms of costs and duration.

###1. POST        /trips   # Plan a trip
####Request
{  
"starting_from_location_id": "5651836515ffb720b8d81ca7",  
"location_ids": [  
    "565183ce15ffb720b8d81ca8",  
    "56518b8215ffb7231c161d6c",  
    "5651848e15ffb720b8d81cab",  
    "565185b315ffb720b8d81cac"  
  ]  
}

where   
5651836515ffb720b8d81ca7	fairmont hotel(Starting Point)  
565183ce15ffb720b8d81ca8	Golden Gate Bridge  
56518b8215ffb7231c161d6c	Pier 39  
5651848e15ffb720b8d81cab	Golden Gate Park  
565185b315ffb720b8d81cac	Twin Peaks  

####Response HTTP 201

{  
  "id": "5652836715ffb726589a0ca5",  
  "status": "planning",  
  "starting_from_location_id": "5651836515ffb720b8d81ca7",  
  "best_route_location_ids": [  
    "56518b8215ffb7231c161d6c",  
    "565183ce15ffb720b8d81ca8",  
    "5651848e15ffb720b8d81cab",  
    "565185b315ffb720b8d81cac"  
  ],  
  "total_uber_costs": 64,  
  "total_uber_duration": 4350,  
  "total_distance": 22.83  
}  

###2. GET        /trips/{trip_id} # Check the trip details and status
####Request
GET             /trips/5652836715ffb726589a0ca5  
####Response
{  
  "id": "5652836715ffb726589a0ca5",  
  "status": "planning",  
  "starting_from_location_id": "5651836515ffb720b8d81ca7",  
  "best_route_location_ids": [  
    "56518b8215ffb7231c161d6c",  
    "565183ce15ffb720b8d81ca8",  
    "5651848e15ffb720b8d81cab",  
    "565185b315ffb720b8d81cac"  
  ],  
  "total_uber_costs": 64,  
  "total_uber_duration": 4350,  
  "total_distance": 22.83  
}  

###3. PUT        /trips/{trip_id}/request # Start the trip by requesting UBER for the first destination. You will call UBER request API to request a car from starting point to the next destination.
####Request
PUT             /trips/5652836715ffb726589a0ca5/request  
####Response
{  
  "id": "5652836715ffb726589a0ca5",  
  "status": "requesting",  
  "starting_from_location_id": "5651836515ffb720b8d81ca7",  
  "next_destination_location_id": "56518b8215ffb7231c161d6c",  
  "best_route_location_ids": [  
    "56518b8215ffb7231c161d6c",  
    "565183ce15ffb720b8d81ca8",  
    "5651848e15ffb720b8d81cab",  
    "565185b315ffb720b8d81cac"  
  ],  
  "total_uber_costs": 64,  
  "total_uber_duration": 4350,  
  "total_distance": 22.83,  
  "uber_wait_time_eta": 15  
}  

######Subsequent calls to /trips/5652836715ffb726589a0ca5/request will book uber from last destination reached to next destination from the best route planned before.
######As soon as user reaches the last destination i.e. 565185b315ffb720b8d81cac(Twin Peaks) The next PUT call requests the Uber to the starting point from where the journey began.
PUT             /trips/5652836715ffb726589a0ca5/request  
{  
  "id": "5652836715ffb726589a0ca5",  
  "status": "requesting",  
  "starting_from_location_id": "565185b315ffb720b8d81cac",  
  "next_destination_location_id": "5651836515ffb720b8d81ca7",  
  "best_route_location_ids": [  
    "56518b8215ffb7231c161d6c",  
    "565183ce15ffb720b8d81ca8",  
    "5651848e15ffb720b8d81cab",  
    "565185b315ffb720b8d81cac"  
  ],  
  "total_uber_costs": 64,  
  "total_uber_duration": 4350,  
  "total_distance": 22.83,  
  "uber_wait_time_eta": 15  
}  

######Now the Uber has reached the starting point and next PUT call will return the status as finished to show the the same.
PUT             /trips/5652836715ffb726589a0ca5/request  
{  
  "id": "5652836715ffb726589a0ca5",  
  "status": "finished",  
  "starting_from_location_id": "5651836515ffb720b8d81ca7",  
  "next_destination_location_id": "",  
  "best_route_location_ids": [  
    "56518b8215ffb7231c161d6c",  
    "565183ce15ffb720b8d81ca8",  
    "5651848e15ffb720b8d81cab",   
    "565185b315ffb720b8d81cac"  
  ],  
  "total_uber_costs": 64,  
  "total_uber_duration": 4350,  
  "total_distance": 22.83,  
  "uber_wait_time_eta": 0  
}  

######Now Any further PUT call for the same request ID will return an error showing that the journey has been finished. Please plan a new journey.
{  
  "error_message": "This trip has already been completed. Please start a new trip."  
}  
