package main

import (
	"fmt"
	"net/http"
	"io"
	"os"
	"time"
	"strconv"
	"strings"
	"encoding/csv"
	"sort"
	"bufio"
	"regexp"
	"log"
)

func main() {
	//Refer to https://golang.org/src/time/format.go for format constants
	urldir := "http://web.mta.info/developers/data/nyct/turnstile/"
	filename := "turnstile_" + getDateOfLastSaturday().Format("060102") + ".txt"

	// _ is a blank identifier, used to "throw away" a value. Go compiler will
	// not allow you to have an assigned variable go unused.
	if _, err := os.Stat(filename); err != nil {
		println("Getting " + filename + "...")
		downloadFile(urldir + filename, filename)
	}

	processFile(filename)

	os.Exit(0)
}

//Save some typing...
func println(s string){
	fmt.Println(s)
}

//Download the file and save it to a directory
func downloadFile(getUrl string, toFile string){
	out, err := os.Create(toFile)
	defer out.Close()
	if err != nil{
		log.Fatal("Error: Could not create the output file!")
		os.Exit(1)
	}

	resp, err := http.Get(getUrl)
	defer resp.Body.Close()
	if err != nil{
		log.Fatal("Error: Could not download the file.")
		os.Exit(1)
	}
	
	n, err := io.Copy(out, resp.Body)
	if err != nil{
		log.Fatal("Error: Could not copy file to local filesystem.")
		os.Exit(1)
	}
	
	println(strconv.FormatInt(n,10) + " bytes downloaded.")
}

//Function to get the date of the last Saturday.
func getDateOfLastSaturday() time.Time {
	currentTime := time.Now()

	if(currentTime.Weekday() == 6){
		return currentTime
	} else {
		//6 is the max day of week, so simply subtract the current day's number + 1 to get last saturday
		return currentTime.AddDate(0,0,-(int(currentTime.Weekday())+1))
	}
}

//Get an int from the user. Error message is displayed if a non-digit is entered or
//the integer is below zero or greater than max
func getIntInput(max int) int{
	//Read in the user's input
	inpReader := bufio.NewReader(os.Stdin)
	inp, _ := inpReader.ReadString('\n')
	inp = strings.Replace(inp,"\n","",-1)
	
	//Check that the user entered a valid number
	userNum, err := strconv.ParseInt(inp,10,32)
	if err != nil || int(userNum) > max || userNum <= 0 {
		log.Fatal("Error: enter a valid number.")
		os.Exit(1)
	}

	return int(userNum)
}

//Get the user input for which station and train line they want to view data for
func processFile(filename string){
	println("Processing " + filename + "...")
	f,err := os.Open(filename)
	if err != nil{
		log.Fatal("Error: File could not be opened.")
		os.Exit(1)
	}
	
	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil{
		log.Fatal("Error: could not read the input file");
		os.Exit(1)
	}
	
	//Use a map to get the unique values for station
	//Stations are mapped to an array of lines
	var stations = make(map[string][]string)
	for i := 1; i < len(records); i++{
		
		//Check if the subway line exists for the station
		found := false
		for _, val := range stations[records[i][3]] {
			if(val == records[i][4]){
				found = true
				break
			}
		}

		if(!found){
			stations[records[i][3]] = append(stations[records[i][3]],records[i][4])
		}
	}
	
	//Have to dump the keys into an array to sort them...
	var stationNames []string
	for key, _ := range stations{
		stationNames = append(stationNames,key)
	}
	
	sort.Strings(stationNames)
	for i := 0; i < len(stationNames); i++{
		println("(" + strconv.FormatInt(int64(i+1),10) + ") " + stationNames[i])
	}
	
	//Get the number of the station the user wants to view
	println("Enter the number of the station: ")
	stNum := getIntInput(len(stationNames))
	selectedStation := stationNames[stNum-1]

	//If there's only one line for this station, dont bother with getting input
	var selectedLine string
	if(len(stations[selectedStation]) > 1){
		//Get the number of the line the user wants to view
		println("Which line at the " + selectedStation + " station?")
		lineArr := stations[selectedStation]
		for i := 0; i < len(lineArr); i++ {
			println("(" + strconv.FormatInt(int64(i+1),10) + ") " + lineArr[i])
		}
		
		lNum := getIntInput(len(lineArr))
		selectedLine = lineArr[lNum-1]
	} else {
		selectedLine = stations[selectedStation][0]
	}

	//Begin coalescing results
	var statsIn = make(map[string][]int64)
	var statsOut = make(map[string][]int64)
	re := regexp.MustCompile("(^0+| +)")
	
	for i := 1; i < len(records); i++{
		//Check if the subway line exists for the station
		if(records[i][3] == selectedStation && records[i][4] == selectedLine){
			tmpKey := records[i][0] + "," + records[i][1] + "," + records[i][2]
			
			//This mess is to trim whitespace and replace leading zeroes, the convert to an int64
			ingress, _ := strconv.ParseInt(re.ReplaceAllString(records[i][9],""),10,64)
			egress, _ := strconv.ParseInt(re.ReplaceAllString(records[i][10],""),10,64)
			
			statsIn[tmpKey] = append(statsIn[tmpKey], ingress)
			statsOut[tmpKey] = append(statsOut[tmpKey], egress)
		}
	}

	//Sort the keys
	var tsUnits []string
	for key, _ := range statsIn{
		tsUnits = append(tsUnits,key)
	}
	sort.Strings(tsUnits)	
	
	//Now lets print some results...
	println("Turnstiles for the " + selectedLine + " line(s) at the " + selectedStation + " station.")
	println("====")
	for i := 0; i < len(tsUnits); i++ {
		ingressArr := statsIn[tsUnits[i]]
		egressArr := statsOut[tsUnits[i]]
		
		//Subtract the Last entry from the first to get the total for the week
		totalIngress := strconv.Itoa(int(ingressArr[len(ingressArr)-1] - ingressArr[0]))
		totalEgress := strconv.Itoa(int(egressArr[len(egressArr)-1] - egressArr[0]))
		fmt.Println(tsUnits[i] + " " + totalIngress + " entries for this week; " + totalEgress + " exits for this week.")
	}
}
