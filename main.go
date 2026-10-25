package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand/v2"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type Satellite struct {
	ObjectName         string  `json:"OBJECT_NAME"`
	ObjectID           string  `json:"OBJECT_ID"`
	Epoch              string  `json:"EPOCH"`
	MeanMotion         float64 `json:"MEAN_MOTION"`
	Eccentricity       float64 `json:"ECCENTRICITY"`
	Inclination        float64 `json:"INCLINATION"`
	RaOfAscNode        float64 `json:"RA_OF_ASC_NODE"`
	ArgOfPericenter    float64 `json:"ARG_OF_PERICENTER"`
	MeanAnomaly        float64 `json:"MEAN_ANOMALY"`
	EphemerisType      int     `json:"EPHEMERIS_TYPE"`
	ClassificationType string  `json:"CLASSIFICATION_TYPE"`
	NORADCatID         int     `json:"NORAD_CAT_ID"`
	ElementSetNo       int     `json:"ELEMENT_SET_NO"`
	RevAtEpoch         int     `json:"REV_AT_EPOCH"`
	BSTAR              float64 `json:"BSTAR"`
	MeanMotionDot      float64 `json:"MEAN_MOTION_DOT"`
	MeanMotionDDot     float64 `json:"MEAN_MOTION_DDOT"`
}

var satellite Satellite
var satellites []Satellite
var watchlist []Satellite
var suggestedSatellites = []Satellite{
	{ObjectName: "ISS (ZARYA)", NORADCatID: 25544},
	{ObjectName: "CSS (TIANHE)", NORADCatID: 48274},
	{ObjectName: "HST", NORADCatID: 20580},
	{ObjectName: "AURA", NORADCatID: 28376},
	{ObjectName: "NOAA 20", NORADCatID: 43013},
	{ObjectName: "SUOMI NPP", NORADCatID: 37849},
	{ObjectName: "TERRA", NORADCatID: 25994},
	{ObjectName: "AQUA", NORADCatID: 27424},
	{ObjectName: "LANDSAT 8", NORADCatID: 39084},
	{ObjectName: "LANDSAT 9", NORADCatID: 49260},
	{ObjectName: "SENTINEL-2A", NORADCatID: 40697},
	{ObjectName: "ICESAT-2", NORADCatID: 43613},
	{ObjectName: "COSMOS 2564 (GLONASS)", NORADCatID: 54379},
	{ObjectName: "GSAT0101 (PRN E11)", NORADCatID: 37846},
	{ObjectName: "BEIDOU 3 M1", NORADCatID: 42984},
	{ObjectName: "STARLINK-1007", NORADCatID: 44713},
	{ObjectName: "STARLINK-30299", NORADCatID: 59379},
	{ObjectName: "ONEWEB-0010", NORADCatID: 44058},
	{ObjectName: "IRIDIUM 100", NORADCatID: 43059},
	{ObjectName: "GOES 18", NORADCatID: 51850},
	{ObjectName: "SENTINEL-2B", NORADCatID: 42063},
	{ObjectName: "SWOT", NORADCatID: 54740},
	{ObjectName: "AMAZONAS 5", NORADCatID: 42934},
}
var offlineMode = false

func lookUpSatellite(url string) bool {
	clearTerminal()
	fmt.Println("Fetching satellite data...")
	client := http.Client{
		Timeout: 10 * time.Second,
	}

	response, err := client.Get(url)
	if err != nil {
		fmt.Println("Could not contact CelesTrak: ", err)
		return false
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		fmt.Printf("CelesTrak returned HTTP status %d. \n", response.StatusCode)
		fmt.Println("No Satellite data was retrieved")
		return false
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Could not read response body: ", err)
		return false
	}

	err = json.Unmarshal(body, &satellites)
	if err != nil {
		fmt.Println("Could not interpret the satellite data: ", err)
		return false
	}

	if len(satellites) == 0 {
		fmt.Println("No satellite data found.")
		return false
	}
	if len(satellites) == 1 {
		satellite = satellites[0]
		clearTerminal()
		displaySatellite(satellite)
		return true
	} else {
		for i, s := range satellites {
			fmt.Printf("%d. %s, (%d)\n", i+1, s.ObjectName, s.NORADCatID)
		}
			fmt.Print("\nChoose a satellite: ")
			var choice int
			fmt.Scanln(&choice)
			if choice < 1 || choice > len(satellites) {
				fmt.Println("Invalid option")
				return false
			}
			satellite = satellites[choice-1]
			clearTerminal()
			displaySatellite(satellite)
			return true
	}
}

