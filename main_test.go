package main

import (
	"github.com/jarcoal/httpmock"
	"testing"
)

func TestEndToEnd(t *testing.T) {
	// Arrange

	// Open DB connection
	conn, err := OpenDbConn()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// Create service
	service := MonitoringService{
		Db: conn,
	}

	// Setup new DB
	service.DropTable()
	service.SetupDB()

	// Create configurations for two websites
	cfg_1 := urlConfiguration{
		id:       1,
		interval: 2,
		url:      "https://foo.com",
		regexp:   "foo.*",
	}

	cfg_2 := urlConfiguration{
		id:       2,
		interval: 5,
		url:      "https://bar.com",
		regexp:   "bar.*",
	}

	cfgs := urlConfigurations{cfg_1, cfg_2}

	// Mock HTTP Responses
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "https://foo.com", httpmock.NewStringResponder(200, `[{"id": 1, "name": "Body should foo."}]`))
	httpmock.RegisterResponder("GET", "https://bar.com", httpmock.NewStringResponder(400, `[{"id": 1, "name": "Should not match regexp."}]`))

	// Expected data to be written to the DB
	expected_foo := dbEntry{
		url:    "https://foo.com",
		status: 200,
		regexp: 1,
	}

	expected_bar := dbEntry{
		url:    "https://bar.com",
		status: 400,
		regexp: 0,
	}

	// Act

	// Read configurations and send out requests
	ch := make(chan response)
	for _, cfg := range cfgs {
		go service.MakeRequest(cfg, ch)
	}

	// Consume responses and make sure they arrive in the right order based on configured interval
	// since in this controlled environment the rt (response time) is 0.
	// Assert url, response status and regexp match are correct.
	// Skip timestamp since this is hard to mock and also skip rt since this will always be 0.
	var responses responses
	for {
		responses = append(responses, <-ch)

		// Read responses after we receive the 3rd response
		if len(responses) == 3 {

			// Write data to the DB
			service.saveToDB(responses)

			// Read data from the DB
			dbEntries, err := service.readFromDB()
			if err != nil {
				t.Errorf("Reading from DB failed, err: %v", err)
			}

			// Assert 1st response is correct
			if dbEntries[0].url != expected_foo.url {
				t.Errorf("Expected %v, got %v", expected_foo.url, dbEntries[0].url)
			}
			if dbEntries[0].status != expected_foo.status {
				t.Errorf("Expected %v, got %v", expected_foo.status, dbEntries[0].status)
			}
			if dbEntries[0].regexp != expected_foo.regexp {
				t.Errorf("Expected %v, got %v", expected_foo.regexp, dbEntries[0].regexp)
			}

			// Assert 2nd response is correct
			if dbEntries[1].url != expected_foo.url {
				t.Errorf("Expected %v, got %v", expected_foo.url, dbEntries[1].url)
			}
			if dbEntries[1].status != expected_foo.status {
				t.Errorf("Expected %v, got %v", expected_foo.status, dbEntries[1].status)
			}
			if dbEntries[1].regexp != expected_foo.regexp {
				t.Errorf("Expected %v, got %v", expected_foo.regexp, dbEntries[1].regexp)
			}

			// Assert 3rd response is correct
			if dbEntries[2].url != expected_bar.url {
				t.Errorf("Expected %v, got %v", expected_bar.url, dbEntries[2].url)
			}
			if dbEntries[2].status != expected_bar.status {
				t.Errorf("Expected %v, got %v", expected_bar.status, dbEntries[2].status)
			}
			if dbEntries[2].regexp != expected_bar.regexp {
				t.Errorf("Expected %v, got %v", expected_bar.regexp, dbEntries[2].regexp)
			}

			return
		}
	}

}
