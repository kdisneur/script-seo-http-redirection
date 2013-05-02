package main

import "bufio"
import "flag"
import "fmt"
import "net/http"
import "os"
import "regexp"

const MaxThreads            = 100
var redirectStatusRegexp, _ = regexp.Compile("^(301|302|307)")
var cleanStatusRegexp, _    = regexp.Compile("^([0-9]+).*")
var inputFileName           = flag.String("input",  "", "file containing all the urls (one by line)")
var threadsSemaphore        = make(chan int, MaxThreads)
var loggerChannel           = make(chan string)

func main() {
  for i := 0; i < MaxThreads; i++ {
    threadsSemaphore <- 1
  }
  CheckParameters()
  go StartLogging()
  AnalyzeUrls()
}

func AnalyzeUrl(url string) {
  request, error := http.NewRequest("GET", url, nil)
  if (error != nil) {
    LogError(fmt.Sprintf("Error on url fetching %s", error))
    threadsSemaphore <- 1
    return;
  }

  response, error := http.DefaultTransport.RoundTrip(request)
  if (error != nil) {
    LogError(fmt.Sprintf("Error on url fetching %s", error))
    threadsSemaphore <- 1
    return;
  }

  response.Body.Close()
  statusBytes := []byte(response.Status)
  statusCode  := string(cleanStatusRegexp.ReplaceAll(statusBytes, []byte("$1")))

  if redirectStatusRegexp.Match(statusBytes) {
    redirection := response.Header.Get("Location")
    Write(url, statusCode, redirection)
  } else {
    Write(url, statusCode, "")
  }

  threadsSemaphore <- 1
}

func AnalyzeUrls() {
  inputFile, error := os.Open(*inputFileName)
  if error != nil {
    RaiseError(fmt.Sprintf("Error on file opening %s", error))
  }

  inputStream := bufio.NewReader(inputFile)
  for {
    bytes, _, error := inputStream.ReadLine()
    url             := string(bytes)
    if error != nil {
      break
    }
    if url == "" {
      continue
    }

    <-threadsSemaphore
    go AnalyzeUrl(url)
  }

  for i := 0; i < MaxThreads; i++ {
    <-threadsSemaphore
  }
}

func CheckParameters() {
  flag.Parse()
  if *inputFileName == "" {
    flag.PrintDefaults()
    os.Exit(1)
  }
}

func LogError(message string) {
  loggerChannel <- fmt.Sprintf("Error -- %s", message)
}

func RaiseError(message string) {
  LogError(message)
  os.Exit(1)
}

func StartLogging() {
  for {
    message := <-loggerChannel
    if message == "" {
      break;
    }
    fmt.Println(message)
  }
}

func Write(url string, status string, redirection string) {
  loggerChannel <- fmt.Sprintf("%s;%s;%s", url, status, redirection)
}