func displaySatellite(s Satellite) {
	fmt.Println("-------------------")
	fmt.Println(" Satellite Details")
	fmt.Println("-------------------")
	fmt.Println("Epoch:", s.Epoch)
	fmt.Println("Name:", s.ObjectName)
	fmt.Println("Object ID:", s.ObjectID)
	fmt.Println("NORAD ID:", s.NORADCatID)
	fmt.Printf("Classification: %s\n", s.ClassificationType)
	fmt.Printf("Element set number: %d\n", s.ElementSetNo)
	fmt.Printf("Ephemeris type: %d\n", s.EphemerisType)
	fmt.Println()
	fmt.Println("----------------------------")
	fmt.Println(" Basic Orbital Measurements")
	fmt.Println("----------------------------")
	fmt.Printf("Orbits per day: %.2f\n", s.MeanMotion)
	fmt.Printf("Orbital Period: ≈ %.2f minutes (%.0fh %.0fm %.0fs)\n", 1440/s.MeanMotion, math.Floor(1440/s.MeanMotion/60), math.Mod(math.Floor(1440/s.MeanMotion), 60), math.Mod(math.Round(1440/s.MeanMotion*60), 60))
	fmt.Printf("Inclination: %.2f ", s.Inclination)
	if s.Inclination < 30 {
		fmt.Println("(Low inclination)")
	} else if s.Inclination < 60 {
		fmt.Println("(Moderate inclination)")
	} else {
		fmt.Println("(High inclination)")
	}
	fmt.Println("Revolutions at epoch:", s.RevAtEpoch)
	fmt.Println()
	fmt.Println("--------------------------")
	fmt.Println(" Orbit shape and geometry")
	fmt.Println("--------------------------")
	fmt.Printf("Eccentricity: %.6f ", s.Eccentricity)
	if s.Eccentricity < 0.01 {
		fmt.Println("(Nearly Circular)")
	} else if s.Eccentricity < 0.1 {
		fmt.Println("(Slightly Elliptical)")
	} else {
		fmt.Println("(Elliptical)")
	}
	const earthMu = 398600.4418
  const earthRadius = 6378.137
	meanMotionRad := s.MeanMotion * 2 * math.Pi / 86400
	semiMajorAxis := math.Cbrt(earthMu / math.Pow(meanMotionRad, 2))
	perigee := semiMajorAxis*(1-s.Eccentricity) - earthRadius
	apogee := semiMajorAxis*(1+s.Eccentricity) - earthRadius
	fmt.Printf("Semi-major axis: ≈ %.2f km\n", semiMajorAxis)
	fmt.Printf("Perigee altitude: ≈ %.2f km\n", perigee)
	fmt.Printf("Apogee altitude: ≈ %.2f km\n", apogee)
	fmt.Printf("Altitude range: ≈ %.2f km\n", apogee-perigee)
	averageAltitude := semiMajorAxis-earthRadius
	fmt.Printf("Average altitude: ≈ %.2f km\n", averageAltitude)
	if averageAltitude < 2000 {
		fmt.Println("Orbit region: Low Earth Orbit (LEO)")
	} else if averageAltitude < 35786 {
		fmt.Println("Orbit region: Medium Earth Orbit (MEO)")
	} else {
		fmt.Println("Orbit region: High Earth Orbit (HEO)")
	}
}

