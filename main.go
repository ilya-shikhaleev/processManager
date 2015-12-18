package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

const DEFAULT_NUMBER_OF_WORKERS int = 1
const DEFAULT_CMD string = "sh"

type arguments []string

func (a *arguments) String() string {
	return strings.Join(*a, ", ")
}

func (a *arguments) Set(value string) error {
	*a = append(*a, value)
	return nil
}

var countOfWorkers int
var cmdString string
var args arguments

func startCmd(killWorkersChan chan bool) error {
	command := exec.Command(cmdString, args...)
	success := make(chan bool)
	command.Start()
	go func() {
		select {
		case <-killWorkersChan:
			command.Process.Kill()
		case <-success:
		}
	}()
	err := command.Wait()

	if err != nil {
		fmt.Fprintln(os.Stderr, "There was an error running command: ", err)
		return err
	}
	success <- true

	return nil
}

func worker(finishCh chan bool, killWorkersChan chan bool) {
	if err := startCmd(killWorkersChan); err != nil {
		finishCh <- false
	} else {
		finishCh <- true
	}
}

func getErrorDuration(errorsCount int) time.Duration {
	pauseCoefficient := errorsCount / countOfWorkers
	return time.Duration(pauseCoefficient * pauseCoefficient) * time.Second
}

func parseFlags() {
	flag.IntVar(&countOfWorkers, "countOfWorkers", DEFAULT_NUMBER_OF_WORKERS, "Count of workers")
	flag.StringVar(&cmdString, "cmd", DEFAULT_CMD, "Worker command")
	flag.Var(&args, "args", "Worker command arguments")
	flag.Parse()

	if flag.NFlag() == 0 {
		flag.Usage()
		os.Exit(1)
	}
}

func main() {
	parseFlags()

	workerEndChan := make(chan bool)
	workersCount := 0
	errorsCount := 0

	osKillSignalChan := make(chan os.Signal, 1)
	signal.Notify(osKillSignalChan, os.Interrupt, os.Kill, syscall.SIGTERM)

	killWorkersChan := make(chan bool)
	go func() {
		<-osKillSignalChan
		close(killWorkersChan)
	}()

	for {
		select {
		case <-killWorkersChan:
			time.Sleep(1 * time.Second)
			fmt.Println("sleep and die")
			os.Exit(1)
		case <-time.After(getErrorDuration(errorsCount)):
			if workersCount < countOfWorkers {
				workersCount++
				go worker(workerEndChan, killWorkersChan)
			}
		case success := <-workerEndChan:
			if !success {
				errorsCount++
			} else {
				errorsCount = 0
			}
			workersCount--
		}
	}
}
