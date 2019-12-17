# Tools

This contains some internal tooling for computing, debugging, and analyzing cost and route matrices. All of these tools with with stdin and stdout like every other Linux tool so you can pipe commands together however you need to.

## Building Tools

```
> make build-tools
```

## Help

Each tool honors the `-h` or `-help` flag to see its usage.

```
> ./dist/optimize -h
Usage of ./dist/optimize:
  -threshold-rtt int
        set the threshold RTT (default 1)
```

## Example Piping Tools Together

```
> ./dist/cost -url=http://localhost:30000/cost_matrix | \
./dist/optimize -threshold-rtt=4 | \
./dist/route -relay=r1 -datacenter=d1

> ./dist/cost | ./dist/optimize | ./dist/debug -relay=r1

> cat ./path/to/cost.bin | ./dist/optimize > ./write/to/optimize.bin

> cat ./path/to/optimize.bin | ./dist/analyze
```