func compareSatellites(s1, s2 Satellite) {
	fmt.Println("-------------------")
	fmt.Println(" Satellite Details")
	fmt.Println("-------------------")
	fmt.Printf("%-25.25s | %-25.25s | %-25.25s\n", "Epoch", s1.Epoch, s2.Epoch)
	fmt.Printf("%-25.25s | %-25.25s | %-25.25s\n", "Name", s1.ObjectName, s2.ObjectName)
	fmt.Printf("%-25.25s | %-25.25s | %-25.25s\n", "Object ID", s1.ObjectID, s2.ObjectID)
	fmt.Printf("%-25.25s | %-25.25s | %-25.25s\n", "NORAD ID", strconv.Itoa(s1.NORADCatID), strconv.Itoa(s2.NORADCatID))
	fmt.Printf("%-25.25s | %-25.25s | %-25.25s\n", "Classification", fmt.Sprintf("%s", s1.ClassificationType), fmt.Sprintf("%s", s2.ClassificationType))
	fmt.Printf("%-25.25s | %-25.25s | %-25.25s\n", "Element set number", strconv.Itoa(s1.ElementSetNo), strconv.Itoa(s2.ElementSetNo))
	fmt.Printf("%-25.25s | %-25.25s | %-25.25s\n", "Ephemeris type", fmt.Sprintf("%d", s1.EphemerisType), fmt.Sprintf("%d", s2.EphemerisType))
	fmt.Println()
	fmt.Println("----------------------------")
	fmt.Println(" Basic Orbital Measurements")
	fmt.Println("----------------------------")
	fmt.Printf("%-25.25s | %-25.25s | %-25.25s\n", "Orbits per day", fmt.Sprintf("%.2f", s1.MeanMotion), fmt.Sprintf("%.2f", s2.MeanMotion))
	fmt.Printf("%-25.25s | %-25.25s | %-25.25s\n", "Orbital Period", fmt.Sprintf("%.2f min (%.0fh %.0fm %.0fs)", 1440/s1.MeanMotion, math.Floor(1440/s1.MeanMotion/60), math.Mod(math.Floor(1440/s1.MeanMotion), 60), math.Mod(math.Round(1440/s1.MeanMotion*60), 60)), fmt.Sprintf("%.2f min (%.0fh %.0fm %.0fs)", 1440/s2.MeanMotion, math.Floor(1440/s2.MeanMotion/60), math.Mod(math.Floor(1440/s2.MeanMotion), 60), math.Mod(math.Round(1440/s2.MeanMotion*60), 60)))
	var s1Inclination string
	if s1.Inclination < 30 {
		s1Inclination = "(Low)"
	} else if s1.Inclination < 60 {
		s1Inclination = "(Moderate)"
	} else {
		s1Inclination = "(High)"
	}
	var s2Inclination string
	if s2.Inclination < 30 {
		s2Inclination = "(Low)"
	} else if s2.Inclination < 60 {
		s2Inclination = "(Moderate)"
	} else {
		s2Inclination = "(High)"
	}
	fmt.Printf("%-25.25s | %-25.25s | %-25.25s\n", "Inclination", fmt.Sprintf("%.2f %s", s1.Inclination, s1Inclination), fmt.Sprintf("%.2f %s", s2.Inclination, s2Inclination))
	fmt.Printf("%-25.25s | %-25.25s | %-25.25s\n", "Revolutions at Epoch", strconv.Itoa(s1.RevAtEpoch), strconv.Itoa(s2.RevAtEpoch))
	fmt.Println()
	fmt.Println("--------------------------")
	fmt.Println(" Orbit shape and geometry")
	fmt.Println("--------------------------")
	fmt.Printf("%-25.25s | %-25.25s | %-25.25s\n", "Eccentricity", fmt.Sprintf("%.6f", s1.Eccentricity), fmt.Sprintf("%.6f", s2.Eccentricity))
	const earthMu = 398600.4418
  const earthRadius = 6378.137
	s1meanMotionRad := s1.MeanMotion * 2 * math.Pi / 86400
	s1semiMajorAxis := math.Cbrt(earthMu / math.Pow(s1meanMotionRad, 2))
	s1perigee := s1semiMajorAxis*(1-s1.Eccentricity) - earthRadius
	s1apogee := s1semiMajorAxis*(1+s1.Eccentricity) - earthRadius
	s1averageAltitude := s1semiMajorAxis-earthRadius
	s2meanMotionRad := s2.MeanMotion * 2 * math.Pi / 86400
	s2semiMajorAxis := math.Cbrt(earthMu / math.Pow(s2meanMotionRad, 2))
	s2perigee := s2semiMajorAxis*(1-s2.Eccentricity) - earthRadius
	s2apogee := s2semiMajorAxis*(1+s2.Eccentricity) - earthRadius
	s2averageAltitude := s2semiMajorAxis-earthRadius
	fmt.Printf("%-25.25s | %-25.25s | %-25.25s\n", "Semi-major axis", fmt.Sprintf("%.2f", s1semiMajorAxis), fmt.Sprintf("%.2f", s2semiMajorAxis))
	fmt.Printf("%-25.25s | %-25.25s | %-25.25s\n", "Perigee altitude", fmt.Sprintf("%.2f", s1perigee), fmt.Sprintf("%.2f", s2perigee))
	fmt.Printf("%-25.25s | %-25.25s | %-25.25s\n", "Apogee altitude", fmt.Sprintf("%.2f", s1apogee), fmt.Sprintf("%.2f", s2apogee))
	fmt.Printf("%-25.25s | %-25.25s | %-25.25s\n", "Altitude range", fmt.Sprintf("%.2f", s1apogee-s1perigee), fmt.Sprintf("%.2f", s2apogee-s2perigee))
	var s1OrbitRange string
	if s1averageAltitude < 2000 {
		s1OrbitRange = "(LEO)"
	} else if s1averageAltitude < 35786 {
		s1OrbitRange = "(MEO)"
	} else {
		s1OrbitRange = "(HEO)"
	}
	var s2OrbitRange string
	if s2averageAltitude < 2000 {
		s2OrbitRange = "(LEO)"
	} else if s2averageAltitude < 35786 {
		s2OrbitRange = "(MEO)"
	} else {
		s2OrbitRange = "(HEO)"
	}
	fmt.Printf("%-25.25s | %-25.25s | %-25.25s\n", "Average altitude", fmt.Sprintf("%.2f %s", s1averageAltitude, s1OrbitRange), fmt.Sprintf("%.2f %s", s2averageAltitude, s2OrbitRange))
}

