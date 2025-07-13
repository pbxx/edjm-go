package edjm

// import "fmt"
import (
	"fmt"
	"testing"

	"github.com/rs/zerolog/log"
)

var config map[string]interface{}

func handleJournalEvent(ec EventCallback) {
	switch ec.EventType {
	case "journalEvent":
		logger.Info().Msg(fmt.Sprintf("Handling journal event: %s", ec.EventName))
	case "dataFile":
		logger.Info().Msg(fmt.Sprintf("Handling dataFile event: %s", ec.EventName))
		logger.Debug().Msg(fmt.Sprintf("Data file entry: %v", string(ec.RawEntry)))
	default:
		log.Info().Msg(fmt.Sprint("Unknown event type:", ec.EventType))
	}
}

func TestLoad(t *testing.T) {
	// readJournalFile("C:\\Users\\dlewis\\Saved Games\\Frontier Developments\\Elite Dangerous\\Journalog.2025-06-01T140511.01.log")
	// import JSON config file if exists
	// zerolog.SetGlobalLevel(zerolog.TraceLevel)
	// zerolog.SetGlobalLevel(3)

	// Start the journal manager
	go Start(handleJournalEvent)

	log.Info().Msg("Journal manager started. Listening for events...")

	// Block main goroutine forever to keep the program running
	<-make(chan struct{})
	got := "DST: CloseLand22"
	want := "DST: CloseLand22"

	if got != want {
		t.Errorf("got %q, wanted %q", got, want)
	}
}
