package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegisterHandlerMethods(t *testing.T) {
	// Test GET method
	req, err := http.NewRequest("GET", "/register", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(registerHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("GET method returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	// Test POST method
	req, err = http.NewRequest("POST", "/register", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusSeeOther {
		t.Errorf("POST method returned wrong status code: got %v want %v", status, http.StatusSeeOther)
	}

	// Test invalid method
	req, err = http.NewRequest("PUT", "/register", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("PUT method returned wrong status code: got %v want %v", status, http.StatusMethodNotAllowed)
	}
}

func TestLogUserInHandlerMethods(t *testing.T) {
	// Test POST method
	req, err := http.NewRequest("POST", "/loguserin", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(logUserInHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusSeeOther {
		t.Errorf("POST method returned wrong status code: got %v want %v", status, http.StatusSeeOther)
	}

	// Test invalid method
	req, err = http.NewRequest("GET", "/loguserin", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("GET method returned wrong status code: got %v want %v", status, http.StatusMethodNotAllowed)
	}
}

func TestLogoutHandlerMethods(t *testing.T) {
	// Test POST method
	req, err := http.NewRequest("POST", "/logout", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(logoutHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusSeeOther {
		t.Errorf("POST method returned wrong status code: got %v want %v", status, http.StatusSeeOther)
	}

	// Test invalid method
	req, err = http.NewRequest("GET", "/logout", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("GET method returned wrong status code: got %v want %v", status, http.StatusMethodNotAllowed)
	}
}