func viewWatchlist() {
	if len(watchlist) == 0 {
		fmt.Println("Watchlist is empty")
		return
	}
	fmt.Println("-----------")
	fmt.Println(" Watchlist")
	fmt.Println("-----------")
	fmt.Println("0. Back to main menu")
	for i, s := range watchlist {
		fmt.Printf("%d. %s\n", i+1, s.ObjectName)
	}
	fmt.Println()
	var satelliteNumber int
	fmt.Print("Choose an option: ")
	fmt.Scanln(&satelliteNumber)
	clearTerminal()
	if satelliteNumber < 1 || satelliteNumber > len(watchlist) {
		if satelliteNumber != 0{
			fmt.Println("Invalid option")
		}
		return
	} else {
		fmt.Printf("Selected Satellite: %s (%d)\n", watchlist[satelliteNumber-1].ObjectName, watchlist[satelliteNumber-1].NORADCatID)
		fmt.Println("1. View details")
		fmt.Println("2. Remove from watchlist")
		fmt.Println("3. Back to main menu")
		var choice string
		fmt.Print("\nChoose an option: ")
		fmt.Scanln(&choice)
		clearTerminal()
		if choice == "1" {
			if !offlineMode {
				id := watchlist[satelliteNumber-1].NORADCatID
				lookUpSatellite(fmt.Sprintf("https://celestrak.org/NORAD/elements/gp.php?CATNR=%d&FORMAT=JSON", id))
				if satellite.NORADCatID == id {
					watchlist[satelliteNumber-1] = satellite
					saveWatchlist()
				}
			} else {
				displaySatellite(watchlist[satelliteNumber-1])
			}
			viewWatchlist()
		} else if choice == "2" {
			watchlist = append(watchlist[:satelliteNumber-1], watchlist[satelliteNumber:]...)
			saveWatchlist()
		} else if choice == "3" {
			return
		} else {
			fmt.Println("Invalid option")
			return
		}
	}

	fmt.Println()
}

func saveWatchlist() {
	data, err := json.MarshalIndent(watchlist, "", "  ")
	if err != nil {
		fmt.Println(err)
		return
	}
	if err := os.WriteFile("watchlist.json", data, 0644); err != nil {
		fmt.Println(err)
		return
	}
}

func clearTerminal() {
	fmt.Print("\033[H\033[2J\033[3J")
}

func searchForSatellite(searchType string) bool {
	if searchType == "NoradID" {
		fmt.Print("Enter NORAD catalog ID: ")
	} else if searchType == "Name" {
		fmt.Print("Enter satellite name: ")
	}
	var input string
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil && len(input) == 0 {
		fmt.Println("Could not read input")
		return false
	}
	input = strings.TrimSpace(input)
	if input == "" {
		fmt.Println("Name cannot be empty")
		return false
	}

	if searchType == "NoradID" {
		id, err := strconv.Atoi(input)
		if err != nil || id <= 0 && searchType == "NoradID" {
			fmt.Println("Enter a valid NORAD ID")
			return false
		}
		return lookUpSatellite(fmt.Sprintf("https://celestrak.org/NORAD/elements/gp.php?CATNR=%d&FORMAT=JSON", id))
	} else if searchType == "Name" {
		return lookUpSatellite(fmt.Sprintf("https://celestrak.org/NORAD/elements/gp.php?NAME=%s&FORMAT=JSON", url.QueryEscape(input)))
	}
	return false
}

