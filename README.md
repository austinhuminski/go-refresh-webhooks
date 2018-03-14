# go-refresh-webhooks
Current company has leveraged webhooks with a project management website to send all the info a new tech support ticket to a AWS lambda function which will then clone the ticket infomation into another project management website. 

 These webhooks would intermittently stop working. The solution was to delete the webhook and resubmit it to the API for all projects. Since this is happening so often now, I built a script that runs once a day on AWS Lambda to delete all the current webhooks and then resubmit them via the websites API
 
This was written in Go mostly as a learning exercise to help better understand the language. (Coming from a python background). 



## Steps
1) `GET` request to retrieve a list of all the webhooks currently setup. Response:

	```		
	{
    "data": [
        {
            "id": 112233445566778, # Webhook ID
            "target": "<URL-TO-SEND-WEBHOOK-DATA>",
            "active": true,
            "resource": {
                "id": 123456789123456, # Project ID
                "name": "Project Name "
            }
        },
    ]}
    ```
    
2) `DELETE` request to remove all active webhooks.

3) `POST` request to recreate webhooks for each project

## Thoughts & TODO's
* It was getting annoying doing error almost every few lines. I wrapped the baisc error check for `nil` into a function and called it throughout the script. Is that idomatic?
* Found it a bit annoying to have to unpack every HTTP JSON response into a struct. Is there another way to quickly see and act on the JSON response without a struct?
* The HTTP `POST` I had a lot of trouble with. Before the current solution I had, I put all the `POST` body data into a `map[string][string]`. When to use `Http.Post` vs `Http.NewRequest("POST"...)`?


	```
	values := map[string]string{"resource": resource, "target": target}
	jsonValue, _ := json.Marshal(values)
	resp, err := http.Post(URL, "application/json", bytes.NewBuffer(jsonValue))
	```

* TODO: run in parallel. I know it's overkill for this example but thought it would be good practice. Not sure how many CPUs available for AWS lambdas. This currently takes ~8 seconds to run with just some goroutines

