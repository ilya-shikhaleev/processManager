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

const PARAM_ARGS_FLAG_NAME string = "args"
const PARAM_COUNT_OF_WORKERS_FLAG_NAME string = "countOfWorkers"
const PARAM_CMD_FLAG_NAME string = "cmd"

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
	successChan := make(chan bool)
	command.Start()

	go func() {
		select {
		case <-killWorkersChan:
			command.Process.Kill()
		case <-successChan:
		}
	}()

	err := command.Wait()

	if err != nil {
		fmt.Fprintln(os.Stderr, "There was an error running command: ", err)
		return err
	}
	successChan <- true

	return nil
}

func worker(finishChan chan bool, killWorkersChan chan bool) {
	if err := startCmd(killWorkersChan); err != nil {
		finishChan <- false
	} else {
		finishChan <- true
	}
}

func getErrorDuration(errorsCount int) time.Duration {
	pauseCoefficient := errorsCount / countOfWorkers
	return time.Duration(pauseCoefficient * pauseCoefficient) * time.Second
}

func parseFlags() {
	flag.IntVar(&countOfWorkers, PARAM_COUNT_OF_WORKERS_FLAG_NAME, DEFAULT_NUMBER_OF_WORKERS, "Count of workers")
	flag.StringVar(&cmdString, PARAM_CMD_FLAG_NAME, DEFAULT_CMD, "Worker command")
	flag.Var(&args, PARAM_ARGS_FLAG_NAME, "Worker command arguments")
	flag.Parse()

	if flag.NFlag() == 0 {
		flag.Usage()
		os.Exit(1)
	}
}

func catchKillSignal() (killWorkersChan chan bool) {
	osKillSignalChan := make(chan os.Signal, 1)
	signal.Notify(osKillSignalChan, os.Interrupt, os.Kill, syscall.SIGTERM)

	go func() {
		<-osKillSignalChan
		close(killWorkersChan)
	}()

	return
}

func main() {
	parseFlags()

	workerEndChan := make(chan bool)
	workersCount := 0
	errorsCount := 0
	killWorkersChan := catchKillSignal()

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
