package edjm

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/rs/zerolog"
)

type State struct {
	JournalEntries int
	JournalDir     string
	SpecifyJournal string // User-specified journal file
}

var JournalState State = State{
	JournalEntries: 0,
	JournalDir:     "",
	SpecifyJournal: "",
}

// var logger = zerolog.New(os.Stdout).Level(zerolog.InfoLevel).With().Timestamp().Logger()

var logger = zerolog.New(os.Stdout).Level(3).With().Timestamp().Logger()

func GetJournalDir() (string, error) {
	// get the user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// logger.Info().Msg(fmt.Sprint("Error getting home directory:", err)
		return "", err
	}
	// logger.Info().Msg(fmt.Sprint("User's home directory:", homeDir)
	// construct the path to the journal directory
	dirPath := filepath.Join(homeDir, "/Saved Games/Frontier Developments/Elite Dangerous/")
	return dirPath, nil
}

func LoadLatestJournal() ([]map[string]interface{}, [][]byte, error) {
	var latestFile string
	// get the latest journal file
	if JournalState.SpecifyJournal != "" {
		// use a specific journal file instead of the latest if user prefers
		logger.Info().Msg(fmt.Sprint("DEV: Using specified journal file:", JournalState.SpecifyJournal))
		latestFile = filepath.Join(JournalState.JournalDir, JournalState.SpecifyJournal)
	} else {
		// find the latest journal file in the directory
		lFile, err := FindLatestJournalFile()
		if err != nil {
			// logger.Info().Msg(fmt.Sprint("Error getting latest journal file:", err)
			return nil, [][]byte{}, err
		}
		latestFile = lFile
	}

	// logger.Info().Msg(fmt.Sprint("Latest journal file:", latestFile)

	// read the latest journal file
	bytesRead, _ := os.ReadFile(latestFile)
	fileContent := string(bytesRead)
	lines := strings.Split(fileContent, "\n")
	count := len(lines) - 1 // subtract 1 for the last empty line

	//instantiate an array to hold the parsed JSON objects
	var parsedLines []map[string]interface{}
	var rawLines [][]byte
	// print each line with its index
	// this loop starts from JournalState.JournalEntries to avoid reprocessing already read entries
	for i := JournalState.JournalEntries; i < count; i++ {
		line := lines[i]
		if line != "" {
			// append raw line to rawLines
			rawLines = append(rawLines, []byte(line))
			// parse the line as JSON
			var result map[string]interface{}
			err := json.Unmarshal([]byte(line), &result)
			if err != nil {
				return nil, [][]byte{}, err
			}
			// append the parsed result to the journal
			parsedLines = append(parsedLines, result)

		}
	}
	// update the state with the new count
	JournalState.JournalEntries = count
	return parsedLines, rawLines, nil
}

func FindLatestJournalFile() (string, error) {
	dirPath := JournalState.JournalDir
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return "", err
	}

	var latestFile string
	var latestModTime int64

	for _, file := range files {
		filepath := filepath.Join(dirPath, file.Name())
		if strings.HasSuffix(filepath, ".log") {
			fileInfo, err := os.Stat(filepath)
			if err != nil {
				return "", err
			}
			if fileInfo.ModTime().Unix() > latestModTime {
				latestModTime = fileInfo.ModTime().Unix()
				latestFile = filepath
				// logger.Info().Msg(fmt.Sprint("Found log file:", latestFile, "with mod time:", fileInfo.ModTime())
			}
		}
	}

	if latestFile != "" {
		return latestFile, nil
	} else {
		return "", fmt.Errorf("no log files found in the directory")
	}
}

func ReadJournalFile(filename string) ([]map[string]interface{}, error) {
	// read file into lines
	bytesRead, _ := os.ReadFile(filename)
	fileContent := string(bytesRead)
	lines := strings.Split(fileContent, "\n")
	count := len(lines) - 1 // subtract 1 for the last empty line

	//instantiate an array to hold the parsed JSON objects
	var parsedObjects []map[string]interface{}
	// print each line with its index
	for i := JournalState.JournalEntries; i < count; i++ {
		line := lines[i]
		if line != "" {

			var result map[string]interface{}
			err := json.Unmarshal([]byte(line), &result)
			if err != nil {
				return nil, err
			}
			parsedObjects = append(parsedObjects, result)

		}
	}

	JournalState.JournalEntries = count // update the state with the new count
	return parsedObjects, nil
}

