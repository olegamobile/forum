package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Unit Test: Directly calls handler w/mocked ResponseWriter and Request to test logic in isolation.
func TestHandler(t *testing.T) {
	testCases := []struct {
		name               string
		method             string
		url                string
		expectedStatusCode int
	}{
		// check Home URLs "/" and "/forum" for diff methods
		{
			name:               "Home / GET",
			method:             http.MethodGet,
			url:                "/",
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Home / POST",
			method:             http.MethodPost,
			url:                "/",
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Home / Invalid PUT",
			method:             http.MethodPut,
			url:                "/",
			expectedStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name:               "Home / Invalid DELETE",
			method:             http.MethodDelete,
			url:                "/",
			expectedStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name:               "Home /forum GET",
			method:             http.MethodGet,
			url:                "/forum",
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Home /forum POST",
			method:             http.MethodPost,
			url:                "/forum",
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Home /forum Invalid PUT",
			method:             http.MethodPut,
			url:                "/forum",
			expectedStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name:               "Home /forum Invalid DELETE",
			method:             http.MethodDelete,
			url:                "/forum",
			expectedStatusCode: http.StatusMethodNotAllowed,
		},
		// check "/thread/" for diff methods
		{
			name:               "Thread /thread/ GET",
			method:             http.MethodGet,
			url:                "/thread/",
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Thread /thread/ Invalid POST",
			method:             http.MethodPost,
			url:                "/thread/",
			expectedStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name:               "Home /thread/ Invalid PUT",
			method:             http.MethodPut,
			url:                "/thread/",
			expectedStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name:               "Home /thread/ Invalid DELETE",
			method:             http.MethodDelete,
			url:                "/thread/",
			expectedStatusCode: http.StatusMethodNotAllowed,
		},
		// check "/add" for diff methods
		{
			name:               "Add Thread /add Invalid GET",
			method:             http.MethodGet,
			url:                "/add",
			expectedStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name:               "Add Thread /add POST",
			method:             http.MethodPost,
			url:                "/add",
			expectedStatusCode: http.StatusOK,
		},
		// check "/reply" for diff methods
		{
			name:               "Add Thread /reply Invalid GET",
			method:             http.MethodGet,
			url:                "/reply",
			expectedStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name:               "Add Thread /reply POST",
			method:             http.MethodPost,
			url:                "/reply",
			expectedStatusCode: http.StatusOK,
		},
		// check "/login" for different methods
		{
			name:               "Login /login GET",
			method:             http.MethodGet,
			url:                "/login",
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Login /login Invalid POST",
			method:             http.MethodPost,
			url:                "/login",
			expectedStatusCode: http.StatusMethodNotAllowed,
		},
		// check "/loguserin" for different methods
		{
			name:               "Login /loguserin Invalid GET",
			method:             http.MethodGet,
			url:                "/loguserin",
			expectedStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name:               "Login /loguserin POST",
			method:             http.MethodPost,
			url:                "/loguserin",
			expectedStatusCode: http.StatusOK,
		},
		// check Register URL path "/register" for different methods
		{
			name:               "Register /register GET",
			method:             http.MethodGet,
			url:                "/register",
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Register /register POST",
			method:             http.MethodPost,
			url:                "/register",
			expectedStatusCode: http.StatusOK,
		},
		{
			name:               "Register /register Invalid DELETE",
			method:             http.MethodDelete,
			url:                "/register",
			expectedStatusCode: http.StatusMethodNotAllowed,
		},
	}

	// Loop over each test case
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new HTTP request for this test case
			req, err := http.NewRequest(tc.method, tc.url, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			// Create a ResponseRecorder to record the response
			rr := httptest.NewRecorder()

			// Call the appropriate handler based on the URL
			if tc.url == "/" || tc.url == "/forum" {
				indexHandler(rr, req, "")
			} else if tc.url == "/thread/" {
				threadPageHandler(rr, req)
			} else if tc.url == "/add" {
				addThreadHandler(rr, req)
				} else if tc.url == "/reply" {
					addReplyHandler(rr, req)
			} else if tc.url == "/login" {
				logInHandler(rr, req)
			} else if tc.url == "/register" {
				registerHandler(rr, req)
			} else {
				goToErrorPage("", tc.expectedStatusCode, rr, req)
			}

			// Check the status code
			if rr.Result().StatusCode != tc.expectedStatusCode {
				t.Errorf("Expected status code is %d; but program gave %d", tc.expectedStatusCode, rr.Result().StatusCode)
			}

			// Close the response body
			rr.Result().Body.Close()
		})
	}
}
	// Loop over each test case
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a new HTTP request for this test case
			req, err := http.NewRequest(tc.method, tc.url, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			// Create a ResponseRecorder to record the response
			rr := httptest.NewRecorder()

			// Call the appropriate handler based on the URL
			if tc.url == "/" {
				indexHandler(rr, req, "")
			} else {
				goToErrorPage("", tc.expectedStatusCode, rr, req)
			}

			// Check the status code
			if rr.Result().StatusCode != tc.expectedStatusCode {
				t.Errorf("Expected status code is %d; but program gave %d", tc.expectedStatusCode, rr.Result().StatusCode)
			}

			// Close the response body
			rr.Result().Body.Close()
		})
	}
}
