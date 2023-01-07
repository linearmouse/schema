package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	"golang.org/x/sync/singleflight"
)

var errSchemaNotFound = errors.New("schema not found")

var (
	schemaSingleFlightGroup singleflight.Group
	schemaCache             sync.Map
)

func getSchema(schemaURL string) ([]byte, error) {
	schema, err, _ := schemaSingleFlightGroup.Do(schemaURL, func() (any, error) {
		cache, ok := schemaCache.Load(schemaURL)
		if ok {
			log.Printf("Load schema from cache: %s", schemaURL)
			return cache.([]byte), nil
		}

		log.Printf("Load schema from GitHub: %s", schemaURL)
		resp, err := client.Get(schemaURL)
		if err != nil {
			return nil, fmt.Errorf("client.Get: %w", err)
		}
		defer resp.Body.Close()

		switch resp.StatusCode {
		case http.StatusOK:

		case http.StatusNotFound:
			return nil, errSchemaNotFound

		default:
			return nil, fmt.Errorf("client.Get: StatusCode: %d", resp.StatusCode)
		}

		schema, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("ioutil.ReadAll: %w", err)
		}

		schemaCache.Store(schemaURL, schema)

		return schema, nil
	})

	if err != nil {
		return nil, err
	}

	return schema.([]byte), nil
}