func LoadDataFileRaw(name string) ([]byte, error) {
	// read file into raw json bytes
	cfile, err := os.ReadFile(filepath.Join(JournalState.JournalDir, fmt.Sprintf("%s.json", name)))
	if err != nil {
		return []byte{}, err
	}
	if len(cfile) == 0 {
		return []byte{}, fmt.Errorf("empty data file")
	}
	return cfile, nil
}

func LoadDataFile(name string) (map[string]interface{}, []byte, error) {
	// helper function to load and parse a data file into unmarshaled JSON
	rawFile, err := LoadDataFileRaw(name)
	if err != nil {
		return nil, []byte{}, err
	}

	result, err := JsonParse(rawFile)
	if err != nil {
		return nil, []byte{}, err
	}

	return result, rawFile, nil
}

func JsonParse(jBytes []byte) (map[string]interface{}, error) {
	// Parse the JSON bytes into a map
	var result map[string]interface{}
	err := json.Unmarshal(jBytes, &result)
	if err != nil {
		return nil, fmt.Errorf("error parsing JSON: %v", err)
	}
	return result, nil
}

type EventCallback struct {
	// EventType is either "journalEvent" or "dataFile"
	EventType string
	EventName string
	Entry     map[string]interface{}
	RawEntry  []byte
}

func StartSpecific(JournalDir string, specific string, callback func(EventCallback)) {
	// Initialize the state with journal directory
	JournalState.JournalDir = JournalDir

	// If a specific journal file is provided, set it to be used
	if specific != "" {
		JournalState.SpecifyJournal = specific
	}

	// Create new watcher.
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// Add the journal directory to the watcher
	err = watcher.Add(JournalDir)
	if err != nil {
		log.Fatal(err)
	}

	// Start listening for events.
	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			// log.Println("event:", event)
			if event.Has(fsnotify.Write) {
				// A file was written to, handle the change

				filename := strings.Split(event.Name, string(os.PathSeparator))
				filenameOnly := filename[len(filename)-1] // get the last part of the path

				// check if the file is a journal file or a data file
				splitFilename := strings.Split(filenameOnly, ".")
				extension := splitFilename[len(splitFilename)-1]
				name := splitFilename[0]

				switch extension {
				case "log":
					// this is a journal file change
					// Load a fresh journal
					parsedLines, rawLines, err := LoadLatestJournal()
					if err != nil {
						logger.Info().Msg(fmt.Sprint("Error loading journal:", err))
						return
					}

					// if new entries were found, call the callback for each entry
					if len(parsedLines) != 0 {
						logger.Info().Msg(fmt.Sprint("Handling ", len(parsedLines), " new entries from journal file..."))
						for i, entry := range parsedLines {
							eventType := entry["event"].(string)
							logger.Info().Msg(fmt.Sprint("Journal file changed:", eventType))
							callback(EventCallback{
								EventType: "journalEvent",
								EventName: eventType,
								Entry:     entry,
								RawEntry:  rawLines[i],
							})
						}
					}
				case "json":
					// this is a data file change
					// Load the data file
					logger.Info().Msg(fmt.Sprint("Data file changed:", name))
					dataFile, rawDataFile, err := LoadDataFile(name)
					if err != nil {
						logger.Info().Msg(fmt.Sprint("Error loading data file:", err))
						if err.Error() == "empty data file" {
							logger.Info().Msg(fmt.Sprint("Skipping empty data file:", name))
						} else {
							return
						}
					}
					// error wont be nil if the file was empty
					// check so we can safely call the callback
					if err == nil {
						// call the callback with the data file
						callback(EventCallback{
							EventType: "dataFile",
							EventName: filenameOnly,
							Entry:     dataFile,
							RawEntry:  rawDataFile, // raw entry is not used for data files
						})
					}
				default:
					logger.Info().Msg(fmt.Sprint("Unknown file type changed:", filenameOnly))
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Println("error:", err)
		}
	}

}

func Start(callback func(EventCallback)) {
	// Start the journal manager using the default location, and latest journal file always
	journalDir, err := GetJournalDir()
	if err != nil {
		logger.Info().Msg(fmt.Sprint("Error getting journal directory:", err))
		return
	}
	StartSpecific(journalDir, "", callback)
}
