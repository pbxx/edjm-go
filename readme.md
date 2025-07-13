![EDJM Icon](./assets/icon.png)

# Elite Dangerous Journal Manager for Golang
An easy, pure-go module for watching and parsing Elite Dangerous journal and JSON file events

## Installation
```sh
go get github.com/pbxx/edjm-go
```

## Usage
Start EDJM by calling `edjm.Start` in a goroutine, passing a function that accepts an `EventCallback`:
```go
package main

import (
	edjm "github.com/pbxx/edjm-go"
)

func handleJournalEvent(ec edjm.EventCallback) {
	switch ec.EventType {
	case "journalEvent":
		// Latest journal.log event received
        fmt.Printf("New %s event received from journal\n", ec.EventName)
	case "dataFile":
		// JSON data file update received
        fmt.Printf("Data file %s.json updated\n", ec.EventName)

	default:
		log.Info().Msg(fmt.Sprint("Unknown event type:", ec.EventType))
	}
}

main () {
    // start EDJM in a goroutine
    go edjm.Start(handleJournalEvent)

    // block main forever so program stays on
    <-make(chan struct{})
}
```

`EventCallback` structure:
```go
type EventCallback struct {
	// EventType is either "journalEvent" or "dataFile"
	EventType string
    // Event name is either the name of the journal event, or name of the JSON file updated
	EventName string
    // Entry contains the parsed JSON journal entry or data file
	Entry     map[string]interface{}
    // RawEntry contains the JSON string of the journal event or data file, when parsed values aren't needed (e.g. sending to a server)
	RawEntry  []byte
}
```