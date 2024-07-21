	
## BigTable Sample Application 
Simple Web Application using Bigtable and golang (echo)

### Dependencies
* [Echo](https://echo.labstack.com/)
* [Golang](https://go.dev/)
* [GCP Bigtable client libraries](https://cloud.google.com/bigtable/docs/reference/libraries)

This program is tested by Mac OS

### Prerequisite
1. Can create bigtable
2. Run with Local machine
3. Use gcloud command to access GCP

### Preparation
1. Create Big Instance from GCP console
2. Create Table and Column Family from GCP console
3. Access GCP from terminal using gcloud command


Need to prepare following information
* Project ID
* BigTable Instance ID

Use this parameters when running a sample application

### Run
Please use same terminal from preparation

0. Prepare required library
```bash
go mod tidy
```

1. Build with go build
```bash
go build
```
2. Run with parameters
```bash
./worldt --project (GCP Project name) --instance (Big Table Instance name)
```

## Request Test
This application has 3 endpoints
<pre>
/update  POST : Post parameters to insert temperature data into BigTable  
/get      GET : Get latest(last log) temperature using city and day
/getall     GET : Get all temperature using city and day(return array)
</pre>
Example)
1. Save temperature data
```bash
curl --location --request POST 'localhost:1323/update' \
--header 'Content-Type: application/json' \
--data-raw '{
    "city": "osaka",
    "day": "2024-07-21",
    "hour": "11",
    "temperature": 36 
}'
```
2. Get latest temperature data
```bash
curl --location --request GET 'localhost:1323/get?city=tokyo&day=2024-07-21'
```


3. Get all temperature data on the day
```bash
curl --location --request GET 'localhost:1323/getall?city=osaka&day=2024-07-21'
```