package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// logDataGet returns the log context
func logDataGet(r *http.Request) map[string]interface{} {
	if r == nil {
		return make(map[string]interface{})
	}
	ctx := r.Context()
	data := ctx.Value("log")
	switch v := data.(type) {
	case map[string]interface{}:
		return v
	}
	return make(map[string]interface{})
}

// logDataAdd adds a single value to the log context
func logDataAdd(r *http.Request, key string, value interface{}) {
	var data map[string]interface{}

	ctx := r.Context()
	d := ctx.Value("log")
	switch v := d.(type) {
	case map[string]interface{}:
		data = v
	default:
		data = make(map[string]interface{})
	}

	data[key] = value

	r = r.WithContext(context.WithValue(ctx, "log", data))
}

// logDataReplace replaces the current log context with the provided log data
func logDataReplace(r *http.Request, data map[string]interface{}) {
	ctx := r.Context()
	r = r.WithContext(context.WithValue(ctx, "log", data))
}

// logAsString returns the string version of the logs (ie, json marshal)
func logAsString(l map[string]interface{}) string {
	l["app"] = AppName
	l["version"] = Version
	b, err := json.Marshal(l)
	if err != nil {
		log.Printf("unable to marshap map[string]interface{}. Wtf. %v \n %#v", err, l)
	}
	return string(b)
}

// logEvent allows us to track novel happenings
func logEvent(r *http.Request, event string, msg string) {
	logData := logDataGet(r)
	logData["event"] = event
	logData["message"] = msg

	log.Println(logAsString(logData))
}

// logError is similar to logEvent but has an error field
func logError(r *http.Request, err error, msg string) {
	logData := logDataGet(r)
	logData["event"] = "error"
	logData["message"] = msg
	if err == nil {
		err = fmt.Errorf("internal error condition")
	}
	logData["error"] = err.Error()

	log.Println(logAsString(logData))
}

// logDebug captures debug messages and only passes them through if Debug is true
func logDebug(r *http.Request, msg string) {
	if Debug {
		logEvent(r, "debug", msg)
	}
}
