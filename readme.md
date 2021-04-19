## To run this project

You can run one of the included binaries (built for windows amd64 and linux amd64) by navigating to the root of this project and running.


```
 ./bin/<binary-name>

```

you can change the port the application runs on with a runtime flag EX

```
./bin/<binary-name> -p 6000

```

If you have go 1.16 on your computer you can run the file without the binary by typing the command below in the root of this project

```
go run cmd/main.go

```

or build using 
```
go build  -o ./bin cmd/main.go

```