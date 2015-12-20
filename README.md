# processManager
Simple process manager

#How to use:
1) for run 10 fake job you should write
./processManager -cmd="./fakeJob" -countOfWorkers=10

2) for run 1 "some.sh"
./processManager -cmd="sh" -args="some.sh" 

3) for run 5 php with args
./processManager -cmd="php" -args="some.php" -args="mySimpleArgument=1" -countOfWorkers=5