func suggestSatellites() []Satellite {	
	indexes := rand.Perm(len(suggestedSatellites))[:3]
	var suggestions []Satellite
	for _, i := range indexes{
		suggestions = append(suggestions, suggestedSatellites[i])
	}
	return suggestions
}

func main() {
	data, err := os.ReadFile("watchlist.json")
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Println(err)
			return
		}
		watchlist = []Satellite{}
	} else if err := json.Unmarshal(data, &watchlist); err != nil {
		fmt.Println(err)
		return
	}
	for {
		if !offlineMode {
			fmt.Println("-------------------")
			fmt.Println(" Satellite Tracker")
			fmt.Println("-------------------")

			fmt.Println("1. Look up a satellite")
			fmt.Println("2. Compare satellites")
			fmt.Println("3. View watchlist")
			fmt.Println("4. Exit")
			fmt.Println("5. Offline Mode")
			if satellite.ObjectName != "" {
				alreadyExists := false
				for _, s := range watchlist {
					if s.NORADCatID == satellite.NORADCatID {
						alreadyExists = true
						break
					}
				}
				if !alreadyExists {
					fmt.Printf("6. Add %s to watchlist\n", satellite.ObjectName)
				}
			}
			fmt.Println()
			fmt.Print("Choose an option: ")

			var choice string
			fmt.Scanln(&choice)
			clearTerminal()

			if choice == "1" {
				fmt.Println("Search Type")
				fmt.Println("1. NORAD ID")
				fmt.Println("2. Name")
				fmt.Println("\nSuggested satellites:")
				suggestions := suggestSatellites()
				for i,s := range suggestions{
					fmt.Printf("%d. %s \n", i+3, s.ObjectName)
				}
				fmt.Print("\nChoose an option: ")
				var choice int
				fmt.Scanln(&choice)
				if choice == 1 {
					searchForSatellite("NoradID")
				} else if choice == 2 {
					searchForSatellite("Name")
				}  else if 3 <= choice && choice <= 5 {
					lookUpSatellite(fmt.Sprintf("https://celestrak.org/NORAD/elements/gp.php?CATNR=%d&FORMAT=JSON", suggestions[choice-3].NORADCatID))
				} else {
					fmt.Println("Invalid option")
				}
			} else if choice == "2" {
				var s1 Satellite
				var s2 Satellite
				success := false
				for attempt := 1; attempt <= 5; attempt++{
					fmt.Println("Satellite 1")
					if !searchForSatellite("NoradID") {
						fmt.Println()
						continue
					}
					s1 = satellite
					success = true
					break
				}
				clearTerminal()
				if !success {
					clearTerminal()
					fmt.Println("Too many failed attempts. Returning to main menu.")
					continue
				}
				success = false
				for attempt := 1; attempt <= 5; attempt++{
					fmt.Println("Satellite 2")
					if !searchForSatellite("NoradID") {
						fmt.Println()
						continue
					}
					s2 = satellite
					if s1.NORADCatID != s2.NORADCatID{
						clearTerminal()
						success = true
						break
					} else {
						clearTerminal()
						fmt.Println("Select two different satellites")
						continue
					}
				}
				if !success {
					clearTerminal()
					fmt.Println("Too many failed attempts. Returning to main menu.")
					continue
				}
				compareSatellites(s1, s2)
			} else if choice == "3" {
				viewWatchlist()
			} else if choice == "4" {
				fmt.Println("Goodbye")
				return
			} else if choice == "5" {
				offlineMode = true
			} else if choice == "6" && satellite.ObjectName != ""{
				alreadyExists := false
				for _, s := range watchlist {
					if s.NORADCatID == satellite.NORADCatID {
						alreadyExists = true
						break
					}
				}
				if !alreadyExists {
					watchlist = append(watchlist, satellite)
					saveWatchlist()
					fmt.Printf("%s added to watchlist\n", satellite.ObjectName)
				}
			} else {
				fmt.Println("Invalid option")
			}
		} else {
			fmt.Println("----------------------------------")
			fmt.Println(" Satellite Tracker - Offline Mode")
			fmt.Println("----------------------------------")

			fmt.Println("1. Normal Mode")
			fmt.Println("2. View watchlist")
			fmt.Println("3. Exit")
			fmt.Println()
			fmt.Print("Choose an option: ")

			var choice string
			fmt.Scanln(&choice)
			clearTerminal()
			if choice == "1" {
				offlineMode = false
			} else if choice == "2" {
				viewWatchlist()
			} else if choice == "3" {
				fmt.Println("Goodbye")
				return
			} else {
				fmt.Println("Invalid option")
			}
		}
	}
}